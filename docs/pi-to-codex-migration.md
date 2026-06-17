## Phase 6: Prompt Template Migration

Goal: convert Pi slash-style prompts into Codex skills or fold them into
existing skills.

- [x] `merge.md`
  - Target: fold into or replace with `manual-branch-integrator`.
  - Rewrite Pi positional placeholders such as `$1` and `${@:2}` into skill
    instructions that work with normal Codex prompts.
- [x] `merge-review.md`
  - Target: add an audit mode to `manual-branch-integrator` or create a focused
    `merge-auditor` skill.
- [x] `progress-check.md`
  - Target: new read-only progress-doc skill.
- [x] `progress-handoff.md`
  - Target: new handoff-plan skill.
- [x] `progress-run.md`
  - Target: new plan-runner skill that updates the plan but does not commit.
- [x] `progress-run-full.md`
  - Target: separate explicit long-run plan executor.
  - Keep WIP commit behavior opt-in and obvious.
- [x] `progress-commit.md`
  - Target: separate commit-oriented skill or leave as an explicit custom
    prompt if it is too command-like.
- [x] `bug-retro.md`
  - Target: new bug-retro skill or fold into `diet`/`structure-review` only if
    that does not blur the workflow.
- [x] `broader-refactor.md`
  - Target: new refactor-assessment skill or fold into `diet`.
- [x] Archive prompt ticket notes under docs only if they are still useful.
  - Candidate: `prompts/tickets/2026-06-16-concise-progress-handoff.md`

Implemented targets:

- `merge.md`: folded into `skills/manual-branch-integrator/SKILL.md` as default
  integration plus dry planning mode.
- `merge-review.md`: folded into `skills/manual-branch-integrator/SKILL.md` as
  merge audit mode.
- `progress-check.md`: migrated to `skills/progress-doc/`.
- `progress-handoff.md`: migrated to `skills/handoff-plan/`.
- `progress-run.md`: migrated to `skills/plan-runner/`.
- `progress-run-full.md`: migrated to `skills/plan-executor/`.
- `progress-commit.md`: migrated to `skills/progress-committer/`.
- `bug-retro.md`: migrated to `skills/bug-retro/`.
- `broader-refactor.md`: migrated to `skills/refactor-assessment/`.
- Prompt ticket note archived at
  `docs/prompt-tickets/2026-06-16-concise-progress-handoff.md`.
