---
name: plan-runner
description: "Execute the next clear slice of a plan file while keeping the same plan updated. Use when continuing from PLAN, TODO, ROADMAP, docs/plans, or handoff files; do not stage, commit, or push unless explicitly asked."
---

# Plan Runner

Use the plan file as the working progress tracker. Continue the next clear task, update the file as work happens, validate when practical, and stop when complete, blocked, or a decision is needed.

## Workflow

1. Read the plan file first.
   - If no plan file is provided or the target is ambiguous, ask for the exact file.
   - Treat the plan as authoritative unless repository evidence proves it is stale.

2. Inspect current state before editing.
   - Check git status and relevant diffs so unrelated user work is not mixed into the run.
   - Read only the files needed for the next clear plan item.

3. Execute one coherent progress slice.
   - Prefer small, safe, validated changes over broad rewrites.
   - If the next step needs a product, design, or ownership decision, ask before continuing.
   - If missing requirements block progress, ask for the specific information needed.

4. Update the same plan file as work proceeds.
   - Keep `Current state` and the next checklist accurate.
   - Mark completed items, adjust next steps, and record blockers.
   - Add a compact dated `Progress log` only after actual work occurs and only when it helps continuation.
   - Do not copy full command output into the plan.
   - Collapse older progress-log detail once `Current state` captures what future sessions need.

5. Validate.
   - Run the relevant existing validation when practical.
   - Record important validation results in the plan file, especially failures or skipped checks that affect continuation.

6. Stop at the right boundary.
   - Stop when the current slice is done, the plan is complete, progress is blocked, or a decision is required.
   - Do not stage, commit, or push unless explicitly asked.

## Final Response

Include:

- work completed
- plan file updates
- validation run and outcome
- blockers or next plan item
