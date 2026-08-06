# Goal: Profile-aware global skill reconciler

Status: active
Planning scope: ROADMAP.md

## Original contract

Goal contract
- Outcome: Deliver an operator-ready exact-state global skill reconciler for the public `dev` and `kicpa` profiles with strict auditing, safe apply/prune behavior, current operator guidance, and complete validation.
- Goal state: goals/profile-aware-global-skill-reconciler.md
- Included results and sources (semantic results define scope; paths supply detail):
  - Exact-state strict audit — plans/profile-aware-skill-sync.md; scripts/audit-global-skills; scripts/lib/global-skill-state.js
  - Remote-only profile reconciliation with explicit Codex and Claude targets — plans/profile-aware-skill-sync.md; skill-registry.json
  - Recoverable quarantine-based pruning and restoration — plans/profile-aware-skill-sync.md; scripts/lib/global-skill-state.js
  - Updated operator documentation, skills, and bounded validation — plans/profile-aware-skill-sync.md; AGENTS.md
- Complete when: Every included result achieves Phase 4 and its applicable Phase 7/8 criteria within its named semantic boundary; repository-required validation and review pass; real-home validation remains read-only; planning is truthful; Delivery finishes.
- Excluded: Private KICPA source support; real-machine apply/prune, legacy-root migration, and rollout.
- Authority: Execute only included results and necessary supporting work; resolve remaining decisions within that closed outcome using best judgment; record anything else and ask before scope expansion or external authority.
- Resume: Initialize this contract with $progress goal mode before work; recover it before every resume, continuation, compaction, or handoff; stop if recovery fails.
- Delivery: PR delivery — use $progress's PR lifecycle and the fewest sequential reviewable PRs; finish each through $create-pr and $address-pr-feedback before starting the next, including the final implementation slice.

## Authorized amendments

_None._

## Execution status

### Completed included results

- Profile-aware registry schema and resolver (phases 1-2).
- Fixture-only exact filesystem inventory, classification, hashing, and operation planning model (phase 3).
- Exact-state strict audit and remote-only public-profile apply with explicit Codex and Claude targets (phase 4).
- Canonically confined, manifest-backed quarantine and crash-resilient restore within the goal boundary.
- Operator guidance for profile selection, publication verification, verified-state adoption, apply, prune, and restore.

### Current in-scope result

PR #6 feedback resolution and merge delivery.

### Next in-scope action

Commit and push the validated feedback fixes, resolve the review threads, and
merge PR #6 after its checks remain green.

### Evidence and blockers

- Inherited pre-goal work is present in commits `1964df1` and `4b2dc11`; both
  are included in pushed PR #6 with the goal implementation commits.
- Real-home validation is authorized only in read-only audit mode.
- Bounded pre-PR review identified symlink confinement, restore idempotence,
  provenance, source sanitization, stale-copy replacement, and rollback-
  guidance gaps; all received fixture-backed fixes.
- PR review identified moving-ref reuse, publication verification, executable-
  bit hashing, approval binding, subprocess bounds, and operator-documentation
  gaps. The local feedback patch addresses every actionable thread with a
  single verified staged snapshot per skill, recoverable verified-update
  quarantine, digest-bound candidate and replacement-content approval, and
  updated guidance.
- Current validation passes 49 dependency-free tests, skill and registry
  validation, JavaScript syntax checks, `git diff --check`, and local catalog
  discovery. The final real-home `dev` audit remained read-only and failed
  strictly on the existing protected, legacy, and unclassified state as
  intended.
- Private KICPA source support, real-machine apply/prune, legacy-root migration, and rollout remain excluded.
