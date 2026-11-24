# Structural assessment evaluation, 2026-11-24

Revision for [issue #20](https://github.com/sjunepark/agent-scripts/issues/20):
the design step now assesses reader paths, document boundaries, and net benefit
and reaches a change-or-retain decision, and the report states the structural
disposition. Baseline: `SKILL.md` at `1670f7a`
(SHA-256 `bfee5d0d…d476`). Candidates: v1 (`9b9fb558…0378`) and the adopted
v2 (`d8612b63…086a`), which adds one sentence to the reader-path criterion.

## Method

`make-fixtures.sh` builds four disposable Git repositories; cases 3 to 6 in
`evals.json` describe them. Assertions were frozen in
`assertions-2026-11-24.json` before any trial. Each run was a fresh Claude Code
subagent (Opus 5.5, default reasoning) given the user prompt and one skill copy
under a neutral directory name, with no access to the issue, assertions, or
other runs. Condition order alternated across fixture sets. Objective checks
(`check_a.py`, output in `check-a-2026-11-24.json`) cover content preservation,
AGENTS.md routing, decision-record bytes, and links. Judgment assertions were
graded once by the candidate's author, unblinded. All raw output diffs are in
`raw-diffs-2026-11-24.txt`.

## Results

| Case | Baseline | Candidate v1 | Candidate v2 |
| --- | --- | --- | --- |
| A: unstated structural problem (A1, the motivating assertion) | 0 of 3 | 0 of 1 | 2 of 2 |
| A: other assertions (A2 to A8) | pass in all 3 | pass | pass in both |
| B: healthy small repository | 1 of 1 | 1 of 1 | 1 of 1 |
| C: scoped change | 1 of 1 | 1 of 1 | not rerun |
| D: scattered reference (holdout) | 1 of 1 | 1 of 1 | not rerun |

Every baseline run of A and the v1 run linked the unsurfaced references and
fixed factual errors, but kept development, code layout, and release steps in
the user-facing README. v1 considered the structure and retained it ("one
README for users and contributors"), so a generic net-benefit check alone did
not change the decision. v2 states that a user-facing entry point carrying
maintainer procedures is friction even when accurate and small. Both v2 runs
moved those sections to a new `CONTRIBUTING.md`, repointed the AGENTS.md
release instruction, kept the decision record byte-identical, and lost no
content. The v2 rerun of B retained the topology with no edits.

Both conditions passed C (only `docs/cli.md` and the README example changed;
the overlapping operations documents were reported as a follow-up) and D
(settings consolidated under one owner, FAQ folded in). These two cases show
no headroom and serve as regression guards only.

## Limits

- Sample sizes are one to three runs per cell. The A1 difference (0 of 4
  without the v2 sentence, 2 of 2 with it) is a case-level finding, not a
  reliability rate.
- v2 was revised after seeing v1 fail case A, and case A is the case it was
  scored on. The holdout D does not exercise the audience-mix criterion, and C
  and D were not rerun against v2; its one-sentence change touches only the
  reader-path bullet.
- Case 7 (unresolved consequential choice) has a written case but was not run.
- Grading of judgment assertions was single-pass and unblinded.
- Discovery is unchanged: the description and `allow_implicit_invocation:
  false` still express explicit invocation. No trigger trials were run.
- Fixture A also contains incidental factual errors (placeholder version bump,
  missing packaging). Runs handled them differently. The assertions do not
  score those choices.
