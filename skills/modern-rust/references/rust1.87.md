# Rust 1.87

Released May 2025.

## Important changes

- Precise return-position `impl Trait` capture syntax, `+ use<...>`, is stable inside trait definitions. The same syntax outside traits had already stabilized in 1.82.
- Inline assembly can use label operands to jump to Rust blocks. Combining output and label operands remains unstable.
- Most `std::arch` intrinsics that were unsafe only because they require target features became safe inside functions with the appropriate features enabled. Pointer-taking or otherwise unsafe intrinsics remain unsafe.
- `std::io::pipe` and its reader and writer types provide portable anonymous pipes that integrate with `Command` and `Stdio`. Cargo also added `cargo package --exclude-lockfile`.

## Compatibility and migration

- Using these language or library additions raises MSRV to 1.87.
- The `i586-pc-windows-msvc` target was removed; migrate to `i686-pc-windows-msvc`. i686 targets now require SSE2 and use it when passing SIMD values.
- Newly safe intrinsics can produce `unused_unsafe` warnings. Do not remove unsafe blocks without checking why each intrinsic was unsafe. When capturing substantial child-process output through a pipe, read before or concurrently with waiting so a full operating-system buffer cannot block the child.

Official details: [Rust 1.87 announcement](https://blog.rust-lang.org/2025/05/15/Rust-1.87.0/).
