# Rust 1.86

Released April 2025.

## Important changes

- Trait-object upcasting is stable: references, smart pointers, and raw pointers can coerce from `dyn Subtrait` to `dyn Supertrait`.
- Safe functions may use `#[target_feature]`. They are safely callable only from contexts enabling the required feature; ordinary callers still need a justified unsafe call.
- Slices and `HashMap` gained safe disjoint mutable access. The compiler also made `missing_abi` warn by default, changed direct `rustc -O` to optimization level 3, and added debug assertions for null raw-pointer access.

## Compatibility and migration

- Code using these features or APIs has MSRV 1.86. Trait-object raw pointers still require valid vtables, and debug pointer assertions neither legalize undefined behavior nor provide release-build safety.
- The `wasm_c_abi` future-compatibility lint became a hard error; projects using `wasm-bindgen` need at least 0.2.89. Long-deprecated `#![no_start]` and `#![crate_id]` were removed, and casting a fieldless enum that implements `Drop` to an integer is now rejected.
- Rust 1.86 only warned that `i586-pc-windows-msvc` would be removed; removal occurred in 1.87. Broad safe calls to feature-gated `std::arch` intrinsics likewise arrived in 1.87, not 1.86.

Official details: [Rust 1.86 announcement](https://blog.rust-lang.org/2025/04/03/Rust-1.86.0/).
