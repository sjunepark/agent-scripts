---
name: codex-cleanup
description: "Audit and safely reduce local Codex Desktop and CLI disk or memory usage. Explicit invocation only. Use for assessing or cleaning task history, archived sessions, logs_2.sqlite, cached standalone releases, shell snapshots, or leftover runtime processes. Do not use for deleting cloud ChatGPT chats, general operating-system cleanup, or uninstalling and resetting Codex."
---

# Codex Cleanup

Reclaim measured local storage or memory without corrupting task history, SQLite
state, an installed release, or an active run. Audit first, propose exact targets,
and mutate only the scope the user approved.

## Preserve these boundaries

- Treat audit and planning as read-only.
- Treat an explicit cleanup scope in the current request as authority for that
  scope. Otherwise obtain confirmation after showing exact targets, descendants,
  estimated bytes, reversibility, and whether a restart is required.
- Use the supported task lifecycle capability for permanent task deletion.
  Verify the exact installed interface through its help, schema, or primary
  protocol documentation before presenting or invoking it. Never invent a
  convenience CLI command, and never remove rollout files or task rows directly.
- Quiesce every process that can write the target database before checkpointing,
  copying, or vacuuming it.
- Never terminate the current execution process or one of its ancestors. If the
  current run lives inside the target Desktop or CLI process tree, prepare the
  exact next action and have the user quit or continue from a separate process.
- Never blanket-delete `.tmp`, `plugins`, `worktrees`, `shell_snapshots`, state
  databases, credentials, or configuration. Prove ownership and staleness for
  each individual target first.
- Preserve unrelated user changes and data. Stop when a target cannot be tied to
  Codex or its safe lifecycle is unknown.

## Audit

1. Resolve the effective Codex home from an explicit path, then `CODEX_HOME`,
   then the platform's user-home `.codex` directory. Record the resolved path.
2. Run `scripts/audit_codex_home.py` from this skill directory. It is read-only
   and reports top-level sizes, history age, SQLite free pages, release-cache
   candidates, and the relevant process tree. Use `--json` when retaining or
   comparing results. Add `--quick-check` only when the extra database scan is
   acceptable and the client has been quiesced. Ordinary page metrics come only
   from the main-file header. For `--quick-check`, the script refuses to proceed
   when it detects matching runtime processes and checks a temporary copy so it
   never creates or changes database sidecars in the Codex home.
3. If a compatible Codex executable provides a read-only doctor command, run it
   and capture rollout/database parity before mutation. Prefer the executable
   used by the active Desktop or CLI; do not assume the shell's `codex` is the
   same build.
4. Read [references/storage-contract.md](references/storage-contract.md) when
   interpreting candidates or performing any cleanup. Re-check installed help
   or current primary source if its version notes do not match the installed
   build.
5. Rank opportunities by measured recoverable bytes and likely responsiveness
   benefit. Separate disk storage, database internal free pages, and resident
   memory; do not add them into one misleading total.

Report the audit before acting. Include the current size, candidate size,
proposed action, risk/restart requirement, and default recommendation for every
material category.

## Build an approved cleanup plan

Select only actions supported by the evidence:

- **Task history:** Select exact root task IDs by age, archive state, project, or
  user-supplied criteria. Expand each root to its descendant/subagent tasks and
  show that set. Count a task as completed only when the lifecycle API reports it
  inactive or completion is otherwise proven; do not trust a stale spawn-edge
  flag alone. Exclude the current task and every active writer.
- **Runtime memory:** Identify the owning Desktop or CLI tree, active work, the
  current process ancestry, and stale residual helpers. Prefer a graceful app or
  CLI exit and restart. Propose exact PIDs only for helpers that remain after
  their owner exits.
- **Standalone releases:** Resolve the `current` target. Keep it and, by default,
  the newest distinct rollback release. Consider only other complete release
  directories while installers and updaters are stopped.
- **Log database:** Compact only when free pages are materially useful and the
  filesystem has enough working space for both the safety copy and SQLite's
  temporary rewrite. Database row retention and file compaction are distinct.
- **Other caches:** Act only when the installed version documents ownership and
  cleanup semantics or the user approves exact disposable files after review.

Do not describe archiving as space reclamation: it normally moves task history
and updates its index without removing the rollout content.

## Apply in a safe order

1. Permanently delete approved tasks through a discovered and verified native
   task-delete capability, such as app-server `thread/delete`. Record the
   executable or connected capability, version, method name, and discovery
   evidence. Include the approved completed descendants. Let the lifecycle operation
   update rollout files, indexes, logs, queues, memories, goals, dynamic tools,
   and spawn edges through its consistency and retry ordering. If the API
   refuses an externally referenced fork,
   report it and ask whether to expand the set; do not bypass the check.
   If no verified callable delete interface is available, stop at a plan that
   names the missing capability; do not translate the method into an assumed
   `codex delete` command or fall back to filesystem/SQLite deletion.
2. Re-run the parity diagnostic and history inventory. Stop on missing,
   duplicate, mismatched, or stale state instead of continuing into lower-value
   cleanup.
3. Quiesce runtime writers. Ask the user to finish or stop active work, exit the
   owning client normally, and verify that relevant PIDs exited. Send a graceful
   termination only to exact approved residual PIDs; escalate to forced
   termination only with separate explicit authority.
4. Remove approved obsolete release directories while quiesced. Re-resolve the
   `current` target immediately before deletion and abort if it changed or any
   candidate equals it.
5. For approved database compaction, make a restorable copy while offline, run
   an integrity check, checkpoint and truncate the WAL, run `VACUUM`, then run
   the integrity check again. Abort if free space is insufficient or any writer
   reappears.
6. Restart the intended client and verify it can list and resume retained tasks.
   Re-run the audit with the same options and compare measured bytes and
   processes with the baseline.

When the current run cannot survive quiescence, stop after producing a complete
offline handoff: exact resolved paths, approved targets, excluded current
ancestry, prerequisite exit checks, commands or API operations, validation, and
recovery instructions. Do not schedule an unmonitored destructive job.

## Finish

Finish only when the approved actions have completed or a precise offline handoff
exists. Report:

- bytes reclaimed by category and remaining disk headroom;
- tasks deleted, including descendant count, and tasks deliberately retained;
- database integrity and rollout/database parity results;
- release directories retained and removed;
- runtime processes stopped or still active;
- any skipped target, blocker, or recovery artifact.

Distinguish measured results from estimates and recommendations.
