# Rust 1.92

Released December 2025.

## Important changes

- The `never_type_fallback_flowing_into_unsafe` and `dependency_on_unit_never_type_fallback` lints became deny-by-default. Code that silently depended on `()` fallback around unsafe operations now needs explicit types.
- Linux binaries built with `panic=abort` now emit unwind tables by default, enabling backtraces through Rust frames despite aborting. This can increase binary size; `-Cforce-unwind-tables=no` opts out.
- Taking `&raw const` or `&raw mut` pointers to union fields is allowed in safe code. Zero-initialized `Box`, `Rc`, and `Arc` constructors also stabilized, but return `MaybeUninit`.
- Invalid arguments to `macro_export` are denied by default and can now be reported from dependencies.

## Compatibility and migration

- New syntax, lint expectations, and APIs require MSRV 1.92. The general-position `!` type is still not fully stable; do not treat the fallback lint changes as that stabilization.
- Call `assume_init` on zeroed allocations only when the all-zero bit pattern is valid for the element type. `Repeat::last` and `Repeat::count` now panic instead of looping forever.
- There was no Rust 1.92 patch release; do not attribute fixes from later lines to 1.92.

Official details: [Rust 1.92 announcement](https://blog.rust-lang.org/2025/12/11/Rust-1.92.0/).
