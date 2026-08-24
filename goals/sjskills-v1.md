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

- 2026-08-24: The user selected the active thread contract as governing and
  authorized it to supersede every differing field in the original contract.
  The effective outcome, included-result wording, completion condition,
  exclusions, authority, resume rule, and delivery rule are therefore the
  active thread contract. In particular, substantive implementation no longer
  requires `$delegate`; repository-required review delegation still applies.

## Execution status

### Completed included results

- Version 4 desired-state contracts and pure resolution.
- Isolated Skills CLI materialization.
- Recoverable project exact-state reconciliation.

### Current in-scope result

Fixed global baseline and legacy-engine cutover.

### Next in-scope action

Port the applicable project safety invariants into one fixed global baseline
and keep all real-home validation read-only.

### Evidence and blockers

- Boundary classification: initialization and delivery-base preflight are
  necessary prerequisites for the contract's PR lifecycle.
- For this goal only, `main` is the explicitly preflighted integration branch:
  the repository has no remote `dev` branch to target, `main` is unprotected
  and directly pushable, and there is no automatic production deployment.
  The repository-wide default remains to use an existing remote `dev`; this
  goal-specific exception avoids creating an unauthorized aggregation branch.
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
  checks.
- `sjskills plan` and the planning portion of `apply` now consume one verified
  materialization session, emit stable expected-content evidence, clean all
  temporary state before exit, and leave project and global sentinel roots
  byte-for-byte unchanged in external-process tests.
- The bounded phase review applied two safe fixes: adapter-owned staging paths
  are sanitized before diagnostic truncation, and direct cleanup failures use
  an exact non-leaking lifecycle message. The clean review loop and full Go,
  Windows compile, legacy Node, skill, catalog, formatting, and diff validation
  matrix passed. Bounded real-source validation remains assigned to the final
  validation phase.
- The project inventory reads only the two derived skill roots and strict
  bounded provenance under a proven canonical project boundary. It reports
  unsafe or malformed state with stable reasons, never follows child symlinks,
  and leaves temporary-project sentinels unchanged.
- The pure project classifier uses only verified expected hashes and trusted
  reconciler provenance. It deterministically distinguishes all eight planned
  state classes, preserves unknown content, refuses byte-equality ownership,
  proposes update/quarantine only for unchanged managed bytes, and blocks
  symlinks or otherwise unverifiable desired copy placements. Version 1 accepts
  only `mode: copy` for Skills CLI declarations; legacy v3 registry/JS behavior
  remains unchanged.
- The read-only project `plan` and planning portion of `apply` now materialize
  exactly once, verify once, copy verified expected hashes into the classifier,
  translate deterministic desired operations, preserve non-conflicting unknown
  entries as warnings, and clean the temporary session on every path. Project,
  `.sjskills`, quarantine, and supplied home sentinels remain byte-for-byte
  unchanged.
- The bounded project-plan review closed two safe gaps: the materializer now
  rejects non-copy desired input instead of silently invoking `--copy`, and
  warning text escapes invalid observed filenames. Full Go unit/race/vet,
  Windows compile-only, all 55 legacy Node tests, skill/catalog validation,
  formatting, and diff checks passed.
- The bounded read-only slice review applied the source-identity compatibility
  fix for credential-free HTTPS remotes with explicit ports. Unit, race, vet,
  Windows compile-only, formatting, and diff validation passed; one external
  CLI test pass saw a non-reproducing preflight mismatch, followed by five
  clean standalone CLI-package runs and a clean race run.
- The internal install-only transaction now locks before its fresh approval
  check, rejects non-install mutations and non-portable staged symlinks before
  writing, publishes copy trees without replacement, syncs their directory
  chain, proves every desired managed placement exact against candidate
  provenance, and preserves ambiguous rollback content instead of recursively
  deleting a raced live path. Non-conflicting unknown entries remain preserved
  and reported.
- Parent review caught and closed stale no-op convergence, unchanged-placement
  commit races, rollback ownership TOCTOU, lock loss, dynamic-warning
  duplication, and ancestor durability gaps. Full Go unit/race/vet, repeated
  Darwin no-replace, Windows compile/build, all 55 legacy Node tests,
  `scripts/validate-skills` (35 skills), local catalog, formatting, and diff
  validation passed.
