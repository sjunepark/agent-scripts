---
name: codex-plan-loop
description: "Supervise codex-plan-loop workflows from inside a Codex conversation. Use only when the user explicitly asks to run, supervise, monitor, stop, inspect, or explain codex-plan-loop/codex-plan-log runs, saved .git/codex-plan-loop artifacts, phase transcripts, worker Codex phases, review-gated plan execution, or Bucket II decisions from this workflow."
---

# Codex Plan Loop

Supervise `codex-plan-loop` as the chat-session controller. Let fresh worker
`codex exec` phases do the progress and review work, while the active Codex
session launches the wrapper, monitors it, explains it, and stops for decisions.

## Start A Run

1. Confirm the target repository and plan file.
   - Work from the target git repository root.
   - If no plan file is provided or multiple candidates are plausible, ask for the exact file.

2. Run preflight checks before asking to start:

```bash
git rev-parse --show-toplevel
git status --porcelain=v1
command -v codex-plan-loop
command -v codex-plan-log
test -f <plan-file>
codex-plan-loop <plan-file> --print-effective-config
```

3. Require explicit confirmation before starting.
   - Report the repository, plan file, clean-worktree result, effective model, effective reasoning effort, whether Fast mode will apply, and exact command.
   - State that `codex-plan-loop` auto-commits each accepted progress/review slice.
   - Use existing CLI defaults unless the user specified options: max 20 cycles, `workspace-write`, Codex configured/default network access, review limit 5, live logs, Codex configured/default model, Codex configured/default reasoning effort, and Codex configured/default service tier.
   - Do not start when the worktree is dirty unless the user explicitly instructs a different cleanup or isolation plan.

4. Start the command in a long-running terminal session so the conversation can continue:

```bash
codex-plan-loop <plan-file>
```

Use user-provided options exactly, for example:

```bash
codex-plan-loop <plan-file> --max-cycles 1 --network --model <model>
codex-plan-loop <plan-file> --no-network
codex-plan-loop <plan-file> --model <model> --reasoning-effort high
codex-plan-loop <plan-file> --fast
codex-plan-loop <plan-file> --service-tier fast
```

## Monitor A Live Run

- Poll the running session and report phase-level updates: progress phase start/finish, review phase start/finish, commits, stop statuses, Bucket II decisions, and validation failures.
- Keep chat updates compact. Do not mirror verbose transcripts unless the user asks.
- If the user asks for status, report the current phase, latest visible event, latest saved run directory if known, and whether the process is still running.
- If the user asks to stop, interrupt the running process, then inspect saved state with `codex-plan-log show latest --json`. Explain that partial runs can lack `final-summary.json`.
- Do not leave a required command session running silently at the end of a response.

## Inspect Saved Runs

Prefer the deterministic helper before opening raw files:

```bash
codex-plan-log show latest --json
codex-plan-log list
codex-plan-log path <run-id|latest>
codex-plan-log transcript latest
codex-plan-log transcript latest cycle-1-progress
codex-plan-log events latest
codex-plan-log events latest cycle-1-review-1 --json
```

Explain from the summary outward:

- run id, run directory, and final status
- stop reason or completion summary
- phase timeline: progress phases, review passes, validation, and commits
- Bucket II decisions and whether they block remaining plan work
- raw paths for follow-up inspection

Use raw artifacts only when the helper output is not enough:

- `final-summary.json`: wrapper final state and aggregated Bucket II items
- `cycle-*-progress-parsed.json`: structured output from `$progress-run`
- `cycle-*-review-*-parsed.json`: structured output from the internal review phase
- `*.transcript.log`: readable phase transcript
- `*.stderr.log`: Codex progress stream and wrapper messages
- `*.jsonl`: raw `codex exec --json` event stream

## Stop And Decision Handling

When the wrapper exits with `blocked`, `decision_needed`, `failed`, or
`validation_failed`:

1. Inspect `codex-plan-log show latest --json`.
2. Read the relevant transcript or events for the stopped phase.
3. Explain the evidence and distinguish wrapper-level status from worker phase status.
4. Ask focused questions for the needed decision or external change.
5. Do not auto-rerun, auto-recover, or implement Bucket II items unless the user explicitly asks.

## Interpretation Notes

- `codex-plan-loop` requires a clean worktree before starting and writes private run logs under `.git/codex-plan-loop/`.
- If `--model`, `--reasoning-effort`, `--service-tier`, or network flags are omitted, the wrapper lets `codex exec` use Codex configuration or Codex defaults. Use `codex-plan-loop <plan-file> --print-effective-config` to see what the wrapper can resolve before starting.
- `--network` and `--no-network` explicitly override Codex network configuration for that run.
- `--fast` is a shortcut for `--service-tier fast`. It uses Codex's documented Fast mode path, `service_tier = "fast"`.
- The wrapper uses fresh `codex exec` turns; the active chat session is the supervisor, not the worker.
- A phase can succeed while the wrapper later stops because review found blocking Bucket II, validation failed, or the review limit was reached.
- Missing `.git/codex-plan-loop/` means no wrapper runs were saved in that repository.
- Missing `final-summary.json` usually means the wrapper exited before its final reporting path or was interrupted.
