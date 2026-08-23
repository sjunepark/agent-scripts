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

_None._

### Current in-scope result

Version 4 desired-state contracts and pure resolution.

### Next in-scope action

Define the registry v4, project-manifest, resolution, operation, and process
contracts in fixtures, then implement the pure Go resolver and CLI shell.

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
