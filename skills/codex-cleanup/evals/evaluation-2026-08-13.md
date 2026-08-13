# Codex Cleanup trigger revision — 2026-08-13

## Intended behavior

`codex-cleanup` is manual-only. Codex may load it when the user explicitly
invokes `$codex-cleanup` or names the `codex-cleanup` skill, but must not select
it merely because a request concerns Codex storage, memory, tasks, releases, or
runtime processes.

## Candidate

- The portable description says `Explicit invocation only` while retaining the
  cleanup scope and near-miss boundaries.
- `agents/openai.yaml` sets `policy.allow_implicit_invocation` to `false`.
- All four positive trigger fixtures explicitly invoke or name the skill.
- Two in-scope but uninvoked cleanup requests are now near misses, alongside the
  four existing out-of-scope requests.

## Validation

- `scripts/validate-skills`: passed for 33 skills and the registry.
- `bunx skills add ./skills/codex-cleanup --list`: discovered exactly
  `codex-cleanup` and displayed the manual-only description.
- JSON parsing and `git diff --check`: passed.

The Codex adapter policy provides the deterministic manual-only gate. The
revised trigger fixture records 4 explicit positives and 6 near misses; a fresh
probabilistic metadata-classifier run was not needed to enforce the adapter
policy and was not performed.

## Compaction clarification

The storage contract was rechecked against upstream Codex commit
`902bd9e06b3ecb32cbf7f8e64cd23b956be3e7fe`. Current app-server and TUI
processes open `logs_2.sqlite` read-write in WAL mode, maintain connection pools,
and start asynchronous log writers; startup maintenance also deletes expired
rows and checkpoints the WAL. The skill now blocks compaction until every
process sharing the target Codex home has exited, regardless of whether its
tasks appear idle. A process may remain only when its effective Codex home is
proven different.

The revision also states that `VACUUM` preserves retained logical rows and only
returns already-free pages to the filesystem. It recommends compaction when the
measured saving is material, uses full `PRAGMA integrity_check` before and after
the rewrite, and preserves the offline backup until restart validation passes.

The follow-up revision moves the fragile operation into a directly routed
compaction workflow. It adds canonical-home and open-handle gates, records
logical invariants in addition to structural integrity, stages the restart with
one client before reopening the rest, and defines recovery behavior for lock,
interruption, integrity, or content-comparison failures. The audit now states
that its privacy-preserving process inventory cannot attribute effective Codex
homes, so its runtime block is intentionally conservative. Evaluation case 6
captures the distinction between multiple sessions sharing one home and a
separately proven session using another home.

SQLite documentation review clarified recovery-copy semantics: the main file
and nonempty WAL are persistent state, while `-shm` is a reconstructible index.
The workflow creates and validates an offline main-plus-WAL recovery set before
opening the source, excluding `-shm` from recovery data; the audit's temporary
snapshot likewise copies the WAL but not `-shm`.

### Follow-up validation

- `scripts/validate-skills`: passed for 33 skills and the registry.
- `bunx skills add ./skills/codex-cleanup --list`: discovered exactly
  `codex-cleanup` with the manual-only description.
- Registry and reconciler tests: 55/55 passed.
- Python compilation, JSON parsing, and `git diff --check`: passed.
- A synthetic WAL-mode database with committed content in a nonempty WAL and a
  present `-shm` passed the audit helper's temporary-snapshot `quick_check`
  after copying only the main file and WAL.
- A live audit reported that Codex-home attribution was unavailable and blocked
  `quick_check` conservatively while matching runtime processes were active.

The new multiple-session behavior case was added to `evals/evals.json`; a fresh
model behavior replay was not run in this revision.
