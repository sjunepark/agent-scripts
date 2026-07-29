---
name: write-go-docs
description: "Go source documentation: write, revise, prune, or audit package and declaration comments, examples, and deprecation notices."
---

# Write Go Docs

Treat information gain as the admission test: source documentation earns its place when it reduces uncertainty beyond the name, signature, types, and immediate context. Keep changes inside Go source and Go example tests; route README and architecture work to their own documentation workflows.

## Workflow

### 1. Bound the work and gather evidence

- Follow the user's requested files, packages, declarations, or diff. Expand to callers, tests, interfaces, implementations, and history only as needed to establish public behavior or durable rationale.
- Honor audit-only or review-only requests by keeping source unchanged and returning evidence-backed decisions.
- Treat files carrying the standard `// Code generated ... DO NOT EDIT.` marker as read-only and report the generator or source artifact that owns their documentation.
- Inspect `go.mod`, nearby documentation style, and configured validation when syntax or lint policy may depend on them.
- Treat code and tests as evidence of behavior, not proof of unstated intent. Record a focused maintainer question when important rationale remains unresolved.

Continue once the documentation boundary, relevant conventions, enforced checks, and unresolved questions are explicit.

### 2. Apply the information-gain test

For each package, exported declaration, existing doc comment, and non-obvious internal declaration in scope, choose retain, add, revise, remove, or skip.

Use a burden-of-proof rule:

- Default each candidate to remove or skip. Promote it only when you can name specific information that the declaration does not make obvious.
- Apply a deletion test to every existing or proposed comment: if its removal loses no non-obvious information, choose remove or skip.
- Give each fact one owning comment instead of repeating a type's contract on its constructor or methods.
- Treat exported status as an audit trigger, not evidence that prose is useful. When configured validation enforces coverage, write the shortest accurate convention-compliant synopsis and report the lint-driven compromise.
- Let conventional constructors, accessors, setters, wrappers, and `(value, ok)` lookups pass only when they add semantics such as ownership, synchronization, validation, caching, or unusual zero-value behavior.

Document reader-relevant knowledge that the declaration does not already reveal:

- semantic guarantees, valid states, invariants, and boundary conditions;
- zero-value behavior, ownership, lifecycle, mutation, and resource release;
- concurrency safety, blocking, cancellation, ordering, and idempotency;
- errors, partial results, panics, side effects, and retry implications;
- package responsibilities, cross-package relationships, and durable tradeoffs;
- non-obvious composition or misuse risks that justify an example.

Continue once every documentation candidate in scope has an evidence-backed decision.

For an audit-only request, skip editing, run applicable read-only validation from step 4, and report the decisions. Otherwise continue.

### 3. Write idiomatic Go documentation

- Place a doc comment immediately before the package clause or declaration it documents. Start package comments with `Package <package-name>`; center a declaration's opening synopsis on its declared name and follow any stricter configured prefix rule.
- Prefer one package comment that explains purpose and boundaries. Use `doc.go` when substantial package guidance deserves a dedicated file.
- Give a declaration group one comment only when its members share one semantic contract; otherwise document the meaningful members individually.
- Keep volatile literals, defaults, member inventories, and implementation sequences in code. State the stable semantic rule.
- Use Go documentation links such as `[Type]` and `[pkg.Type]` when they clarify relationships. Add headings or lists only when a longer package comment needs structure.
- Start a separate deprecation paragraph with `Deprecated:` and name the supported replacement or migration constraint when repository evidence provides one.
- Add an `Example`, `ExampleFoo`, or `ExampleFoo_Bar` only when executable usage communicates more clearly than prose. Keep its output assertion meaningful.

Continue once each changed artifact is associated with the intended Go symbol, follows Go formatting conventions, and passes the information-gain test.

### 4. Validate and report

- Run `gofmt` on changed Go files.
- Run the narrowest relevant `go test` command for documentation examples and the repository's configured documentation or lint checks.
- Inspect association, synopsis, and grouping with `go doc`. Use a local `pkgsite` when available to verify rendered links or headings.
- If executable validation is unavailable, inspect comment placement and example naming directly and report the limitation.

Complete when relevant validation passes or its limitation is reported, and the response identifies comments added, revised, or removed; notable omissions; unresolved maintainer questions; and every lint-driven low-information comment.
