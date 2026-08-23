# Delegate candidate evaluation — 2026-08-23

## Intended behavior

When explicitly invoked for substantive implementation, the parent prepares a
self-contained contract and delegates detailed edits to GPT-5.6 Luna at maximum
reasoning with no inherited conversation. The parent remains accountable for
review and validation, iterates with the same worker when material changes are
needed, and may directly make obvious localized fixes when another worker turn
would be less efficient.

The exact model pin is deliberate user policy. It makes this runtime contract
Codex-specific, so the registry recommends it only for the `.agents` target.

## Fresh-context behavior comparison

Two fresh GPT-5.6 Sol/high orchestrators received identical isolated copies of
a small JavaScript cache task with one failing integration test.

- The baseline received no candidate instructions. It edited `src/cache.js`
  directly and passed the supplied test plus `git diff --check`.
- The candidate read `delegate/SKILL.md`, then spawned one nested implementer
  with `model: gpt-5.6-luna`, `reasoning_effort: max`, and
  `fork_turns: none`. The parent made no concurrent edits, waited for the
  worker, independently reviewed its diff, and reran validation.
- The candidate artifact passed the supplied test and additional runtime
  assertions for rejected-load retry, key isolation, and invalidated in-flight
  loads. Only the intended source file changed and no commit was created.

This comparison establishes the discriminating behavior: the candidate moved
detailed implementation to the pinned worker while retaining parent review and
validation. It does not establish comparative performance or cost from one
small fixture.

## Trigger evaluation

Two independent fresh GPT-5.6 Sol/high classifiers received only the catalog
description and `trigger-cases.json`, not the runtime instructions. Both
classified all seven cases as intended:

- 3/3 explicit positive invocations activated;
- 2/2 uninvoked but otherwise in-scope implementation requests did not
  activate; and
- 2/2 adjacent workflows—analysis-only fan-out and a user-owned Codex
  task—did not activate.

## Static validation

- `scripts/validate-skills`: passed for 34 skills and the registry.
- `bunx skills add ./skills/delegate --list`: discovered exactly `delegate`.
- Registry and reconciler Node suite: 55/55 passed.
- `git diff --check`: passed.

No installation, commit, publication, or machine reconciliation was performed.
