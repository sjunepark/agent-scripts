# Rust 1.82

Released October 2024.

## Important changes

- `impl Trait + use<'a, T>` stabilized precise capture for return-position `impl Trait`, replacing lifetime capture and outlives tricks.
- Native `&raw const` and `&raw mut` create raw pointers without first creating references. Rust also gained `unsafe extern` blocks with safe foreign items and `#[unsafe(...)]` forms for `no_mangle`, `link_section`, and `export_name`.
- Floating-point arithmetic became available in `const fn`, with documented NaN semantics that intentionally do not guarantee identical sign or payload bits across architectures, optimization levels, or compile-time and runtime evaluation.
- Cargo gained `cargo info`; commonly useful library additions included `Option::is_none_or`, `iter::repeat_n`, and slice and iterator sortedness checks.

## Compatibility and migration

- These features raise the MSRV to 1.82. Precise capture was unavailable in trait definitions or implementations and had to list every in-scope type and const parameter.
- Forming a raw pointer is safe; dereferencing it is not. Removing now-redundant `unsafe` blocks can conflict with support for older compilers.
- Rust 2024 requirements mentioned for extern blocks and unsafe attributes were not stable in 1.82. No 1.82 patch release followed.

Official details: [Rust 1.82 announcement](https://blog.rust-lang.org/2024/10/17/Rust-1.82.0/).
