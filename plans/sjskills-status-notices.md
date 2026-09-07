# Show project and global skill status on CLI use

## Outcome

Normal `sjskills` commands tell the user when the current project's skills or
fixed global baseline need attention, naming the affected skills and pointing
to the appropriate plan command. Notices reuse reconciliation's ownership and
content rules, remain separate from command success and approval evidence, and
avoid fetching upstream content on every invocation.

## Current state

Planning only; implementation has not started. The user requested this feature
and a detailed plan. Daily upstream refresh and coverage of all actionable
drift are proposed defaults from the design discussion, not explicit answers
to the earlier clarification questions. The contract below makes those defaults
concrete for implementation review.

Repository evidence:

- `cmd/sjskills/main.go`: `application.prepare` resolves one scope, materializes
  expected content, inventories installed files, and classifies a plan.
  `executeWithInput` owns dispatch and final output; help and version bypass
  ordinary command execution. There is no automatic cross-scope notice service.
- `internal/sjskills/classify.go`: a verified managed copy differing from
  expected content is `verified-update`; local modifications and unowned
  desired placements are blocked separately. The global classifier reuses
  managed-state classification.
- `internal/sjskills/materialize.go`: expected hashes come from verified
  temporary copies produced through the Skills CLI. Fetching is substantially
  more work than reading a version string; there is no advisory fingerprint
  cache to reuse today.
- `internal/sjskills/types.go` and `reviewed_plan.go`: JSON uses one `Envelope`;
  global approval loads it strictly and compares normalized envelopes, including
  top-level warnings. Time-varying notices cannot be appended to those warnings.
- `internal/sjskills/discover.go` finds the nearest parent `sjskills.toml`.
  The registry is embedded in the executable; checking source content does not
  update registry membership or check for a new CLI release.

## Next action

When implementation is requested, redispatch through `$progress`, then implement
the advisory result model and approval-boundary regression tests described in
the first implementation slice. Treat the proposed product defaults below as
visible planning assumptions and incorporate any user corrections before coding.

## Proposed behavior

### Invocation and scope

- After a successfully dispatched `init`, `profiles`, `plan`, `apply`, or
  `restore`, inspect the fixed global baseline and the nearest configured
  project, even when the command selected only one scope.
- Skip automatic work for help, version, invalid invocations, cancelled
  commands, and unsuccessful operations. Preserve the primary diagnostic and
  exit status; an explicit failed plan is not converted into a cached success.
- No project manifest means global-only inspection. Do not initialize a
  project or infer profiles. An invalid project manifest produces an advisory
  check error while global inspection remains independent.
- Findings cover the current project's configured `.agents/skills` and
  `.claude/skills` roots plus the existing fixed-global roots. Reuse existing
  read-only global inventory observations of protected and legacy locations,
  but exclude those locations from update or reconciliation recommendations.
  Do not add scans of other projects, built-in skills, or plugin contents.
- Add one root flag, `--no-status-check`, to disable ancillary inspection,
  fetching, and cache writes for scripts or deliberate offline use. It does
  not disable the requested command's own verification or network needs.

### Findings and output

- Distinguish available updates, missing desired skills, undeclared extras,
  and conflicts requiring attention. Preserve stable reason codes and targets.
  A local modification, untrusted provenance, or source mismatch is never
  described as a routine update.
- Manual and workflow-managed skills retain their existing classification;
  do not invent an upstream version or an automatic update recommendation.
- Group human notices by scope and category. Count distinct skills within each
  group, retain affected targets in structured data, and sort deterministically.
  A skill may have an update in one target and a conflict in another.
- Write human notices to stderr after normal command output. Show at most five
  names per category followed by a remaining count. JSON contains full findings
  in a separate optional `advisories` field in the same result document; JSON
  mode does not additionally print human notices to stderr.
- Include the upstream observation time and freshness in structured output.
  Human findings based on cached evidence show its age; failed refreshes clearly
  say that upstream status could not be refreshed. Never imply that an old or
  unavailable observation establishes current upstream state.
- Recommend `sjskills plan` for project findings and `sjskills plan --global`
  for global findings. These review instructions cover extras and conflicts
  without suggesting an unconditional apply.
- Suppress duplicate human findings already printed by a fresh explicit plan
  for that scope. Emit nothing advisory when both scopes are verified exact;
  unavailable evidence must not silently masquerade as exact state.
- Repeat actionable findings on eligible invocations until resolved. Do not
  add notification history, dismissals, or a separate scheduling service.

Illustrative output:

```text
sjskills: project — updates available: code-review, modern-go (checked 2h ago)
sjskills: global — missing: clarify
sjskills: project — local changes need attention: teach
sjskills: review with `sjskills plan` and `sjskills plan --global`.
```

### Freshness, delay, and failure

- Reinspect local content and provenance on each eligible invocation. Cache
  verified expected hashes, not final findings, so local edits, deletions, and
  successful syncs become visible without waiting for the refresh interval.
