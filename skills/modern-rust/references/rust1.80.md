# Rust 1.80

Released July 2024.

## Important changes

- `LazyCell` and thread-safe `LazyLock` are stable, providing standard-library lazy initialization without requiring `lazy_static` or `once_cell` for the common cases.
- Cargo now enables conditional-configuration checking for every build. Unknown `cfg` names or values trigger the warn-by-default `unexpected_cfgs` lint; declare custom values through `[lints.rust]` or `cargo::rustc-check-cfg`.
- Exclusive range patterns (`a..b` and `..b`) are stable, with gap and overlap lints helping prevent off-by-one mistakes.
- `Box<[T]>` implements `IntoIterator` in every edition, allowing by-value iteration. Useful additions also include `Option::take_if` and checked slice and string splitting.

## Compatibility and migration

- These features require an MSRV of 1.80. In Editions 2015–2021, `boxed.into_iter()` remains by-reference; in Edition 2024 it becomes by-value. Use `.iter()` to preserve references, or run `cargo fix --edition`.
- New `FromIterator` implementations for `Box<str>` can make inference ambiguous; upgrade `time` to at least 0.3.35 or add type annotations.
- Prefer patch release 1.80.1: it fixes a float-comparison miscompilation originating in the 1.78 jump-threading optimization and reverts `dead_code` false positives introduced by 1.80.0.

Official details: [Rust 1.80 release notes](https://doc.rust-lang.org/stable/releases.html#version-1801-2024-08-08).
