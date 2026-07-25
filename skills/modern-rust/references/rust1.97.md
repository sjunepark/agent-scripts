# Rust 1.97

Released July 2026.

## Important changes

- Rust’s v0 symbol-mangling scheme became the stable default, preserving more useful generic information in symbols than the legacy scheme.
- Cargo stabilized `build.warnings`, with `allow`, `warn`, and `deny` levels for adjustable lint warnings in local packages. Unlike `RUSTFLAGS=-Dwarnings`, changing it does not invalidate the build cache.
- Cargo stabilized `resolver.lockfile-path` for builds from read-only source trees; the configured path must end in `Cargo.lock`. `cargo clean --target-dir` also refuses directories that do not resemble Cargo target directories.
- Successful linker stderr is now surfaced by the warn-by-default `linker_messages` lint. Integer and `NonZero` types also gained stable bit-width and highest/lowest-set-bit operations.

## Compatibility and migration

- New source APIs require MSRV 1.97; new Cargo configuration sets a development-toolchain floor of 1.97.
- Older debuggers and profilers may not demangle v0 symbols, and backtraces may change. Configure `linker_messages` independently when intentional linker output is noisy.
- Prefer Rust 1.97.1 over 1.97.0. The patch fixes an LLVM optimization miscompilation whose likelihood increased in 1.97.0; this is a compiler fix, not a base-release feature.

Official details: [Rust 1.97 release notes](https://doc.rust-lang.org/stable/releases.html#version-1971-2026-07-16).
