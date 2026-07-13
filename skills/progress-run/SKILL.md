---
name: progress-run
description: "Execute the next substantial coherent slice from a plan or handoff file while keeping progress updated in that same file. Use when continuing from PLAN, TODO, ROADMAP, docs/plans, handoff files, or .pi plan files."
---

# Progress Run

Use the provided plan or handoff file as the working progress tracker.

## Workflow

1. Resolve and read the plan file first.
   - Use the specified file. Otherwise inspect repository conventions, use the sole plausible plan, and ask only when multiple candidates remain or no target can be inferred safely.
   - Treat the plan as authoritative unless repository evidence proves it is stale.

2. Inspect current state before editing.
   - Check git status and relevant diffs, then inspect only the files needed for the next slice.
   - Use focused subagents only for reconnaissance that materially reduces uncertainty; verify their evidence and keep implementation in one agent.

3. Execute one substantial coherent progress slice.
   - Prefer the broadest connected slice that remains bounded, reviewable, and validated.
   - Treat checklist items as planning hints. Include adjacent items when a shared phase, subsystem, files, ownership boundary, validation suite, or decision context makes them one slice.
   - After finishing a tiny subtask, inspect the next adjacent item before stopping; continue when it is low-risk and connected.
   - Before asking for a product, design, API, data-shape, compatibility, or other user decision, check existing code, tests, docs, or product behavior; ask only when the answer is external or not implied by that evidence.

4. Update the same plan file as work proceeds.
   - Keep `Current state` and the checklist accurate: mark completed items, adjust next steps, and record blockers.
   - Add a compact dated `Progress log` only after actual work and when it helps continuation; include the completed slice, meaningful validation, blocker if any, and next slice.
   - Keep `Current state` as the summary: collapse older log detail and omit full command output.

5. Validate.
   - Run the repository's relevant existing validation and record meaningful results in the plan; record skipped checks and why.

6. Stop at the right boundary.
   - Stop when the current slice is done, the plan is complete, progress is blocked, or a decision is required; do not start unrelated phases or broad rewrites.
   - Do not stage, commit, or push unless explicitly asked.

## Final Response

Report the completed slice, plan updates, validation, stopping reason, adjacent work intentionally included or left out, and the next item or blocker.
