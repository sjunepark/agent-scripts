# Rust 1.81

Released September 2024.

## Important changes

- The `Error` trait is now available as `core::error::Error`, letting `#![no_std]` libraries share a standard error abstraction.
- `#[expect(lint)]` is stable: it suppresses an anticipated lint but warns when that lint no longer occurs. Lint-level attributes also accept `reason = "..."`, keeping policy explanations in diagnostics.
- Uncaught panics crossing non-unwind ABIs such as `extern "C"` now abort; use a `-unwind` ABI when unwinding is intentional. New implementations back both `sort*` and `sort_unstable*`; “unstable” here describes equal-element ordering, not a nightly feature.
- `std::process::Command` fixes Windows batch-file argument escaping for CVE-2024-43402.

## Compatibility and migration

- New APIs and attributes require an MSRV of 1.81. `hint::assert_unchecked` is now stable but unsafe: passing `false` is immediate undefined behavior, so do not use it as a normal assertion.
- `std::panic::PanicInfo` was renamed to `PanicHookInfo`; the old alias starts warning in 1.82, while `core::panic::PanicInfo` remains distinct. Migrate `wasm32-wasi` builds to the identical `wasm32-wasip1` target.
- The new sorts may panic for inconsistent `Ord` implementations or comparators. Cargo packaging now rejects declared README or license files that do not exist.

Official details: [Rust 1.81 announcement](https://blog.rust-lang.org/2024/09/05/Rust-1.81.0/).
