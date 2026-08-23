#!/usr/bin/env python3
"""Safely convert one local PDF to Markdown with the Xberg CLI."""

from __future__ import annotations

import argparse
from dataclasses import dataclass
import json
import os
from pathlib import Path
import re
import secrets
import shutil
import subprocess
import sys
import tempfile
from typing import Any, BinaryIO


ANSI_ESCAPE_RE = re.compile(r"\x1b(?:[@-_][0-?]*[ -/]*[@-~])")
DICTIONARY_STREAM_WARNING = (
    "Dictionary used where Stream expected, treating as empty stream"
)
ORT_VERSION_RE = re.compile(
    r"The requested API version \[(\d+)\] is not available, only API versions "
    r"\[([^\]]+)\] are supported in this build\. Current ORT Version is: "
    r"([^\r\n]+)"
)
MAX_DIAGNOSTIC_LINES = 5
MAX_DIAGNOSTIC_CHARS = 1_000
MAX_DIAGNOSTIC_FRAGMENT_BYTES = 64 * 1024


@dataclass
class DiagnosticSummary:
    dictionary_stream_warning_count: int
    ort_version_mismatch: tuple[str, str, str] | None
    ort_initialization_failed: bool
    useful_lines: list[str]
    fallback_lines: list[str]
    omitted_line_count: int


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Convert one local PDF to Markdown without silent overwrite."
    )
    parser.add_argument("source", type=Path, help="Path to the source PDF")
    parser.add_argument(
        "--output",
        type=Path,
        help="Destination Markdown path (default: source basename with .md)",
    )
    parser.add_argument(
        "--korean-ocr",
        action="store_true",
        help="Use PaddleOCR with Korean on pages classified as scans",
    )
    parser.add_argument(
        "--force-ocr",
        action="store_true",
        help="OCR every page; reserve for image-only or broken text layers",
    )
    parser.add_argument(
        "--layout",
        action="store_true",
        help="Enable adaptive layout-informed Markdown",
    )
    parser.add_argument(
        "--no-page-markers",
        action="store_true",
        help="Omit source-page comments from Markdown",
    )
    parser.add_argument(
        "--overwrite",
        action="store_true",
        help="Replace an existing destination (requires prior user authority)",
    )
    parser.add_argument(
        "--xberg",
        help="Explicit Xberg executable path (primarily for controlled environments)",
    )
    return parser.parse_args()


def resolve_xberg(explicit: str | None) -> str:
    if explicit:
        candidate = Path(explicit).expanduser()
        if candidate.is_file() and os.access(candidate, os.X_OK):
            return str(candidate.resolve())
        raise SystemExit(f"Xberg executable is not runnable: {candidate}")

    executable = shutil.which("xberg")
    if executable:
        return executable
    raise SystemExit(
        "Xberg is not installed or not on PATH. Obtain approval before installing it."
    )


def validate_paths(source: Path, output: Path, overwrite: bool) -> tuple[Path, Path]:
    source = source.expanduser().resolve()
    output = Path(os.path.abspath(output.expanduser()))

    if not source.is_file():
        raise SystemExit(f"Source PDF does not exist: {source}")
    if source.suffix.lower() != ".pdf":
        raise SystemExit(f"Source must have a .pdf extension: {source}")
    if source == output:
        raise SystemExit("Output path must differ from the source PDF")
    if (output.exists() or output.is_symlink()) and not overwrite:
        raise SystemExit(
            f"Refusing to overwrite existing output: {output}. "
            "Obtain explicit authority before using --overwrite."
        )
    if not output.parent.is_dir():
        raise SystemExit(f"Output directory does not exist: {output.parent}")
    return source, output


def build_command(
    executable: str,
    source: Path,
    args: argparse.Namespace,
    private_marker_format: str | None,
) -> list[str]:
    command = [
        executable,
        "extract",
        str(source),
        "--no-config-discovery",
        "--format",
        "json",
        "--content-format",
        "markdown",
        "--extract-pages",
        "true",
        "--page-markers",
        "false" if private_marker_format is None else "true",
    ]

    if private_marker_format is not None:
        command.extend(
            [
                "--config-json",
                json.dumps(
                    {"pages": {"marker_format": private_marker_format}},
                    separators=(",", ":"),
                ),
            ]
        )

    if args.korean_ocr:
        command.extend(
            [
                "--ocr",
                "true",
                "--ocr-backend",
                "paddle-ocr",
                "--ocr-language",
                "korean",
            ]
        )
        if not args.force_ocr:
            command.append("--ocr-scanned-pages")
    elif args.force_ocr:
        command.extend(["--ocr", "true"])

    if args.force_ocr:
        command.extend(["--force-ocr", "true"])

    if args.layout:
        command.extend(
            [
                "--layout",
                "--layout-strategy",
                "auto",
                "--use-layout-for-markdown",
            ]
        )

    return command


