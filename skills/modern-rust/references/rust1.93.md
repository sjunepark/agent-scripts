# Rust 1.93

Released January 2026.

## Important changes

- Bundled Linux musl targets moved to musl 1.2.5, improving DNS behavior but raising the minimum supported external musl version and exposing stale `libc` dependency assumptions.
- Statement-level `#[cfg]` attributes inside `asm!`, `global_asm!`, and `naked_asm!` became stable. C-style variadic declarations also became stable for the `system` ABI; defining variadic Rust functions remains unsupported on stable.
- Global allocators may now use `thread_local!` and `std::thread::current` without triggering standard-allocator reentrancy.
- Cargo's `CARGO_CFG_DEBUG_ASSERTIONS` now reflects the active profile, and `cargo clean --workspace` cleans all workspace members.

## Compatibility and migration

- These facilities require MSRV 1.93. For musl builds, update `libc` to at least 0.2.146; older releases may resolve `open64` incorrectly, so `cargo update` is usually sufficient.
- Emscripten `panic=unwind` switched to WebAssembly exception handling. Compile linked C or C++ objects with `-fwasm-exceptions`; the former escape hatch is not stable.
- Rust 1.93.1 fixed a compiler ICE affecting rustfmt, a Clippy false positive, and a `wasm32-wasip2` file-descriptor leak.

Official details: [Rust 1.93 release notes](https://doc.rust-lang.org/stable/releases.html#version-1931-2026-02-12).
