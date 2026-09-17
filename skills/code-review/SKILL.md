---
name: code-review
description: "Review changes or implementation against intended behavior, system constraints, design, and unnecessary complexity."
---

# Code Review

Deliver one bounded review with evidenced, actionable findings or a defensible
clean result. Apply the implementation, system, design, and diet perspectives
proportionately; no finding is preferable to a speculative one.

## Scope and guidance

Use the named files, subsystem, branch, commit, PR, or diff. A roadmap,
specification, issue, ADR, or migration review covers its governed implementation,
not just the document. Establish the intended outcome, claimed milestone,
implementation range, and edit policy. Without a named target, use uncommitted
changes when present; ask only if no target can be inferred.

Read the [implementation lens](lenses/implementation-review.md) for the core
review and safe-fix policy. Load the other detailed lenses when their dimensions
are material to the target:

- [System](lenses/system.md): initiative delivery, technical choices, or
  cross-cutting consequences. Check that the requested outcome is actually delivered.
- [Design](lenses/design.md): changed responsibilities, interfaces, dependencies,
  or consumers. Check that contracts and invariants have clear owners.
- [Diet](lenses/diet.md): added concepts, states, wrappers, compatibility paths,
  or other complexity. Check that the machinery earns its maintenance cost.

For a routine local edit, these can be brief checks without loading every lens
or expanding into a repository-wide audit. Follow affected requirements and
consumer paths; separate introduced issues from pre-existing debt.

## Findings and fixes

Record each issue once, where its remedy belongs, with a trigger scenario,
affected path or governing requirement, consequence, evidence, and likely action.
Architectural taste and hypothetical flexibility are not findings.

- **Bucket I — Safe fixes:** narrow, in-scope, mechanically verifiable changes
  without unresolved product, design, rollout, compatibility, or risk judgment.
  Apply these by default unless the user requested review-only; then mark proposed.
- **Bucket II — Needs decision:** real issues with unresolved tradeoffs or
  authority. Explain the smallest viable alternative and decision needed. Apply
  an already supplied decision within the authorized scope; ask only for what
  remains unresolved.

Recheck fixes and affected validation until the evidenced problems are resolved
or a concrete limitation prevents continuation. Do not turn the pass into an
autonomous review loop or repeat passing checks without new evidence. Do not
stage, commit, or push unless asked.

## Report

Lead with consequential findings and their applied/proposed status. Include
delivery coverage for initiative reviews; validation and material residual risk;
and meaningful keep-as-is decisions when useful. A clean review needs only the
result and verification limits, not empty buckets or a checklist transcript.
