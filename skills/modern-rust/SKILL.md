---
name: modern-rust
description: "Modern Rust release awareness and plain-Rust style preferences for any Rust source, test, macro, Cargo workspace, manifest, build script, toolchain, FFI, upgrade, review, or API work from Rust 1.72 onward."
---

# Modern Rust

Anchor Rust work in the repository's compatibility contract, not the model's
training cutoff.

## Orient

1. Inspect workspace and package `rust-version`, `edition`, and resolver
   settings in `Cargo.toml`; `rust-toolchain` or `rust-toolchain.toml`; and CI,
   container, and release toolchain pins when relevant.
2. Keep these signals distinct:

   | Signal | Meaning |
   | --- | --- |
   | `edition` | Per-crate language semantics; not an MSRV declaration |
   | `rust-version` | Minimum supported Rust compiler and Cargo |
   | Toolchain pin | Compiler selected for builds; not permission to raise MSRV |

3. Use syntax, standard-library APIs, Cargo features, and lints available at
   the declared MSRV and edition. If the MSRV is unstated, infer it
   conservatively from project policy and CI rather than raising it silently.
4. Read every release crossed by a toolchain or MSRV upgrade and any release
   whose feature affects the task. For an edition migration, run
   `cargo fix --edition` before changing `edition`, audit its safety-related
   edits, and test the supported toolchain range.
5. Use the latest patch in a selected `1.N` family. Point releases can fix
   miscompilations, security issues, and regressions without adding features.

## Style

When writing, refactoring, or reviewing Rust code, read
[plain Rust style](guides/style.md) for the user's design, ownership, macro,
and Clippy preferences.

## Release Index

Read a file when its feature affects the task or an upgrade crosses it.

- **2023-2025:** [1.72-1.87](references/rust1.72-1.87.md) MSRV, Rust 2024,
  resolver 3, build-script, target, and patch-release gotchas only.
- **2025:** [1.88](references/rust1.88.md) let-chains and cache collection ·
  [1.89](references/rust1.89.md) const inference and the Wasm C ABI ·
  [1.90](references/rust1.90.md) `rust-lld` and workspace publishing ·
  [1.91](references/rust1.91.md) C variadics and strict arithmetic ·
  [1.92](references/rust1.92.md) never fallback and zeroed allocation.
- **2026:** [1.93](references/rust1.93.md) musl and conditional assembly ·
  [1.94](references/rust1.94.md) `array_windows` and Cargo config inclusion ·
  [1.95](references/rust1.95.md) `if let` guards and `cfg_select!` ·
  [1.96](references/rust1.96.md) core ranges and stricter Wasm linking ·
  [1.97](references/rust1.97.md) v0 mangling and Cargo warning controls ·
  [1.98](references/rust1.98.md) integer formatting and runtime-symbol lints ·
  [1.99](references/rust1.99.md) C-variadic definitions and raw vector parts.

Rust remains on SemVer major 1; this skill treats each stable `1.N` feature
release as a version. Coverage starts at Rust 1.72, with 1.72-1.87 condensed
to decision notes, and ends at 1.99. For a newer or uncertain toolchain, check
the current [official release announcements](https://blog.rust-lang.org/releases/)
before making version-sensitive decisions. Use these files for orientation, then
verify exact APIs and behavior in the official documentation for the relevant
toolchain.
