# Active-review waiting evaluation, 2026-11-24

Revision for [issue #27](https://github.com/sjunepark/agent-scripts/issues/27).

## Changes

- The active-review gate defines waiting as staying in the task: re-check
  about every 60 to 90 seconds for up to about 20 minutes, then report intake
  as blocked. It replaces "a reasonable wait window".
- Review polling is authorized by the gate even where a host discourages
  polling CI; CI results stay out unless the task needs green checks.
- CodeRabbit's pending or concluded check counts as a state signal. An explicit
  failure ends that service's wait instead of blocking intake.
- `collect_pr_feedback.py` decodes `gh` output as UTF-8. On a cp949 console it
  previously raised `UnicodeDecodeError` in the reader thread, leaving stdout
  `None` and failing with `'NoneType' object has no attribute 'strip'`.
- The collector command notes `python` for hosts without `python3`.
- After review, the gate gained a blocked fallback when no bounded wait is
  permitted, chained shorter waits, an explicit no-retrigger after a failure,
  and a note to read only the CodeRabbit check. The authority wording became a
  distinction ("this gate's wait, not CI polling"). These post-trial edits are
  not covered by the trials below.

## Encoding regression

`test_collector_encoding.py` runs the real `gh_json` decoding path against a
stand-in `gh` that writes UTF-8 bytes cp949 cannot decode. On this Windows host
(locale encoding cp949), `python -X utf8=0 evals/test_collector_encoding.py`
failed before the fix with the exact `AttributeError` from the issue and passes
after it. The stand-in also asserts that `encoding="utf-8"` is passed, so a
revert fails on UTF-8 locales and Linux CI too, where decoding alone would pass. A live run against a GitHub PR was
not possible because the network was unavailable during this session.

## Waiting trial

`packet-active-review.txt` reproduces the reported situation: CodeRabbit check
pending, a Codex "something went wrong" summary, a host rule forbidding CI
polling, and a PR monitor that will not wake the agent. Assertions W1 to W6 in
`evals.json` were frozen before trials. Each run was a fresh Claude Code
subagent (Opus 5.5) reading only the packet and one skill copy under a neutral
directory name. Baseline is `SKILL.md` at `origin/main` (SHA-256
`0658f41c…4f00`); candidate SHA-256 `ba52b120…7657`.

| Assertion | Baseline run 1 | Baseline run 2 | Candidate run 1 | Candidate run 2 |
| --- | --- | --- | --- | --- |
| W1 stays in task | fail | fail | pass | pass |
| W2 bounded loop | fail | fail | pass | pass |
| W3 review polling is not CI polling | fail | fail | pass | pass |
| W4 Codex failure ends its wait, no retrigger | partial | partial | pass | pass |
| W5 no edits before CodeRabbit completes | pass | pass | pass | pass |
| W6 blocked at bound | n/a | n/a | pass | pass |

Both baseline runs ended the turn, reasoning that the host's no-polling rule
covered the CodeRabbit check, and asked the user to re-invoke: the failure
from the issue. Both treated the Codex failure as a blocker needing a
user decision instead of continuing with existing feedback, though neither
retriggered Codex. Both candidate runs planned a bounded loop over review
surfaces only (comments, reviews, reactions, head SHA), recorded the Codex
failure as a non-blocking report item, and named the blocked report at about
20 minutes.

Neither condition resolves whether polling CI for a later merge step is
allowed. Both candidate runs read checks once and would end the turn if they
were still pending, so a merge-and-release chain can still pause on CI.
Merging is outside this skill's authority, and the issue scopes the fix to
review waiting, so that is left alone.

## Limits

Two runs per condition, graded once by the revision's author without blinding.
These are decision simulations: no agent actually waited on a live review, and
the host's real tool restrictions were described rather than enforced. Full
transcripts were not retained; the table and summaries above are the record.
Discovery is unchanged and was not retested.
