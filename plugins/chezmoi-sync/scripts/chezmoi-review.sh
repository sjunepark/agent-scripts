#!/usr/bin/env bash
set -u

usage() {
	cat <<'USAGE'
Usage: chezmoi-review.sh [--fetch] [--diff [target...]]

Review chezmoi sync state for Codex. This script is read-only except --fetch,
which updates git remote refs for the chezmoi source repository.
USAGE
}

fetch=0
show_diff=0
targets=()

while [ "$#" -gt 0 ]; do
	case "$1" in
		--fetch)
			fetch=1
			shift
			;;
		--diff)
			show_diff=1
			shift
			while [ "$#" -gt 0 ]; do
				targets+=("$1")
				shift
			done
			;;
		-h|--help)
			usage
			exit 0
			;;
		*)
			echo "Unknown argument: $1" >&2
			usage >&2
			exit 2
			;;
	esac
done

if ! command -v chezmoi >/dev/null 2>&1; then
	echo "chezmoi is not installed or not on PATH." >&2
	exit 1
fi

source_dir="$(chezmoi source-path 2>/dev/null || true)"

if [ "$fetch" -eq 1 ]; then
	if [ -z "$source_dir" ]; then
		echo "Cannot fetch: chezmoi source path is unknown." >&2
		exit 1
	fi
	if ! command -v git >/dev/null 2>&1; then
		echo "Cannot fetch: git is not installed or not on PATH." >&2
		exit 1
	fi
	echo "# Fetch chezmoi source remote"
	git -C "$source_dir" fetch --quiet
	echo
fi

echo "# Chezmoi sync review"
echo
CHEZMOI_SYNC_VERBOSE=1 "$(dirname "$0")/chezmoi-check.sh"

if [ "$show_diff" -eq 1 ]; then
	echo
	echo "# Chezmoi diff"
	if [ "${#targets[@]}" -gt 0 ]; then
		chezmoi diff "${targets[@]}"
	else
		chezmoi diff
	fi
fi
