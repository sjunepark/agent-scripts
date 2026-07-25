# Rust 1.79

Released June 2024.

## Important changes

- Inline `const { ... }` expressions are stable and may use in-scope generics, enabling patterns such as `[const { None::<T> }; N]` without helper constants. Associated-type bounds such as `Iterator<Item: Copy>` are also stable, replacing many projection-heavy `where` clauses.
- References to temporaries produced by tail blocks in `if` and `match` expressions now receive the same automatic lifetime extension as block expressions. Older workarounds that bind each temporary first may be unnecessary.
- `rustc --check-cfg` is stable for checking conditional-configuration names and values. Separately, `cargo add` now respects `package.rust-version` when selecting a dependency version, unless explicitly overridden.

## Compatibility and migration

- New syntax and library APIs require an MSRV of 1.79. Cargo 1.79 does not yet automatically enable `check-cfg`; that integration arrived in 1.80.
- Cargo now rejects versionless optional dependencies such as `foo = { optional = true }`. For WebAssembly, the `wasm_c_abi` future-incompatibility lint may require upgrading generated bindings to `wasm-bindgen` 0.2.88 or newer.

Official details: [Rust 1.79 announcement](https://blog.rust-lang.org/2024/06/13/Rust-1.79.0/).
