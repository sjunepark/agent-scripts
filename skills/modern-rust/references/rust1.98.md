# Rust 1.98

Released August 2026.

## Important changes

- Standard-library additions include `str::substr_range`, `[T]::subslice_range`, `strip_circumfix` for slices and strings, `bool::ok_or` and `ok_or_else`, endian-specific `String::from_utf16le`/`from_utf16be` constructors and lossy variants, `NonZero::from_str_radix`, and `Atomic<T>::from_mut` with slice conversions.
- Allocation-free integer formatting is stable through `core::fmt::NumBuffer` and `{integer}::format_into`; floats gain `algebraic_*` arithmetic that permits reassociation-style optimizations.
- `std::range::legacy` names the `core::ops` range types alongside the 1.96 `core::range` types. `{core,std}::derive` is now an intentionally stable path.
- New runtime-symbol lints reject or warn on definitions of core symbols such as `memcmp`, `memset`, and `strlen`; `c_void_returns` warns on `c_void` return types. Several bare-metal Armv7-A/R and Armv8-R targets reach Tier 2.

## Compatibility and migration

- New APIs require MSRV 1.98, except `{core,std}::derive`, which already worked from 1.96.
- More ambiguous glob imports become hard errors, and `repr(transparent)` no longer treats `repr(C)`, private-field, or `#[non_exhaustive]` fields as trivial.
- `std::env::Vars` and `VarsOs` are no longer `Send` or `Sync`. Inconsistent `PartialOrd` and `Ord` implementations can misbehave under the new derived `PartialOrd` fast path. Solaris `File::lock` now returns unsupported.
- Prefer Rust 1.98.1 over 1.98.0: it fixes a vtable-generation miscompilation.

Official details: [Rust 1.98 release notes](https://doc.rust-lang.org/stable/releases.html#version-1981-2026-09-03).
