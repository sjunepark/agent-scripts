# Include the sjskills CLI version in automatic status checks

## Outcome

Implemented and locally validated on 2026-09-08. The authorized implementation
uses the proposed latest stable published sjskills release as its comparison
baseline. Branch freshness is outside this feature. CLI evidence appears once
alongside skill status in explicit reports, incidental notices, and JSON.

The [status contract](../docs/skill-registry.md#automatic-status-evidence) owns
comparison states, release selection, caching, output, and approval compatibility.
The [release guide](../docs/sjskills-releases.md) owns installation and publication.
Before publication on 2026-09-08, GitHub inspection returned no releases.

## Implementation and validation

- The existing CLI reported only project/global skill status. Integration now
  collects release evidence independently within the existing shared deadline,
  including when project setup, the registry, or skill inspection is unavailable.
- Service tests cover numeric ordering, filtering, pagination bounds, malformed
  and oversized responses, distribution assets, embedded-identity comparison,
  warm reuse, explicit no-release caching, stale preservation, cooldown,
  clock rollback, corruption, symlink preservation, and concurrent cancellation.
- Real CLI subprocess tests cover human/JSON placement, update links, opt-out and
  exclusion paths. Collector coverage blocks all three peers and verifies their
  identical deadline. Renderer tests cover comparison/freshness states and
  terminal escaping. Tests use isolated caches and blocked external HTTP.
- Reviewed-plan tests preserve strict structure, exact artifact digests, and
  stable evidence rechecks while excluding only ancillary CLI metadata from
  semantic equality. Older field-absent artifacts remain supported.
- Full Go and race suites, vet, Node registry/audit tests, and release regression
  tests passed. Release archives cross-built for the supported targets; native
  macOS arm64 consumer and installation-preservation checks passed without
  Go, Bun, Git, or gh on PATH. The hosted release run subsequently passed native
  macOS Intel, macOS arm64, and Windows amd64 tests and installer checks.
- A warm release-only check measured approximately 0.04 ms without HTTP. The
  slow-server test cancels under its 100 ms test deadline and verifies prompt
  return; production shares the existing 30-second budget.
- One bounded independent code review found that missing/null cache fields could
  impersonate a successful no-release observation. Required fields are now
  explicit and regression-tested. No actionable review findings remain.
  Affected documentation was harmonized.

## Delivery boundary

Committed and pushed to main as `7e5e2ba`. Published in the immutable
[sjskills-v1.0.0 release](https://github.com/sjunepark/agent-scripts/releases/tag/sjskills-v1.0.0)
on 2026-09-08. All supported native targets passed before publication; the
[release record](../plans/sjskills-release.md) records the authenticated manual
publication fallback. Installed-binary replacement and other-machine rollout
remain separate operations.
