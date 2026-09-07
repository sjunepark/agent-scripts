---
name: macos-storage-cleanup
description: "Audit and safely reclaim local macOS storage through measured targets and authorized actions. Explicit invocation only; use when the user names $macos-storage-cleanup for low disk space, broad Storage categories, large files, application caches, developer package caches or build artifacts, containers, cloud-local files, device backups, or local snapshots. Do not use for Windows cleanup, Codex-owned task and database cleanup, disk repair, malware response, device reset, secure erasure, or blind deletion of System Data."
compatibility: "Requires macOS and read-only filesystem, process, and command inspection; optional cleanup requires the supported owning app or CLI."
---

# macOS Storage Cleanup

Measure what occupies local storage, identify exact targets and effects, and
complete authorized cleanup. Stop with an itemized proposal for choices that
remain outside the user's request or delegated authority.

When invoked without a narrower scope, default to a read-only staged audit of the
startup volume and current user's data. Do not scan other users, request Full
Disk Access, run an intensive whole-filesystem traversal, or change anything.

## Preserve the authority boundary

- Treat every audit and proposal as read-only. Moving an item to Trash, emptying
  Trash, deleting a cache, pruning a store, uninstalling software, stopping a
  process, changing a retention setting, removing a local cloud copy, deleting a
  snapshot, or running a command that garbage-collects data is a state change.
- Use the current request and applicable earlier authority. “Remove JetBrains
  caches” authorizes verified regenerable JetBrains caches, including resolving
  exact product/version paths; it does not require a second user turn. Honor
  explicit delegation to choose cleanup candidates within its stated limits.
- A broad “clean my Mac” or “free as much space as possible” leaves data-retention
  choices unresolved. Inspect and propose those choices before deleting. A clear
  “remove them” can authorize previously identified targets; resolve ambiguous
  referents from context or ask when materially different targets remain.
- Preserve scope: a JetBrains-only request does not authorize other categories.
  Settings, plugins, Local History, user files, backups, container volumes, and
  unrelated Trash contents are not implied by cache-cleanup authority. Verify
  command flags and descendants stay within the authorized data effects.
- Recheck changed paths, symlinks, contents, processes, sync state, and command
  behavior before mutation. Restore safe preconditions within existing authority;
  ask again only when the changed target or effects exceed that authority. Never
  transfer approval through an unexpected symlink.
- Work without elevation during discovery. Treat macOS privacy or permission
  errors as **not inspected**; report them instead of bypassing them. Full Disk
  Access or elevation is a separate user decision, not a routine prerequisite.
- Never recursively delete `/System`, `/Library`, `/private`, all of
  `~/Library`, an application-support root, a package-manager home, a container
  or virtual-machine disk, a Photos library package, a Git object database, or
  an unknown directory merely because its name or measured size looks disposable.
- Stop cleanup when storage hardware reports a critical condition, the system is
  unstable, an active security incident is suspected, or the proposed target is
  the only verified backup. Preserve data and hand off to recovery, hardware,
  security, or backup work.

Read [references/cleanup-catalog.md](references/cleanup-catalog.md) during every
task before interpreting candidates, proposing actions, or executing cleanup.
It distinguishes supported inspection and cleanup surfaces, version-sensitive
commands, regeneration costs, sync effects, and protected data.

## Establish a read-only baseline

1. Record the user's goal, the exact requested roots or categories, current free
   bytes on the startup volume, and what amount of headroom would count as success.
2. Inspect the APFS/container layout and macOS Storage categories without adding
   their figures together. APFS volumes share space, purgeable bytes and snapshots
   complicate “available” space, and the Storage **Documents** category is not the
   same thing as `~/Documents`.
3. Scan in stages. Start with the requested path or current user's top-level
   directories, then inspect only material categories. Use allocated size where
   available, stay on the intended filesystem, do not follow symlinks, and record
   skipped or timed-out paths. Do not use `sudo` to turn an access failure into a
   broader scan.
