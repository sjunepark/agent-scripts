---
name: progress
description: "Manage repo-local roadmaps, ordered plans, unordered tasks, continuation, handoffs, and status reviews. Use only when the user explicitly invokes $progress."
---

# Progress

Run this skill only after the user explicitly invokes `$progress`. Treat the
planning system as manually selected project memory: reconcile it with direct
instructions and repository evidence before relying on it.

## Resolve the Work System

1. Read applicable `AGENTS.md` files and inspect git status.
2. Use a user-named central index when it matches the work-queue model.
   Otherwise prefer root `ROADMAP.md`. Use root `PLAN.md` only when it clearly
   routes project work. When only standalone legacy progress documents exist,
   report that they need the organize branch's explicit migration before
   another workflow can use them.
3. Read the index and every work-item file needed by the selected branch.
4. Inspect recent history, code, tests, or diffs only far enough to verify
   material status claims. Treat conflicts as staleness to report or repair,
   not as restrictions imposed by the documents.

Resolution is complete when the current item, the next scheduled item, and any
material conflict between documentation and repository state are known.

## Keep One Work Queue

- Keep the root index compact: a `Current` slot, an ordered `Plans` queue, and
  an unordered `Tasks` pool.
- Store scheduled work in `plans/<stable-slug>.md` and unscheduled work in
  `tasks/<stable-slug>.md`. Scheduling, not size or dependency, distinguishes
  them.
- Persist plan order only in the root index. Keep sequence numbers out of
  filenames and item files.
- List every active item exactly once: current, queued, or pooled. An item in
  `Current` is absent from its former list until it stops being current.
- Make each item file the source of truth for its `Outcome`, `Current state`,
  and `Next action`. Add scope, checklist, decisions, or validation only when
  they carry continuation value.
- On completion, record the result in the item, remove it from the active
  index, and retain the unlisted file as concise history.

## Select a Workflow

Read exactly one primary workflow:

- [Organize work](workflows/organize.md) — initialize the system; capture,
  start, reorder, schedule, unschedule, or complete items.
- [Orient to work](workflows/orient.md) — report current, next, blocked, or
  stale work without modifying the repository.
- [Continue work](workflows/continue.md) — select and implement one substantial
  coherent slice while keeping its item current.
- [Hand off work](workflows/handoff.md) — compress the present session into a
  durable item and next action without implementing.

When the user is actually coordinating multiple git worktrees, also read
[Parallel worktrees](workflows/parallel-worktrees.md). Keep that exception out
of ordinary runs.

## Finish Coherently

Before ending a mutating workflow, verify that every indexed path exists, each
active item appears exactly once, plan numbering is continuous, and every
touched item states an evidence-backed current state and concrete next action.
Keep transcripts and routine session logs out of the files. Report changed
planning files, validation performed, and the next intended work or blocker.
