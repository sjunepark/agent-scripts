"""Regression test: the collector decodes gh output as UTF-8 regardless of locale.

Run with the locale encoding in effect (UTF-8 mode off) to exercise the bug on a
non-UTF-8 console such as Windows cp949:

    python -X utf8=0 evals/test_collector_encoding.py

A stand-in `gh` written in Python emits UTF-8 bytes that cp949 cannot decode;
`subprocess.run` is wrapped so the collector's real `gh_text`/`gh_json` decoding
path runs against it without network access.
"""
import importlib.util
import pathlib
import subprocess
import sys
import tempfile
import unittest

SCRIPT = pathlib.Path(__file__).resolve().parent.parent / "scripts" / "collect_pr_feedback.py"
PAYLOAD = '{"body": "LGTM \\u2014 \\ud83d\\udc4d caf\\u00e9 \\ud55c\\uad6d\\uc5b4"}'


def load_collector():
    spec = importlib.util.spec_from_file_location("collect_pr_feedback", SCRIPT)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class CollectorEncodingTest(unittest.TestCase):
    def test_gh_json_decodes_utf8_output(self):
        collector = load_collector()
        with tempfile.TemporaryDirectory() as tmp:
            fake_gh = pathlib.Path(tmp) / "fake_gh.py"
            fake_gh.write_text(
                "import json, sys\n"
                f"sys.stdout.buffer.write(json.dumps(json.loads({PAYLOAD!r}), ensure_ascii=False).encode('utf-8'))\n",
                encoding="utf-8",
            )
            real_run = subprocess.run

            def run_fake_gh(command, **kwargs):
                self.assertEqual(command[0], "gh")
                # Catches a revert on UTF-8 locales, where decoding alone passes.
                self.assertEqual(kwargs.get("encoding"), "utf-8")
                return real_run([sys.executable, str(fake_gh), *command[1:]], **kwargs)

            collector.subprocess.run = run_fake_gh
            try:
                result = collector.gh_json(["api", "repos/o/r/pulls/1"])
            finally:
                collector.subprocess.run = real_run
        self.assertEqual(result["body"], "LGTM — \U0001f44d café 한국어")


if __name__ == "__main__":
    unittest.main()
