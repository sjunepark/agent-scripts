---
name: progress-run
description: "Execute the next substantial coherent slice from a plan or handoff file while keeping progress updated in that same file. Use when continuing from PLAN, progress handoff, TODO, ROADMAP, docs/plans, handoff files, or .pi plan files; do not stage, commit, or push unless explicitly asked."
---

# Progress Run

Use the provided plan or handoff file as the working progress tracker. Read it first, continue the next substantial coherent slice, update it as work happens, validate when practical, and stop when the plan is complete, blocked, or a decision is needed.

## Workflow

1. Read the plan file first.
   - If no plan file is provided or the target is ambiguous, ask for the exact file.
   - Treat the plan as authoritative unless repository evidence proves it is stale.

2. Inspect current state before editing.
   - Check git status and relevant diffs so unrelated user work is not mixed into the run.
   - Read only the files needed for the next clear plan item.

3. Execute one substantial coherent progress slice.
   - Complete the smallest useful milestone that leaves the repo in a coherent, validated state.
   - Include tightly coupled adjacent checklist items when they share files, ownership boundaries, or validation.
   - Do not stop after a tiny subtask if the next connected work is clear and low-risk.
   - Prefer safe, reviewable progress over broad rewrites.
   - If the next step needs a product, design, ownership, API, data-shape, compatibility, or other user decision, ask before continuing.
   - When missing context or requirements block progress, identify the specific information needed and ask focused questions.

4. Use subagents for focused context gathering when useful.
   - Use none, one, or a few focused subagents only when they materially reduce uncertainty or search cost.
   - Good uses include broad repo reconnaissance, unfamiliar subsystem mapping, dependency tracing, test discovery, and identifying likely integration risks before edits.
   - Ask for concise findings with evidence, affected paths, relevant commands, and next-step recommendations.
   - Treat subagent output as leads; verify evidence yourself before acting on it.
   - Keep subagents out of concurrent edits and autonomous implementation loops.

5. Update the same plan file as work proceeds.
   - Keep `Current state` and the next checklist accurate.
   - Mark completed items, adjust next steps, and record blockers.
   - Add a compact dated `Progress log` only after actual work occurs and only when it helps continuation.
   - Each progress entry should include the completed slice, meaningful validation result, blocker if any, and next slice.
   - Do not copy full command output into the plan.
   - Collapse or remove older progress-log detail once `Current state` captures what future sessions need.

6. Validate.
   - Run the relevant existing validation when practical.
   - Record important validation results in the plan file, especially failures or skipped checks that affect continuation.

7. Stop at the right boundary.
   - Stop when the current slice is done, the plan is complete, progress is blocked, or a decision is required.
   - Stop before unrelated phases, broad rewrites, unclear requirements, or work that needs a user decision.
   - Do not stage, commit, or push unless explicitly asked.

## Final Response

Include:

- work completed
- plan file updates
- blockers or next plan item
