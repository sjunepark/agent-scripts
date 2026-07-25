# Rust 1.76

Released February 2024.

## Important changes

- Rust now documents function-signature ABI compatibility and guarantees that `char` and `u32` are ABI-compatible. Do not generalize this guarantee to unrelated types or to unspecified Rust-layout structs.
- `ptr::addr_eq` offers an explicit data-address-only comparison for wide pointers. The new `ambiguous_wide_pointer_comparisons` lint helps identify comparisons that may unexpectedly include slice metadata or trait-object vtables.
- Cargo rejects `[lints]` in a virtual manifest because it was silently ignored; use `[workspace.lints]`. `cargo add --optional foo` now creates a same-name feature referencing `dep:foo`, and package-ID specifications can express Git and path sources more precisely.

## Compatibility and migration

- APIs introduced here, including `Option::inspect`, `Result::inspect`, and `type_name_of_val`, require MSRV 1.76. Prefer the older equivalent when maintaining a lower MSRV.
- `IMPLIED_BOUNDS_ENTAILMENT` became a hard error. Procedural macros must also avoid depending on the exact token grouping or text emitted by `TokenStream::to_string()`.
- Custom rustc builds now require LLVM 16 or newer; the `x86_64-sun-solaris` and `asmjs-unknown-emscripten` targets were removed.

Official details: [Rust 1.76 announcement](https://blog.rust-lang.org/2024/02/08/Rust-1.76.0/).
