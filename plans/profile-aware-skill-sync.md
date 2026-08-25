# Profile-aware cross-machine skill sync

Status: complete; historical and superseded

## Purpose

This plan delivered the predecessor to `sjskills`: a strict global reconciler
that selected either a `dev` or `kicpa` machine profile, managed explicit
`.agents` and `.claude` destinations, and preserved uncertain content.

## Implemented results

- Registry versions 2 and 3 introduced machine audiences, explicit profiles,
  and destination-oriented targets.
- Exact-state inventory distinguished missing, outdated, modified, misplaced,
  unclassified, protected, and verified legacy copies.
- Remote desired content was materialized in temporary state and verified
  before installation.
- Apply updated desired entries without deleting unrelated content. Prune was
  separately confirmed and moved only revalidated managed content into
  manifest-backed quarantine.
- Restore was path-confined and overwrite-refusing.
- The implementation and review fixes merged in PR #6. Validation covered the
  registry, reconciler fixtures, skill metadata, local catalog discovery, and
  read-only inspection of the development home.

## Deferred work

The plan deliberately did not perform these operations:

- create or configure a private KICPA skill repository;
- mutate either real machine;
- remove former-profile skills or Pi-specific copies;
- classify and clean every legacy or externally owned skill placement.

Those unchecked rollout items are not active work under this historical plan.
The accepted product direction changed before they were authorized.

## Replacement

[`sjskills` v1](sjskills-skill-reconciler.md) replaced machine-global profiles
with one fixed global baseline and moved `dev`, `go`, `rust`, and `kicpa`
selection into project manifests. It also replaced the JavaScript mutation
engine; `scripts/audit-global-skills` remains only as a read-only transition
wrapper.

Current behavior is documented in the [README](../README.md) and
[registry contract](../docs/skill-registry.md). Any real global mutation must
follow the proposed, separately authorized
[global rollout plan](sjskills-global-rollout.md).
