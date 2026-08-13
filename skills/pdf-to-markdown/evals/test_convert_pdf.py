from __future__ import annotations

import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import unittest


WRAPPER = Path(__file__).parents[1] / "scripts" / "convert_pdf.py"
PAGE_MARKER_RE = re.compile(r"<!-- PAGE (\d+) -->")


class ConvertPdfTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temporary_directory = tempfile.TemporaryDirectory()
        self.root = Path(self.temporary_directory.name)
        self.source = self.root / "한국어 보고서.pdf"
        self.source_bytes = b"%PDF-1.4\nfixture\n"
        self.source.write_bytes(self.source_bytes)
        fake_script = self.root / "xberg-fake.py"
        fake_script.write_text(
            """#!/usr/bin/env python3
import json, os, pathlib, sys

pathlib.Path(os.environ["FAKE_XBERG_LOG"]).write_text("\\n".join(sys.argv[1:]))

if os.environ.get("FAKE_XBERG_ORT_FAIL"):
    print(
        "The requested API version [18] is not available, only API versions [1, 17] "
        "are supported in this build. Current ORT Version is: 1.17.1",
        file=sys.stderr,
    )
    print("Failed to initialize ORT API", file=sys.stderr)
    print("Mutex poisoned", file=sys.stderr)
    raise SystemExit(17)
if os.environ.get("FAKE_XBERG_FAIL"):
    print("fatal error: fixture extraction failed", file=sys.stderr)
    raise SystemExit(17)
if os.environ.get("FAKE_XBERG_INVALID_JSON"):
    print("not JSON")
    raise SystemExit(0)

if os.environ.get("FAKE_XBERG_WARNINGS"):
    for _ in range(3):
        print(
            "Dictionary used where Stream expected, treating as empty stream",
            file=sys.stderr,
        )

include_markers = True
if "--page-markers" in sys.argv:
    marker_index = sys.argv.index("--page-markers")
    include_markers = sys.argv[marker_index + 1] == "true"

if os.environ.get("FAKE_XBERG_SHIFTED_MARKERS"):
    pages = [
        {"page_number": 1, "is_blank": False},
        {"page_number": 2, "is_blank": True},
        {"page_number": 3, "is_blank": False},
    ]
    content = "<!-- PAGE 1 -->first\\n<!-- PAGE 2 -->third"
elif os.environ.get("FAKE_XBERG_BAD_MARKERS"):
    pages = [
        {"page_number": 1, "is_blank": False},
        {"page_number": 2, "is_blank": True},
        {"page_number": 3, "is_blank": False},
    ]
    content = "<!-- PAGE 1 -->first\\n<!-- PAGE 1 -->third"
else:
    pages = [{"page_number": 1, "is_blank": False}]
    marker = "<!-- PAGE 1 -->\\n" if include_markers else ""
    content = marker + "# 한국어 보고서\\n금액 12,345원, 날짜 2026-08-13\\n"

if os.environ.get("FAKE_XBERG_EMPTY"):
    content = ""
if os.environ.get("FAKE_XBERG_WHITESPACE"):
    content = "   "

page_count = len(pages)
if os.environ.get("FAKE_XBERG_BAD_PAGE_COUNT"):
    page_count += 1
result = {
    "content": content,
    "pages": pages,
    "metadata": {
        "format": {"page_count": page_count},
        "pages": {"total_count": page_count},
    },
}
json.dump({"result": result}, sys.stdout, ensure_ascii=False)
""",
            encoding="utf-8",
        )
        if os.name == "nt":
            self.fake_xberg = self.root / "xberg-fake.cmd"
            self.fake_xberg.write_text(
                f'@echo off\n"{sys.executable}" "%~dp0xberg-fake.py" %*\n'
            )
        else:
            self.fake_xberg = fake_script
            self.fake_xberg.chmod(0o755)
        self.log = self.root / "args.log"
        self.environment = {
            **os.environ,
            "FAKE_XBERG_LOG": str(self.log),
            "PYTHONIOENCODING": "utf-8",
        }

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
            encoding="utf-8",
            capture_output=True,
            env=environment or self.environment,
            check=False,
        )

    def test_ordinary_conversion_is_nonempty_and_preserves_source(self) -> None:
        completed = self.run_wrapper()
        output = self.source.with_suffix(".md")

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assertIn("한국어 보고서", output.read_text(encoding="utf-8"))
        self.assertEqual(self.source.read_bytes(), self.source_bytes)
        arguments = self.log.read_text().splitlines()
        for expected in (
            "--no-config-discovery",
            "--format",
            "json",
            "--extract-pages",
            "true",
        ):
            self.assertIn(expected, arguments)
        self.assertNotIn("--ocr", arguments)
        self.assertIn("Validated 1 PDF page(s)", completed.stderr)

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
        self.assertIn("한국어 보고서", output.read_text(encoding="utf-8"))
        self.assertEqual(self.source.read_bytes(), self.source_bytes)

    def test_overwrite_replaces_a_named_symlink_not_its_target(self) -> None:
        target = self.root / "target.md"
        target.write_text("target must remain\n")
        output = self.root / "report.md"
        try:
            output.symlink_to(target)
        except OSError as error:
            self.skipTest(f"symlinks are unavailable in this environment: {error}")

        completed = self.run_wrapper("--output", str(output), "--overwrite")

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assertFalse(output.is_symlink())
        self.assertIn("한국어 보고서", output.read_text(encoding="utf-8"))
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

    def test_no_page_markers_keeps_canonical_markdown_without_comments(self) -> None:
        completed = self.run_wrapper("--no-page-markers")
        output = self.source.with_suffix(".md")

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assertNotIn("<!-- PAGE", output.read_text(encoding="utf-8"))
        arguments = self.log.read_text().splitlines()
        marker_index = arguments.index("--page-markers")
        self.assertEqual(arguments[marker_index + 1], "false")

    def test_blank_page_markers_are_restored_without_changing_body(self) -> None:
        environment = {**self.environment, "FAKE_XBERG_SHIFTED_MARKERS": "1"}
        completed = self.run_wrapper(environment=environment)
        output = self.source.with_suffix(".md")

        self.assertEqual(completed.returncode, 0, completed.stderr)
        markdown = output.read_text(encoding="utf-8")
        self.assertEqual(
            [int(match.group(1)) for match in PAGE_MARKER_RE.finditer(markdown)],
            [1, 2, 3],
        )
        self.assertEqual(PAGE_MARKER_RE.sub("", markdown), "first\nthird")
        self.assertIn("restored 1 blank-page marker(s)", completed.stderr)

    def test_repeated_parser_warnings_are_aggregated(self) -> None:
        environment = {**self.environment, "FAKE_XBERG_WARNINGS": "1"}
        completed = self.run_wrapper(environment=environment)

        self.assertEqual(completed.returncode, 0, completed.stderr)
        self.assertIn("encountered 3 dictionary-as-stream object(s)", completed.stderr)
        self.assertEqual(completed.stderr.count("Dictionary used where"), 0)

    def test_layout_runtime_mismatch_reports_root_cause_not_mutex(self) -> None:
        output = self.source.with_suffix(".md")
        environment = {**self.environment, "FAKE_XBERG_ORT_FAIL": "1"}
        completed = self.run_wrapper("--layout", environment=environment)

        self.assertNotEqual(completed.returncode, 0)
        self.assertIn("ONNX Runtime API 18", completed.stderr)
        self.assertIn("1.17.1", completed.stderr)
        self.assertIn("ORT_DYLIB_PATH", completed.stderr)
        self.assertNotIn("Mutex poisoned", completed.stderr)
        self.assertFalse(output.exists())

    def test_invalid_json_or_page_metadata_publishes_nothing(self) -> None:
        for variable in (
            "FAKE_XBERG_INVALID_JSON",
            "FAKE_XBERG_BAD_PAGE_COUNT",
            "FAKE_XBERG_BAD_MARKERS",
        ):
            with self.subTest(variable=variable):
                output = self.root / f"{variable}.md"
                environment = {**self.environment, variable: "1"}
                completed = self.run_wrapper(
                    "--output", str(output), environment=environment
                )

                self.assertNotEqual(completed.returncode, 0)
                self.assertFalse(output.exists())

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