def finalize(temp_path: Path, output: Path, overwrite: bool) -> None:
    if overwrite:
        os.replace(temp_path, output)
        return

    try:
        os.link(temp_path, output)
    except FileExistsError as error:
        raise SystemExit(
            f"Output appeared during conversion; refusing to overwrite: {output}"
        ) from error
    else:
        temp_path.unlink()


def summarize_diagnostics(diagnostics_stream: BinaryIO) -> DiagnosticSummary:
    diagnostics_stream.seek(0)
    warning_count = 0
    version_mismatch: tuple[str, str, str] | None = None
    ort_initialization_failed = False
    useful_lines: list[str] = []
    fallback_lines: list[str] = []
    retained_lines: set[str] = set()
    omitted_line_count = 0
    pattern_tail = ""

    while fragment := diagnostics_stream.readline(MAX_DIAGNOSTIC_FRAGMENT_BYTES):
        cleaned = ANSI_ESCAPE_RE.sub("", fragment.decode("utf-8", errors="replace"))
        warning_count += cleaned.count(DICTIONARY_STREAM_WARNING)

        pattern_text = pattern_tail + cleaned
        if version_mismatch is None:
            version_match = ORT_VERSION_RE.search(pattern_text)
            if version_match:
                version_mismatch = version_match.groups()
        ort_initialization_failed = (
            ort_initialization_failed or "Failed to initialize ORT API" in pattern_text
        )
        pattern_tail = pattern_text[-MAX_DIAGNOSTIC_CHARS:]

        stripped = cleaned.strip()
        if not stripped or DICTIONARY_STREAM_WARNING in stripped:
            continue
        retained = stripped[:MAX_DIAGNOSTIC_CHARS]
        if retained in retained_lines:
            continue
        if len(fallback_lines) < MAX_DIAGNOSTIC_LINES:
            fallback_lines.append(retained)
            retained_lines.add(retained)
        else:
            omitted_line_count += 1
        if (
            len(useful_lines) < MAX_DIAGNOSTIC_LINES
            and any(word in stripped.lower() for word in ("error", "failed", "panic"))
            and retained not in useful_lines
        ):
            useful_lines.append(retained)

    return DiagnosticSummary(
        dictionary_stream_warning_count=warning_count,
        ort_version_mismatch=version_mismatch,
        ort_initialization_failed=ort_initialization_failed,
        useful_lines=useful_lines,
        fallback_lines=fallback_lines,
        omitted_line_count=omitted_line_count,
    )


def failure_message(returncode: int, summary: DiagnosticSummary) -> str:
    if summary.ort_version_mismatch is not None:
        requested_api, supported_apis, runtime_version = summary.ort_version_mismatch
        return (
            "Xberg ONNX Runtime mismatch: it requested ONNX Runtime API "
            f"{requested_api}, but the loaded runtime ({runtime_version.strip()}) supports "
            f"only {supported_apis}. On Windows, an older "
            r"C:\Windows\System32\onnxruntime.dll may have been selected; configure "
            "a compatible runtime with ORT_DYLIB_PATH. No output was created."
        )

    if summary.ort_initialization_failed:
        return (
            "Xberg could not initialize the ONNX Runtime API for layout extraction. "
            "Check the loaded ONNX Runtime library and ORT_DYLIB_PATH. "
            "No output was created."
        )

    selected_lines = summary.useful_lines or summary.fallback_lines
    detail = " | ".join(selected_lines) or "no diagnostic was emitted"
    if len(detail) > MAX_DIAGNOSTIC_CHARS:
        detail = f"{detail[: MAX_DIAGNOSTIC_CHARS - 3]}..."
    return f"Xberg failed with exit status {returncode}: {detail}"


def emit_success_diagnostics(summary: DiagnosticSummary) -> None:
    if summary.dictionary_stream_warning_count:
        print(
            "Xberg warning: pdf_oxide encountered "
            f"{summary.dictionary_stream_warning_count} dictionary-as-stream object(s) "
            "and treated them as empty "
            "streams; inspect affected visual content if fidelity is uncertain.",
            file=sys.stderr,
        )

    for line in summary.fallback_lines:
        print(f"Xberg diagnostic: {line}", file=sys.stderr)
    if summary.omitted_line_count:
        print(
            "Xberg diagnostic: "
            f"{summary.omitted_line_count} additional line(s) omitted",
            file=sys.stderr,
        )


