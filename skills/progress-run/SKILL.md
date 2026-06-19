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
   - Prefer the broadest connected slice that remains bounded, reviewable, and validated.
   - Treat checklist items as planning hints, not automatic stop boundaries.
   - Execute the next coherent checklist cluster, not just the next checkbox.
   - Include adjacent plan items when they share a phase, subsystem, files, ownership boundary, validation suite, or decision context.
   - After finishing a tiny subtask, inspect the next adjacent item before stopping; continue when it is low-risk and connected.
   - Prefer safe, reviewable progress over broad rewrites.
   - If the next step appears to need a product, design, ownership, API, data-shape, compatibility, or other user decision, first do a quick evidence check in existing code, tests, docs, or product behavior unless the decision is clearly external to the repo.
   - Stop for a decision only when the answer is not already implied by that evidence.
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
   - Stop when the current coherent cluster is done, the plan is complete, progress is blocked, or a decision is required.
   - Stop before unrelated phases, broad rewrites, unclear requirements, or work that needs a user decision.
   - Do not stage, commit, or push unless explicitly asked.

## Final Response

Include:

- work completed
- plan file updates
- why you stopped
- adjacent work intentionally included or left out
- blockers or next plan item
