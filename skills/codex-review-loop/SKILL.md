---
name: codex-review-loop
description: "Supervise codex-review-loop workflows from inside a Codex conversation. Use only when the user explicitly asks to run, supervise, monitor, stop, inspect, or explain codex-review-loop/codex-review-log runs, saved .git/codex-review-loop artifacts, review/fix phase transcripts, Bucket I safe fixes, Bucket II decisions, or validation outcomes from this workflow."
---

# Codex Review Loop

Supervise `codex-review-loop` as the chat-session controller. Let fresh worker
`codex exec` phases do review and safe-fix work, while the active Codex session
launches the wrapper, monitors it, explains it, and stops for decisions.

## Start A Run

1. Confirm the target repository and review scope.
   - Work from the target git repository root.
   - Default scope is `uncommitted changes`.
   - If the user names a branch, commit range, feature, or file list, pass that as `--scope`.

2. Run preflight checks before asking to start:

```bash
git rev-parse --show-toplevel
git status --porcelain=v1
command -v codex-review-loop
command -v codex-review-log
codex-review-loop --scope "<scope>" --print-effective-config
```

3. Require explicit confirmation before starting.
   - Report the repository, scope, current worktree status, effective model, effective reasoning effort, whether Fast mode will apply, and exact command.
   - State that `codex-review-loop` may edit files by applying safe Bucket I fixes, but never stages, commits, or pushes.
   - Use existing CLI defaults unless the user specified options: max 5 iterations, `workspace-write`, no network, live logs, Codex configured/default model, Codex configured/default reasoning effort, and Codex configured/default service tier.
   - For read-only review, include `--review-only`.

4. Start the command in a long-running terminal session so the conversation can continue:

```bash
codex-review-loop --scope "uncommitted changes"
```

Use user-provided options exactly, for example:

```bash
codex-review-loop --scope "HEAD~1..HEAD" --review-only --max-iterations 1
codex-review-loop --scope "current uncommitted changes" --network --model <model>
codex-review-loop --scope "current uncommitted changes" --model <model> --reasoning-effort high
codex-review-loop --scope "current uncommitted changes" --fast
codex-review-loop --scope "current uncommitted changes" --service-tier fast
```

## Monitor A Live Run

- Poll the running session and report phase-level updates: review phase start/finish, fix phase start/finish, applied fixes, stop statuses, Bucket II decisions, and validation failures.
- Keep chat updates compact. Do not mirror verbose transcripts unless the user asks.
- If the user asks for status, report the current phase, latest visible event, latest saved run directory if known, and whether the process is still running.
- If the user asks to stop, interrupt the running process, then inspect saved state with `codex-review-log show latest --json`. Explain that partial runs can lack `final-summary.json`.
- Do not leave a required command session running silently at the end of a response.

## Inspect Saved Runs

Prefer the deterministic helper before opening raw files:

```bash
codex-review-log show latest --json
codex-review-log list
codex-review-log path <run-id|latest>
codex-review-log transcript latest
codex-review-log transcript latest iteration-1-review
codex-review-log events latest
codex-review-log events latest iteration-1-fix --json
```

Explain from the summary outward:

- run id, run directory, final status, and scope
- applied Bucket I fix count and validation result
- phase timeline: review and fix phases
- Bucket II decisions and whether they block completion
- raw paths for follow-up inspection

Use raw artifacts only when the helper output is not enough:

- `final-summary.json`: wrapper final state and aggregated Bucket II items
- `iteration-*-review-parsed.json`: structured review output
- `iteration-*-fix-parsed.json`: structured safe-fix output
- `*.transcript.log`: readable phase transcript
- `*.stderr.log`: Codex progress stream and wrapper messages
- `*.jsonl`: raw `codex exec --json` event stream

## Stop And Decision Handling

When the wrapper exits with `decision_needed`, `blocked`, `failed`,
`validation_failed`, or `max_iterations_reached`:

1. Inspect `codex-review-log show latest --json`.
2. Read the relevant transcript or events for the stopped phase.
3. Explain the evidence and distinguish wrapper-level status from worker phase status.
4. Ask focused questions for the needed decision or external change.
5. Do not auto-rerun, auto-recover, or implement Bucket II items unless the user explicitly asks.

## Interpretation Notes

- `codex-review-loop` allows an initially dirty worktree because the default scope is uncommitted changes.
- The wrapper writes private run logs under `.git/codex-review-loop/`.
- If `--model`, `--reasoning-effort`, or `--service-tier` is omitted, the wrapper lets `codex exec` use Codex configuration or Codex defaults. Use `codex-review-loop --print-effective-config` to see what the wrapper can resolve before starting.
- `--fast` is a shortcut for `--service-tier fast`. It uses Codex's documented Fast mode path, `service_tier = "fast"`.
- The wrapper uses fresh `codex exec` turns; the active chat session is the supervisor, not the worker.
- The wrapper applies only Bucket I safe fixes by default and never stages, commits, or pushes.
- A phase can succeed while the wrapper later stops because Bucket II needs a decision, validation failed, or the iteration limit was reached.
- Missing `.git/codex-review-loop/` means no wrapper runs were saved in that repository.
- Missing `final-summary.json` usually means the wrapper exited before its final reporting path or was interrupted.
