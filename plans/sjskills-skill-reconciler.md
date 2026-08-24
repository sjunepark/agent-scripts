# Reconcile global and project skills with `sjskills`

## Outcome

Deliver `sjskills`, a stable Go CLI that makes this repository the control
plane for a minimal machine-global skill baseline and lets each project commit
its own reproducible selection of profiles and third-party skills.

The CLI must turn declared intent into an explainable, recoverable exact state
for `.agents` and `.claude` without making users select harnesses, maintain a
central profile for every project-specific skill, or trust an upstream
installer with deletion and ownership decisions.

## Current state

- The product intent has been clarified and approved. Project manifests are
  committed, reconciliation is managed-exact, unknown entries are preserved,
  and profile or global name collisions fail closed.
- `skill-registry.json` version 3 fixes every skill to `global`, `project`, or
  `catalog` scope and composes the machine-global `dev` and `kicpa` profiles
  from audiences. That model cannot express a single machine-independent
  global baseline plus project-selected `dev + go` or `kicpa` profiles.
- Project applicability is currently explanatory prose in `recommendation.when`;
  it is not executable desired-state input.
- `scripts/audit-global-skills` and `scripts/lib/global-skill-state.js` already
  implement valuable global safety behavior: isolated Skills CLI
  materialization, deterministic tree hashing, exact-root inventory,
  provenance checks, path confinement, verified replacement, quarantine,
  restore, and mutation-race refusal.
- The existing global reconciler is JavaScript and global-only. It requires a
  machine profile and has no committed project-intent model.
- The implementation branch now has a Go module, typed `cmd/sjskills` entry
  point, and source wrapper at `bin/sjskills`; generated binaries remain
  uncommitted and repository maintenance helpers remain in `scripts/`.
- Skills CLI 1.5.23 is present through `bunx` at planning time. It can discover
  and materialize remote skills, but its install process is not the public
  process contract for `sjskills`.
- The historical machine-profile plan is complete within its authorized goal.
  Its unstarted migration and rollout work has not mutated either real machine.

The first implementation result is complete on `codex/sjskills-v1`:
fixture-backed registry v4 and strict project-manifest contracts, pure global
and project resolution, project-root discovery, derived project layout and
minimal provenance shapes, stable plan/process contracts, and a typed Kong CLI
shell all validate without touching a managed root or network. The live
registry remains version 3 until the later atomic cutover.

The isolated materialization result is complete and independently validated:
the adapter invokes exactly Skills CLI 1.5.23 through `bunx`, confines all
home/config/cache/temp writes to an owned temporary root, bounds subprocess and
tree resources, computes legacy-compatible verified hashes, reuses one staged
tree per desired skill, rejects unsafe symlinks and diagnostics, detects
tampering, and cleans up recoverably. The read-only CLI planning path consumes
one verified session, emits stable expected-content evidence, sanitizes private
staging paths before diagnostic truncation, and leaves managed-root sentinels
byte-for-byte unchanged.

The project read-only state boundary is also complete: `sjskills` inventories
only the canonical project `.agents/skills` and `.claude/skills` roots, hashes
real directory entries without following symlinks, and loads strict bounded
reconciler provenance. A pure classifier now distinguishes missing, exact,
outdated, modified, unmanaged, malformed, misplaced, and protected state
without granting ownership from byte equality alone. Symlink placement policy,
planning integration, and every mutation remain deliberately unimplemented.

## Next action

Define and fixture the canonical project placement policy, including the
portable meaning of registry copy and symlink modes, then integrate the
reviewed inventory/classifier into the read-only project plan. Keep mutation,
quarantine, and restore behind later independently reviewed slices; reserve
bounded real-source validation for the final validation phase.

## Accepted product contract

### Desired state and scope

- Global state is one fixed, minimal baseline shared by every machine. There
  are no `dev` or `kicpa` machine-global profiles and no hostname inference.
- Project state is declared by a committed `sjskills.toml` at the project root.
  Its desired set is the union of centrally defined profiles and direct skill
  declarations in that manifest.
- Central project profiles are composable sets such as `dev`, `go`, `rust`,
  and `kicpa`. They do not determine installation scope.
- Direct declarations let a project use third-party skills without first
  changing this repository's central registry or profiles. A direct entry
  records enough source and skill identity to materialize it independently.
