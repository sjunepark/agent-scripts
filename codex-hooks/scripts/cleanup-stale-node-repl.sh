#!/bin/sh
#
# Pressure-relief workaround for https://github.com/openai/codex/issues/26984.
# Codex can retain node_repl MCP children and their stdio pipes after the
# owning session is gone. Only prune old helpers when an app-server is close
# to macOS's default descriptor soft limit.

set -eu

readonly HIGH_WATER="${CODEX_NODE_REPL_CLEANUP_HIGH_WATER:-200}"
readonly TARGET="${CODEX_NODE_REPL_CLEANUP_TARGET:-160}"
readonly MIN_AGE_SECONDS="${CODEX_NODE_REPL_CLEANUP_MIN_AGE_SECONDS:-1800}"
readonly EMERGENCY_WATER="${CODEX_NODE_REPL_CLEANUP_EMERGENCY_WATER:-240}"
readonly EMERGENCY_MIN_AGE_SECONDS="${CODEX_NODE_REPL_CLEANUP_EMERGENCY_MIN_AGE_SECONDS:-60}"
readonly CODEX_STATE_DIR="${CODEX_HOME:-$HOME/.codex}"
readonly LOCK_DIR="$CODEX_STATE_DIR/.node-repl-cleanup.lock"
readonly LOG_DIR="$CODEX_STATE_DIR/log"
readonly LOG_FILE="$LOG_DIR/node-repl-cleanup.log"

case "$HIGH_WATER:$TARGET:$MIN_AGE_SECONDS:$EMERGENCY_WATER:$EMERGENCY_MIN_AGE_SECONDS" in
  *[!0-9:]* | :* | *: | *::*)
    exit 0
    ;;
esac

if [ "$TARGET" -ge "$HIGH_WATER" ] || [ "$EMERGENCY_WATER" -lt "$HIGH_WATER" ]; then
  exit 0
fi

# SessionStart hooks from separate Codex clients can overlap.
if ! /bin/mkdir "$LOCK_DIR" 2>/dev/null; then
  exit 0
fi
trap '/bin/rmdir "$LOCK_DIR" 2>/dev/null || true' EXIT HUP INT TERM

/bin/mkdir -p "$LOG_DIR"

elapsed_seconds() {
  value=$1
  days=0

  case "$value" in
    *-*)
      days=${value%%-*}
      value=${value#*-}
      ;;
  esac

  old_ifs=$IFS
  IFS=:
  set -- $value
  IFS=$old_ifs

  if [ "$#" -eq 3 ]; then
    hours=$1
    minutes=$2
    seconds=$3
  else
    hours=0
    minutes=$1
    seconds=$2
  fi

  # Shell arithmetic may interpret a leading zero as octal.
  days=$(printf '%s' "$days" | /usr/bin/sed 's/^0*//')
  hours=$(printf '%s' "$hours" | /usr/bin/sed 's/^0*//')
  minutes=$(printf '%s' "$minutes" | /usr/bin/sed 's/^0*//')
  seconds=$(printf '%s' "$seconds" | /usr/bin/sed 's/^0*//')
  days=${days:-0}
  hours=${hours:-0}
  minutes=${minutes:-0}
  seconds=${seconds:-0}

  echo $((days * 86400 + hours * 3600 + minutes * 60 + seconds))
}

fd_count() {
  # lsof emits one header row.
  count=$(/usr/sbin/lsof -p "$1" 2>/dev/null | /usr/bin/wc -l | /usr/bin/tr -d ' ')
  if [ "$count" -gt 0 ]; then
    echo $((count - 1))
  else
    echo 0
  fi
}

is_codex_app_server() {
  pid=$1
  comm=$(/bin/ps -p "$pid" -o comm= 2>/dev/null || true)
  args=$(/bin/ps -p "$pid" -o command= 2>/dev/null || true)

  case "$comm" in
    codex | */codex) ;;
    *) return 1 ;;
  esac

  case "$args" in
    *app-server*) return 0 ;;
    *) return 1 ;;
  esac
}

