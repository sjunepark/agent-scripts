---
name: plan-executor
description: "Long-run plan execution. Use when explicitly asked to execute an entire plan file end to end with progress updates; commit only when explicitly requested, and never push unless asked."
---

# Plan Executor

Run an entire plan file as far as possible in one session. This is broader than `progress-run`: continue through multiple clear tasks until the plan is complete, blocked by a real external issue, or a decision is required.

## Workflow

1. Confirm the plan target.
   - If no plan file is provided or the target is ambiguous, ask for the exact file before changing files.
   - Read the plan completely before acting.

2. Preflight repository state.
   - Inspect `git status --short` and relevant diffs.
   - Keep unrelated user changes separate from plan work.
   - If unrelated changes cannot be separated safely, ask before proceeding with edits or commits.

3. Execute through the plan.
   - Continue through the next clear tasks without stopping after a single slice.
   - Prefer small, safe, validated changes over broad rewrites.
   - Ask before product, design, ownership, API, data-shape, or compatibility decisions.
   - Stop only when the whole plan is complete, blocked, or a decision is required.

4. Maintain the plan file.
   - Update the same plan file as the authoritative tracker.
   - Keep current status, completed steps, blockers, validation, commits, and next steps accurate.
   - Keep progress history compact; summarize command output instead of pasting logs.

5. Validate after meaningful chunks.
   - Run relevant existing checks when practical.
   - Separate unrelated existing failures from failures introduced by the plan work.
   - Record important validation results in the plan file.

## Commit Mode

Checkpoint commits are allowed only when the user explicitly asks for commits or checkpoint commits.

- Before committing, inspect status and relevant diffs.
- Stage only intentional changes for the current plan chunk, including the plan-file progress update.
- Do not stage unrelated user changes.
- Use clear normal project commit messages. `WIP:` is acceptable for incomplete intermediate checkpoints.
- Use a non-WIP final commit message when the committed chunk is complete.
- Do not push unless explicitly asked.

## Final Response

Before stopping, make the plan file accurately state whether the plan is complete, what validation ran, which commits were created, and any remaining blockers or follow-up work. Then report the same points briefly to the user.
