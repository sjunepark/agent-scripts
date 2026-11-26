# Plain Rust Style

Use when writing, refactoring, or reviewing Rust code. These are the user's
preferences where they lean further toward plainness than common idiomatic
Rust. Apply them within the repository's contracts, conventions, and measured
performance needs. In reviews, assess without rewriting unless asked, and
preserve choices that are already clear.

## Plain over abstract

- Start with functions, structs, and enums. Add registries, factories, plugin
  layers, or framework-like indirection only when current variability or a
  public contract requires it.
- Start concrete. Add a trait, generic parameter, or trait object only for a
  contract current code uses; keep bounds to what the implementation needs.
- Keep small, stable duplication when extracting it would couple unrelated
  flows or make readers jump between files.
- Prefer `if`, `match`, `for`, and named intermediates when branches, mutation,
  or error context matter. Keep only short, linear iterator chains; split dense
  or side-effectful ones into named steps or a loop.
- Split functions and modules at responsibility boundaries, not size limits.
- Use the simplest enum, struct, or newtype that rules out meaningful invalid
  states. Reserve typestate and generic state machines for invariants that
  justify the machinery.
- Stay synchronous unless concurrency is a requirement.

## Ownership and failure

- An intentional, bounded clone is fine when it makes ownership clearer and the
  cost is acceptable. Investigate clones of large or hot-path data rather than
  using them to silence the borrow checker.
- Name lifetimes only when the compiler requires it or the relationship itself
  needs explanation.
- Keep `Rc`, `Arc`, `RefCell`, locks, and atomics inside a narrow boundary.
- Return `Result` for anticipated failures even when the layer can only report
  and exit. Limit `unwrap`/`expect` to tests, prototypes, and locally evident
  impossible cases; put a non-obvious invariant in the `expect` message.
- Do not collapse distinct failures into strings early.

## Macros

- Express project-owned behavior with functions, types, and traits when they
  can represent it clearly.
- Write a declarative macro only when the interface needs syntax, variable
  arity, or mechanical repetition and the expansion stays obvious. Procedural
  macros and embedded DSLs need a stronger present need.
- Test representative custom-macro invocations and failure cases.

## Lint policy

- Configured lints plus Clippy's default groups are the baseline. Enable
  `pedantic`, `restriction`, or `nursery` lints only when the project already
  does or after evaluating each lint for this codebase.
- Prefer a justified local allowance over contorting clear code for an
  opinionated lint, and record why the exception is correct.
