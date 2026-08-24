# Goal: Deliver sjskills v1

Status: active
Planning scope: ROADMAP.md

## Original contract

Goal contract

- Outcome: Deliver the complete `sjskills` v1 Go CLI as the single supported control plane for safe, reproducible project and global skill reconciliation.
- Goal state: goals/sjskills-v1.md
- Included results and sources (semantic results define scope; paths supply detail):
  - Version 4 desired-state contracts and pure resolution — plans/sjskills-skill-reconciler.md; skill-registry.json
  - Isolated Skills CLI materialization — plans/sjskills-skill-reconciler.md; scripts/audit-global-skills
  - Recoverable project exact-state reconciliation — plans/sjskills-skill-reconciler.md; scripts/lib/global-skill-state.js
  - Fixed global baseline and legacy-engine cutover — plans/sjskills-skill-reconciler.md; skill-registry.json; scripts/audit-global-skills
  - Complete operator workflow, documentation, and validation — plans/sjskills-skill-reconciler.md; README.md; AGENTS.md
- Complete when: Every included result achieves its cited outcome and applicable completion criteria within its named semantic boundary; repository-required validation and review pass; planning is truthful; real-home state remains unmodified; Delivery finishes.
- Excluded: Real-home apply, quarantine, prune, migration, or rollout; that requires a separate reviewed plan and explicit authorization.
- Authority: Execute only included results and necessary supporting work; resolve remaining decisions within that closed outcome using best judgment; record anything else and ask before scope expansion or external authority. Invoke $delegate for every substantive implementation slice; the parent retains scope, decisions, independent review, and final validation.
- Resume: Initialize this contract with $progress goal mode before work; recover it before every resume, continuation, compaction, or handoff; stop if recovery fails.
- Delivery: PR delivery — use $progress's PR lifecycle and the fewest sequential reviewable PRs; finish each through $create-pr and $address-pr-feedback before starting the next, including the final implementation slice.

## Authorized amendments

_None._

## Execution status

### Completed included results

- Version 4 desired-state contracts and pure resolution.

### Current in-scope result

Isolated Skills CLI materialization.

### Next in-scope action

Integrate the verified materialization session into the read-only planning
path so `sjskills plan` can establish expected content without touching real
managed roots.

### Evidence and blockers

- Boundary classification: initialization and delivery-base preflight are
  necessary prerequisites for the contract's PR lifecycle.
- `main` is the integration branch: it is unprotected, directly pushable,
  current with `origin/main`, and the repository has no automatic production
  deployment. This avoids an unnecessary staging branch and aggregation PR.
- The preexisting goal at
  `goals/profile-aware-global-skill-reconciler.md` is complete; this is the
  only active goal in the `ROADMAP.md` planning scope.
- Real-home mutation remains excluded. All development mutation and validation
  must use repository fixtures, temporary projects, or temporary homes.
- Registry v4, strict manifest, project discovery, derived layout/provenance,
  pure resolution, operation/envelope, and CLI-shell contracts are implemented
  in Go fixtures while the live version 3 registry and legacy engine remain
  operational.
- Parent validation passed: Go formatting, unit tests, race tests, vet, all 55
  legacy registry/global-state tests, `scripts/validate-skills` (35 skills),
  local published-catalog validation, and `git diff --check`.
- The pinned Skills CLI 1.5.23 adapter now materializes each installable skill
  once into an isolated, bounded temporary root; computes the legacy-compatible
  `tree-sha256-v2` digest; rejects unsafe sources, paths, trees, and diagnostics;
  detects staged tampering; and owns idempotent cleanup. Manual/workflow entries
  remain explicitly skipped.
- Parent validation of the adapter passed fresh Go unit and race runs, a
  repeated real subprocess-bound test, vet, Windows compile-only validation,
  all 55 legacy safety tests, `scripts/validate-skills`, formatting, and diff
  checks. The phase remains active until the CLI planning path consumes the
  verified session.
