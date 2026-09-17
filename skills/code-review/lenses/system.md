# System and Technical-Decision Lens

Determine whether the implementation delivers the intended outcome within the
system's constraints. Scope follows affected requirements and consequences,
not unrelated repository opportunities.

## Map delivery

Read governing intent and relevant architecture or operational evidence. Extract
goals, non-goals, constraints, success signals, and tradeoffs; expose conflicting
sources. Classify relevant outcomes as `delivered`, `partial`, `missing`,
`deferred`, or `unverified`, and material work outside intent as `extra`.
Use `deferred` only when authoritative scope places it beyond the claimed
milestone. Identify affected users, operators, consumers, and owners.

Inspect choices introduced, replaced, or locked in when they create durable
contracts or meaningful data, security, operational, support, or reversal costs:
languages, runtimes, dependencies, protocols, storage, services, build and
deployment mechanisms.

## Follow material consequences

- **Outcome and completeness:** trace callers, configuration, packaging,
  deployment, docs, and upgrade or removal paths needed to deliver the result.
- **Technical fit:** compare with repository conventions, platform constraints,
  operating context, and the simplest existing mechanism. Weak rationale
  documentation is different from an unsound choice.
- **Dependencies:** check compatibility, support, security, licensing, footprint,
  portability, lock-in, and exit costs when material. Follow transitive packages
  only on a concrete signal.
- **Data:** examine ownership, consistency, idempotency, migration/backfill,
  rollback, backup/restore, retention, deletion, and privacy where data changes.
- **Security:** trace trust, authorization, secrets, untrusted input, and unsafe
  defaults. Findings require a reachable misuse path.
- **Operations:** check partial failure, retries, timeouts, cancellation,
  backpressure, recovery, degradation, and useful operator diagnostics.
- **Performance and cost:** examine hot paths, bounds, scale assumptions, and
  operating costs. Use measurements or a credible measurement plan for claims.
- **Compatibility and people:** follow public APIs, formats, configuration,
  platforms, rollout/coexistence, reversibility, user workflows, accessibility,
  and maintainer burden where affected.

## Judge and report

For a criticized choice, compare retaining the existing path and the strongest
credible alternative against actual requirements, including transition,
migration, training, operating, and exit costs. Verify time-sensitive claims
with current primary sources and label inference or unavailable evidence.
Recommend replacement only for a concrete net advantage.

Trace representative success and material failure, migration, rollback, or
removal paths. State the governing intent, affected scenario, consequence,
evidence, viable alternative, and tradeoff. Durable choices belong in Bucket II
under the entry point's authority policy. Documented decisions are evidence,
not immunity; reopen them for changed assumptions or observed consequences.

For initiative reviews, report compact delivery coverage with evidence. Distinguish
ambiguous intent, absent implementation, and unavailable verification. Record
significant justified choices as keep-as-is and unexamined material boundaries as
residual risk. Do not produce a wishlist to fill the categories.
