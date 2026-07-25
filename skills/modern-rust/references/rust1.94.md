# Rust 1.94

Released March 2026.

## Important changes

- Slices gained stable `array_windows::<N>()`, yielding overlapping `&[T; N]` windows without conversion. Calling it with `N = 0` panics.
- Cargo configuration files can recursively `include` required or optional TOML files. This is a top-level `.cargo/config.toml` feature, not the unrelated `package.include` manifest field.
- Cargo accepts TOML 1.1 syntax, including multiline inline tables, trailing commas in inline tables, and additional escapes. Cargo rewrites published manifests for registry consumers.
- Some FP16 intrinsics stabilized, but the primitive `f16` type itself remains unstable; do not present stable `f16` support.

## Compatibility and migration

- New Rust APIs require MSRV 1.94. Merely adopting TOML 1.1 raises the developer Cargo requirement to 1.94 and may break third-party TOML parsers, although it need not raise a published crate's consumer MSRV.
- Standard macros are now imported through the prelude, so conflicting glob imports can become ambiguous. More precise closure captures can also change borrow errors or drop timing.
- Rust 1.94.1 fixed WASI thread spawning, a Clippy ICE, unstable Windows API regressions, and Cargo tar vulnerabilities; crates.io was unaffected.

Official details: [Rust 1.94 release notes](https://doc.rust-lang.org/stable/releases.html#version-1941-2026-03-26).
