# Implementation Review Gate

Use this reference for completed implementations, current uncommitted changes,
or any code review where correctness, regressions, validation, and safe cleanup
matter.

This is prompt-level loop guidance, not a deterministic workflow controller:
do not promise persistent phase state, do not spawn iterative worker loops, and
do not stage, commit, or push unless explicitly asked.

## Workflow

1. Inspect the review target.
   - Check `git status --short` when reviewing a repository change set.
   - Inspect relevant diffs and nearby code needed to judge the change.
   - Distinguish issues introduced by the reviewed change from pre-existing
     debt.

2. Review for material issues.
   - Prefer no finding over a weak finding.
   - Record only discrete, actionable issues with concrete affected paths,
     trigger scenarios, and likely maintainer action.
   - Check correctness, state/lifecycle ordering, validation and invariants,
     type/schema fit, ownership boundaries, integration contracts, tests, and
     user-visible regressions.

3. Use subagents when useful.
   - Use the lightest review shape that materially improves confidence: often
     none, sometimes one focused reviewer, and multiple only for broad or risky
     changes with distinct review angles.
   - Good subagent tasks include fresh-context review of correctness,
     lifecycle/concurrency, integration contracts, validation gaps,
     security/privacy, performance, UI behavior, or broad reconnaissance across
     an unfamiliar subsystem.
   - Keep subagents review-only. Ask for concise findings with evidence,
     affected paths, trigger scenarios, likely validation, and Bucket I /
     Bucket II / Keep As-Is recommendations.
   - Treat subagent output as leads, not verdicts. Verify evidence yourself
     before recording findings or applying fixes.
   - Do not use subagents to create autonomous review/fix loops, edit
     concurrently, or keep reviewing after the bounded pass is complete.

4. Run the Bucket I clean loop.
   - Bucket I fixes are narrow, local, in-scope, mechanically verifiable or
     self-evident, and free of product, design, architecture, rollout,
     compatibility, churn, or risk judgment.
   - Apply current Bucket I fixes by default unless the user says `review-only`,
     `do not edit`, or equivalent.
   - After Bucket I edits, re-inspect the new diff and affected code paths. If
     new Bucket I items appear, apply them and recheck again.
   - Keep going while useful until no Bucket I items remain, only Bucket II
     decisions remain, validation/scope/context blocks safe continuation, or
     another pass would become repetitive, broad, or speculative.
   - Batch related Bucket I work that shares files, modules, ownership
     boundaries, or state domains when safe.
   - Keep edits tightly scoped to the reviewed change.
   - Do not apply removals of intentional-looking code unless the artifact is
     trivially dead, such as an unused import or broken doc link.

5. Preserve human decision boundaries.
   - Put broader refactors, ambiguous fixes, risky behavior changes, and design
     choices in Bucket II.
   - Do not implement Bucket II without explicit user approval.
   - If validation is blocked, unavailable, or failing for an external reason,
     report it and stop; do not compensate with broader static churn.

6. Validate.
   - Run the most relevant existing validation when practical.
   - Prefer targeted checks for touched areas over broad expensive suites unless
     the change risk warrants them.
   - Separate unrelated existing failures from failures introduced by review
     fixes.

## Bucket Meanings

### Bucket I - Applied Safe Fixes

Use for fixes already applied during this pass. Include the issue, changed
files, and validation evidence.

### Bucket II - Needs Decision

Use for real improvements that need user judgment before implementation.
Include the decision needed and the main tradeoff.

### Keep As-Is

Use for meaningful concerns inspected and intentionally rejected.
