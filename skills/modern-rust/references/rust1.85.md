# Rust 1.85

Released February 2025.

## Important changes

- Rust 2024 became stable. Its semantic changes include broader return-position `impl Trait` lifetime capture, shorter `if let` and tail-expression temporary scopes, new never-type fallback behavior, and broader `macro_rules!` `expr` matching.
- Edition 2024 tightens unsafe boundaries: `extern` blocks are unsafe; `no_mangle`, `export_name`, and `link_section` require `#[unsafe(...)]`; unsafe operations inside an unsafe function need explicit unsafe blocks; references to `static mut` are denied; and process-environment mutation becomes unsafe.
- Async closures and the `AsyncFn*` traits are stable.
- Edition 2024 implies Cargo resolver 3, which prefers dependencies compatible with `package.rust-version`. Rustdoc also attempts to merge 2024-edition doctests.

## Compatibility and migration

- Setting `edition = "2024"` raises MSRV to 1.85. Different-edition crates interoperate. Reserving `gen` does not make generator blocks stable.
- Run `cargo update` and then `cargo fix --edition` before changing the manifest. Fixes conservatively preserve old semantics, skip doctests and generated code, and cannot verify inserted unsafe blocks. Virtual workspaces must set `resolver = "3"` explicitly.
- Rust 1.85.1, not 1.85.0, fixed merged doctests falling back to separate executables, plus several regressions.

Official details: [Rust 1.85 release notes](https://doc.rust-lang.org/stable/releases.html#version-1851-2025-03-18).
