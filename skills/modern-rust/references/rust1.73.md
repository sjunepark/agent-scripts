# Rust 1.73

Released October 2023.

## Important changes

- The default panic handler and `assert_eq!`/`assert_ne!` changed their output layout, putting the panic message on a new line and simplifying assertion diagnostics.
- Thread-local `Cell` and `RefCell` values gained direct `get`, `set`, `take`, `replace`, and borrow helpers. In particular, `set` can avoid evaluating the declared initializer.
- Unsigned integers gained `div_ceil`, `next_multiple_of`, and `checked_next_multiple_of`; `Arc<File>` gained `Read`, `Write`, and `Seek`.
- `invalid_reference_casting` became deny-by-default, while `noop_method_call` became warn-by-default. Code previously accepted around invalid reference casts can now fail compilation.

## Compatibility and migration

- Update panic-output snapshots and parsers; the text and line structure are intentionally different.
- Cargo 1.73 deliberately rejects build-script output using `cargo::KEY`. For an MSRV below 1.77, emit the older `cargo:KEY` form.
- Source builds using an external LLVM now require LLVM 15, and rustc rejects previously accepted non-defining uses of return-position `impl Trait`. Using the new APIs sets MSRV 1.73. There was no 1.73.x patch release.

Official details: [Rust 1.73 announcement](https://blog.rust-lang.org/2023/10/05/Rust-1.73.0/).
