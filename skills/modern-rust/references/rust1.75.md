# Rust 1.75

Released December 2023.

## Important changes

- Return-position `impl Trait` and `async fn` became stable in traits and trait implementations. This removes the need for a proc macro in many statically dispatched async traits, but not for every public or dynamically dispatched design.
- Raw pointers gained byte-granularity `byte_add`, `byte_offset`, `byte_sub`, and wrapping variants, avoiding casts through byte pointers.
- `Option` gained `as_slice` and `as_mut_slice`; atomic values can be viewed through `Atomic*::from_ptr`; file timestamp setters also became stable.
- Cargo permits manifests without `package.version`, defaulting them to `0.0.0`, and `cargo new` adds nested packages to workspace members.

## Compatibility and migration

- Callers cannot add bounds such as `Send` to a trait's opaque return type. Bare `async fn` in a public trait therefore warns, and these methods are not dyn-compatible. Adding bounds later is an API break; native async traits set MSRV 1.75.
- Version-less packages cannot be published, and Cargo before 1.75 requires `package.version`.
- FreeBSD targets now require version 12, const misalignment is a hard error, and legacy compiler-plugin support was removed. There was no 1.75.x patch release.

Official details: [Rust 1.75 announcement](https://blog.rust-lang.org/2023/12/28/Rust-1.75.0/).
