# Rust 1.84

Released January 2025.

## Important changes

- Cargo stabilized MSRV-aware dependency selection through a `fallback` policy or resolver 3, preferring releases compatible with the package’s declared `rust-version`.
- The next-generation trait solver began handling coherence only. It could newly accept provably disjoint implementations or reject previously missed overlaps.
- Strict and exposed-provenance pointer APIs stabilized address inspection and mapping without integer-pointer round trips. Integer square-root and additional const-capable pointer, atomic, and `Pin` operations also stabilized.
- Creating a raw reference to a raw-pointer dereference no longer required `unsafe`; actually reading or writing through the pointer remained unsafe.

## Compatibility and migration

- Resolver 3 was opt-in and requires Rust 1.84; its Rust 2024 default was still future behavior. Fallback may choose an incompatible release when no compatible candidate satisfies the requirement.
- The `wasm32-wasi` target was removed in favor of `wasm32-wasip1`, and unsupported calling conventions became hard errors.
- Patch 1.84.1 fixed false overlapping-implementation errors in incremental builds, new-solver compilation slowdowns, an ICE, and a debuginfo regression; these were fixes, not 1.84.0 features.

Official details: [Rust 1.84 release notes](https://doc.rust-lang.org/stable/releases.html#version-1841-2025-01-30).
