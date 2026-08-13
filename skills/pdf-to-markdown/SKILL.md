---
name: pdf-to-markdown
description: "Convert PDF inputs into clean, reusable whole-document Markdown for agents, LLM ingestion, search, or archival text, using Xberg with safe output handling, Korean scan OCR, page markers, validation, and layout escalation. Use when the user asks to create, save, generate, normalize, reconvert, or extract a Markdown or whole-document agent-readable text artifact from a PDF. Do not use for summaries or questions with no converted artifact, PDF editing or rendering, parser research, or targeted extraction into CSV, JSON, images, or other structured formats."
---

# PDF to Markdown

Create a validated Markdown sibling while preserving the PDF as the visual
authority. Default to `report.pdf` plus `report.md`; do not retain parser JSON,
coordinates, extracted assets, or chunks unless the user names a consumer for
them.

## Convert safely

1. Resolve each source PDF and intended output path. Process multiple PDFs one
   at a time so each result can be validated independently.
2. If the intended output exists, stop before conversion. Do not replace it or
   invent an alternate destination. Pass `--overwrite` only after the user
   explicitly authorizes replacing that exact file; use another path only when
   the user selects or approves it.
3. Check for the `xberg` executable. If it is missing, show the current official
   installation method for the host platform and request approval before
   installing it. Do not run an install script, package manager, or model
   download prewarming without approval.
4. Resolve `scripts/convert_pdf.py` relative to this `SKILL.md` and run:

   ```bash
   python3 scripts/convert_pdf.py /absolute/path/report.pdf
   ```

   The wrapper writes `report.md` atomically, adds page markers, preserves the
   PDF, rejects empty output, and never invokes a shell.

   Its ordinary Xberg extraction command is:

   ```bash
   xberg extract /absolute/path/report.pdf \
     --no-config-discovery \
     --content-format markdown \
     --page-markers true
   ```

   The wrapper captures that command's standard output into a temporary file
   and publishes it only after validation. Do not replace the wrapper with a
   direct shell redirect when creating the final artifact.
5. For a PDF that contains scanned Korean pages, add `--korean-ocr`. This uses
   PaddleOCR with explicit Korean selection and OCRs only pages classified as
   scans while retaining native text elsewhere:

   ```bash
   python3 scripts/convert_pdf.py /absolute/path/report.pdf --korean-ocr
   ```

   The corresponding Xberg flags are:

   ```bash
   xberg extract /absolute/path/report.pdf \
     --no-config-discovery \
     --content-format markdown \
     --page-markers true \
     --ocr true \
     --ocr-backend paddle-ocr \
     --ocr-language korean \
     --ocr-scanned-pages
   ```

   The first PaddleOCR use may download models as part of the requested
   conversion. Tell the user before starting when network use or model storage
   is material in the current environment.

Do not OCR a born-digital PDF merely because its language is Korean. Use
`--force-ocr` only when the entire document is image-only or its text layer is
demonstrably broken; combine it with `--korean-ocr` for Korean documents.

If Python is unavailable but Xberg is present, reproduce the wrapper's exact
flags with a command runner while preserving the same non-overwrite, temporary
output, nonempty-result, and atomic-finalization guarantees. If those guarantees
cannot be preserved, stop and report the missing capability.

## Validate and escalate

After every conversion:

1. Inspect the beginning, middle, and end of the Markdown. Check headings,
   paragraph order, tables, lists, code, Hangul where expected, numbers, dates,
   missing pages, suspicious repetition, mojibake, and abrupt density changes.
2. Compare page-marker coverage with the source page count when a PDF inspection
   capability is available. Treat this as a diagnostic: a page can legitimately
   contain no text.
3. If ordinary extraction loses reading order, headings, tables, lists, or
   figure placement, rerun to a temporary comparison file with `--layout`:

   ```bash
   python3 scripts/convert_pdf.py /absolute/path/report.pdf \
     --output /absolute/path/report.layout.md \
     --layout
   ```

   This adds `--layout --layout-strategy auto
   --use-layout-for-markdown` to the ordinary Xberg command.

   Compare content fidelity, then replace the intended output only with the
   user's overwrite authority. Do not assume that the larger output is better.
4. If Korean OCR still omits or corrupts content, retry with `--force-ocr` only
   when the native text layer is the cause. For complex visual pages that remain
   unreliable, report the failing pages and offer a Docling or document-VLM
   comparison rather than silently switching tools or installing another
   parser.

## Finish

Finish only when the Markdown exists, is nonempty, passed representative source
checks, and the source PDF remains unchanged. Report the PDF and Markdown paths,
Xberg version, OCR or layout options used, any overwritten file explicitly
authorized by the user, and any pages or structures that remain uncertain.
