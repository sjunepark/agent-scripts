# Compact the log database

Use this workflow only when the audit reports material free pages in
`logs_*.sqlite`, the user approves compaction, and the installed Codex version's
storage contract is compatible. Compaction is physical maintenance: it preserves
retained logical rows and returns already-free pages to the filesystem. It does
not select tasks or logs for deletion.

## 1. Decide whether to compact

- Record the canonical Codex home and exact database, WAL, and SHM paths.
- Record the main-file size, sidecar sizes, page size, free-page count, estimated
  reclaimable bytes, filesystem headroom, and expected downtime.
- Recommend `VACUUM` when the measured saving is material relative to the
  database and available disk. Skip it when the likely saving does not justify
  an offline rewrite; do not invent a universal percentage or byte threshold.
- Require working space for the private recovery copy and SQLite's temporary
  rewrite in addition to ordinary system headroom. Stop if space is marginal or
  cannot be measured reliably.

## 2. Prove the target is offline

Use the canonical Codex home—not task, workspace, client, or terminal identity—as
the sharing boundary.

1. Inventory every Desktop, CLI, IDE extension, app-server, updater, and helper
   process that may use the target home. An idle task or client remains a possible
   writer because Codex keeps a read-write WAL connection pool and asynchronous
   log inserter for the process lifetime.
2. Allow a live Codex process only when its effective Codex home is positively
   proven to be different. Process names and ancestry alone do not prove this;
   fail closed when attribution is ambiguous.
3. When the platform exposes open-file ownership, verify that no process holds
   the database, WAL, or SHM path open. Treat an unavailable handle check as
   missing corroboration, not proof of quiescence; use other direct ownership
   evidence or stop.
4. Prevent clients, login items, extensions, and updaters from automatically
   relaunching during maintenance. Repeat the process and handle checks
   immediately before opening the maintenance connection and abort if anything
   reappears.

If the current run belongs to the target process tree, stop here and produce an
offline handoff. Do not launch a delayed background command: the user must quit
the clients, run the monitored maintenance from a process outside that Codex
home, and observe every gate and result.

## 3. Create the recovery copy and baseline

- Create a private, access-restricted recovery directory outside the Codex home
  on storage with sufficient headroom. Record its exact path and required restore
  permissions.
- Repeat the process and open-handle gates, then copy the main database and any
  nonempty WAL together while fully offline and before opening a SQLite
  connection to the source. Do not omit or separate a nonempty WAL. Do not copy
  `-shm` as recovery data; it contains no persistent content and SQLite
  reconstructs it from the WAL.
- Open only the copied set and require `PRAGMA integrity_check` to return exactly
  `ok`. Record schema object types, names, owning tables, and SQL definitions
  without treating physical root-page numbers as stable. Also record migration
  rows, `PRAGMA user_version`, and `PRAGMA application_id` as the logical
  baseline.
- Record row counts for every user table. For the current log schema, also record
  the `logs` ID range, `sqlite_sequence`, and the complete migration-row
  identities. Adapt these invariants to the installed schema instead of assuming
  the current table set.
- Record source and recovery byte sizes plus page and freelist metrics. Label
  header-derived source metrics as excluding uncheckpointed WAL frames. Stop if
  the copied set is not independently restorable.
- Treat the recovery set as sensitive because free pages or WAL frames may
  contain deleted log content. Keep it only through restart validation unless
  the user asks to retain it longer.

Structural integrity does not prove application-level content equality; both the
integrity result and logical baseline are required. Creating the recovery copy
before opening the source also prevents an automatic recovery or close-time
checkpoint from becoming the first mutation.

## 4. Checkpoint and compact

1. Re-run the process and open-handle gates.
2. Run `PRAGMA wal_checkpoint(TRUNCATE)` and require a completed, non-busy
   result. A busy or locked result proves quiescence was incomplete; stop and
   identify the owner instead of retrying blindly.
3. Confirm the sidecar state and filesystem headroom again.
4. Run plain `VACUUM` without schema changes, row deletion, retention commands,
   or concurrent maintenance. Capture the exit status and errors.

SQLite locking normally prevents conflicting writes, but lock failure is a last
line of defense, not the quiescence mechanism.

Monitor for client or updater reappearance throughout the operation and stop if
one is detected.

## 5. Validate before reopening Codex

Keep all Codex clients stopped while validating:

1. Require `PRAGMA integrity_check` to return exactly `ok`.
2. Compare logical schema definitions, migration records, version pragmas, table
   row counts, key ranges, and sequence values with the baseline. Ignore expected
   physical page-location changes only. Investigate every logical difference; do
   not explain it away as compaction.
3. Re-measure file size, sidecars, page count, freelist count, and filesystem
   headroom. Report measured reclaimed bytes separately from the preflight
   estimate.

If compaction or validation fails, do not start Codex. Preserve both the failed
database and recovery set, re-establish quiescence, and restore only from the
validated recovery set. Run the full integrity and logical comparison on the
restored database before opening a client.

## 6. Restart in stages

1. Start one intended Codex client and verify database initialization completes
   without migration, locking, or corruption errors.
2. Verify retained tasks can be listed and a representative retained task can be
   resumed. Re-run the storage audit and compare it with the baseline.
3. Only after those checks pass, reopen other clients that share the home.
4. Remove the recovery set after successful restart validation when its cleanup
   was included in the approved operation; otherwise report its path, size,
   sensitivity, and exact removal decision as outstanding.

Report the verified quiescence evidence, logical comparisons, integrity results,
reclaimed bytes, restart checks, and any retained recovery artifact.
