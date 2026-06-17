#!/usr/bin/env bash
set -u

# Read-only startup check for Codex lifecycle hooks.
# This script intentionally never mutates chezmoi state, git refs, or files.

main() {
	local verbose="${CHEZMOI_SYNC_VERBOSE:-0}"

	if ! command -v chezmoi >/dev/null 2>&1; then
		[ "$verbose" = "1" ] && echo "chezmoi-sync: chezmoi is not installed or not on PATH."
		exit 0
	fi

	local status
	if ! status="$(chezmoi status 2>&1)"; then
		echo "chezmoi-sync: chezmoi status failed:"
		echo "$status"
		exit 0
	fi

	local source_dir=""
	source_dir="$(chezmoi source-path 2>/dev/null || true)"

	local managed_count script_count
	managed_count="$(printf '%s\n' "$status" | awk 'length($0) >= 3 && substr($0, 2, 1) != "R" { count++ } END { print count + 0 }')"
	script_count="$(printf '%s\n' "$status" | awk 'length($0) >= 3 && substr($0, 2, 1) == "R" { count++ } END { print count + 0 }')"

	local repo_summary=""
	if [ -n "$source_dir" ] && command -v git >/dev/null 2>&1 && git -C "$source_dir" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
		repo_summary="$(git_summary "$source_dir")"
	fi

	if [ "$managed_count" -eq 0 ] && [ "$script_count" -eq 0 ] && ! printf '%s\n' "$repo_summary" | grep -q 'needs attention'; then
		[ "$verbose" = "1" ] && echo "chezmoi-sync: clean."
		exit 0
	fi

	echo "chezmoi-sync: attention needed."
	if [ "$managed_count" -gt 0 ]; then
		echo "- $managed_count managed file change(s) differ between target files and the chezmoi source."
	fi
	if [ "$script_count" -gt 0 ]; then
		echo "- $script_count chezmoi script action(s) are pending."
	fi
	if [ -n "$source_dir" ]; then
		echo "- chezmoi source: $source_dir"
	fi
	if [ -n "$repo_summary" ]; then
		printf '%s\n' "$repo_summary"
	fi
	if [ -n "$status" ]; then
		echo
		echo "chezmoi status:"
		printf '%s\n' "$status" | sed 's/^/  /'
	fi
	echo
	echo "Ask Codex to review chezmoi sync state before taking explicit next steps. The startup hook only reports; it does not sync."
	exit 0
}

git_summary() {
	local repo="$1"
	local branch upstream porcelain ahead behind counts

	branch="$(git -C "$repo" branch --show-current 2>/dev/null || true)"
	upstream="$(git -C "$repo" rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null || true)"
	porcelain="$(git -C "$repo" status --porcelain 2>/dev/null || true)"
	ahead=0
	behind=0

	if [ -n "$upstream" ]; then
		counts="$(git -C "$repo" rev-list --left-right --count "HEAD...@{upstream}" 2>/dev/null || true)"
		if [ -n "$counts" ]; then
			ahead="${counts%%[[:space:]]*}"
			behind="${counts##*[[:space:]]}"
		fi
	fi

	if [ -z "$porcelain" ] && [ "${ahead:-0}" -eq 0 ] && [ "${behind:-0}" -eq 0 ]; then
		return 0
	fi

	echo "- chezmoi source git needs attention."
	[ -n "$branch" ] && echo "  - branch: $branch"
	[ -n "$upstream" ] && echo "  - upstream: $upstream"
	[ "${behind:-0}" -gt 0 ] && echo "  - behind upstream by $behind commit(s), based on existing local remote refs."
	[ "${ahead:-0}" -gt 0 ] && echo "  - ahead of upstream by $ahead commit(s)."
	if [ -n "$porcelain" ]; then
		echo "  - uncommitted source changes:"
		printf '%s\n' "$porcelain" | sed 's/^/    /'
	fi
}

main "$@"
