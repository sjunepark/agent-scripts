# Rust 1.77

Released March 2024.

## Important changes

- C-string literals such as `c"hello"` are stable and produce `&'static CStr`; termination and interior-NUL validation happen at compile time.
- Recursive calls in an `async fn` are accepted when the recursive edge is indirect, such as `Box::pin(f()).await`. Direct recursion still creates an infinitely sized future. `mem::offset_of!` is also stable for struct fields.
- Cargo stabilized Package ID Spec and the `cargo::key=value` build-script directive syntax. When debuginfo is disabled and `strip` is unset, release profiles now strip debuginfo inherited from the standard library.

## Compatibility and migration

- These syntax and library additions require MSRV 1.77. Build scripts supporting older Cargo must keep emitting legacy `cargo:key=value` directives. `offset_of!` reports the selected layout; it does not make an otherwise unspecified layout stable.
- On x86, `i128` and `u128` alignment changed from 8 to 16 bytes. FFI calling-convention bugs can remain when rustc uses LLVM older than 18.
- Patch-only: 1.77.1 reverted default stripping on Windows MSVC because it broke backtraces. 1.77.2 fixed CVE-2024-24576; unsafe-to-escape Windows batch arguments can now fail with `InvalidInput`.

Official details: [Rust 1.77 release notes](https://doc.rust-lang.org/stable/releases.html#version-1772-2024-04-09).
