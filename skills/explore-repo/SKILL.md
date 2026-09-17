---
name: explore-repo
description: "Inspect external Git source at pinned refs or compare upstream behavior using the centralized ~/.repos cache."
---

# Explore Repo

Use external source when implementation details, version-specific behavior, or
history are needed; prefer installed source or primary documentation when they
answer the question. Requires Git and access to the target repository.

## Cache and ref

Use `~/.repos/<host>/<owner-or-group>/<repo>`, stripping a trailing `.git`.
Do not clone into the current project, a tracked worktree, or `.tmp/` unless the
user requests that location. Do not add ignore rules to hide external clones or
wire cached repositories into the project as dependencies.

Before reusing a clone, inspect its status and remote identity:

```bash
git -C <cache-path> status --short --branch
git -C <cache-path> remote -v
```

For a dirty clone, inspect read-only or use an isolated worktree; do not reset,
clean, pull, check out, or overwrite it without explicit authorization. Update
a clean clone with `git -C <cache-path> fetch --prune --tags` when fresh refs are
needed. For a new clone, prefer `git clone --filter=blob:none <url> <cache-path>`.
Use `--depth=1` only when history and older refs are unnecessary.

Fetch and inspect the named branch, tag, commit, release, or dependency version.
Resolve a more exact ref from local evidence when needed, and record the actual
commit with `git -C <cache-path> rev-parse HEAD`.

## Inspection and experiments

Use focused search and file/history reads. Delegate independent exploration when
useful, choosing a fast, low-reasoning configuration for routine scans and a more
capable configuration for complex logic or architecture. Request concise
findings with paths, refs, and supporting evidence.

Keep cached clones read-only by default. For builds, edits, generated files, or
risky checkouts, create an isolated worktree under `~/.repos/.worktrees/`:

```bash
git -C <cache-path> worktree add --detach <worktree-path> <ref>
```

Remove only worktrees created for this task and no longer needed. Cached clones
are retained; delete them only for requested cleanup, after inspecting the
specific candidates and their state, size, and age.

Report the answer with remote URL, cache path, inspected commit, and relevant
files or symbols. Distinguish direct source observations from inference.