is_ancestor_of_self() {
  wanted=$1
  current=$$

  while [ "$current" -gt 1 ]; do
    if [ "$current" = "$wanted" ]; then
      return 0
    fi
    current=$(/bin/ps -p "$current" -o ppid= 2>/dev/null | /usr/bin/tr -d ' ' || true)
    [ -n "$current" ] || break
  done

  return 1
}

server_pids=$(
  /bin/ps -axo pid=,comm=,command= |
    /usr/bin/awk '$2 == "codex" || $2 ~ /\/codex$/ {
      if ($0 ~ /app-server/) print $1
    }'
)

for server_pid in $server_pids; do
  if ! is_codex_app_server "$server_pid"; then
    continue
  fi

  before=$(fd_count "$server_pid")
  if [ "$before" -lt "$HIGH_WATER" ]; then
    continue
  fi

  active_min_age=$MIN_AGE_SECONDS
  if [ "$before" -ge "$EMERGENCY_WATER" ]; then
    active_min_age=$EMERGENCY_MIN_AGE_SECONDS
  fi

  candidates=""
  while read -r child_pid child_ppid child_pgid child_age child_comm; do
    [ -n "${child_pid:-}" ] || continue
    [ "$child_ppid" = "$server_pid" ] || continue

    case "$child_comm" in
      node_repl | */node_repl) ;;
      *) continue ;;
    esac

    age_seconds=$(elapsed_seconds "$child_age")
    if [ "$age_seconds" -lt "$active_min_age" ]; then
      continue
    fi

    # node_repl is launched as its own process group. Never signal a group
    # shared with the app-server or an unrelated process.
    if [ "$child_pgid" != "$child_pid" ] || [ "$child_pgid" = "$server_pid" ]; then
      continue
    fi
    if is_ancestor_of_self "$child_pid"; then
      continue
    fi

    candidates="${candidates}${age_seconds} ${child_pid} ${child_pgid}
"
  done <<EOF
$(/bin/ps -axo pid=,ppid=,pgid=,etime=,comm=)
EOF

  [ -n "$candidates" ] || continue

  needed=$(((before - TARGET + 2) / 3))
  killed=0

  sorted=$(
    printf '%s' "$candidates" |
      /usr/bin/sort -k1,1nr
  )

  while read -r age_seconds child_pid child_pgid; do
    [ -n "${child_pid:-}" ] || continue
    [ "$killed" -lt "$needed" ] || break

    # Revalidate immediately before signalling to avoid PID-reuse mistakes.
    current_ppid=$(/bin/ps -p "$child_pid" -o ppid= 2>/dev/null | /usr/bin/tr -d ' ' || true)
    current_pgid=$(/bin/ps -p "$child_pid" -o pgid= 2>/dev/null | /usr/bin/tr -d ' ' || true)
    current_comm=$(/bin/ps -p "$child_pid" -o comm= 2>/dev/null || true)
    case "$current_comm" in
      node_repl | */node_repl) ;;
      *) continue ;;
    esac
    if [ "$current_ppid" != "$server_pid" ] || [ "$current_pgid" != "$child_pgid" ]; then
      continue
    fi

    if /bin/kill -TERM -- "-$child_pgid" 2>/dev/null; then
      killed=$((killed + 1))
    fi
  done <<EOF
$sorted
EOF

  if [ "$killed" -gt 0 ]; then
    after=$(fd_count "$server_pid")
    timestamp=$(/bin/date -u '+%Y-%m-%dT%H:%M:%SZ')
    printf '%s server_pid=%s fd_before=%s fd_after=%s helpers_terminated=%s\n' \
      "$timestamp" "$server_pid" "$before" "$after" "$killed" >>"$LOG_FILE"
  fi
done

exit 0
