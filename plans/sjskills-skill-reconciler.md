# Reconcile global and project skills with `sjskills`

Status: complete

## Outcome

`sjskills` is the repository's supported reconciliation engine for one minimal
machine-global skill baseline and reproducible project selections. It turns
declared intent into explainable, recoverable exact state for `.agents` and
`.claude` while preserving content it cannot safely own.

The implementation is merged. The separate
[global rollout plan](sjskills-global-rollout.md) remains proposed and does not
authorize mutation on any real machine.

## Product contract

### Desired state

- `skill-registry.json` owns one fixed global baseline and composable project
  profiles such as `dev`, `go`, `rust`, and `kicpa`.
- A project's committed `sjskills.toml` selects profiles and may add direct
  third-party declarations. Installed trees and reconciler provenance are
  derived machine-local state.
- Project scope is the default. Global scope requires explicit `--global` and
  never infers a machine profile or hostname.
- `.agents` and `.claude` are the default targets. Per-skill target exceptions
  remain centralized in the registry.
- Selected names must be unique across the baseline, profiles, and direct
  declarations. Duplicate identities or contradictory sources fail before
  materialization or managed-root writes.

### Reconciliation

- There is one managed-exact behavior; there is no additive mode.
- Plan is read-only with respect to managed roots. It may use the network and
  disposable storage to establish verified expected content.
- Apply recomputes the plan, obtains one confirmation unless `--yes`, and
  revalidates approved filesystem and provenance evidence under lock.
- Missing desired content is installed. Trusted outdated content is replaced
  only after the old tree is retained in manifest-backed quarantine. Trusted
  former desired content is quarantined rather than deleted.
- Unknown content is preserved. A non-conflicting unknown entry is a warning;
  an unmanaged desired path, locally modified managed tree, malformed
  provenance, symlink escape, or unsafe filesystem boundary blocks the
  affected operation.
- Restore re-proves the whole committed quarantine and refuses to overwrite an
  occupied destination.
- Interrupted transactions recover from strict private journals. Ambiguity is
  retained with recovery evidence rather than overwritten or reported as
  converged.

### Sources and ownership

- `sjskills` follows the registry's current remote sources and records source
  identity and verified tree hashes; v1 does not promise offline installs or a
  committed content lock.
- One pinned Skills CLI release is invoked through `bunx` only inside isolated,
  bounded temporary homes. It never receives authority over real managed
  roots.
- `sjskills` verifies staged trees and owns final placement, provenance,
  quarantine, rollback, and reporting.
- Manual and workflow-managed entries remain externally owned. The CLI reports
  their required action without claiming to reconcile them.

## Command surface

```console
sjskills init dev go
sjskills profiles
sjskills plan
sjskills apply
sjskills restore <quarantine-id>
sjskills plan --global
sjskills apply --global
sjskills restore --global <quarantine-id>
```

`--json` emits one newline-terminated structured stdout document; progress and
diagnostics remain on stderr. JSON mutation requires `--yes`.

## Implementation results

1. Version 4 registry and strict project-manifest parsing provide pure,
   collision-checking global and project resolution.
2. The Go CLI and `bin/sjskills` wrapper provide typed commands without a
   committed generated binary.
3. The materializer confines, bounds, verifies, reuses, and cleans staged
   Skills CLI output.
4. Project inventory and classification cover missing, exact, outdated,
   modified, unmanaged, malformed, misplaced, and protected state.
5. Project apply, removal quarantine, restore, locking, journaling, and crash
   recovery share one exact-state transaction engine.
6. Global plan, apply, provenance migration, quarantine, and restore reuse that
   engine with home-scoped paths. Former-profile and Pi-specific placements
   remain report-only.
7. The JavaScript profile mutation engine was retired;
   `scripts/audit-global-skills` is a read-only wrapper for
   `sjskills plan --global`.
8. Operator docs, registry guidance, tests, platform builds, bounded review,
   and PR feedback were completed before merge.

## Completion evidence

The implementation and final review fixes merged through PR #13. Completion
validation covered Go unit and race tests, vet, external-process contracts,
supported-platform builds, dependency-free registry and wrapper tests, skill
validation, catalog discovery, and remote-source materialization in disposable
homes. Real-home validation remained read-only.

Current operator guidance belongs in the [README](../README.md); registry and
ownership details belong in
[`docs/skill-registry.md`](../docs/skill-registry.md). This plan is the durable
design and delivery record, not an active progress tracker.
