#!/usr/bin/env python3
"""Safely convert one local PDF to Markdown with the Xberg CLI."""

from __future__ import annotations

import argparse
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile


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


def build_command(executable: str, source: Path, args: argparse.Namespace) -> list[str]:
    command = [
        executable,
        "extract",
        str(source),
        "--no-config-discovery",
        "--content-format",
        "markdown",
        "--page-markers",
        "false" if args.no_page_markers else "true",
    ]

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


def contains_non_whitespace(path: Path) -> bool:
    with path.open("rb") as stream:
        return any(chunk.strip() for chunk in iter(lambda: stream.read(64 * 1024), b""))


def main() -> int:
    args = parse_args()
    source_arg = args.source.expanduser()
    output_arg = args.output or source_arg.with_suffix(".md")
    source, output = validate_paths(source_arg, output_arg, args.overwrite)
    executable = resolve_xberg(args.xberg)
    command = build_command(executable, source, args)

    file_descriptor, temp_name = tempfile.mkstemp(
        prefix=f".{output.name}.", suffix=".tmp", dir=output.parent
    )
    temp_path = Path(temp_name)
    try:
        with os.fdopen(file_descriptor, "wb") as output_stream:
            completed = subprocess.run(command, stdout=output_stream, check=False)
        if completed.returncode != 0:
            raise SystemExit(f"Xberg failed with exit status {completed.returncode}")
        if not contains_non_whitespace(temp_path):
            raise SystemExit("Xberg produced empty Markdown; output was not created")
        os.chmod(temp_path, 0o644)
        finalize(temp_path, output, args.overwrite)
    finally:
        temp_path.unlink(missing_ok=True)

    print(output)
    return 0


if __name__ == "__main__":
    sys.exit(main())