def emit_processing_warnings(result: dict[str, Any]) -> None:
    warnings = result.get("processing_warnings")
    if not isinstance(warnings, list):
        return

    rendered: list[str] = []
    seen: set[tuple[str, str]] = set()
    omitted_count = 0
    for warning in warnings:
        if not isinstance(warning, dict):
            continue
        source = warning.get("source")
        message = warning.get("message")
        if not isinstance(source, str) or not isinstance(message, str):
            continue
        warning_key = (source.strip(), message.strip())
        if warning_key in seen:
            continue
        if len(rendered) < MAX_DIAGNOSTIC_LINES:
            seen.add(warning_key)
            rendered.append(
                f"[{warning_key[0] or 'unknown'}] "
                f"{warning_key[1] or 'unspecified warning'}"
            )
        else:
            omitted_count += 1

    for warning in rendered:
        if len(warning) > MAX_DIAGNOSTIC_CHARS:
            warning = f"{warning[: MAX_DIAGNOSTIC_CHARS - 3]}..."
        print(f"Xberg processing warning: {warning}", file=sys.stderr)
    if omitted_count:
        print(
            "Xberg processing warning: "
            f"{omitted_count} additional warning(s) omitted",
            file=sys.stderr,
        )


def load_result(json_stream: BinaryIO) -> dict[str, Any]:
    try:
        json_stream.seek(0)
        document = json.loads(json_stream.read().decode("utf-8"))
    except (OSError, UnicodeError, json.JSONDecodeError) as error:
        raise SystemExit(
            f"Xberg returned invalid JSON; output was not created: {error}"
        ) from error

    if not isinstance(document, dict) or not isinstance(document.get("result"), dict):
        raise SystemExit(
            "Xberg JSON does not contain a single object result; output was not created"
        )
    return document["result"]


def validate_pages(result: dict[str, Any]) -> tuple[list[dict[str, Any]], int]:
    pages = result.get("pages")
    if not isinstance(pages, list) or not pages:
        raise SystemExit("Xberg JSON contains no page metadata; output was not created")

    blank_count = 0
    for expected_number, page in enumerate(pages, start=1):
        if not isinstance(page, dict):
            raise SystemExit(
                f"Xberg page metadata entry {expected_number} is invalid; output was not created"
            )
        page_number = page.get("page_number")
        is_blank = page.get("is_blank")
        if page_number != expected_number or isinstance(page_number, bool):
            raise SystemExit(
                "Xberg page metadata is not a contiguous 1-based sequence; "
                "output was not created"
            )
        if not isinstance(is_blank, bool):
            raise SystemExit(
                f"Xberg page {expected_number} has no valid blank-page status; "
                "output was not created"
            )
        blank_count += int(is_blank)

    declared_counts: list[Any] = []
    metadata = result.get("metadata")
    if isinstance(metadata, dict):
        format_metadata = metadata.get("format")
        page_metadata = metadata.get("pages")
        if isinstance(format_metadata, dict):
            declared_counts.append(format_metadata.get("page_count"))
        if isinstance(page_metadata, dict):
            declared_counts.append(page_metadata.get("total_count"))

    valid_declared_counts = [
        count
        for count in declared_counts
        if isinstance(count, int) and not isinstance(count, bool)
    ]
    if not valid_declared_counts:
        raise SystemExit(
            "Xberg JSON contains no valid declared page count; output was not created"
        )
    if any(count != len(pages) for count in valid_declared_counts):
        raise SystemExit(
            "Xberg page-count metadata disagrees with its page records; "
            "output was not created"
        )

    return pages, blank_count


