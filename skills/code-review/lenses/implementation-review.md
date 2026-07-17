# Implementation Review Gate

## Workflow

1. Inspect the review target.
   - Check `git status --short` when reviewing a repository change set.
   - Inspect relevant diffs and nearby code needed to judge the change.

2. Review for material issues.
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
   - After Bucket I edits, re-inspect the new diff and affected code paths. If
     new Bucket I items appear, apply them and recheck again.
   - Keep going while useful until no Bucket I items remain, only Bucket II
     decisions remain, validation/scope/context blocks safe continuation, or
     another pass would become repetitive, broad, or speculative.
   - Batch related Bucket I work that shares files, modules, ownership
     boundaries, or state domains when safe.
   - Do not apply removals of intentional-looking code unless the artifact is
     trivially dead, such as an unused import or broken doc link.

5. Preserve human decision boundaries.
   - Put broader refactors, ambiguous fixes, risky behavior changes, and design
     choices in Bucket II.
   - If validation is blocked, unavailable, or failing for an external reason,
     report it and stop; do not compensate with broader static churn.

6. Validate.
   - Run the most relevant existing validation when practical.
   - Prefer targeted checks for touched areas over broad expensive suites unless
     the change risk warrants them.
   - Separate unrelated existing failures from failures introduced by review
     fixes.
