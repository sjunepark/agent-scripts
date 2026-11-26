---
name: code-review
description: "Review changes or implementation against intended behavior, system constraints, design, and unnecessary complexity. Use when asked to review a diff, branch, PR, or finished implementation, or for the review pass after a substantive change."
---

# Code Review

Deliver one bounded review with evidenced, actionable findings or a defensible
clean result. No finding is preferable to a speculative one.

## Target

Use the named files, subsystem, branch, commit, PR, or diff. A roadmap,
specification, issue, ADR, or migration review covers its governed implementation,
not just the document. Establish the intended outcome, claimed milestone,
implementation range, and edit policy. Without a named target, use uncommitted
changes when present; ask only if no target can be inferred.

Read the [implementation lens](lenses/implementation-review.md) for every review.
For an initiative, milestone, or durable technical choice (dependency, runtime,
protocol, storage, deployment), also read the [system lens](lenses/system.md).

Follow affected requirements, contracts, and consumers rather than auditing the
repository, and separate introduced issues from pre-existing debt. Design and
complexity findings need a concrete correctness, maintenance, testability, or
change-locality cost and a viable simpler shape; taste, file size, hypothetical
flexibility, and useful explicitness or duplication are not findings. Repository
decisions outweigh generic patterns. Compatibility layers, aliases, fallbacks,
and dual paths need a current rollout, external contract, or migration window;
identify their consumers before proposing removal.

## Findings and fixes

Record each issue once, where its remedy belongs, with a trigger scenario,
affected path or governing requirement, consequence, evidence, and likely action.

- **Bucket I — Safe fixes:** narrow, in-scope, mechanically verifiable changes
  without unresolved product, design, rollout, compatibility, or risk judgment.
  Apply these by default unless the user requested review-only; then mark proposed.
- **Bucket II — Needs decision:** real issues with unresolved tradeoffs or
  authority, including broader refactors, public contracts, migrations, and
  durable technical choices. Explain the smallest viable alternative and the
  decision needed. Apply an already supplied decision within the authorized
  scope; ask only for what remains unresolved.

Recheck fixes and affected validation until the evidenced problems are resolved
or a concrete limitation prevents continuation. Do not turn the pass into an
autonomous review loop or repeat passing checks without new evidence. Do not
stage, commit, or push unless asked.

## Report

Lead with consequential findings and their applied/proposed status. Include
delivery coverage for initiative reviews; validation and material residual risk;
and meaningful keep-as-is decisions when useful. A clean review needs only the
result and verification limits, not empty buckets or a checklist transcript.
