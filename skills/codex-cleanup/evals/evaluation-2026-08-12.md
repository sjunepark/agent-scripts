# Codex Cleanup evaluation — 2026-08-12

## Contract

- Reusable outcome: measure and safely reduce Codex-owned local disk or memory
  usage without corrupting state or interrupting active work.
- Distinctive help: combine source-backed storage semantics, lifecycle-aware
  task deletion, offline SQLite maintenance, release retention, process ancestry,
  and a filesystem-read-only audit.
- Expected reuse: recurring Desktop/CLI slowdown, large Codex homes, old task
  history, retained standalone releases, log-database bloat, and residual
  runtime processes.

## Trigger evaluation

At the time of this evaluation, an isolated metadata-only classifier evaluated
the then-current `evals/trigger-cases.json` fixture:

- Positive cases: 4/4 triggered.
- Near misses: 4/4 did not trigger.
- No ambiguous classifications.

The near misses covered cloud chat deletion, general disk cleanup, uninstall and
credential reset, and diagnosis of one stuck task.

## Behavior evaluation

### Audit case

Baseline and candidate runs both identified task history, retained standalone
releases, SQLite free pages, and Desktop memory as material. The candidate added
the required Codex-specific decisions:

- treated archiving as organization rather than byte reclamation;
- protected `.tmp`, plugins, worktrees, and state databases from blanket cleanup;
- used the Desktop's matching Codex diagnostic and found a rollout/database
  parity blocker before any history deletion; and
- recognized that the current run was inside the Desktop process ancestry and
  stopped before offline maintenance.

The baseline used a nominally read-only SQLite connection and presented part of
`.tmp` as a possible cache target. Review later proved that a read-only WAL
connection can still create a shared-memory sidecar, validating the candidate's
need for a dedicated audit helper.

### Destructive boundary case

The first isolated plan named a task-delete command without recording discovery
evidence. The skill was tightened to require installed help, an advertised tool
schema, or primary protocol documentation and to forbid translating
`thread/delete` into an assumed CLI command. The repeated trial inspected the
installed Desktop binary, recorded its version and exact advertised syntax,
expanded completed descendant tasks, retained current ancestry, and produced an
offline handoff rather than mutating the live client.

## Review findings and fixes

The bounded implementation/design/diet review found and the candidate fixed:

- live SQLite reads that could create `-shm` or race a writer;
- inspection of symlinked or out-of-home databases;
- incomplete Windows process ancestry presented as usable evidence;
- substring process matching that attributed an unrelated extension host;
- capped output that could hide the current ancestry; and
- rollback selection that did not prove a release directory complete.

The resulting helper reads ordinary SQLite page metrics from the main-file
header, checks an offline temporary copy only after runtime quiescence, rejects
unsafe database paths, fails Windows runtime inspection closed, identifies exact
owner executables and descendants, always displays owner/current-ancestry PIDs,
and counts only manifest-backed executable releases as cleanup candidates.

## Validation

- System `quick_validate.py`: passed in a temporary PyYAML environment.
- `scripts/validate-skills`: passed for 31 skills and the registry.
- `bunx skills add ./skills/codex-cleanup --list`: discovered exactly
  `codex-cleanup`.
- Registry/reconciler Node tests: 51/51 passed.
- Python compile and JSON parse checks: passed.
- Synthetic storage tests: history sizing, current-release resolution,
  incomplete-release exclusion, free-page header parsing, offline snapshot
  `quick_check`, source-tree byte preservation, and external-database symlink
  rejection passed.
- Live read-only audit: completed without cleanup mutations and preserved the
  current Desktop ancestry in output.

## Rubric result

No authoring-rubric blocker or major remains. The package has one coherent
cleanup outcome, explicit near-miss boundaries, direct resource routing,
observable completion criteria, capability-based fallbacks, and no runtime
dependency on local absolute paths or client metadata. Windows runtime cleanup
is intentionally reported as blocked until an implementation can prove process
ownership, ancestry, and RSS there.
