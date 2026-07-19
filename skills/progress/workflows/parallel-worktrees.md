# Parallel Worktrees

Read this only when the user is actually coordinating work in multiple git
worktrees.

1. Inspect `git worktree list`, branch names, existing planning files, and the
   user's ownership instructions. Assign one stable planning scope to every
   concurrently edited worktree. A user-named scope wins; otherwise derive a
   short label from the branch or worktree name. Normalize `<scope>` to
   lowercase kebab-case and `<SCOPE>` to uppercase snake case.
2. Give every scope its own planning namespace:
   - `ROADMAP_<SCOPE>.md`
   - `plans/<scope>/`
   - `tasks/<scope>/`

   For example, Rust and dev work use `ROADMAP_RUST.md` with `plans/rust/` and
   `tasks/rust/`, and `ROADMAP_DEV.md` with `plans/dev/` and `tasks/dev/`. If
   the repository already uses `todo/` or `todos/` as its task pool, namespace
   that directory the same way.
3. Assign each active item to exactly one scope. Keep one `Current` slot in
   each scoped roadmap; parallel work is represented by the separate indexes,
   not by multiple worktrees editing a shared `Current` section.
4. Treat the current worktree's namespace as its writable planning boundary.
   Read other namespaces only for coordination or a requested aggregate status
   report. Record a cross-scope handoff in the originating item; the receiving
   worktree creates any follow-up item in its own namespace.
5. Before merging a worker branch, update that worker's roadmap and item files
   to their durable current state. The integration worktree maintains its own
   namespace and does not reconcile a shared roadmap after the merge.

When parallel work starts from a shared `ROADMAP.md`, propose an
ownership-preserving mapping into scoped indexes and item directories. Ask
before moving files or replacing links unless the user already authorized the
migration. Finish only when every concurrent worktree has a distinct namespace,
every active item has one owner, and no planning file is assigned for
concurrent edits.
