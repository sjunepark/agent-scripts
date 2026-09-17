# Implementation Review

Inspect the diff and nearby behavior, plus any governing issue, roadmap,
specification, ADR, or migration plan. Check working-tree state when reviewing
repository changes so fixes preserve unrelated work.

## Evidence

Trace relevant outcomes, constraints, and non-goals into the implementation.
Check correctness, lifecycle and concurrency ordering, invariants, schema/type
fit, ownership, integration contracts, errors, tests, and user-visible behavior.
Follow security, persisted-data safety, compatibility, resource bounds,
diagnostics, docs, or accessibility where the target exposes those risks.

Every material concern needs evidence, a finding, or an explicit validation
limit. Absent implementation can itself be a finding against a governing
requirement; do not invent a changed path.

## Independent review

Use the lightest delegation that improves confidence: none for a small obvious
change, a focused reviewer for shared behavior, and separate reviewers only for
distinct broad or risky areas. Keep reviewers read-only and ask for concise
findings with paths, trigger scenarios, evidence, and likely validation.
Verify their evidence before accepting a finding. Do not start autonomous
review/fix loops or concurrent edits under the review.

## Safe fixes and completion

Bucket I is narrow, local, in scope, mechanically verifiable or self-evident, and
free of unresolved product, design, architecture, rollout, compatibility, churn,
or risk judgment. Batch related fixes where safe. Preserve intentional-looking
code; only trivially dead artifacts qualify for automatic removal.

Honor review-only. Otherwise apply safe fixes, inspect their effects, and run
relevant validation. Broader refactors and ambiguous changes belong in Bucket II;
reuse an explicit decision already supplied, and ask only for unresolved choices.
A validation blocker limits what can safely be changed or claimed; report it
and continue independent review without compensating with unrelated churn.

Finish after findings are applied or proposed and relevant checks are accounted
for. Repeat checks only for changed behavior, failures, or unresolved risks.
Separate pre-existing failures from those introduced by review fixes.
