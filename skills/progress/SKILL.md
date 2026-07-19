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
2. Determine the planning scope before selecting files. Use the user-named
   scope or index. During concurrent worktree work, select the current
   worktree's isolated namespace from the parallel-worktree workflow. Otherwise
   use the ordinary unscoped namespace.
3. Within that scope, use an existing index when it matches the work-queue
   model. Prefer `ROADMAP.md` for ordinary work and the scoped roadmap for
   parallel work. Use `PLAN.md` only when it clearly routes project work. When
   only standalone legacy progress documents exist, report that they need the
   organize branch's explicit migration before another workflow can use them.
4. Read the selected index and every work-item file needed by the workflow.
5. Inspect recent history, code, tests, or diffs only far enough to verify
   material status claims. Treat conflicts as staleness to report or repair,
   not as restrictions imposed by the documents.

Resolution is complete when the current item, the next scheduled item, and any
material conflict between documentation and repository state are known.

## Keep One Work Queue per Scope

- Keep each index compact: a `Current` slot, an ordered `Plans` queue, and an
  unordered `Tasks` pool.
- For ordinary work, store the index at `ROADMAP.md`, scheduled work at
  `plans/<stable-slug>.md`, and unscheduled work at
  `tasks/<stable-slug>.md`. Preserve an established `todo/` or `todos/`
  directory as the task pool instead of creating a duplicate convention.
- During concurrent worktree work, give every concurrently edited worktree a
  distinct roadmap and distinct plan and task subdirectories. Apply the naming
  and ownership rules in the parallel-worktree workflow.
- Persist plan order only in the selected index. Keep sequence numbers out of
  filenames and item files.
- List every active item in a scope exactly once: current, queued, or pooled.
  An item in `Current` is absent from its former list until it stops being
  current.
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

Before ending a mutating workflow, verify that every path in each touched index
exists, each active item appears exactly once within its scope, plan numbering
is continuous, and every touched item states an evidence-backed current state
and concrete next action. During parallel work, also verify that the current
worktree changed only its own planning namespace. Keep transcripts and routine
session logs out of the files. Report changed planning files, validation
performed, and the next intended work or blocker.
