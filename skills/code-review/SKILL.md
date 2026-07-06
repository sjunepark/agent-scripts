---
name: code-review
description: "Review code or completed implementation changes for correctness, regressions, validation gaps, maintainability risks, and unearned complexity. Use when the current task is code review, review-only critique, or complexity/simplification review. Do not use merely because you are editing code."
---

# Code Review

## Workflow

1. Establish scope and edit policy.
   - If the user names files, a branch, a commit, a PR patch, or a diff, review
     that target.
   - If no target is named and the repository has uncommitted changes, default
     to the current uncommitted change set.
   - If no target is named and no review target can be inferred, ask for the
     target.
   - Honor `review-only`, `do not edit`, and equivalent instructions.
   - Distinguish issues introduced by the reviewed change from pre-existing
     debt.
   - Continue once the target and edit policy are explicit.

2. Run the implementation gate when reviewing completed work, current
   uncommitted changes, or any code review where correctness, regressions, or
   validation matter. Read [references/implementation-review.md](references/implementation-review.md).
   - Continue once material issues, obvious safe fixes, validation gaps, and
     remaining decision points have been checked.

3. Add the diet lens when the user asks about overengineering, simplification,
   deletion, trimming, wrappers, helpers, schema/config surface, compatibility
   paths, modes, flags, generic layers, or bolted-on design. Also add it when
   the review target itself introduces those surfaces. Read
   [references/diet.md](references/diet.md).
   - Do not load the diet lens for a routine small edit with no complexity
     signal unless the user asked for that lens.
   - Continue once unearned complexity candidates are either reported or
     explicitly kept as-is.

4. Report in buckets.
   - `Bucket I - Applied Safe Fixes`: fixes applied during the pass, with
     changed files and validation evidence, or `none`.
   - `Bucket II - Needs Decision`: real issues or simplifications that need
     user judgment, with the decision needed and the main tradeoff, or `none`.
   - `Keep As-Is`: meaningful concerns inspected and intentionally rejected,
     when useful.
   - `Validation`: commands run and results, or skipped/blocked reason.
   - `Residual Risk`: only when validation, scope, or context leaves a material
     gap.

## Review Rules

- Prefer no finding over a weak finding.
- Record only discrete, actionable issues with affected paths, trigger
  scenarios, and likely maintainer action.
- Apply obvious safe Bucket I fixes by default unless the edit policy forbids
  edits.
- Do not implement Bucket II without explicit user approval.
- Keep edits tightly scoped to the reviewed target.
- Do not stage, commit, or push unless explicitly asked.
