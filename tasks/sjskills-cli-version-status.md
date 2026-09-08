# Include the sjskills CLI version in automatic status checks

## Outcome

The existing status check reports whether the running `sjskills` CLI has a
newer available version alongside project and global skill status. A machine
with current skills must still surface an outdated CLI. Both incidental notices
and the default/explicit status report use the same version-check result.

Checking is advisory. It identifies the running version, available version,
evidence freshness, and a usable update destination without installing anything
or changing the primary command's result.

## Current state

Planning only, captured on 2026-09-08 against `cddd448` and committed as
`580a0ef`. The user subsequently authorized registering it in the roadmap.
Implementation, release, installation, pushing, and selecting this task for
execution remain outside the current authorization.

- `c186e3a` added automatic project/global skill notices; PR #19 merged as
  `57dd2f1`, and `6b45e83` recorded completion. These checks do not inspect the
  CLI's available version.
- `internal/sjskills/version.go` embeds `internal/sjskills/VERSION`, currently
  `1.0.0`. The checkout wrapper also builds with that value; it does not establish
  whether checkout code matches a published binary.
- The [release guide](../docs/sjskills-releases.md) names GitHub Releases as the
  canonical binary source, with `sjskills-vX.Y.Z` tags. Its recorded state says
  first publication is pending. This is local documentation evidence, not a
  fresh inspection of remote release availability.
- Concurrent uncommitted work described in the local task file
  `tasks/sjskills-default-status.md` adds `sjskills` /
  `sjskills status` reports and refactors collection and rendering. Its code,
  tests, task file, and `ROADMAP.md` belong to that ongoing work.
- The ordinary queue has no current item, a proposed global rollout in Plans,
  and default status plus this item in Tasks. Existing goal contracts are
  complete. This plan starts no goal and does not select or displace queued work.

This is an unscheduled draft listed once under Tasks in
[the roadmap](../ROADMAP.md). Registration does not change the current item or
the order of scheduled plans.

## Next action

Resolve the comparison baseline below before implementing the release lookup.
After implementation is authorized, re-read the landed default-status contract
and current diff, and coordinate ownership of shared files before selecting
this task for execution without taking over the parallel work.

## Product decision still open

The user has not answered whether “outdated” means a newer published stable
release exists or newer code exists on `main`.

**Proposed target: latest stable published sjskills release.** The rest of this
plan specifies that target conditionally. It fits the documented distribution
contract and gives users an installable update. It is not recorded as approved.

If `main` freshness is selected, revise this plan before implementation: the
embedded version alone cannot distinguish multiple commits carrying `1.0.0`.
That outcome needs a defined build/source identity and update path for checkout
and standalone users. Do not silently add both channels or fall back from a
missing release to branch comparison. Publishing a first release remains a
separate task under either decision.

## Proposed behavior contract

### Invocation and presentation

| Invocation | CLI-version behavior |
| --- | --- |
| Successful eligible named command | Check once alongside skill checks; print an actionable update or evidence failure on stderr |
| `sjskills` / `sjskills status` | Check once; show CLI status on stdout before project and global status |
| Either surface with `--json` | One existing result envelope containing structured CLI evidence; no human notice |
| `--no-status-check` | Skip CLI lookup and all status-cache work, together with existing skill checks |
| Help, exact version requests, invalid invocation, unsuccessful command | Preserve existing check exclusions and output |

Use the final default-status routing as the integration point. Preserve its
setup guidance, root-flag placement, stream rules, cancellation behavior, and
advisory exit semantics. Do not collect again after the status command has
already produced its report. CLI comparison must work without a project
manifest, valid skill registry, usable global home, Bun, Git, or GitHub CLI.

With fresh evidence, explicit output distinguishes:

- An available update: running version, newest stable version, and the exact
  release page with installation guidance.
- An equal version: no newer stable release, with the compared versions clear.
- A higher running version: ahead of the latest stable release; no downgrade
  advice and no claim of an exact published-build match.
- No published stable release: a successful absence observation, never “up to
  date,” “update available,” or a network failure.
- An uncomparable running version: local version identity cannot be compared;
  skill status remains usable.
- Unavailable evidence: a concise reason, without claiming freshness.

Incidental output stays quiet for fresh equal/ahead/no-release results. A failed
refresh reports stale or unavailable evidence concisely, following existing
skill-notice behavior. A previously observed update may still be shown with an
explicit stale label. Explicit output and JSON expose observation age whenever
cached; stale evidence never produces an unqualified current-version claim.
Sanitize external text and construct update links from trusted identity; do not
print release-body content or offer invented install commands.

### Release selection and version identity

Read release metadata for `sjunepark/agent-scripts` directly over HTTPS. Query
only public metadata and introduce no runtime authentication requirement.
Verify the official release API contract during implementation before selecting
an endpoint; documentation lookup is separate from the shipped runtime.

