# Show useful status when sjskills runs without a subcommand

## Outcome

Implemented and locally validated on 2026-09-08. `sjskills` and `sjskills status`
share an explicit project/global report, including setup guidance when no
project manifest exists. JSON uses `operation: "status"` with setup metadata and
existing typed advisories. Reports are advisory and exit 0 for drift, missing
setup, stale evidence, and per-scope unavailability.

The current operator contract lives in
[status evidence](../docs/skill-registry.md#automatic-status-evidence), with
[usage](../README.md#skill-installs) and
[native consumer requirements](../docs/sjskills-releases.md#verification-and-publication)
linked from their entry points. The
[automatic-notice plan](../plans/sjskills-status-notices.md) remains historical
and owns the preserved cache/materializer infrastructure decisions.

## Implementation and validation

The isolated baseline reproduced exit 64, empty stdout, and an expected-subcommand
error before implementation. The pinned Kong parser's `default:"1"` route now
selects the same command as explicit `status`; parse errors remain errors.
Collection retains project discovery alongside independent scope advisories.
The explicit renderer shares grouping, escaping, age, and review-command rules
with incidental notices, while showing every checked scope.

| Acceptance area | Evidence |
| --- | --- |
| Bare/named routing, root flags, JSON and help | `TestDefaultStatusCLI`, `TestDefaultStatusDispatchNoWork`; one status envelope, no plan/evidence |
| Missing setup, nearest nested root, unsafe/unreadable manifest | `TestDefaultStatusCLI`, `TestDefaultStatusNestedAndUnsafeConfiguration`; global peer survives, no initialization |
| Invalid/deleted/inaccessible start, registry/home failures | `TestDefaultStatusDiscoveryFailuresAndOptOut`; removed-cwd subprocess also checked during review |
| Fresh exact state and incidental silence | `TestDefaultStatusFreshAndApprovalBoundary`; both scopes synchronized only inside isolated fixtures |
| Four drift categories, duplicate targets, escaping, caps, complete JSON | `TestDefaultStatusPresentation` and existing status classifier/renderer tests |
| Fresh/stale/unavailable, empty/nonempty findings and observation age | `TestDefaultStatusPresentation` and existing internal cache/freshness tests |
| Opt-out and invalid invocations do no status work | `TestDefaultStatusDispatchNoWork`, injectable opt-out discovery assertions |
| Cancellation, shared budget and staging cleanup | `TestDefaultStatusAlreadyCancelled`, `TestDefaultStatusSignalCleanup`, `TestStatusScopesShareDeadlineAndCancellation`, existing materializer descendant tests |
| Cold/warm status and no duplicate collection | `TestDefaultStatusCLI`; each selected skill materialized once cold, no upstream calls warm |
| Existing named commands and notices | Existing CLI status eligibility, plan/apply reuse, restore and full command suites |
| Strict reviewed-global-plan boundary | `TestDefaultStatusFreshAndApprovalBoundary`; status operation and status fields, including case variants and null, rejected before materialization |
| Native archive without Bun | Extended `scripts/test-release.py`; temporary archive build, verification and native Darwin/arm64 consumer passed |

Completed checks:

- Focused CLI/internal tests, `go test ./...`, `go test -race ./...`,
  and `go vet ./...` passed.
- Node registry/audit tests, `scripts/validate-skills`, release unit tests,
  and whitespace checks passed.
- Existing warm benchmark passed at approximately 17 ms per check on this
  Darwin/arm64 machine; refresh policy was unchanged, so no real-network timing
  was needed.
- One bounded code review found a case-insensitive JSON status-field rejection
  gap. It was fixed with case-folded presence checks and regression coverage;
  no actionable findings remain. Affected documentation was harmonized afterward.

Tests use fake Skills CLI materialization, isolated homes, cache and staging.
The existing 24-hour refresh, 15-minute cooldown, shared 30-second budget,
provenance, protected paths and apply approval boundaries remain intact.

## Delivery boundary

Implementation and local validation are complete. The user authorized committing,
pushing, and updating the local CLI on 2026-09-08. Binary release publication and
other-machine rollout remain separate operations. Supported-target workflow
validation belongs to release delivery; local native consumer validation covered
Darwin/arm64.
