#!/usr/bin/env python3
"""Inject a versioned Codex session identifier as developer context."""

from __future__ import annotations

import json
import sys


MARKER_PREFIX = "codex.runtime.session_id.v1="
VALID_SOURCES = frozenset({"startup", "resume", "clear", "compact"})


def response_for(payload: object) -> str | None:
    if not isinstance(payload, dict):
        return None
    if payload.get("hook_event_name") != "SessionStart":
        return None
    if payload.get("source") not in VALID_SOURCES:
        return None

    session_id = payload.get("session_id")
    if (
        not isinstance(session_id, str)
        or not session_id.strip()
        or "\n" in session_id
        or "\r" in session_id
    ):
        return None

    output = {
        "hookSpecificOutput": {
            "hookEventName": "SessionStart",
            "additionalContext": f"{MARKER_PREFIX}{session_id}",
        }
    }
    return json.dumps(output, ensure_ascii=False, separators=(",", ":"))


def main() -> int:
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, UnicodeError, OSError):
        return 0

    output = response_for(payload)
    if output is not None:
        try:
            sys.stdout.write(output)
        except OSError:
            return 0
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
