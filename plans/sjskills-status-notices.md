# Show project and global skill status on CLI use

## Outcome

Normal `sjskills` commands tell the user when the current project's skills or
fixed global baseline need attention, naming the affected skills and pointing
to the appropriate plan command. Notices reuse reconciliation's ownership and
content rules, remain separate from command success and approval evidence, and
avoid fetching upstream content on every invocation.

## Current state

Complete under [the goal contract](../goals/sjskills-status-notices.md).
[PR #19](https://github.com/sjunepark/agent-scripts/pull/19) merged as `57dd2f1`
on 2026-09-07. The feature, documentation, review feedback, and native acceptance
checks are delivered. Release, installation, and real-machine rollout remain excluded.

## Next action

None — goal complete.

## Performance evidence

Measured on an Apple M1 Pro (darwin/arm64), using isolated temporary homes and
projects, with the embedded `dev`/`go` selection and fixed global baseline:

- Warm local inspection: 11.4–18.1 ms per two-scope check across four ten-iteration
  benchmark runs. Fixture: 45 placements, 360 files, 2.8125 MiB; upstream calls
  are forbidden in the benchmark. This satisfies the 250 ms target.
- Actual remote cold checks after source/option batching: 7.35 s and 6.29 s;
  both scopes were fresh and staging directories were empty afterward. The
  initial per-skill strategy exceeded 30 seconds and was replaced.

The selected defaults are 24-hour refresh, 15-minute failure cooldown, and a
shared 30-second foreground budget plus bounded process cleanup. CI uses
isolated fake-CLI call assertions rather than network timing thresholds.

## Decisions and boundaries

- Notices cover all actionable updates, missing skills, undeclared extras, and
  conflicts in the nearest configured project and fixed global baseline.
  Manual/workflow ownership and protected/legacy boundaries remain unchanged.
- Reuse reconciliation inventory and classification. Cache complete expected
  hashes, never local findings or apply authority. Command snapshots become
  reusable only after verification and cleanup; notices run after mutation locks
  are released and describe post-operation state.
- Use one bounded foreground check with two independent scopes. Source/option
  batching reuses the pinned materializer and individually verifies every tree.
  Unix process groups and Windows jobs contain subprocesses through cleanup.
- Root flag `--no-status-check` skips all ancillary work. Human notices use
  stderr after command output; JSON has one typed optional `advisories` field.
  Advisory failures preserve primary success and exit status.
- Exact artifact SHA-256 includes notices. Only fresh-plan semantic comparison
  ignores advisories; stable warnings and all plan evidence remain enforced.
  Older artifacts remain readable, and plan/apply must use the same executable.
- No release, installation, real-machine rollout, automatic apply, daemon,
  CLI self-update, registry migration, all-project scan, or replacement
  ownership policy belongs to this goal.

The [registry contract](../docs/skill-registry.md#automatic-status-evidence) owns
current cache, output, freshness, and approval behavior. The
[README](../README.md#skill-installs) owns command usage; the
[sjskills skill](../skills/sjskills/SKILL.md) and its global procedure own operator
authority and evidence preparation.

## Acceptance coverage

All tests use isolated temporary homes, projects, and caches. CI fake-CLI tests
must never fall through to real tooling or mutate a real home.

| Obligation | Evidence |
| --- | --- |
| Correct scope, category, skill, target, reason, and review command for updates/missing/extras/conflicts | Internal status classification tests and CLI renderer/JSON tests |
| Local edits, deletions, successful sync, unowned desired copies, and extra directories detected on warm checks; no invented manual/workflow/protected updates | Provenance/local drift fixtures; no-upstream assertions |
| Cold/expired refresh complete; identity changes invalidate immediately; failed scope independent; stale/unavailable explicit; retry and rollback bounded | Status freshness, identity, cooldown, malformed/missing manifest, and independent scope tests |
| No partial/hostile snapshot; bounded regular-file reads, links rejected, strict format/hash checks, safe locking/pruning, newer publication preserved | Cache adversarial, contention, abandoned-file, rollback, publication, and pruning regressions |
| Successful init/profiles/plan/apply/restore eligible; help/version/invalid/failure/cancel/opt-out skip; same-scope snapshot reuse and post-mutation state | CLI fake-bunx command/call assertions, restore and reviewed global apply tests |
| Human output follows primary output, exact silent, names capped/deduplicated, fresh explicit plan duplicate suppressed; one JSON document | Renderer and CLI integration tests |
| Advisory changes ignored only semantically; byte digest and stable warnings/current/ownership/expected comparisons enforced; old artifacts accepted and unknown fields rejected | Reviewed-plan tests and global approved-apply CLI test |
| Shared foreground deadline, cancellation, parent exit and descendants holding staging handles; staging retained if termination unverifiable | Process-tree and materializer lifecycle tests on supported native targets |
| Warm overhead below 250 ms; representative remote cold checks finish both scopes within 30 s | Recorded performance evidence above; reproducible BenchmarkStatusWarm |

## Delivery evidence

One feature PR delivered the connected model, cache, CLI, documentation, and
acceptance work. Initial and follow-up bounded reviews completed. CodeRabbit
confirmed four fixes and withdrew its batching finding after pinned-parser and
remote-run evidence; all threads are resolved. Codex completed without findings.
Documentation harmonization covers the README, registry contract, sjskills entry
point/global procedure, and goal/roadmap/plan status.

Local validation passed:

```sh
go test ./internal/sjskills ./cmd/sjskills
go test -race ./internal/sjskills ./cmd/sjskills
go vet ./...
node --test scripts/lib/skill-registry.test.js scripts/audit-global-skills.test.js
scripts/validate-skills
python3 scripts/release_test.py
```

Final [run 34108846117](https://github.com/sjunepark/agent-scripts/actions/runs/34108846117)
passed Linux source/build checks, native Go suites on macOS Intel/ARM and Windows,
and temporary installer/consumer checks. Validation exposed and corrected a
cache-pruning race, Darwin process-group disappearance handling, a Windows hash
golden assumption, and the Windows installer's null backup-path binding.
The repeated process-cleanup run passed 20 times. Temporary CI artifact
installation did not activate or install the feature on a user's machine.
