from __future__ import annotations

import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


WRAPPER = Path(__file__).parents[1] / "scripts" / "convert_pdf.py"


class ConvertPdfTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary_directory = tempfile.TemporaryDirectory()
        self.root = Path(self.temporary_directory.name)
        self.source = self.root / "한국어 보고서.pdf"
        self.source_bytes = b"%PDF-1.4\nfixture\n"
        self.source.write_bytes(self.source_bytes)
        self.fake_xberg = self.root / "xberg-fake"
        self.fake_xberg.write_text(
            """#!/usr/bin/env python3
import os, pathlib, sys
pathlib.Path(os.environ["FAKE_XBERG_LOG"]).write_text("\\n".join(sys.argv[1:]))
if os.environ.get("FAKE_XBERG_FAIL"):
    raise SystemExit(17)
if os.environ.get("FAKE_XBERG_WHITESPACE"):
    print("   ")
    raise SystemExit(0)
if not os.environ.get("FAKE_XBERG_EMPTY"):
    print("<!-- source-page: 1 -->")
    print("# 한국어 보고서")
    print("금액 12,345원, 날짜 2026-08-13")
"""
        )
        self.fake_xberg.chmod(0o755)
        self.log = self.root / "args.log"
        self.environment = {**os.environ, "FAKE_XBERG_LOG": str(self.log)}

    def tearDown(self) -> None:
        self.temporary_directory.cleanup()

    def run_wrapper(
        self, *arguments: str, environment: dict[str, str] | None = None
    ) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [
                sys.executable,
                str(WRAPPER),
                str(self.source),
                *arguments,
                "--xberg",
                str(self.fake_xberg),
            ],
            text=True,
            capture_output=True,
            env=environment or self.environment,
            check=False,
        )

    def test_ordinary_conversion_is_nonempty_and_preserves_source(self) -> None:
        completed = self.run_wrapper()
        output = self.source.with_suffix(".md")

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assertIn("한국어 보고서", output.read_text())
        self.assertEqual(self.source.read_bytes(), self.source_bytes)
        arguments = self.log.read_text().splitlines()
        self.assertIn("--no-config-discovery", arguments)
        self.assertNotIn("--ocr", arguments)

    def test_existing_output_is_not_overwritten(self) -> None:
        output = self.source.with_suffix(".md")
        output.write_text("existing\n")

        completed = self.run_wrapper()

        self.assertNotEqual(completed.returncode, 0)
        self.assertIn("Refusing to overwrite", completed.stderr)
        self.assertEqual(output.read_text(), "existing\n")
        self.assertFalse(self.log.exists())

    def test_explicit_overwrite_replaces_only_the_output(self) -> None:
        output = self.source.with_suffix(".md")
        output.write_text("existing\n")

        completed = self.run_wrapper("--overwrite")

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assertIn("한국어 보고서", output.read_text())
        self.assertEqual(self.source.read_bytes(), self.source_bytes)

    def test_overwrite_replaces_a_named_symlink_not_its_target(self) -> None:
        target = self.root / "target.md"
        target.write_text("target must remain\n")
        output = self.root / "report.md"
        output.symlink_to(target)

        completed = self.run_wrapper(
            "--output", str(output), "--overwrite"
        )

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assertFalse(output.is_symlink())
        self.assertIn("한국어 보고서", output.read_text())
        self.assertEqual(target.read_text(), "target must remain\n")
        self.assertEqual(self.source.read_bytes(), self.source_bytes)

    def test_korean_ocr_uses_selective_paddle_flags(self) -> None:
        completed = self.run_wrapper("--korean-ocr")
        arguments = self.log.read_text().splitlines()

        self.assertEqual(completed.returncode, 0, completed.stderr)
        for expected in (
            "--ocr",
            "true",
            "--ocr-backend",
            "paddle-ocr",
            "--ocr-language",
            "korean",
            "--ocr-scanned-pages",
        ):
            self.assertIn(expected, arguments)

    def test_layout_uses_adaptive_markdown_flags(self) -> None:
        completed = self.run_wrapper("--layout")
        arguments = self.log.read_text().splitlines()

        self.assertEqual(completed.returncode, 0, completed.stderr)
        for expected in (
            "--layout",
            "--layout-strategy",
            "auto",
            "--use-layout-for-markdown",
        ):
            self.assertIn(expected, arguments)

    def test_empty_or_failed_extraction_publishes_nothing(self) -> None:
        for variable in (
            "FAKE_XBERG_EMPTY",
            "FAKE_XBERG_WHITESPACE",
            "FAKE_XBERG_FAIL",
        ):
            with self.subTest(variable=variable):
                output = self.root / f"{variable}.md"
                environment = {**self.environment, variable: "1"}
                completed = self.run_wrapper(
                    "--output", str(output), environment=environment
                )

                self.assertNotEqual(completed.returncode, 0)
                self.assertFalse(output.exists())


if __name__ == "__main__":
    unittest.main()