- Project `apply` now retains exactly one verified materialization session
  through confirmation and the locked install transaction, requires `--yes`
  for JSON automation, installs missing copies idempotently, preserves unknown
  entries, blocks unsupported mutations before prompting or writing, and maps
  conflicts through a stable process issue. Global apply remains read-only and
  unavailable.
- Parent review corrected ambiguous failure evidence so the CLI reports only
  known committed placements rather than inferring unchanged roots. Packaged
  executable tests cover default decline, stdin confirmation, JSON silence,
  idempotence, strict provenance, unknown preservation, temporary cleanup, and
  isolated project/home sentinels.
- The clean review loop passed fresh Go unit/race/vet, repeated Darwin
  no-replace behavior, Windows test compilation and CLI build, all 55 legacy
  Node tests, `scripts/validate-skills` (35 skills), temporary-home local
  catalog validation, formatting, and diff checks. The affected-doc audit
  updated this goal and implementation plan; operator docs remain deferred
  until the complete workflow is live.
- The internal project executor now applies verified installs and updates in
  one locked transaction. For each update it proves the reviewed provenance
  and old tree, publishes a strict path-free manifest before moving content,
  moves the old tree without replacement into a unique `0700` quarantine run,
  publishes only the verified staged inode, commits deterministic provenance,
  and retains the old tree under a committed recovery handle.
- Partial and raced updates restore only re-proven content. Ambiguous
  destinations preserve both external bytes and quarantined originals with a
  `recovery-required` manifest; preparation failures after run creation still
  return durable rollback evidence. Install-only and no-op transactions create
  no quarantine; at that internal-only checkpoint, the public CLI still
  rejected update operations.
- The bounded code-review loop fixed a post-publication ownership race by
  requiring the destination inode to match the staged inode before recording
  rollback ownership, and kept new fault-injection seams package-private.
  Fresh Go unit/race/vet, twenty repeated Darwin lifecycle and swap tests,
  Windows test compilation and CLI build, all 55 legacy Node tests,
  `scripts/validate-skills` (35 skills), temporary-home catalog validation,
  formatting, and diff checks passed without real-home mutation.
- Public project `apply` now accepts verified updates as well as installs,
  presents one mutation-count confirmation, emits separate truthful installed
  and updated evidence, and renders only a validated path-free quarantine
  handle when durable recovery state exists. Removed-skill quarantine remains
  blocked before prompting or writing.
- Packaged-process coverage proves default-declined updates are byte-preserving,
  confirmed updates retain exact old bytes and provenance in a strict committed
  manifest, reruns are idempotent, and a raced external replacement produces a
  `recovery-required` handle while preserving both external and quarantined
  content. The clean bounded review passed ten repeated high-risk update runs,
  full Go unit/race/vet, all 55 legacy Node tests, 35 skill validations,
  temporary-home catalog validation, formatting, and diff checks.
- The internal project transaction now accepts a freshly reviewed removal only
  when its canonical source and exact current/expected hash match one trusted
  provenance record. It moves the unchanged managed tree without replacement,
  deletes that record atomically, and records an action-typed strict manifest
  whose removal entries omit replacement fields and remain restorable.
- Removal-only and mixed install/update/remove transactions share one durable
  quarantine and preserve unknown or modified content. Rollback restores only
  re-proven originals; raced external replacements retain both the external
  bytes and quarantined original under `recovery-required` evidence. Public
  removal remains blocked before prompting or writing.
- The bounded review corrected strict update-source identity, exact removal
  expected-hash validation, and direct provenance-drift coverage. Twenty
  repeated removal lifecycle and race tests, full Go unit/race/vet, Windows
  compilation/build, all 55 legacy Node tests, 35 skill validations,
  temporary-home catalog validation, formatting, and diff checks passed
  without real-home mutation.
