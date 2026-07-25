# Rust 1.89

Released August 2025.

## Important changes

- `_` may infer a const-generic argument in bodies, such as `[false; _]`; it remains forbidden in item signatures and constant type declarations.
- The warn-by-default `mismatched_lifetime_syntaxes` lint exposes hidden output lifetimes; use forms such as `Iter<'_, T>` when lifetimes connect. `dangerous_implicit_autorefs` became deny-by-default for raw-pointer dereferences.
- Cargo now runs cross-compiled doctests through the configured target runner. `cargo fix` and `cargo clippy --fix` now use normal build-target selection rather than every target.
- `wasm32-unknown-unknown` adopted the standard WebAssembly C ABI. Rust `i128`/`u128` may be used in `extern "C"`, but only match C `__int128` where that type and ABI exist; they are not generally `_BitInt(128)` compatible.

## Compatibility and migration

- These features require MSRV 1.89. New lints can break `-D warnings`; raw-pointer code may need an explicit, safety-audited borrow.
- Cross-target doctests may newly fail; configure a runner or use a target-specific `ignore-*` annotation. Rebuild both sides of affected Wasm FFI.
- No Rust 1.89 patch release followed; the macOS x86_64 Tier 2 demotion occurred in 1.90, not 1.89.

Official details: [Rust 1.89 announcement](https://blog.rust-lang.org/2025/08/07/Rust-1.89.0/).