- Select the highest numeric `X.Y.Z` version among published, non-draft,
  non-prerelease releases whose tags match the sjskills release contract.
  Compare numeric components rather than strings or publication timestamps.
- Ignore other tools' releases and malformed/nonmatching tags. Complete any
  required pagination within bounded request, time, and response-size limits.
  An incomplete listing is unavailable evidence, not proof of no release.
- Do not assume a repository-wide “latest” release belongs to sjskills. Tag
  existence alone is insufficient to establish a published installable release.
- Compare against the running executable's embedded identity on every check;
  do not probe another binary on PATH or use the registry/Skills CLI version.
- Check compatibility with the existing release packaging contract before
  offering update guidance. A malformed release or missing expected target
  asset must not be advertised as an installable update for this platform.
  Report the distribution problem rather than silently claiming current status.
- Report a version comparison only. A checkout with the same embedded version
  may contain different code; binary integrity and commit freshness are not
  established by version equality.

Handle rate limits, offline access, API errors, malformed data, and deadlines
as advisory failures. An HTTP error is never equivalent to a verified empty
release set. Avoid downloading executable archives during checks.

### Evidence lifecycle and resource bounds

Preserve the existing 24-hour upstream refresh interval, 15-minute failure
cooldown, and one shared 30-second foreground refresh budget. CLI lookup runs
independently alongside both skill scopes within that budget; it must not add
a second serial timeout or prevent peer results from being reported.

Cache upstream release evidence once per repository/channel and any platform
inputs used to verify installability, independent of project roots and skill
configuration. Recompare it with the running version each time, so replacing
the executable immediately changes the result even while metadata is cached.

- Cache successful no-release observations as explicit data. Distinguish them
  from a failed cold lookup with no successful observation.
- Failed refresh preserves the last valid observation as stale and records a
  bounded retry time. Never overwrite good evidence with malformed output.
- Validate schema, source identity, sizes, timestamps, and release identity.
  Future timestamps must not extend freshness; clock rollback must not prolong
  retry suppression indefinitely.
- Store only disposable metadata and retry state. Use bounded locking, safe
  regular-file access, atomic replacement, and bounded retention consistent
  with existing status cache guarantees. Concurrent commands must not corrupt
  each other's entries or wait beyond the foreground budget.
- Keep release entries separate from skill expected-hash snapshots and their
  pruning namespace. Do not make a release record impersonate a skill scope or
  force a migration of existing valid skill-cache files.
- If cache access fails, report the limitation without changing command
  success; any direct fallback lookup must still obey the shared budget.

## Integration design and retained contracts

1. Add one focused CLI-version service in `internal/sjskills` owning release
   selection, comparison, cache lifecycle, and typed results. Inject clock and
   HTTP transport/client boundaries for deterministic tests. Avoid a general
   updater framework, background daemon, extra flags, or channel configuration.
2. Extend the existing collector in `cmd/sjskills/status.go` to collect the CLI
   result once with the two scope results. Share invocation eligibility and
   deadlines; keep each domain's result independent. Do not copy the collector
   or reuse `ScopeProject`/`ScopeGlobal` for the executable.
3. Keep skill advisories unchanged. Add one optional typed CLI advisory to the
   envelope for both incidental and explicit results, separate from
   status-only project setup metadata. Model comparison and evidence freshness
   separately with validated combinations; avoid nullable fields that imply
   contradictory states. Settle field names against the landed envelope.
4. Treat CLI metadata as non-authoritative in reviewed-plan handling. Validate
   its shape when present and exclude only that field from fresh-plan semantic
   comparison, just as skill advisories are ancillary today. Continue hashing
   every byte of the approved artifact, including advisories. Preserve strict
   decoding, successful-global-plan checks, all stable plan evidence, and
   rejection of explicit status envelopes/status-only metadata as approval.
5. Preserve acceptance of older artifacts without the optional CLI field.
   Older binaries may reject new envelopes through strict decoding; document
   generating and applying a reviewed artifact with the same compatible CLI.
   Never introduce permissive unknown-field handling to make this work.
6. Give explicit and incidental renderers one shared version summary/link
   formatter. Explicit status always shows the result; incidental rendering
   suppresses fresh non-actionable results. Preserve human escaping and JSON
   completeness without changing skill finding limits or ordering.

Retain the two skill scopes, provenance checks, live apply evidence, exact
approval digest, pure help/version path, platform support, and existing
installer verification. Consolidate only shared status orchestration and CLI
presentation rules. Release metadata remains a distinct domain because its
identity and lifecycle do not depend on desired skill trees. No old behavior
requires a temporary alternate updater, dual cache migration, or branch-based
fallback.

## Implementation checkpoints

