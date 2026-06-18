---
name: codex-plan-logs
description: "Inspect and explain codex-plan-loop run logs. Use when the user asks what happened inside a codex-plan-loop run, why it stopped, what progress/review phases did, what Bucket II items were found, where saved codex exec JSONL logs are, or wants to interactively investigate .git/codex-plan-loop artifacts."
---

# Codex Plan Logs

Inspect saved `codex-plan-loop` automation runs. Prefer the deterministic
`codex-plan-log` helper first, then read raw artifacts only for deeper follow-up
questions.

## Workflow

1. Confirm the current directory is the target git repository.
   - If the user names another repository, work there.
   - If the repository is ambiguous, ask for the path.

2. Discover runs with the helper.
   - Start with `codex-plan-log show latest --json` for the latest run.
   - Use `codex-plan-log list` when the user asks for older runs or when the latest run is not the intended target.
   - Use `codex-plan-log path <run-id|latest>` before opening raw files.

3. Explain from the summary outward.
   - Report the run status, stop reason, plan summary, and Bucket II count first.
   - Then summarize each phase: progress phases, review passes, validation, and applied fixes.
   - Distinguish wrapper-level status from the inner Codex phase statuses.

4. Inspect raw artifacts only when needed.
   - `final-summary.json`: wrapper final state and aggregated Bucket II items.
   - `cycle-*-progress-parsed.json`: structured output from `$progress-run`.
   - `cycle-*-review-*-parsed.json`: structured output from `$post-review-loop`.
   - `*.transcript.log`: readable live-style transcript rendered from Codex JSONL events.
   - `*.stderr.log`: Codex progress stream and command/status messages.
   - `*.jsonl`: raw `codex exec --json` event stream.

5. For event-level questions, use:

```bash
codex-plan-log transcript latest
codex-plan-log transcript latest cycle-1-progress
codex-plan-log events latest
codex-plan-log events latest cycle-1-progress
codex-plan-log events latest cycle-1-review-1 --json
```

6. When answering, cite file paths or phase names that support the conclusion.
   - For example: `cycle-2-review-1-parsed.json` found one blocking Bucket II item.
   - Keep raw JSONL excerpts short; summarize command/event sequences instead of pasting logs.

## Interpretation Notes

- Missing `.git/codex-plan-loop/` means no wrapper runs were saved in that repo.
- Missing `final-summary.json` usually means the wrapper exited before its final reporting path.
- A phase can succeed while the wrapper stops later because review found blocking Bucket II, validation failed, or the review limit was reached.
- `codex-plan-loop` uses fresh `codex exec` turns; `*.transcript.log` is a readable automation transcript, while `*.jsonl` is the raw event stream.
- The helper is read-only. Do not edit or delete saved run directories unless the user explicitly asks for cleanup.

## Output Shape

For broad "what happened?" questions, answer with:

- run id and final status
- stop reason or completion summary
- phase timeline
- validation result
- Bucket II decisions, if any
- raw log paths for follow-up inspection
