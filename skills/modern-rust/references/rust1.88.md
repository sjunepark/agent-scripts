# Rust 1.88

Released June 2025.

## Important changes

- Let chains became stable in `if` and `while`: multiple `let` patterns and boolean conditions may be joined with `&&`, with earlier bindings available later in the chain. This syntax is stable only in Edition 2024; do not add a feature gate or emit it for older editions.
- Stable naked functions use `#[unsafe(naked)]` and a single `naked_asm!` body. Boolean configuration predicates `cfg(true)` and `cfg(false)` are also stable in Rust attributes, `cfg!`, and Cargo target tables.
- Cargo now cleans its download cache automatically. It skips cleaning under `--offline` and `--frozen`; the separate `cargo gc` command was not stabilized.
- Commonly useful stable APIs include `HashMap::extract_if`, `HashSet::extract_if`, and slice `as_chunks`/`as_rchunks`.

## Compatibility and migration

- Code using these additions has MSRV 1.88; let chains additionally require `edition = "2024"`.
- Stable `#[bench]` usage became a hard error, and `i686-pc-windows-gnu` moved from Tier 1 to Tier 2.
- No Rust 1.88 patch release followed, so do not attribute later fixes to 1.88.x.

Official details: [Rust 1.88 announcement](https://blog.rust-lang.org/2025/06/26/Rust-1.88.0/).
