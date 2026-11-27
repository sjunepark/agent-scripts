# Go Doc Comments

Use when writing, revising, pruning, or auditing Go package and declaration
comments, examples, or deprecation notices. Standard godoc syntax and placement
apply; this guide records the editorial preferences on top of it.

## Scope

- Stay in Go source and example tests. README and architecture docs belong to
  their own workflows.
- For audit-only or review-only requests, leave source unchanged and return
  evidence-backed keep, prune, or add decisions.
- Treat files marked `// Code generated ... DO NOT EDIT.` as read-only; report
  the generator or source artifact that owns their documentation.
- Treat code and tests as evidence of behavior, not of unstated intent. When
  important rationale stays unresolved, record a focused maintainer question
  instead of inventing one.

## Admission test

Comments earn their place by information gain beyond the name, signature,
types, and immediate context.

- Default every candidate, existing or proposed, to remove or skip. Promote it
  only when you can name the specific non-obvious information it carries.
- Deletion test: if removing the comment loses nothing non-obvious, remove it.
- Exported status is an audit trigger, not evidence that prose is useful. When
  configured lint enforces coverage, write the shortest accurate synopsis and
  report it as a lint-driven compromise.
- Let conventional constructors, accessors, setters, wrappers, and `(value, ok)`
  lookups go undocumented unless they add semantics such as ownership,
  synchronization, validation, caching, or unusual zero-value behavior.
- Give each fact one owning comment; do not restate a type's contract on its
  constructor or methods.
- Worth stating: guarantees and invariants, zero-value behavior, ownership and
  lifecycle, concurrency, blocking and cancellation, error and panic behavior,
  side effects, package boundaries, and durable tradeoffs.

## Shape

- Keep volatile literals, defaults, member inventories, and implementation
  step sequences in code; state the stable semantic rule instead.
- Prefer one package comment covering purpose and boundaries; move it to
  `doc.go` only when the guidance is substantial.
- Use one comment for a declaration group only when its members share one
  contract.
- In a `Deprecated:` notice, name a replacement or migration constraint only
  when repository evidence supports it.
- Add an `Example` only when executable usage communicates more clearly than
  prose, and keep its output assertion meaningful.

## Finish

Run `gofmt`, the narrowest relevant `go test` for changed examples, and the
repository's configured doc or lint checks; use `go doc` to confirm changed
associations. Report material changes, intentional omissions, unresolved
rationale, lint-driven low-information comments, and validation limits.