def normalize_page_markers(
    content: str,
    pages: list[dict[str, Any]],
    private_marker_re: re.Pattern[str],
) -> tuple[str, int]:
    matches = list(private_marker_re.finditer(content))
    marker_numbers = [int(match.group(1)) for match in matches]
    expected_numbers = list(range(1, len(pages) + 1))
    if marker_numbers == expected_numbers:
        normalized = private_marker_re.sub(
            lambda match: f"<!-- PAGE {match.group(1)} -->", content
        )
        return normalized, 0

    nonblank_numbers = [
        page["page_number"] for page in pages if not page["is_blank"]
    ]
    ordinal_numbers = list(range(1, len(nonblank_numbers) + 1))
    if marker_numbers not in (nonblank_numbers, ordinal_numbers):
        raise SystemExit(
            "Xberg page markers cannot be reconciled with its page metadata; "
            f"found {len(marker_numbers)} marker(s) for {len(pages)} page(s), "
            f"including {len(nonblank_numbers)} nonblank page(s). Output was not created."
        )

    pieces: list[str] = []
    generated_numbers: list[int] = []
    cursor = 0
    previous_page = 0
    for match, target_page in zip(matches, nonblank_numbers):
        body_piece = content[cursor : match.start()]
        pieces.append(body_piece)
        page_numbers = list(range(previous_page + 1, target_page + 1))
        generated_numbers.extend(page_numbers)
        pieces.extend(f"<!-- PAGE {page_number} -->" for page_number in page_numbers)
        cursor = match.end()
        previous_page = target_page
    final_body_piece = content[cursor:]
    pieces.append(final_body_piece)
    trailing_numbers = list(range(previous_page + 1, len(pages) + 1))
    generated_numbers.extend(trailing_numbers)
    pieces.extend(f"<!-- PAGE {page_number} -->" for page_number in trailing_numbers)
    normalized = "".join(pieces)

    if generated_numbers != expected_numbers:
        raise SystemExit(
            "Internal page-marker normalization failed; output was not created"
        )
    return normalized, len(pages) - len(matches)


def render_markdown(
    result: dict[str, Any], private_marker_re: re.Pattern[str] | None
) -> tuple[str, int, int, int]:
    content = result.get("content")
    if not isinstance(content, str):
        raise SystemExit(
            "Xberg JSON contains no Markdown content; output was not created"
        )
    pages, blank_count = validate_pages(result)
    repaired_count = 0
    if private_marker_re is not None:
        content, repaired_count = normalize_page_markers(
            content, pages, private_marker_re
        )
    if not content.strip():
        raise SystemExit("Xberg produced empty Markdown; output was not created")
    return content, len(pages), blank_count, repaired_count


def make_temp_file(output: Path, suffix: str) -> tuple[int, Path]:
    file_descriptor, temp_name = tempfile.mkstemp(
        prefix=f".{output.name}.", suffix=suffix, dir=output.parent
    )
    return file_descriptor, Path(temp_name)


def make_private_marker_format() -> tuple[str, re.Pattern[str]]:
    token = secrets.token_hex(16)
    marker_format = f"\n\n<!-- XBERG-PAGE-{token} {{page_num}} -->\n\n"
    marker_text = re.escape(marker_format.strip()).replace(
        re.escape("{page_num}"), r"(\d+)"
    )
    return marker_format, re.compile(rf"(?m)^{marker_text}$")


def main() -> int:
    args = parse_args()
    source_arg = args.source.expanduser()
    output_arg = args.output or source_arg.with_suffix(".md")
    source, output = validate_paths(source_arg, output_arg, args.overwrite)
    executable = resolve_xberg(args.xberg)
    private_marker_format: str | None = None
    private_marker_re: re.Pattern[str] | None = None
    if not args.no_page_markers:
        private_marker_format, private_marker_re = make_private_marker_format()
    command = build_command(executable, source, args, private_marker_format)

    markdown_path: Path | None = None
    try:
        with (
            tempfile.TemporaryFile(mode="w+b", dir=output.parent) as json_stream,
            tempfile.TemporaryFile(mode="w+b", dir=output.parent) as diagnostics_stream,
        ):
            completed = subprocess.run(
                command,
                stdout=json_stream,
                stderr=diagnostics_stream,
                check=False,
            )

            diagnostic_summary = summarize_diagnostics(diagnostics_stream)
            if completed.returncode != 0:
                raise SystemExit(
                    failure_message(completed.returncode, diagnostic_summary)
                )
            emit_success_diagnostics(diagnostic_summary)

            result = load_result(json_stream)
            emit_processing_warnings(result)
            content, page_count, blank_count, repaired_count = render_markdown(
                result, private_marker_re
            )

        markdown_descriptor, markdown_path = make_temp_file(output, ".md.tmp")
        with os.fdopen(
            markdown_descriptor, "w", encoding="utf-8", newline=""
        ) as stream:
            stream.write(content)
            if hasattr(os, "fchmod"):
                os.fchmod(stream.fileno(), 0o644)
        if not hasattr(os, "fchmod"):
            os.chmod(markdown_path, 0o644)
        finalize(markdown_path, output, args.overwrite)

        summary = (
            f"Validated {page_count} PDF page(s): {page_count - blank_count} nonblank, "
            f"{blank_count} blank"
        )
        if repaired_count:
            summary += f"; restored {repaired_count} blank-page marker(s)"
        print(f"{summary}.", file=sys.stderr)
    finally:
        if markdown_path is not None:
            markdown_path.unlink(missing_ok=True)

    print(output)
    return 0


if __name__ == "__main__":
    sys.exit(main())
