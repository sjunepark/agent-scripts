# Pass: effect

Applies to areas the ledger marks effect-applicable. Judge against current idioms, not memory.

## Prepare — mandatory

Run `effect-solutions list`, then `effect-solutions show <topic>` for each topic the area touches (services/layers, error handling, runtime, resources, concurrency, schema, config). The rule for writing Effect code extends to reviewing it: do not guess idioms. Cite the guide when a finding depends on one. If the CLI is unavailable, ground judgments in current Effect documentation or source — still never from memory.

## Check

- **Services and layers.** Dependencies arrive via Context/Layer, not module-level singletons or parameter drilling; layers compose at the app's runtime entry, not ad hoc per call site.
- **Runtime at the edges.** `runPromise`/`runSync` only at process boundaries (main entry, request handler, IPC handler). Sprinkled mid-stack runs that re-enter Effect are a smell; track where the area's runtime lives and whether everything funnels through it.
- **Error channel discipline.** Failures typed in the channel (tagged errors), not `unknown`/`Error` everywhere; `catchAll`/`orElse` does not silently flatten distinct failures (cross-reference the errors pass); defects vs failures used as designed.
- **Promise interop.** `Effect.tryPromise` with a typed error at true edges is fine; `await` inside logic that is otherwise Effect, or Promise/Effect ping-pong layering, is not. Flag the seam, propose a direction — not a wholesale rewrite.
- **Resources and interruption.** Files, connections, subprocesses, subscriptions via `acquireRelease`/`Scope`; long-running work is interruptible where cancellation is meaningful (IPC-cancelled operations, window teardown).
- **Concurrency made explicit.** `Effect.forEach` with a stated `concurrency`, fibers over raw `Promise.all` on effects, races/timeouts via combinators.
- **Schema and config at boundaries.** External data validated once at the boundary with one validator (not effect/Schema in one place and hand-rolled guards duplicating it elsewhere); env/config via `Config`, not `process.env` scattered through logic.

## Do not flag

- Pragmatic Promise interop at genuine edges (Electron APIs, third-party SDKs).
- Partial adoption as such — flag only leaky or confusing seams, and propose the boundary, not total conversion.
- Style variants the effect-solutions guides themselves allow.

## Severity and tier

Untyped error channel on a critical path or mid-stack runtime re-entry: major. Resource leaks: major. Idiom drift: minor. Effect rewrites invite opinions — everything here is triage.
