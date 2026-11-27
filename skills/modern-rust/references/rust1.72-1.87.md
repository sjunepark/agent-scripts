# Rust 1.72-1.87 decision notes

Released August 2023 to May 2025. Features from these releases are assumed
known; this file keeps only MSRV, edition, toolchain, and migration facts that
are easy to get wrong. Each feature still sets MSRV to the release that
stabilized it.

## Edition 2024 and resolver 3

- Rust 2024 stabilized in 1.85; `edition = "2024"` sets MSRV 1.85. Crates of
  different editions interoperate.
- Edition 2024 implies resolver 3 (MSRV-aware selection using
  `package.rust-version`), but a virtual workspace must set `resolver = "3"`
  explicitly. Resolver 3, or the Cargo config
  `resolver.incompatible-rust-versions = "fallback"`, is opt-in from 1.84 for
  older editions; fallback can still pick an incompatible release
  when nothing compatible satisfies the requirement.
- Migrate with `cargo update`, then `cargo fix --edition`, then change the
  manifest. Fixes skip doctests and generated code and cannot verify inserted
  unsafe blocks; review them. Use 1.85.1 or later for merged doctests.
- Edition 2024 changes semantics that `cargo fix` may not fully cover: `if let`
  scrutinee and tail-expression temporaries drop earlier, never-type fallback
  becomes `!`, references to `static mut` are denied, and `std::env::set_var`
  and `remove_var` become `unsafe`. Review drop-order-sensitive code (locks,
  guards) after migrating.
- `Box<[T]>::into_iter` is by-value only in Edition 2024; older editions keep
  by-reference method resolution.

## Cargo and build scripts

- `[lints]` tables (1.74) are ignored by older Cargo, so they cannot be the only
  lint policy below MSRV 1.74. Cargo rejects `[lints]` in a virtual manifest
  (1.76); use `[workspace.lints]`.
- Build scripts must emit `cargo:KEY=VALUE` when MSRV is below 1.77; Cargo 1.73
  and older reject `cargo::KEY`.
- Since 1.80 every build checks `cfg` names; declare custom cfgs through
  `[lints.rust] unexpected_cfgs = { check-cfg = [...] }` or
  `cargo::rustc-check-cfg`.
- `cargo add` respects `package.rust-version` since 1.79. Version-less packages
  (1.75) cannot be published.

## Language and API gotchas

- Bare `async fn` in a public trait (1.75): callers cannot add `Send` bounds,
  the method is not dyn-compatible, and adding bounds later is an API break.
- Panics crossing `extern "C"` and other non-unwind ABIs abort since 1.81; use a
  `-unwind` ABI when unwinding is intended.
- The 1.81 sort implementations may panic on inconsistent `Ord` impls or
  comparators. `hint::assert_unchecked(false)` is immediate undefined behavior.
- Standard-library unsafe precondition checks follow the consuming crate's
  debug assertions since 1.78, so latent UB can panic in debug and test builds.
- Removing `unsafe` blocks made redundant by 1.82 raw references or 1.87 safe
  `std::arch` intrinsics breaks older supported compilers; check MSRV first.
- `std::panic::PanicInfo` became `PanicHookInfo` (1.81);
  `core::panic::PanicInfo` is a different type.
- On x86, `i128`/`u128` alignment is 16 bytes since 1.77; FFI layouts that
  assumed 8 need review.
- 1.80 added `FromIterator` impls for `Box<str>` that break type inference in
  older `time` releases; require `time` 0.3.35 or later before raising the
  toolchain.
- When piping child output with `std::io::pipe` (1.87), read before or while
  waiting so a full OS buffer cannot block the child.

## Targets and patch releases

- Tier-1 Windows targets require Windows 10 (1.78); Apple minimums rose to
  macOS 10.12 and iOS/tvOS 10 (1.74). `wasm32-wasi` was removed in 1.84; use
  `wasm32-wasip1`. `wasm-bindgen` must be at least 0.2.89 (1.86).
  `i586-pc-windows-msvc` was removed in 1.87.
- Prefer patched toolchains: 1.77.2 and 1.81 fix Windows batch-file argument
  escaping (CVE-2024-24576, CVE-2024-43402); 1.80.1 fixes a float-comparison
  miscompilation; 1.85.1 fixes merged doctests.

Official details: [Rust release notes](https://doc.rust-lang.org/stable/releases.html).
