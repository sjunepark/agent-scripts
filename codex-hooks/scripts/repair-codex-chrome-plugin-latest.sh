#!/bin/sh
#
# Work around a Codex Chrome plugin installer bug that can leave the native
# messaging host pointed through a missing or stale `latest` cache symlink.
# This script may replace only that symlink; installed plugin versions and
# unexpected real files or directories are never modified.

set -eu

mode=repair
case "${1:-}" in
  "") ;;
  --check) mode=check ;;
  *)
    printf 'Usage: %s [--check]\n' "$0" >&2
    exit 2
    ;;
esac

readonly CODEX_STATE_DIR="${CODEX_HOME:-$HOME/.codex}"
readonly PLUGIN_ROOT="${CODEX_CHROME_PLUGIN_ROOT:-$CODEX_STATE_DIR/plugins/cache/openai-bundled/chrome}"
readonly LATEST_LINK="$PLUGIN_ROOT/latest"
readonly LOCK_DIR="$CODEX_STATE_DIR/.chrome-plugin-latest-repair.lock"
readonly LOG_DIR="$CODEX_STATE_DIR/log"
readonly LOG_FILE="$LOG_DIR/chrome-plugin-latest-repair.log"

report() {
  level=$1
  message=$2
  timestamp=$(/bin/date -u '+%Y-%m-%dT%H:%M:%SZ')

  /bin/mkdir -p "$LOG_DIR" 2>/dev/null || true
  printf '%s level=%s %s\n' "$timestamp" "$level" "$message" >>"$LOG_FILE" 2>/dev/null || true
  printf 'Codex Chrome plugin: %s\n' "$message" >&2
}

fail() {
  report error "$1"
  if [ "$mode" = check ]; then
    exit 1
  fi
  exit 0
}

[ -d "$PLUGIN_ROOT" ] || exit 0

# SessionStart hooks from separate Codex clients can overlap. Check mode stays
# lock-free so it can diagnose a repair already running in another process.
if [ "$mode" = repair ]; then
  if ! /bin/mkdir "$LOCK_DIR" 2>/dev/null; then
    exit 0
  fi
  trap '/bin/rmdir "$LOCK_DIR" 2>/dev/null || true' EXIT HUP INT TERM
fi

valid_versions=""
for candidate in "$PLUGIN_ROOT"/*; do
  [ -d "$candidate" ] || continue
  [ ! -L "$candidate" ] || continue

  version=${candidate##*/}
  case "$version" in
    latest | *[!0-9.]* | .* | *..* | *.) continue ;;
  esac

  [ -f "$candidate/.codex-plugin/plugin.json" ] || continue

  valid_host=false
  for host in "$candidate"/extension-host/macos/*/"ChatGPT for Chrome"; do
    if [ -x "$host" ]; then
      valid_host=true
      break
    fi
  done
  [ "$valid_host" = true ] || continue

  valid_versions="${valid_versions}${version}
"
done

[ -n "$valid_versions" ] || fail "no valid installed Chrome plugin version was found under $PLUGIN_ROOT"

# BSD sort lacks version sorting. Build a fixed-width numeric key for each
# dotted version so a cache directory such as 10.x sorts after 9.x.
latest_version=$(
  printf '%s' "$valid_versions" |
    /usr/bin/awk -F. '
      NF > 0 {
        key = ""
        for (field_index = 1; field_index <= NF; field_index += 1) {
          key = key sprintf("%012d", $field_index)
        }
        print key "\t" $0
      }
    ' |
    /usr/bin/sort |
    /usr/bin/tail -n 1 |
    /usr/bin/cut -f 2
)

[ -n "$latest_version" ] || fail "could not determine the newest valid installed Chrome plugin version"

if [ -L "$LATEST_LINK" ]; then
  current_target=$(/usr/bin/readlink "$LATEST_LINK" 2>/dev/null || true)
  if [ "$current_target" = "$latest_version" ] || [ "$current_target" = "$PLUGIN_ROOT/$latest_version" ]; then
    if [ "$mode" = check ]; then
      printf 'Codex Chrome plugin: latest -> %s (ok)\n' "$latest_version"
    fi
    exit 0
  fi
elif [ -e "$LATEST_LINK" ]; then
  fail "$LATEST_LINK is not a symlink; refusing to overwrite it"
fi

if [ "$mode" = check ]; then
  if [ -L "$LATEST_LINK" ]; then
    fail "latest points to $(/usr/bin/readlink "$LATEST_LINK" 2>/dev/null || printf unknown), expected $latest_version"
  fi
  fail "latest is missing; expected a symlink to $latest_version"
fi

if ! /bin/ln -sfn "$latest_version" "$LATEST_LINK"; then
  fail "failed to update latest to $latest_version"
fi

if [ "$(/usr/bin/readlink "$LATEST_LINK" 2>/dev/null || true)" != "$latest_version" ]; then
  fail "latest did not resolve to $latest_version after repair"
fi

report repaired "updated latest -> $latest_version"
exit 0
