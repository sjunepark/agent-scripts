# Rust 1.74

Released November 2023.

## Important changes

- Cargo stabilized `[lints.rust]`, `[lints.clippy]`, and workspace lint inheritance. These settings are only enforced by Cargo 1.74 or newer, so do not rely on them as the sole lint policy for an older MSRV.
- Cargo credential providers and authenticated private registries became stable. Alternative registries requiring authentication need an explicitly configured provider.
- Opaque return types, including `async fn` and `-> impl Trait`, may now mention `Self` and associated types that capture lifetimes from the surrounding scope.
- Notable library additions include `io::Error::other`, encoded-byte conversions for `OsStr`/`OsString`, `Saturating`, and owned file-descriptor or handle conversions for child-process streams.

## Compatibility and migration

- Minimum Apple targets rose to macOS 10.12 and iOS/tvOS 10.
- Cargo configuration arrays now merge from low to high precedence. Audit order-sensitive `rustflags`; new array-to-`Vec` conversions can also require type annotations in overly generic code.
- Prefer 1.74.1 over 1.74.0: it fixed LLVM access violations and subtyping regressions and clarified `mem::discriminant` guarantees.

Official details: [Rust 1.74 release notes](https://doc.rust-lang.org/stable/releases.html#version-1741-2023-12-07).
