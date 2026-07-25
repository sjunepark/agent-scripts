# Rust 1.78

Released May 2024.

## Important changes

- `#[diagnostic]` and `#[diagnostic::on_unimplemented]` let trait authors improve compiler guidance. They are hints; exact use and rendering are not guaranteed.
- `#[cfg(target_abi = "...")]` is stable, and async trait methods can be implemented with equivalent concrete or desugared signatures. NaN patterns and const patterns whose type lacks `PartialEq` are now hard errors.
- Standard-library unsafe precondition checks now follow the consuming crate's debug assertions, exposing more undefined behavior in debug and test builds. `align_offset` and `align_to` gained deterministic runtime guarantees.
- Cargo stabilized lockfile v4, target-specific `rustdocflags`, and global-cache access tracking. It deprecated `.cargo/config` in favor of `.cargo/config.toml`.

## Compatibility and migration

- New syntax requires MSRV 1.78. Lockfile v3 remained the default, and cache deletion via `-Zgc` was still nightly-only despite stable access tracking.
- Tier-1 `*-pc-windows-*` toolchains and their output now require Windows 10. Bundled LLVM 18 completes the x86 `i128`/`u128` ABI correction; custom older LLVM builds can retain calling bugs.
- `wasm32-wasip1` arrived as Tier 2. `wasm32-wasip2` was only Tier 3 and incomplete, so do not present it as production-ready support.

Official details: [Rust 1.78 announcement](https://blog.rust-lang.org/2024/05/02/Rust-1.78.0/).