- Skill names are unique across the fixed global baseline, selected profiles,
  and direct project declarations. Duplicate names or contradictory sources
  are validation errors; no source silently wins.
- Project scope is the default. Global scope always requires explicit
  `--global` and resolves only the fixed baseline.
- `.agents` and `.claude` are the default targets. Per-skill exceptions may be
  modeled where a skill genuinely supports only one target, but routine users
  do not select harnesses.
- The committed manifest is canonical intent. Installed skill trees and
  reconciler provenance are derived, machine-local state and are not a second
  project configuration source.

### Reconciliation behavior

- There is one behavior: managed-exact. Do not add an additive mode.
- Planning installs missing desired skills and proposes updates for desired
  managed skills whose published content changed.
- Removing a profile or direct declaration proposes quarantine of the
  corresponding previously managed project skill.
- Unknown entries are reported and preserved. A non-conflicting unknown entry
  is a warning and does not prevent convergence of the managed set; an unknown
  entry occupying a desired name or path blocks that operation. Unknown
  content does not become managed merely because its bytes match a desired
  skill.
- A locally modified managed entry blocks replacement until the user resolves
  it; version 1 has no force-replace path.
- Quarantine is manifest-backed and restorable. No operation recursively
  deletes a skills root.
- Every mutation revalidates path confinement, entry type, source identity,
  and content hash after planning and immediately before replacement or move.
- Partial update failure restores the previous managed content or reports the
  precise recoverable state; it never claims convergence after partial work.

### CLI interface

The initial human-facing command set is deliberately small:

```console
sjskills init dev go
sjskills profiles
sjskills plan
sjskills apply
sjskills plan --global
sjskills apply --global
sjskills restore <quarantine-id>
```

- `init` creates project intent without overwriting an existing manifest; it
  does not apply filesystem changes.
- `profiles` provides side-effect-free discovery from the canonical registry.
- `plan` is always read-only with respect to managed roots. It may use the
  network and temporary storage to establish current expected content.
- `apply` recomputes the plan, presents one confirmation, and revalidates the
  approved state before mutation. `--yes` suppresses the prompt for automation
  but never bypasses validation or race checks.
- Global commands need no initialization because the global baseline is
  centrally defined.
- Do not expose `sync`, `--dry`, `--scope`, harness-selection flags, or an
  additive behavior in the initial interface.
- Human output is concise. An explicit JSON mode emits one newline-terminated
  structured stdout document with stable operation, result/error, warning,
  and evidence fields; progress and logs stay on stderr.
- Use a small stable exit-status taxonomy for success, desired-state drift or
  valid execution failure, and invalid invocation.

### Source and dependency policy

- Version 1 follows the latest published registry sources. It records installed
  source identity and verified tree hashes but does not commit a content lock
  or promise offline installation.
- `sjskills` is a Go program using Kong for its typed command model.
- One exactly pinned Skills CLI version is invoked through `bunx` only as a
  private materialization adapter. Do not fall back among `bunx`, `npx`, and
  `pnpm dlx`.
- Preflight verifies `bunx` and the expected Skills CLI version before remote
  work. Version upgrades are explicit code/config changes with validation.
- Every Skills CLI invocation runs with isolated temporary home and harness
  directories, bounded time and output, noninteractive flags, and redacted
  errors. It never receives authority to write the real project or user roots.
- `sjskills` verifies the staged tree and owns final placement, state,
  quarantine, rollback, and reporting.
- Existing `manual` and `workflow` managers remain explicit classifications.
  The CLI reports their required external action and never claims to have
  installed or reconciled them through the Skills CLI adapter.

## Scope boundaries

Included:

- Registry evolution from machine audiences to a fixed global baseline and
  composable project profiles.
- A strict project manifest with direct third-party skill declarations.
- Project and global plan/apply flows for `.agents` and `.claude`.
- Provenance-aware managed state, safe update, recoverable quarantine, and
  restore.
- Migration from the existing global reconciler without losing its safety
  invariants or silently taking ownership of unknown files.
- Human and automation-facing command contracts, installation guidance, and
  validation on the development machine plus portable path fixtures.

Excluded from version 1:

- Additive reconciliation.
- Machine-specific global profiles or automatic mutation based on repository
  detection.
- Arbitrary harness selection, Pi-specific copies, plugin installation, or
  general dotfile management.
