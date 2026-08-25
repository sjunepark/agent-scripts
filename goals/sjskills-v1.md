# Goal: Deliver sjskills v1

Status: complete
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

- 2026-08-24: The user selected the active thread contract as governing and
  authorized it to supersede every differing field in the original contract.
  The effective outcome, included-result wording, completion condition,
  exclusions, authority, resume rule, and delivery rule are therefore the
  active thread contract. In particular, substantive implementation no longer
  requires `$delegate`; repository-required review delegation still applies.

## Execution status

### Completed included results

- Version 4 desired-state contracts and pure global and project resolution.
- Pinned, isolated Skills CLI materialization with verified staged content.
- Recoverable project reconciliation, locking, journaling, quarantine,
  restore, and crash recovery.
- Fixed global baseline reconciliation and retirement of the legacy mutation
  engine.
- Operator documentation, validation, review, and PR delivery.

### Current in-scope result

Delivery complete.

### Next in-scope action

None — goal complete. Real-home apply, restore, migration, cleanup, and rollout
remain outside this goal.

### Evidence and blockers

- Implementation and final review fixes merged through PR #13 at merge commit
  `2453488801068feea2fbb624da07aa74f0c7723c`.
- Completion validation covered Go unit and race tests, vet,
  supported-platform builds, external-process contracts, dependency-free Node
  tests, skill and registry validation, syntax and formatting checks, and
  disposable-home remote materialization.
- Read-only parity between `scripts/audit-global-skills` and
  `sjskills plan --global` was demonstrated in disposable homes. The real home
  was not mutated.
- Current operator behavior is documented in the [README](../README.md) and
  [skill registry contract](../docs/skill-registry.md). The completed design
  and implementation record is summarized in
  [`plans/sjskills-skill-reconciler.md`](../plans/sjskills-skill-reconciler.md).
- Real-machine global rollout remains a separate proposed operation governed by
  [`plans/sjskills-global-rollout.md`](../plans/sjskills-global-rollout.md).
