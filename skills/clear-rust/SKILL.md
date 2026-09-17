---
name: clear-rust
description: "Plain Rust for any source, test, macro, Cargo crate, or API work: visible control flow, concrete design, narrow abstractions, and explicit safety boundaries."
---

# Clear Rust

Write **plain Rust**: make behavior, ownership, failure, and control flow easy
to see.

Use these design defaults within the repository's contracts, MSRV, conventions,
and measured performance requirements. For reviews or diagnosis, assess the code
without rewriting it unless requested. Preserve already-clear choices.

Loops and iterators, borrowing and cloning, enums and traits, and sync and async
are contextual choices: favor the form whose behavior is easiest to trace.

## Shape Code

- Start with functions, structs, and enums over registries, factories, plugin
  layers, or framework-like indirection. Add a layer when current variability
  or a public contract requires it.
- Start concrete. Introduce a trait, generic parameter, or trait object when it
  expresses a real contract used by current code. Keep bounds limited to the
  capabilities the implementation uses.
- Keep small, stable duplication when extracting it would couple unrelated
  flows or force readers to jump between files. Extract when the repeated code
  represents one coherent concept with one reason to change.
- Use `if`, `match`, `for`, and named intermediate values when branches,
  mutation, or error context matter. Keep short, linear iterator pipelines;
  split dense or side-effectful chains into named steps or a loop.
- Split functions and modules at responsibility boundaries, not arbitrary size
  limits. Keep logic together when separation would hide the flow.
- Use the simplest practical enum, struct, or newtype to make meaningful
  invalid states unrepresentable. Reserve typestate and generic state machines
  for invariants whose value justifies their machinery.
- Prefer the standard library and existing dependencies when they keep the
  implementation clear. Add a crate when it removes durable complexity the
  project should not own; check its maintenance, features, MSRV, and public API
  impact.

## Keep Ownership Visible

- Borrow when the caller should retain ownership; own values that must be stored
  or outlive the call. Return values instead of adding out-parameters.
- Accept an intentional, bounded clone when it makes ownership substantially
  clearer and the cost is acceptable. Investigate clones of large or hot-path
  data instead of using them only to silence the borrow checker.
- Let lifetime elision handle conventional signatures. Name lifetimes when the
  compiler requires them or when the relationship itself needs explanation.
- Keep shared ownership and interior mutability within a narrow boundary. Reach
  for `Rc`, `Arc`, `RefCell`, locks, or atomics when the ownership or concurrency
  model actually requires them.

## Make Failure and Safety Explicit

- Return `Result` for anticipated runtime failures, even when the current layer
  can only report the error and exit. Use `Option` for expected absence. Reserve
  deliberate panics for violated invariants and programmer bugs.
- Use `unwrap` or `expect` in tests, prototypes, or a locally evident
  impossible-failure case. State a non-obvious invariant in the `expect`
  message or an adjacent explanation.
- Follow the project's error model. Preserve useful source errors and add
  actionable context at boundaries; avoid collapsing distinct failures into
  undifferentiated strings too early.
- Validate input at the boundary. Encode stable domain distinctions with small
  types when doing so prevents real misuse.
- Stay in safe Rust when a maintained safe abstraction can do the job. When
  `unsafe` is required, keep the block small, state the invariant in a
  `SAFETY` comment, expose a safe interface where possible, and document every
  caller obligation in a public `# Safety` section.
- Keep sequential or synchronous code when concurrency is not a requirement.
  When async is required, preserve the established runtime, keep the sync/async
  boundary visible, and keep blocking work out of async execution paths.

## Use Macros Deliberately

- Use established macros such as standard derives, `println!`, `vec!`,
  assertions, formatting macros, and conventional framework macros normally.
- Express project-owned behavior with ordinary functions, types, traits, and
  control flow when they can represent it clearly.
- Use a small declarative macro when the interface genuinely needs syntax,
  variable arity, or mechanical repetition and the expansion remains obvious.
- Require a stronger present need for procedural macros and embedded DSLs.
  Keep custom syntax close to ordinary Rust, generated behavior unsurprising,
  generated names collision-resistant, paths explicit, and diagnostics
  actionable.
- Test representative custom-macro invocations and failure cases. Make the
  generated behavior understandable without reverse-engineering the macro.

## Keep APIs Conventional

- Keep the public surface no larger than current callers require. Use private
  fields unless direct representation access is part of the contract.
- Use standard naming, conversion traits, and common derived traits where they
  fit. Make operations with a clear receiver into methods, keep operator
  overloads unsurprising, and implement `Deref` only for smart-pointer behavior.
- Document public error, panic, and safety conditions. Add examples where they
  communicate why or how an API should be used.
- Comment on invariants, intent, and tradeoffs rather than restating syntax.

## Validate Proportionally

- Use the repository's commands first. Otherwise, after changing code, run the
  applicable subset of `cargo fmt --check`, `cargo check`, focused tests, and
  `cargo clippy`.
- Treat configured lints and Clippy's default groups as the baseline. Enable
  pedantic, restriction, or nursery lints only when the project already does or
  after evaluating individual lints for the current codebase.
- Prefer a justified local lint allowance over contorting clear code to satisfy
  an opinionated lint. Record why the exception is correct.
