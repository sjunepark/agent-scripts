# Reduce repeated sjskills status work

## Outcome

Implemented and locally validated on 2026-09-08. The local shell prefers the
installed native CLI. Status reuses upstream expected hashes across projects
with identical materialization inputs, while local inventory, provenance, and
classification remain per-invocation work. The current cache contract lives in
[status evidence](../docs/skill-registry.md#automatic-status-evidence).

## Evidence and decisions

- Real configured-project checks took 6.35–9.50 seconds cold and 18–20 ms warm.
  The development wrapper added about 0.7–1.3 seconds per invocation.
- A traced project refresh spent 0.64 seconds preparing the pinned Skills CLI,
  1.99 seconds fetching repository skills, and 4.55 seconds fetching Context7;
  local inspection took 34 ms.
- Key disposable evidence by exact materialization inputs and format versions,
  independent of project root and unrelated registry metadata. Retain strict
  scope matching for command-produced snapshots and all apply boundaries.
- Keep the 24-hour freshness interval, shared 30-second foreground budget,
  failure cooldown, complete-snapshot validation, and filesystem safeguards.

## Validation and delivery

- The new CLI regression first reproduced repeated materialization in a second
  project, then passed with shared evidence and independent local findings.
- Internal tests cover input separation, irrelevant metadata, ordering, shared
  locks and cooldowns, and reuse after a concurrent refresh. Existing hostile
  cache, provenance, freshness, cancellation, and approval checks still pass.
- Focused tests, `go test -race ./...`, `go vet ./...`, and the bounded code
  review passed; no actionable findings remain. Affected docs were harmonized.
- Real-directory validation: after `darty` refreshed its selection, the first
  `ytm` check reused it in 38 ms (previously 6.35 seconds). Warm checks took
  14–21 ms. New selections still took about 7.6–7.7 seconds to refresh.
- Login and non-login shell resolution and syntax checks passed. The managed
  `.zshrc` source was updated in chezmoi; `.zprofile` is unmanaged.

The tested native binary is installed at `~/.local/bin/sjskills`; fresh-shell
verification confirms the native executable and expected status findings.
Binary release publication and other-machine installation remain separate from
these source changes. Persistent package caching and progress rendering were
not part of this implementation slice.
