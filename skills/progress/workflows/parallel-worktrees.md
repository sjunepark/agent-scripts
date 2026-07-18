# Parallel Worktrees

Read this only when the user is actually coordinating work in multiple git
worktrees.

1. Establish one shared root index and one distinct work-item file per
   worktree on their common base before parallel implementation begins.
2. Temporarily allow multiple links under `Current`, annotating each with its
   branch or worktree. Remove those links from the plan queue or task pool so
   each active item still appears once.
3. Give one integration worktree ownership of `ROADMAP.md`. After branching,
   worker worktrees update their assigned item files and implementation, while
   the integration worktree owns queue changes.
4. Merge each worker's code and item update, then reconcile `ROADMAP.md` in the
   integration worktree. Unlist completed items and return unfinished plans or
   tasks to the appropriate queue or pool.

Keep the normal layout: one root index plus `plans/` and `tasks/`. Represent
parallelism with temporary `Current` annotations rather than additional root
`PLAN-*.md` files or per-worktree planning directories. Finish only when the
integration index reflects every merged item and ordinary zero-or-one-current
operation can resume.
