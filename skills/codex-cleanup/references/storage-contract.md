# Codex local storage contract

Use this reference when classifying audit results or performing cleanup. These
semantics were verified against the open-source Codex repository at commit
`902bd9e06b3ecb32cbf7f8e64cd23b956be3e7fe` on 2026-08-13. Treat paths and
method names as version-sensitive and verify them against the installed build
when behavior differs.

## Desktop and CLI relationship

The Desktop UI is distributed separately from the open-source Codex repository.
The open-source app-server is the backend contract for rich clients and exposes
task lifecycle methods. A Desktop build may bundle its own Codex executable or
embed the server; it need not run the `codex` found on `PATH`. Inspect the active
process executable before relying on version-specific behavior.

Primary references:

- Codex open-source surfaces: <https://learn.chatgpt.com/docs/open-source>
- App-server overview: <https://learn.chatgpt.com/docs/app-server>
- Codex source: <https://github.com/openai/codex>

## Known local data

| Data | Meaning | Cleanup contract |
| --- | --- | --- |
| `sessions/`, `archived_sessions/` | Task rollout history | Archive only reorganizes history. Permanently delete through the task lifecycle API. |
| `state_*.sqlite` | Task index and related state | Never delete or edit rows by hand. Use lifecycle operations and parity diagnostics. |
| `logs_*.sqlite` plus WAL/SHM | Structured local logs | Retention deletes rows, but the main file may retain free pages. Compact only offline after integrity and space checks. |
| `packages/standalone/releases/` | Versioned standalone CLI releases | The installer moves `current` but may retain older releases. Keep the resolved current target and a rollback version. |
| `shell_snapshots/` | Per-task shell state | Codex normally prunes stale or orphaned snapshots. Do not blanket-delete while tasks are active. |
| `.tmp/` | Locks, maintenance state, and plugin metadata | Not a disposable temp directory. Inspect exact files and owner semantics. |
| `plugins/` | Installed plugin cache and data | Use the plugin lifecycle; do not treat the whole tree as a cache. |
| `worktrees/` | Potential user or task work | No general deletion contract is established. Prove ownership and preservation first. |

## Task lifecycle guarantees

The app-server protocol includes task/thread archive, unarchive, and delete
operations. Archive takes writer locks, moves the rollout, and updates SQLite;
it usually saves no bytes.

Permanent delete is subtree-aware. The current implementation shuts down loaded
tasks, expands spawned descendants, deletes rollout variants and indexes, then
removes associated task logs, queue items, memories, goals, dynamic tools,
spawn edges, and state rows. It refuses a deletion that would leave a fork
outside the delete set referring to removed history. These consistency rules are
why raw removal from `sessions/` or `archived_sessions/` is unsafe.

Use the installed build's native task-delete capability after verifying its
exact callable surface through installed help, an advertised tool schema, or
primary protocol documentation. App-server names the method `thread/delete`;
that does not imply a shell command named `codex delete`. If no verified
interface is available, stop and report that permanent history cleanup cannot
be done safely in the current environment.

## Diagnostics and database maintenance

Current Codex builds provide a read-only doctor diagnostic that can compare
rollout files with the task database and flag missing, stale, mismatched, or
duplicate entries. Use the matching executable when available and run it before
and after task deletion.

Structured log startup maintenance removes old rows and checkpoints the WAL, but
does not necessarily shrink the main SQLite file. `freelist_count * page_size`
estimates internal free space; it is not additional filesystem usage. A full
`VACUUM` rewrites the database and must not race a Desktop, CLI, app-server,
updater, or helper that can write it.

For a WAL-mode database, the main file and any nonempty `-wal` file together
form persistent database state and must remain together when copied or restored.
The `-shm` file contains only a shared-memory WAL index; SQLite reconstructs it
from the WAL, so do not treat a copied `-shm` file as recovery data.

Each current app-server and TUI process opens `logs_2.sqlite` read-write from its
resolved Codex home in WAL mode, maintains a connection pool, and starts an
asynchronous log inserter. Startup maintenance can also delete expired log rows
and checkpoint the WAL. Multiple Codex processes may normally share the database,
but an apparently idle task or client is not evidence that its process will not
access it. For maintenance, stop every Desktop, CLI, IDE extension, app-server,
and helper process sharing the target Codex home. A process may remain only when
its effective Codex home is proven different; when that cannot be established,
treat it as a possible writer and do not compact.

`VACUUM` is transactional and preserves the retained logical database content
while rebuilding the physical file to return free pages to the filesystem. It
does not choose or delete retained tasks or log rows. SQLite normally protects
the rewrite with a transaction and locking, but relying on lock failure is
insufficient here: a concurrent process can make the prerequisite checkpoint,
safety copy, baseline, or validation inconsistent. Recommend compaction when
measured reclaimable pages are material; skip it when the likely filesystem
saving is negligible.

SQLite may change implicit `ROWID` values in tables without an explicit
`INTEGER PRIMARY KEY`. The current `logs` table uses an explicit integer primary
key, but verify the installed schema and compare logical invariants before and
after maintenance rather than assuming that detail is stable across versions.

## Runtime cleanup

Treat memory cleanup as a process-lifecycle task, not file deletion. Attribute
helpers to an owning Desktop or CLI process tree, let active work finish, exit
the owner normally, and verify descendants disappear. A helper is not stale
merely because it is old or uses substantial memory.

Never terminate the process ancestry executing the cleanup. When quiescence
would end the current run, produce an offline handoff rather than launching a
delayed or unmonitored destructive command.
