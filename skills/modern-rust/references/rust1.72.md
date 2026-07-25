# Rust 1.72

Released August 2023.

## Important changes

- Constant evaluation no longer stops at a fixed statement limit. Rust instead reports progressively slower warnings and uses a deny-by-default lint for likely runaway evaluation; deliberately expensive constants can opt out of the lint.
- rustc now remembers names and conditions of `cfg`-disabled items, so unresolved-item diagnostics can identify a missing crate feature or incompatible target.
- Several Clippy checks moved into rustc, including checks for ineffective `ManuallyDrop` drops, invalid UTF-8 literals, NaN comparisons, and reference-to-mutable-reference casts. `invalid_reference_casting` was still allow-by-default in 1.72.
- `mpsc::Sender<T>` became `Sync`, and Cargo doctests now run from their package root.

## Compatibility and migration

- Cargo rejects invalid feature names rather than warning. Audit private-registry manifests and generated features.
- Doctests that assumed the workspace root, plus snapshots of IPv4-compatible `Ipv6Addr` formatting, may need adjustment.
- Prefer 1.72.1 over 1.72.0: the patch fixed codegen and compile-time regressions, rustdoc lifetime rendering, and several compiler crashes. New 1.72 APIs or trait guarantees set MSRV 1.72.

Official details: [Rust 1.72 release notes](https://doc.rust-lang.org/stable/releases.html#version-1721-2023-09-19).
