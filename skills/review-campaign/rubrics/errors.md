# Errors Pass Rubric

Expected failures are modeled and routed; invariant violations crash loudly; nothing disappears. Language-aware: Effect error channel / typed results in TS, wrapped `error` returns in Go.

## Check

- **Modeled vs stringly.** Expected failures live in types — Effect error channel members, tagged result types, Go sentinel/typed errors — not thrown strings or generic `Error` consumers must parse. The caller should be able to handle cases by type.
- **Nothing swallowed.** Empty `catch`, `.catch(() => {})`, `catchAll` that maps everything to a default, Go `_ =` on errors, logging-then-continuing where the operation actually failed. Every absorbed error needs a reason it is safe to absorb.
- **Broad catches narrowed.** A catch-all at a boundary re-throws or maps with narrowing and added context; it does not flatten distinct failures into one bucket mid-stack.
- **Defect vs expected failure.** Invariant violations should fail fast (Effect `die`/defect, Go `panic` only for true invariants), not be dressed up as handleable errors; expected failures should not crash the process. Flag each direction of confusion.
- **Context accretes at boundaries.** Wrapping adds the operation and ids (`fmt.Errorf("activate workspace %s: %w", ...)` style; Effect `mapError`/`annotate` equivalents) so the top-level error tells the story without a debugger.
- **Edges are deliberate.** Retries/timeouts state which errors are retryable; user-facing messages are actionable while internal detail goes to logs; process-level handlers exist and log (unhandledRejection, desktop crash/render-gone handling, service panic recovery middleware).

## Cross-references

Failure paths untested → tests pass; failures undiagnosable in production → logging pass. Put the finding where the fix is; cross-reference in one line at most.

## Do not flag

- Crashes on genuine invariants — that is the design.
- Missing handling for states the types already make unrepresentable.
- Pragmatic error flattening at true edges (a CLI exit, a toast) where cases genuinely converge.

## Severity and tier

Swallowed error on a critical path: blocker. Flattened error types forcing string matching: major. Missing wrap context: minor unless it guards a hot incident path. Error-model rewrites are design decisions — always triage; auto tier only for adding context to an existing wrap or removing a provably unreachable catch.