1. **Resolve and align:** settle the baseline, recheck remote publication state
   read-only, and inspect final default-status code/tests before changing shared
   files. Confirm executable-version and release tag contracts. Capture current
   output with isolated fixtures to establish the missing CLI notice.
2. **Comparison and evidence:** implement the focused service, typed states,
   bounded metadata retrieval, cache behavior, and failure classification.
   Exercise version ordering, release filtering, and lifecycle tests first.
3. **CLI integration:** connect both status surfaces and JSON, preserve skip
   paths and one collection pass, and update reviewed-plan validation together
   with the envelope. Verify peer independence and primary result preservation.
4. **Consumer validation:** extend isolated CLI and native archive consumer
   checks for the new feature. Update existing status test fixtures to supply
   deterministic release evidence so previously offline tests cannot start
   calling the live network.
5. **Review and documentation:** run one bounded `$code-review` with an
   independent reviewer for the shared CLI/approval contracts. Fix actionable
   issues and validate affected behavior. Then run `$harmonize-docs changes`
   under the repository workflow; locate that skill if not currently loaded.
   Update only documents whose implemented contracts change.

Coordinate shared-file edits with the default-status work. Do not overwrite,
stage, commit, stash, or reset its changes. If its refactor is still in flight,
finish independent service work in an isolated checkout and defer shared
integration until ownership is clear. The present planning commit changes only
this file; no branch switch or worktree-wide staging is needed.

## Acceptance and validation

Use temporary project/home/cache directories, fixture release responses, and
existing fake Skills CLI infrastructure. No real-home reconciliation, real
binary replacement, or live release publication belongs in tests.

| Scenario | Required evidence |
| --- | --- |
| Newer/equal/older stable release | Correct comparison, including `1.9.0` versus `1.10.0`; running version always comes from this executable |
| Mixed releases and pagination | Other tools, drafts, prereleases, malformed tags, and publication ordering cannot select the wrong sjskills release; truncated discovery is unavailable |
| No stable release | Explicit absence result; cached absence remains distinct from a failed lookup |
| Invalid local identity or unusable release | Honest uncomparable/distribution status; no invented update or downgrade instruction |
| Exact embedded version from a checkout | No assertion of commit equality or binary integrity |
| Updated executable with warm metadata | Recomparison immediately reflects new running version without fetching again |
| Warm/expired/failed cache | No warm request; daily refresh; 15-minute failure cooldown; valid stale evidence preserved; age labels accurate |
| Unsafe/corrupt cache, clock changes, concurrent processes | Bounded recovery, no symlink writes, no peer/pruning interference, atomic valid records |
| Missing project, broken registry/home, absent Bun/Git/gh | CLI comparison still completes; skill results accurately retain their own failures/setup state |
| CLI timeout/offline/rate limit/malformed or oversized response | Primary command result and skill results survive; total work stays within shared deadline and existing cleanup bounds |
| Bare/named/incidental and JSON invocations | One collection; correct streams and visibility; one JSON envelope; all states represented deterministically |
| Opt-out/help/version/invalid/unsuccessful invocation | Zero ancillary HTTP requests and status-cache work, preserving each primary command's own requirements |
| Reviewed global plan | Older field-absent artifact accepted; valid optional CLI evidence tolerated; malformed evidence/status artifacts rejected; byte tampering invalidates digest; live evidence changes still block apply |
| Native packaged executable | Feature works without Go/checkout/gh; tests stay offline; supported macOS and Windows consumer behavior is preserved |
| Mutation boundary | No installer invocation, binary writes, skill changes, provenance writes, prompts, or release creation from status |

Run focused service and CLI tests, then `go test ./...`, `go test -race ./...`,
`go vet ./...`, the Node registry/audit tests specified in `AGENTS.md`, and
`scripts/validate-skills` for the completed implementation. Run
`python3 scripts/release_test.py` if packaging/consumer checks change, plus the
existing native consumer validation at its supported delivery boundary.
Measure a warm no-request check and a bounded slow-server case; do not add a
new CI matrix or repeat unrelated suites after success without new evidence.

Implementation documentation targets are `README.md`,
`docs/skill-registry.md`, `docs/sjskills-releases.md`, and affected usage in
`skills/sjskills/SKILL.md`. Explain the comparison source, no-release and stale
states, opt-out, JSON/approval compatibility, update guidance, and shared
latency bound. Update help text that currently promises only skill inspection.
Keep documentation truthful about whether any release has actually shipped.

Completion requires the baseline decision resolved, all applicable acceptance
cases passing, coherent documentation, and no actionable bounded-review
findings. PR delivery, publication, installed-binary upgrades, and real-machine
rollout require their own authorized work. The planning document was reviewed
and committed separately; its roadmap registration keeps it unscheduled.
