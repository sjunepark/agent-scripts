# Structural assessment evaluation, 2026-11-24

Revision for [issue #20](https://github.com/sjunepark/agent-scripts/issues/20).
The design step now assesses reader paths, document boundaries, and net
benefit and reaches a change-or-retain decision. The report states the
structural disposition.

## Versions

Identified by Git blob ID of the LF-normalized `SKILL.md`:

| Label | Blob | Change |
| --- | --- | --- |
| baseline | `5e4380d5` | `1670f7a` |
| v1 | `dfdcd315` | three-part structural assessment and report disposition |
| v2 | `712b6488` | adds "a user-facing entry point that also carries maintainer procedures is friction even when … small" |
| v3 (adopted) | `a46a7c90` | after independent review: softened to "maintainer procedures beyond a short pointer"; the approval clause folded into "Apply the coherent result directly" |

## Method

`make-fixtures.sh` builds five disposable Git repositories; cases 3 to 6 and 8
in `evals.json` describe them. Assertions for A to D were frozen in
`assertions-2026-11-24.json` before round 1, and E in
`assertions-round3-2026-11-24.json` before round 3. Each run was a fresh Claude
Code subagent (Opus 5.5, default reasoning) given the user prompt and one skill
copy under a neutral directory name, with no access to the issue, assertions,
or other runs. Objective checks (`check_a.py`, output in
`check-a-2026-11-24.json`) cover content preservation, AGENTS.md routing,
decision-record bytes, and links. Judgment assertions were graded once by the
candidate's author, unblinded. All raw output diffs are in
`raw-diffs-2026-11-24.txt`.

Run assignments. The round 1 assignment is frozen in the assertions file.
Rounds 2 and 3 are recorded here from the launch prompts.

| Round | Run | Version | A1 / structural outcome |
| --- | --- | --- | --- |
| 1 | set1/A | baseline | kept mixed README: fail |
| 1 | set2/A | v1 | kept mixed README after considering it: fail |
| 1 | set1/B, set2/B | v1, baseline | retained: pass |
| 1 | set1/C, set2/C | baseline, v1 | scoped to CLI docs: pass |
| 1 | set1/D, set2/D | v1, baseline | consolidated settings: pass |
| 2 | set5/A | baseline | kept mixed README: fail |
| 2 | set3/A, set4/A | v2 | moved maintainer sections to CONTRIBUTING.md: pass |
| 2 | set3/B | v2 | retained: pass |
| 3 | set6/A | v3 | moved maintainer sections to CONTRIBUTING.md: pass |
| 3 | set6/E, set7/E | v3 | kept the two-sentence Contributing note: pass |

## Results

| Case | Baseline | v1 | v2 | v3 |
| --- | --- | --- | --- | --- |
| A1: mixed-audience README restructured | 0 of 2 | 0 of 1 | 2 of 2 | 1 of 1 |
| A2 to A8 | pass in both | pass, A3 by judgment (see below) | pass in both | pass |
| B: healthy repository retained | 1 of 1 | 1 of 1 | 1 of 1 | not rerun |
| C: scoped change | 1 of 1 | 1 of 1 | not rerun | not rerun |
| D: scattered reference (holdout) | 1 of 1 | 1 of 1 | not rerun | not rerun |
| E: brief contributor note retained | not run | not run | not run | 2 of 2 |

Every baseline run of A and the v1 run linked the unsurfaced references and
fixed factual errors, but kept development, code layout, and release steps in
the user-facing README. v1 considered the structure and retained it ("one
README for users and contributors"), so a generic net-benefit check alone did
not change the decision. Each v2 and v3 run moved those sections to a new
`CONTRIBUTING.md`, repointed the AGENTS.md release instruction, kept the
decision record byte-identical, and lost no checked content. Both v3 runs of E
kept a two-sentence Contributing section in place, which shows the softened
sentence does not force a split of brief contributor notes.

A3 override: `check_a.py` flags set2/A (v1) for a missing `pip install`. That
run deliberately replaced the unbacked `pip install tally-words` instructions
with checkout instructions and moved packaging to the roadmap, because the
fixture has no packaging metadata. The assertion targets loss of useful
content, so the author graded it a pass.

Both conditions passed C (only `docs/cli.md` and the README example changed;
the overlapping operations documents were reported as a follow-up) and D
(settings consolidated under one owner, FAQ folded in). These cases show no
headroom and serve as regression guards only.

## Limits

- One to two runs per cell. The A1 difference (0 of 3 without the
  audience sentence, 3 of 3 with it) is a case-level finding, not a
  reliability rate.
- v2 was written after seeing v1 fail case A and was scored on that case. E
  was added after review to test the over-reach risk; the holdout D does not
  exercise the audience-mix criterion. C and D were not rerun on v2 or v3,
  whose changes touch only the reader-path bullet and wording.
- Case 7 (unresolved consequential choice) has a written case but was not run.
- Judgment grading was single-pass and unblinded.
- Discovery is unchanged: the description and `allow_implicit_invocation:
  false` still express explicit invocation. No trigger trials were run.
- Fixture A also contains incidental factual errors (placeholder version bump,
  missing packaging). Runs handled them differently, and the assertions do not
  score those choices.