4. Resolve managed locations through the installed owning tool or app where
   possible. Examples include npm, pnpm, Yarn, Bun, `uv`, pip, Cargo, Go, Gradle,
   Maven, Homebrew, Xcode, JetBrains IDEs, Playwright, Docker, Android tools,
   cloud-sync providers, and virtual-machine managers. Re-check local help when
   the catalog marks a command version-sensitive.
5. Use only operations proven read-only during the baseline. A command named
   `verify`, `repair`, `clean`, `prune`, `gc`, `optimize`, or `free up space` may
   mutate data. If its behavior is unclear, inspect the resolved directory and
   stop at a proposal.
6. Identify the owning process, sync provider, backup role, and lifecycle of each
   material candidate. Do not infer that an old file is unused, a cache is fully
   regenerable, a stopped container has no valuable volume, or an archive is a
   duplicate backup.

## Rank measured opportunities

Rank candidates by reclaimable local bytes, confidence in ownership and
regeneration, user value, reversibility, operational cost, and risk. Keep these
classes separate:

- regenerable caches and downloaded toolchains;
- project build outputs and dependency trees;
- app-managed downloads, container data, simulators, and virtual machines;
- large user files, archives, media, and backups;
- cloud-synced local copies;
- macOS-managed data, logs, and snapshots.

Do not combine overlapping directory totals or present logical database size,
cloud size, purgeable space, and immediately reclaimable filesystem bytes as one
number. Label estimates and unknowns explicitly. Prefer a smaller, well-supported
action over a larger target with ambiguous ownership or recovery.

## Preview the exact plan

Give every proposed action a stable row identifier and all of these fields:

| Field | Required detail |
| --- | --- |
| Evidence | Measured size, owning tool or app, and why the target is a candidate |
| Exact target | Resolved path, app item, snapshot ID, container object, or setting |
| Exact operation | Supported command or app action, including every material flag |
| Expected benefit | Estimated reclaimable local bytes without double counting |
| Data effect | What is deleted, retained, stopped, moved, or made online-only |
| Risk and recovery | Reversibility, redownload/rebuild cost, backup, and sync scope |
| Preconditions | App/process shutdown, completed sync, network, time, elevation, or restart |

Keep user files, Trash, synced data, backups, snapshots, archives, container
volumes, IDE Local History, Xcode archives, and project dependency trees as
separate rows. Show the exact resolved targets and operations before execution;
already-authorized rows may proceed in the same turn after verification. Ask
only about rows whose targets or effects are not covered by existing authority.
An unresolved row does not block independent authorized rows.

## Apply authorized rows

1. Immediately before each action, re-resolve the target without following an
   unexpected symlink, remeasure it, verify its owner and filesystem, and check
   required process, sync, backup, and free-space gates. Reassess authority on drift.
2. Use the supported owning app or CLI with exactly the previewed flags. Prefer
   the narrowest reversible mechanism. Never substitute a broad recursive
   deletion because a supported cleanup action is unavailable or slow.
3. Apply one row at a time. Capture the command or app result and remeasure free
   space before starting another row. Stop that branch on an error or unexpected
   effect; do not escalate to force flags, permissions changes, a wider target,
   or permanent deletion beyond the authorized operation.
4. Treat Trash as storage, not completion. Moving an exact item to Trash is one
   authorized action; permanently emptying those exact items needs its own
   applicable authority. Never empty unrelated Trash contents to realize an estimate.
5. Preserve unsaved work before stopping or restarting an app or the Mac. Proceed
   only when the interruption is covered by existing authority; otherwise ask
   after showing its effect. Do not infer permission to schedule future cleanup
   from a request to clean now.

## Verify and finish

Repeat the relevant baseline and report actual free-space change separately from
the sum of item sizes. List completed rows, exact targets retained, measured
bytes reclaimed by category, rebuild or redownload consequences, Trash or
recovery artifacts, failures, skipped paths, permission gaps, and remaining risk.

The task is complete when either:

- the read-only audit and itemized proposal have been delivered with no state
  change; or
- every authorized row has been applied and verified, with all unauthorized rows
  and unrelated data unchanged.

Never claim cleanup succeeded solely because a command exited successfully.
