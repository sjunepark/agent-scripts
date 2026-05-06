#!/usr/bin/env python3
"""
Rasterize PDF pages to PNG for OCR/visual reading.

Use this for scanned or CID-encoded PDFs where text extraction gives garbage.
For text-based PDFs, just use pdftotext or pypdf directly — no need for this.

Usage:
  python pdf_to_images.py <input.pdf> <out_dir> [--dpi 200] [--pages 1,2,5-9]

Output: PNG files named page_001.png, page_002.png, ...
"""
from __future__ import annotations

import argparse
import sys
from pathlib import Path


def parse_pages(spec: str | None, n_pages: int) -> list[int]:
    if not spec:
        return list(range(1, n_pages + 1))
    out: set[int] = set()
    for part in spec.split(","):
        part = part.strip()
        if "-" in part:
            a, b = part.split("-", 1)
            for p in range(int(a), int(b) + 1):
                out.add(p)
        else:
            out.add(int(part))
    return sorted(p for p in out if 1 <= p <= n_pages)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("input")
    ap.add_argument("out_dir")
    ap.add_argument("--dpi", type=int, default=200,
                    help="Render DPI. 200 is a good default; raise for tiny digits.")
    ap.add_argument("--pages", default=None,
                    help="Page spec like '1,3,5-9'. Default: all pages.")
    args = ap.parse_args()

    in_path = Path(args.input)
    out_dir = Path(args.out_dir)
    out_dir.mkdir(parents=True, exist_ok=True)

    # pdf2image requires poppler; pymupdf (fitz) is self-contained — prefer fitz
    try:
        import fitz  # PyMuPDF
    except ImportError:
        print("PyMuPDF not installed. Run: pip install --break-system-packages PyMuPDF",
              file=sys.stderr)
        sys.exit(1)

    doc = fitz.open(in_path)
    n_pages = doc.page_count
    pages = parse_pages(args.pages, n_pages)

    zoom = args.dpi / 72.0
    mat = fitz.Matrix(zoom, zoom)

    for p in pages:
        page = doc.load_page(p - 1)  # 0-indexed
        pix = page.get_pixmap(matrix=mat, alpha=False)
        out_path = out_dir / f"page_{p:03d}.png"
        pix.save(out_path)
        print(f"  {out_path}  ({pix.width}x{pix.height})")

    print(f"Wrote {len(pages)} page(s) to {out_dir}")


if __name__ == "__main__":
    main()
