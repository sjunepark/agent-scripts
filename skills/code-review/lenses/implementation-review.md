# Implementation Review

Inspect the diff and nearby behavior, plus any governing issue, roadmap,
specification, ADR, or migration plan. Check working-tree state when reviewing
repository changes so fixes preserve unrelated work.

## Evidence

Trace relevant outcomes, constraints, and non-goals into the implementation.
Every material concern needs evidence, a finding, or an explicit validation
limit. Absent implementation can itself be a finding against a governing
requirement; do not invent a changed path.

Cover the dimensions the change actually touches: lifecycle, ordering, and
concurrency; schema and type fit; error paths; security, where a finding needs a
reachable misuse path; persisted-data safety, migration, and rollback; resource
bounds and performance, judged from measurements when claimed. For design and
complexity, look for low cohesion, dependency cycles, shallow interfaces,
missing or premature seams, pass-through wrappers, and single-path generics or
configuration.

## Independent review

Use the lightest delegation that improves confidence: none for a small obvious
change, a focused reviewer for shared behavior, and separate reviewers only for
distinct broad or risky areas. Keep reviewers read-only and ask for concise
findings with paths, trigger scenarios, evidence, and likely validation.
Verify their evidence before accepting a finding. Do not start autonomous
review/fix loops or concurrent edits under the review.

## Safe fixes and completion

Batch related Bucket I fixes where safe. Preserve intentional-looking code; only
trivially dead artifacts qualify for automatic removal. After applying fixes,
inspect their effects and run relevant validation. A validation blocker limits
what can safely be changed or claimed; report it and continue independent
review without compensating with unrelated churn.

Finish after findings are applied or proposed and relevant checks are accounted
for. Repeat checks only for changed behavior, failures, or unresolved risks.
Separate pre-existing failures from those introduced by review fixes.
