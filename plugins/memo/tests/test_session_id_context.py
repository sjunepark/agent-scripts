from __future__ import annotations

import json
from pathlib import Path
import subprocess
import sys
import unittest


PLUGIN_ROOT = Path(__file__).resolve().parents[1]
HOOK = PLUGIN_ROOT / "scripts" / "session-id-context.py"
HOOKS_CONFIG = PLUGIN_ROOT / "hooks" / "hooks.json"
MANIFEST = PLUGIN_ROOT / ".codex-plugin" / "plugin.json"
SKILL = PLUGIN_ROOT / "skills" / "memo" / "SKILL.md"
MARKER_PREFIX = "codex.runtime.session_id.v1="


class SessionIdContextTests(unittest.TestCase):
    def run_hook(self, payload: object | None = None, raw: str | None = None) -> subprocess.CompletedProcess[str]:
        hook_input = raw if raw is not None else json.dumps(payload)
        return subprocess.run(
            [sys.executable, str(HOOK)],
            input=hook_input,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

    def expected_output(self, session_id: str) -> str:
        return json.dumps(
            {
                "hookSpecificOutput": {
                    "hookEventName": "SessionStart",
                    "additionalContext": f"{MARKER_PREFIX}{session_id}",
                }
            },
            ensure_ascii=False,
            separators=(",", ":"),
        )

    def test_injects_exact_marker_for_every_start_source(self) -> None:
        for source in ("startup", "resume", "clear", "compact"):
            with self.subTest(source=source):
                result = self.run_hook(
                    {
                        "hook_event_name": "SessionStart",
                        "source": source,
                        "session_id": "session-opaque-123",
                        "cwd": "/ignored",
                    }
                )
                self.assertEqual(result.returncode, 0)
                self.assertEqual(result.stdout, self.expected_output("session-opaque-123"))
                self.assertEqual(result.stderr, "")

    def test_output_is_stable_for_the_same_session(self) -> None:
        outputs = {
            self.run_hook(
                {
                    "hook_event_name": "SessionStart",
                    "source": source,
                    "session_id": "same-session",
                }
            ).stdout
            for source in ("startup", "resume", "clear", "compact")
        }
        self.assertEqual(outputs, {self.expected_output("same-session")})

    def test_invalid_input_is_a_silent_success(self) -> None:
        cases: list[tuple[str, object | None, str | None]] = [
            ("empty", None, ""),
            ("malformed-json", None, "{"),
            ("array", [], None),
            ("missing-fields", {}, None),
            ("wrong-event", {"hook_event_name": "Stop", "source": "startup", "session_id": "id"}, None),
            ("wrong-source", {"hook_event_name": "SessionStart", "source": "other", "session_id": "id"}, None),
            ("missing-id", {"hook_event_name": "SessionStart", "source": "startup"}, None),
            ("empty-id", {"hook_event_name": "SessionStart", "source": "startup", "session_id": ""}, None),
            ("blank-id", {"hook_event_name": "SessionStart", "source": "startup", "session_id": "  "}, None),
            ("non-string-id", {"hook_event_name": "SessionStart", "source": "startup", "session_id": 7}, None),
            ("multiline-id", {"hook_event_name": "SessionStart", "source": "startup", "session_id": "a\nb"}, None),
        ]
        for name, payload, raw in cases:
            with self.subTest(name=name):
                result = self.run_hook(payload, raw)
                self.assertEqual(result.returncode, 0)
                self.assertEqual(result.stdout, "")
                self.assertEqual(result.stderr, "")

    def test_json_escaping_preserves_the_opaque_identifier(self) -> None:
        session_id = 'opaque-"\\\t-한글'
        result = self.run_hook(
            {
                "hook_event_name": "SessionStart",
                "source": "startup",
                "session_id": session_id,
            }
        )
        self.assertEqual(result.returncode, 0)
        self.assertNotIn("\n", result.stdout)
        decoded = json.loads(result.stdout)
        self.assertEqual(decoded["hookSpecificOutput"]["additionalContext"], f"{MARKER_PREFIX}{session_id}")

    def test_hook_is_generic_and_configured_without_status_noise(self) -> None:
        source = HOOK.read_text(encoding="utf-8")
        for forbidden in (".tmp", "CODEX_THREAD_ID", "pathlib", "open("):
            self.assertNotIn(forbidden, source)

        config = json.loads(HOOKS_CONFIG.read_text(encoding="utf-8"))
        group = config["hooks"]["SessionStart"][0]
        handler = group["hooks"][0]
        self.assertEqual(group["matcher"], "startup|resume|clear|compact")
        self.assertIn("$PLUGIN_ROOT", handler["command"])
        self.assertIn("%PLUGIN_ROOT%", handler["commandWindows"])
        self.assertLessEqual(handler["timeout"], 5)
        self.assertNotIn("statusMessage", handler)
        self.assertNotIn("hooks", json.loads(MANIFEST.read_text(encoding="utf-8")))
        self.assertIn("py -3 <plugin-root>\\scripts\\memo.py", SKILL.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()