- Refresh a matching scope's complete expected-hash snapshot when absent or
  at least 24 hours old. A fresh explicit `plan` or verified `apply` preparation
  can supply that snapshot without another materialization.
- Run needed ancillary scope refreshes in-process under a shared 30-second
  deadline, with at most two concurrent refreshes. Each scope has independent
  results; a broken source or manifest in one cannot discard the other's
  evidence. Propagate CLI interruption into the ancillary context. Stop the
  refresh process and its descendants before removing temporary staging;
  bounded pipe waiting alone does not establish that descendants have stopped.
- This is a bounded foreground check: warm invocations use local evidence,
  while a first or expired-cache invocation can wait up to the refresh budget
  plus bounded cleanup. Do not promise invisible background work or spawn a
  detached process. The main command's work has already completed.
- After a failed refresh, keep matching last-successful evidence, label it
  stale, and retry no sooner than 15 minutes later. Without matching evidence,
  mark upstream status unavailable. Do not feed partial expected maps into
  classifiers that require a complete desired set. A failed check is not drift.
- Cache directory errors, corrupt cache data, missing tooling, offline sources,
  and timeouts remain advisory failures. They do not alter the command's result,
  exit code, installed skills, provenance, quarantine, or approval state.
- Measure the warm path and representative cold refreshes before accepting
  this design. Target under 250 ms added warm latency for a representative
  dev/go project plus global baseline on a development machine. Record fixture
  size and measurements; use deterministic dependency-call assertions in CI.
  If normal cold checks cannot complete within the budget, revise fetching
  granularity or the explicit latency contract before declaring completion;
  an implementation that routinely reports unknown is not sufficient.

## Implementation design

### One classification policy, separate advisory lifecycle

Add a small status service in `internal/sjskills` that owns expected-hash cache
lookup, scope inspection, classification, and advisory summaries. Keep command
dispatch, flag parsing, final rendering, and access to command-produced verified
hashes in `cmd/sjskills`. Use the existing resolver, inventories, classifiers,
materializer, and diagnostic sanitization. Do not shell out recursively to
`sjskills plan` or build a second ownership policy.

Extract only the reusable preparation/inspection seams needed by both callers.
Keep `preparedPlan`'s verified staging lifecycle and apply sessions intact;
advisory hashes must never substitute for the live snapshots needed by apply.
Distinguish observation completeness/freshness from finding categories using
explicit types, including an unavailable result without fabricated operations.

### Cache contract

- Store disposable versioned cache files under the platform user-cache
  directory in `sjskills/status/`, outside project and global managed state.
  Keep the clock, cache root, and refresh dependency injectable for tests.
- Identify a scope by canonical project root or selected global home. Validate
  the expected-content identity using a deterministic digest of the resolved
  desired set, registry content, source identities, install options, targets,
  and tree-hash/cache format versions. Do not key only by skill name or CLI
  version. Changed identities cannot reuse old evidence, including when offline.
- Store successful complete expected-hash maps, their observation time, and
  bounded retry metadata separately from installed provenance. Validate all
  cached hashes and the exact expected installable set before classification.
  Reject future timestamps as stale and handle clock rollback explicitly.
- Use bounded regular-file reads, safe cache paths, and atomic replacement;
  never follow a cache entry symlink into another location. No source content,
  credentials, raw subprocess output, or approval artifacts belong in the cache.
- Coordinate concurrent refreshes per scope with a short-lived lock. Contention
  uses valid cached evidence or returns unavailable without waiting on another
  invocation's network work. Clean abandoned temporary files/locks safely and
  prevent an older refresh from overwriting newer evidence. Prune obsolete
  entries under a bounded retention policy during refresh only.
- Deleting the cache requires no migration or recovery. It causes a cold check.
  It never removes managed files or invalidates trusted reconciler provenance.

### JSON and reviewed-plan compatibility

Add `Envelope.Advisories` as an optional, typed field. Preserve existing
`warnings`, `evidence`, `plan`, result, and exit semantics. Keep advisory data
outside the approved semantic comparison by clearing only this field in
`normalizeReviewedEnvelope`; retain all existing stable evidence comparisons.
The SHA-256 still binds the exact artifact bytes, including any advisory bytes.
Only the subsequent fresh-plan semantic comparison ignores advisory content.

The updated loader must accept older artifacts without this field, continue
rejecting unknown fields, and validate the advisory shape when present. Older
executables with strict decoding cannot consume new artifacts containing it;
retain the existing operational requirement to plan and apply with the same
retained executable. Do not add a general unknown-field escape hatch, downgrade
converter, or new provenance schema.

### Command lifecycle

Collect validated command-produced expected hashes before staging cleanup, but
publish them to the advisory cache only after successful verification and
cleanup. Ancillary checks run after the requested operation and release of its
mutation locks. They must not run while an interactive prompt is pending.

