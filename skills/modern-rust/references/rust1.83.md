# Rust 1.83

Released November 2024.

## Important changes

- Const evaluation gained mutable references, raw pointers, interior mutation during evaluation, and references to static items. Many existing mutation, pointer, `Option`, `Result`, slice, and floating-point operations also became const-callable.
- The `expr_2021` macro fragment specifier stabilized in every edition so macro authors could preserve pre-2024 expression matching. Raw lifetimes and labels such as `'r#gen` also became valid.
- Low-level `Waker` construction and inspection stabilized, along with useful `ControlFlow`, `HashMap` entry, buffered-I/O, and error-kind operations. Cargo added `CARGO_MANIFEST_PATH` and `package.autolib`.

## Compatibility and migration

- New syntax and APIs require Rust 1.83. A constant still cannot contain a mutable reference, read mutable or interior-mutable static memory, or retain a reference to it; retaining a raw pointer is permitted.
- `expr_2021` did not stabilize the Rust 2024 Edition. The compiler also stopped accepting implicit coercions from `!`, while `cfg_attr`-selected `crate_name` and `crate_type` became hard errors; pass the corresponding compiler flags instead.
- `Waker::new` is unsafe and requires its vtable invariants to hold. No 1.83 patch release followed.

Official details: [Rust 1.83 announcement](https://blog.rust-lang.org/2024/11/28/Rust-1.83.0/).
