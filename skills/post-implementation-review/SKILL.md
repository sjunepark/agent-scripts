---
name: post-implementation-review
description: "Run a bounded post-implementation review of completed code or document changes, applying only obvious safe fixes by default. Use after an implementation, refactor, plan slice, bug fix, or uncommitted change set needs review for regressions, missed invariants, validation gaps, and small safe corrections without autonomous review/fix loops."
---

# Post-Implementation Review

Review the completed implementation once, apply only obvious safe fixes by
default, and surface everything else for user judgment. This skill replaces
autonomous review/fix loops: do not spawn iterative workers, do not keep
reviewing until clean, and do not stage, commit, or push unless explicitly
asked.

## Workflow

1. Establish scope.
   - Default to current uncommitted changes when the user does not name a scope.
   - Inspect `git status --short`, relevant diffs, and nearby code needed to judge the change.
   - Distinguish issues introduced by the reviewed change from pre-existing debt.

2. Review for material issues.
   - Prefer no finding over a weak finding.
   - Record only discrete, actionable issues with concrete affected paths, trigger scenarios, and likely maintainer action.
   - Check correctness, state/lifecycle ordering, validation and invariants, type/schema fit, ownership boundaries, integration contracts, tests, and user-visible regressions.

3. Apply obvious safe fixes by default.
   - Apply only Bucket I fixes: narrow, local, in-scope, mechanically verifiable or self-evident, and free of product, design, architecture, rollout, compatibility, churn, or risk judgment.
   - Keep edits tightly scoped to the reviewed change.
   - Do not apply removals of intentional-looking code unless the artifact is trivially dead, such as an unused import or broken doc link.
   - If the user says `review-only`, `do not edit`, or equivalent, report findings without changing files.

4. Stop instead of looping.
   - Run one review/fix pass only.
   - If validation is blocked, unavailable, or failing for an external reason, report it and stop; do not compensate with broader static churn.
   - Put broader refactors, ambiguous fixes, risky behavior changes, and design choices in Bucket II.

5. Validate.
   - Run the most relevant existing validation when practical.
   - Prefer targeted checks for touched areas over broad expensive suites unless the change risk warrants them.
   - Separate unrelated existing failures from failures introduced by the review fixes.

## Buckets

### Bucket I - Applied Safe Fixes

Use for fixes already applied during this pass. Include the issue, changed files,
and validation evidence.

### Bucket II - Needs Decision

Use for real improvements that need user judgment before implementation. Include
the decision needed and the main tradeoff.

### Keep As-Is

Use for meaningful concerns inspected and intentionally rejected.

## Final Response

Report:

- Bucket I fixes applied, or `none`
- Bucket II decisions, or `none`
- Keep-as-is notes when useful
- validation run and result
- residual risk if validation was skipped or blocked