- Committed content locks, offline source caches, and source-vendoring policy.
- Creating or administering private repositories and credentials.
- Automatic deletion of unknown, modified, protected, or ambiguous entries.
- Real-home apply, quarantine, prune, or rollout without a separately reviewed
  plan and explicit authorization after read-only validation.

## Implementation plan

### 1. Freeze the version 4 contracts in fixtures

- [x] Define the registry's fixed global baseline, project profile membership,
      catalog/source records, target exceptions, and installation managers
      without duplicating skill declarations.
- [x] Define the strict `sjskills.toml` shape for selected profiles and direct
      source declarations. Reject unknown fields, duplicate names, invalid
      sources, empty selections, and collisions with global/profile skills.
- [x] Define project-root discovery and the relationship between committed
      intent, generated placements, and machine-local provenance.
- [x] Define structured plan operations, issue categories, warnings, evidence,
      JSON result envelopes, and exit statuses before formatting human output.
- [x] Add resolver fixtures for the fixed baseline, `dev + go`, `kicpa`, direct
      third-party entries, target exceptions, manager boundaries, and every
      collision class.
- [x] Keep the live version 3 registry unchanged during this slice so the
      existing global audit remains usable.

Exit condition: pure fixtures describe every accepted product rule and reject
ambiguous desired state without touching a managed root or network.

### 2. Build the deep Go planning module and CLI shell

- [x] Add the repository's Go module and typed Kong command entry point, then
      define a stable installation into a PATH location consistent with the
      repository's `bin/` command convention without committing generated
      binaries.
- [x] Implement strict registry and TOML parsing as project-owned types rather
      than passing raw maps through the program.
- [x] Implement profile/direct/global resolution as a pure operation that
      returns desired placements or typed validation errors.
- [x] Expose planning and application through a small internal interface;
      keep filesystem, process, and presentation details behind it.
- [x] Implement human help plus JSON discovery/validation without granting
      those code paths filesystem-write, subprocess, credential, or network
      capabilities.
- [x] Test the compiled executable as an external process for help, version,
      malformed input, stdout/stderr separation, JSON newline behavior, and
      exit statuses.

Exit condition: the executable can validate and resolve fixtures and manifests
without materializing or installing a skill.

### 3. Add isolated Skills CLI materialization

- [x] Pin one Skills CLI version and verify it through `bunx` during preflight.
- [x] Translate resolved installable skills into bounded, noninteractive
      materialization commands with explicit skill selection and full-depth
      behavior where required; never use `--all`.
- [x] Isolate `HOME`, `USERPROFILE`, `CODEX_HOME`, and Claude configuration so
      subprocesses can write only inside a temporary staging root.
- [x] Materialize each desired remote tree once per plan, compute the
      reconciler-owned tree hash, and reuse those verified bytes for all
      placements in that plan.
- [x] Bound command duration, stdout/stderr bytes, staged tree size, file count,
      path depth, and diagnostic output. Redact credentials and unsafe URLs.
- [x] Test source failure, timeout, oversized output/tree, tampering after
      staging, unsupported source forms, private Git-helper operation, and
      cleanup of temporary state.

Exit condition: a read-only plan can establish verified expected content while
the real project and home remain byte-for-byte unchanged.

### 4. Implement project exact-state reconciliation

- [x] Inventory only the modeled project roots and prove their canonical paths
      remain inside the selected project before reading ownership files or
      planning mutations.
- [x] Define reconciler-owned provenance records for source identity, target,
      verified tree hash, install time, and the minimum recovery evidence.
- [x] Classify missing, exact, outdated, modified, unmanaged, malformed,
      misplaced, and protected state without treating Skills CLI metadata as a
      verified local-content hash.
- [ ] Plan canonical `.agents` and `.claude` placements with portable behavior
      across macOS and Windows; choose copy/symlink details only after fixtures
      prove identical discovery and recovery semantics.
- [ ] Apply additions and verified updates without clobbering changed targets.
      Preserve prior content in manifest-backed quarantine before replacement.
- [ ] Quarantine previously managed skills removed from project intent while
      preserving unknown entries, and implement overwrite-refusing restore.
- [ ] Revalidate approval state before every move or replacement and recover
      coherently from interruption or partial failure.
