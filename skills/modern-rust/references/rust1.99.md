# Rust 1.99

Released October 2026.

## Important changes

- C-variadic function definitions are stable: `extern "C"` and `"C-unwind"` functions may take `...`, read through `core::ffi::VaList::next_arg` for `VaArgSafe` types.
- Attribute and derive macros may apply to outlined modules (`#[my_macro] mod foo;`). The language now guarantees that `UnsafeCell` contents may be accessed without going through `UnsafeCell::get`.
- Standard-library additions include `Vec::into_parts` and `from_parts`, `Box::into_non_null` and `from_non_null`, `VecDeque::retain_back`, `String::from_utf8_lossy_owned`, `std::fs::set_times` and `set_times_nofollow`, raw-pointer `size_of_val_raw`, `align_of_val_raw`, and `Layout::for_value_raw`, and `IntoIterator` for `Box<[T; N]>`.
- Cargo adds a built-in `debug` profile, currently identical to `dev`, ahead of a planned change to `dev` debug settings. Edition 2024 workspace members may override an inherited dependency's `default-features`. `riscv64-unknown-linux-musl` reaches Tier 2 with host tools.

## Compatibility and migration

- New APIs require MSRV 1.99; overriding inherited `default-features` also requires edition 2024, while older editions ignore it with a warning.
- Cargo disables incremental compilation by default when the `CI` environment variable is set, which can change CI build times and caching.
- Legacy integer modules such as `std::i32` are fully deprecated; use associated constants like `i32::MAX`. `no_mangle_generic_items` is a hard error, and `semicolon_in_expressions_from_non_local_macros` now covers macros from other crates.
- Prefer Rust 1.99.1 over 1.99.0: it fixes an x86_64 miscompilation of FFI `bool` returns and a missing `wcslen` symbol on UEFI and embedded targets.

Official details: [Rust 1.99 release notes](https://doc.rust-lang.org/stable/releases.html#version-1991-2026-10-15).