- Public project `apply` now includes reviewed removed-skill quarantine in its
  single mutation confirmation. Success and post-commit failure evidence report
  installed, updated, and quarantined removal counts separately without
  inferring state, while normal and recovery output expose only the validated
  quarantine handle and omit the absolute manifest path.
- Packaged-process tests prove default-declined removal is byte-preserving,
  confirmed removal retains the exact source/hash/bytes in a strict committed
  manifest, unrelated content survives, provenance retains only desired
  placements, modified former managed content blocks before prompting, and
  reruns create no new quarantine. Restore remains explicitly unavailable.
- The clean bounded review passed ten repeated removal and recovery-sensitive
  CLI runs, full Go unit/race/vet, Windows test compilation and CLI build, all
  55 legacy Node tests, 35 skill validations, temporary-home catalog
  validation, formatting, and diff checks without real-home mutation.
- The internal project restore transaction now accepts only exact lower-hex
  quarantine handles whose strict manifest is `committed`, preflights the whole
  run before its first move, refuses every occupied destination, and requires
  exact quarantine bytes plus action-specific current provenance. It moves old
  trees without replacement, commits restored provenance and durable
  `restored` manifest state, and proves idempotent reruns against the old bytes
  and source identity.
- Parent review closed two rollback gaps: partially created managed ancestors
  are now retained as transaction ownership even when preparation fails, and a
  rollback cannot return to `committed` until every quarantined tree, absent
  destination, provenance preimage, and original manifest is re-proven.
  Ambiguity preserves external bytes and records `recovery-required` whenever
  the manifest boundary remains safely writable.
- The clean bounded restore review passed twenty repeated restore lifecycle and
  race tests, full Go unit/race/vet, Windows test compilation and CLI build,
  all 55 legacy Node tests, 35 skill validations, temporary-home catalog
  validation, formatting, and diff checks without real-home mutation. At that
  internal-only checkpoint, public restore dispatch and interruption recovery
  remained deliberately unavailable.
- Public `sjskills restore <quarantine-id>` now validates one exact lower-hex
  recovery handle, discovers the canonical project without remote
  materialization, requires one human confirmation or JSON `--yes`, and maps
  the reviewed internal transaction into truthful path-free human/JSON
  evidence. Packaged-process coverage proves update and removal restore,
  idempotence, occupied-destination and provenance conflicts, nested-project
  discovery, and supplied-home isolation.
- The bounded public-restore review found no safe fixes or decisions. Fresh Go
  unit/race/vet, ten repeated restore-package runs, all 55 legacy Node safety
  tests, skill validation, formatting, and diff checks passed without real-home
  mutation. Interruption recovery remains the next independently reviewable
  project-reconciliation slice.
- Project apply and restore now acquire crash-releasing advisory locks on
  Darwin/Linux/BSD/Solaris, AIX, and Windows while retaining one recognizable
  private lock marker. A killed process no longer strands the project lock;
  malformed or actively held competitors still fail closed.
- A strict bounded private transaction journal commits before the first
  managed mutation and records exact preimage/candidate provenance plus
  path-relative placement/source/hash evidence. Restart rolls interrupted
  install, update, removal, and restore transactions back to their exact
  pre-state; semantic transition validation prevents unrelated journal state
  from claiming an unknown tree.
- Every staged tree, live placement, provenance file, quarantine boundary, and
  manifest is revalidated immediately before publication, movement, or
  replacement. Atomic operations reconcile landed-error outcomes, and raced
  external bytes are returned when possible or retained privately without
  overwrite or deletion.
- Abrupt subprocess fixtures cover install/update and restore death after
  provenance commit, plus death after durable quarantine-run and initial
  manifest staging. Ambiguous recovery preserves the journal and reports a
  path-free quarantine handle when one exists; a missing committed restore run
  blocks before any placement move.
- The independent recovery review closed Windows delete-sharing, AIX locking,
  journal-transition, journal-unlink, recovery-handle, pre-manifest crash, and
  provenance/tree move race gaps. Fresh Go unit/race/vet, Windows compile,
  AIX/Solaris build, 55 legacy Node safety tests, 35 skill validations, local
  catalog validation, formatting, and diff checks passed using temporary
  projects and homes only.