- [ ] Cover symlink ancestors, path escapes, executable bits, source changes,
      stale provenance, plan/apply races, partial updates, collision failures,
      and restore races using temporary projects only.

Exit condition: fixture projects converge to manifest intent repeatedly, and
every removal or replacement remains recoverable.

### 5. Replace machine profiles with the fixed global baseline

- [ ] Port or reuse every applicable safety invariant and fixture from
      `global-skill-state.js`; document any deliberate semantic difference.
- [ ] Teach the Go planner to inventory user-level `.agents` and `.claude`
      roots under an explicitly selected test home, protecting vendor, cache,
      backup, legacy, and runtime-owned locations.
- [ ] Define migration of trusted existing reconciler provenance so previously
      managed entries remain distinguishable from unknown content.
- [ ] Produce a read-only migration plan showing which former `dev` or `kicpa`
      global skills fall outside the new baseline. Do not quarantine them as
      part of validation or registry migration.
- [ ] Migrate the live registry to version 4 only after the Go command resolves
      and audits the fixed baseline with parity. Keep the version 3 command
      functional until that atomic cutover.
- [ ] Remove the profile argument from the new global interface and either
      retire `scripts/audit-global-skills` or leave a short, explicit
      transition wrapper; do not maintain two independent policy engines.

Exit condition: `sjskills plan --global` fully replaces profile selection and
truthfully reports the current machine without mutating it.

### 6. Complete operator workflow and documentation

- [ ] Implement `init`, `profiles`, human plan presentation, confirmation,
      `--yes`, JSON mode, quarantine identifiers, and restore guidance against
      the same canonical command model used by execution.
- [ ] Update `README.md`, `AGENTS.md`, registry documentation, Skills CLI
      guidance, and sync guidance only when their described behavior is live.
- [ ] Explain how projects commit `sjskills.toml`, ignore or regenerate derived
      placements/state, add third-party skills, review updates, resolve local
      modifications, and restore quarantined content.
- [ ] Document the transition from global machine profiles to the fixed
      baseline and the separately authorized real-machine rollout sequence.
- [ ] Provide installation/update instructions for the Go command without
      adding an unneeded release platform or auto-updater.

Exit condition: a fresh session can configure and reconcile a project or audit
the global baseline without relying on this plan or historical documents.

### 7. Validate, review, and prepare cutover

- [ ] Run all Go unit, fixture, race, vet, and external-process tests.
- [ ] Cross-compile the command for the intended macOS and Windows targets and
      run platform-specific path/placement fixtures where execution is
      available.
- [ ] Run the existing dependency-free registry and global-state Node tests
      until their consumer is deliberately retired.
- [ ] Run `scripts/validate-skills`, local published-catalog validation, syntax
      checks, and `git diff --check` after final changes.
- [ ] Run bounded real-source materialization tests only into temporary homes
      and projects.
- [ ] Run both the legacy global audit and `sjskills plan --global` read-only
      during parity validation; do not apply either to the real home.
- [ ] Run `$code-review`, apply safe findings, repeat validation, and run
      `$harmonize-docs changes` if the final implementation changes documented
      behavior, operations, commands, architecture, or delivery status.
- [ ] Prepare a separate digest- or evidence-bound rollout plan before any
      real global removal, replacement, or quarantine.

Exit condition: the Go CLI is reviewable, documented, read-only parity is
demonstrated, and no real-machine mutation is bundled into implementation
validation.

## Completion criteria

- [ ] Every machine resolves the same fixed minimal global baseline without a
      profile argument.
- [ ] A committed project manifest can resolve `dev + go`, `kicpa`, or another
      profile union plus direct third-party skills.
- [ ] Name and source collisions across global, profile, and direct declarations
      fail before network access or filesystem writes.
- [ ] `.agents` and `.claude` work by default without routine harness flags.
- [ ] Plan is read-only; apply is managed-exact, race-safe, and recoverable.
- [ ] Unknown and modified entries are preserved and truthfully reported.
- [ ] Skills CLI is pinned and isolated behind `sjskills`; it never writes a
      real managed root directly.
- [ ] Human and JSON process contracts are tested through the packaged
      executable.
- [ ] The live registry, CLI, tests, and documentation describe one current
      workflow, with the legacy policy engine retired or explicitly transitional.
- [ ] Real-home rollout remains a separate, reviewed, explicitly authorized
      action after implementation and read-only validation.
