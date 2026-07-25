# Rust 1.95

Released April 2026.

## Important changes

- Match arms can use `if let` guards and mixed `&&` guard chains, with new bindings available to later guard operands and the arm body. Guards still do not satisfy exhaustiveness, so retain an unguarded fallback.
- `cfg_select!` provides stable first-match compile-time dispatch over `cfg` predicates in item or expression position, reducing the common need for `cfg-if`. Without a matching predicate or `_` fallback, it emits a compile error.
- Collection insertion methods such as `Vec::push_mut` and `Vec::insert_mut` now return a mutable reference to the inserted value; corresponding `VecDeque` and `LinkedList` operations also stabilized. Atomic types gained `update` and `try_update`.
- `rustc --remap-path-scope` stabilized for scoped path remapping, while rustdoc can hide deprecated items and ranks unstable search results lower.

## Compatibility and migration

- Using the new syntax or library APIs raises the MSRV to 1.95.
- Custom JSON target specifications were removed from stable Rust. They require nightly options; do not present them as stable custom-target support.
- Accidentally accepted `mut ref` and `mut ref mut` shorthand patterns are again rejected on stable.

Official details: [Rust 1.95 announcement](https://blog.rust-lang.org/2026/04/16/Rust-1.95.0/).