After successful apply, re-inventory current files against its verified expected
hashes. After restore, re-inventory against matching cached hashes or perform
the ordinary bounded refresh. Never cache or display pre-mutation findings.
For an unsuccessful operation, skip ancillary checks; the next invocation's
fresh local inventory determines the actual state. Failure to write disposable
cache data cannot turn an otherwise successful mutation into a failed command.

## Implementation slices

1. **Result model and approval contract.** Add typed advisory observations and
   findings, deterministic summarization, and the optional envelope field.
   Prove that changing advisory timestamps/findings alone does not break a
   reviewed recheck, while changed artifact bytes still fail the supplied hash
   and changed stable plan evidence still fails comparison. No automatic hook
   is enabled in this slice.
2. **Scope inspection and cache.** Extract the shared inspection seam; implement
   validated complete-snapshot cache storage, identity invalidation, freshness,
   retry cooldown, locking, cancellation, and cleanup. Extend the materializer's
   process lifecycle only as needed to terminate refresh descendants safely on
   supported platforms; preserve ordinary plan/apply behavior. Reuse current
   ownership classification and materialization. Validate isolated scope
   failures and fresh local classification against cached upstream hashes.
3. **CLI integration and output.** Add the dispatch hook, `--no-status-check`,
   renderer, structured JSON output, command-result reuse, and post-mutation
   inspection. Cover every invocation category and suppress duplicate explicit
   plan notices. Measure warm and cold behavior against the stated budget.
4. **Documentation and final validation.** Update the README's command behavior,
   `docs/skill-registry.md` for advisory versus trusted evidence boundaries,
   and `skills/sjskills/SKILL.md` plus its global procedure where the operator
   workflow is affected. Keep this item current; run one bounded `$code-review`
   pass and `$harmonize-docs changes` after each reviewable behavioral slice.

These are dependent implementation checkpoints within one feature. They do
not mandate multiple PRs, commits, or delivery work. Commit, push, release,
installation, and real-machine reconciliation require a later request.

## Acceptance and validation

Use temporary homes, projects, and cache roots throughout. Extend the existing
fake Skills CLI at `cmd/sjskills/testdata/fakebunx/main.go` where needed; tests
must not depend on external network access or mutate the user's real home.

| Scenario | Required evidence |
| --- | --- |
| Project plus global drift | Correct scope, category, skill names, targets, and review commands; neither scope lost. |
| Local edits, missing files, extras, and unmanaged desired copies | Existing reason/ownership rules preserved; local edits never mislabeled as updates. |
| Exact state | No human advisory; no misleading clean result from incomplete or stale evidence. |
| Warm cache | No upstream subprocess; changed local files detected on the next invocation. |
| Cold/expired cache | Successful complete snapshots refreshed within budget; no partial snapshot accepted. |
| Offline or failed source | Healthy scope still reported; stale age or unavailable state explicit; exit code unchanged. |
| Manifest/source/target/registry change | Incompatible evidence rejected immediately, including under retry cooldown. |
| Missing or malformed project | Missing project skips cleanly; malformed project cannot suppress global results. |
| Apply/restore | Notices describe observed post-operation state; no second same-scope materialization when reusable evidence exists. |
| Help/version/invalid/failure/opt-out | No ancillary inventory, subprocess, or cache write; command output contracts preserved. |
| JSON | Exactly one document; structured findings; no human notice on stderr; old artifacts remain readable by the new binary. |
| Reviewed global apply | Advisory-only changes ignored semantically; byte digest, stable warnings, ownership, current/expected hashes still enforced. |
| Concurrent/cancelled refresh | Bounded waiting; a fake CLI spawning a descendant that retains staging access is fully stopped before cleanup, including on timeout/interruption; no torn writes, stale overwrite, or changed managed roots. |
| Hostile/corrupt cache | Oversize, symlink, malformed hash, wrong identity, future timestamp, and unsupported schema handled safely. |
| Manual/workflow/protected paths | No invented upstream update or expanded reconciliation scope. |

Run focused tests as each slice lands, then the relevant repository checks:

```sh
go test ./internal/sjskills ./cmd/sjskills
go test -race ./internal/sjskills ./cmd/sjskills
go vet ./...
node --test scripts/lib/skill-registry.test.js scripts/audit-global-skills.test.js
scripts/validate-skills
```

Use the existing native release-test workflow for supported-platform coverage
when implementation is delivered; local execution proves only the current host.
Retain warm/cold timing evidence and subprocess counts with the implementation
review. Planning-file validation checks links, queue uniqueness, and whitespace;
it does not claim any of these behavior tests have passed.

## End-state boundaries

Retain the existing strict approval, provenance, protected-root, and same-binary
artifact contracts. Consolidate inspection and classification where shared;
keep the advisory cache permanently distinct from authoritative apply evidence
because their lifetimes and trust differ. No background daemon, automatic
apply, CLI self-update, remote registry migration, all-project scan, generalized
notification framework, or replacement materializer is part of this feature.
