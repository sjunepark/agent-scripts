# Rust 1.96

Released May 2026.

## Important changes

- New `core::range` types are copyable when their bounds are copyable and implement `IntoIterator` rather than `Iterator`. This makes ranges practical fields in `Copy` types without carrying iterator state.
- `assert_matches!` and `debug_assert_matches!` provide pattern assertions that print the mismatching value through `Debug`. They are not in the prelude and must be imported from `core` or `std`.
- Cargo permits a dependency to name both a Git repository for local use and an alternate registry for publication, and supports target-`cfg`-specific `rustdocflags`. It also fixed two vulnerabilities affecting third-party registries; crates.io users were unaffected.
- Rustdoc now renders deprecation notes as ordinary Markdown and separates methods from associated functions in sidebars.

## Compatibility and migration

- Using the new APIs raises the MSRV to 1.96.
- Range syntax such as `0..1` still constructs legacy `core::ops` ranges. A future edition may switch it; do not assume the new `core::range` types are syntax defaults.
- WebAssembly links now reject undefined symbols. Declare intended imports explicitly, or deliberately restore `--allow-undefined` through linker arguments.
- Prefer Rust 1.96.1 over 1.96.0: it fixes a MIR optimization miscompilation, Cargo retry and timeout handling, and vulnerabilities in the bundled `libssh2`.

Official details: [Rust 1.96 release notes](https://doc.rust-lang.org/stable/releases.html#version-1961-2026-06-30).
