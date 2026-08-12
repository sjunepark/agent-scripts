# Profile-aware cross-machine skill sync

Status: Implemented for the authorized phases 1-4 and goal-bounded phase 7/8
work. Phases 5, 6, and 9 remain deliberately deferred.

## Outcome

Make the repository the reproducible control plane for skills on two machines:

- The development machine installs the `common` and `dev` audiences.
- The KICPA machine installs the `common` and `kicpa` audiences.
- Codex, Claude Code, and Pi are assumed on both machines. Machine profiles do
  not configure agents independently in this iteration.
- Machine-global reconciliation is strict: the managed roots must equal the
  selected profile's desired state. Protected and unclassified entries must be
  reported without mutation; an unclassified entry in a managed root prevents
  the audit from passing.
- Project-scoped skills remain selectively installed by their target
  repositories and are never promoted into a machine-global profile.
- KICPA skill contents may stay in a private GitHub repository and install
  directly through authenticated Git access.

The workflow must be safe to run repeatedly, explain every discrepancy, and
make removal recoverable.

## Current state

Snapshot taken on 2026-08-05 from the development machine:

- `main` is clean and aligned with `origin/main` before this plan was added.
- `skill-registry.json` is the authoritative registry. It contains 37 skills:
  29 global recommendations and 8 project-scoped recommendations.
- `scripts/lib/skill-registry.js` understands `global`, `project`, and
  `catalog` scopes, but has one undifferentiated global baseline and no machine
  profile or audience model.
- `scripts/audit-global-skills` compares global skill names and apparent agent
  visibility. Its `--apply` mode installs missing entries; it does not enforce
  profile selection, exact filesystem placement, provenance, updates, or
  pruning.
- The current development-machine audit passes for the existing 29-skill
  baseline.
- The current Skills CLI supports global/project scope and agent targets, but
  not named machine profiles. Profile composition therefore belongs in this
  repository's registry and reconciliation code.
- Codex discovers user skills from `~/.agents/skills`. The installed Pi variant
  also documents both `~/.agents/skills` and `~/.pi/agent/skills`. Claude Code
  uses `~/.claude/skills`.
- The development machine currently has the 29 expected registry skills in
  `~/.agents/skills`, 27 duplicate registry skills in `~/.claude/skills`, and
  21 duplicate registry skills in `~/.pi/agent/skills`. Older all-agent
  installs left additional duplicate copies in many agent-specific roots.
- `~/.agents/.skill-lock.json` records source URL, ref, source subpath, content
  hash, and timestamps, but not all target directories. It is useful
  provenance evidence, not a complete desired-state database.
- Unclassified entries include `planning-doc-triad`, `find-docs`, and the
  `~/.codex/skills/gitlog-html` symlink. They must be classified or preserved,
  not silently removed.
- Runtime-owned locations also exist, including Codex system/vendor/cache and
  backup directories, Pi remote skills, and editor-extension skills. A broad
  filesystem search followed by deletion would be unsafe.
- The tracked root `skills-lock.json` contains a local absolute source path for
  `doc-comment-writer`. Its purpose and portability need resolution.
- GitHub CLI authentication on the development machine currently has private
  repository access and a Git credential helper configured. The KICPA machine
  must be bootstrapped and verified separately.
- The 2026-08-05 schema slice migrated the registry to version 2. The `dev`
  profile resolves 29 entries (8 `common`, 21 `dev`); the current `kicpa`
  profile resolves the 8 `common` entries because no private KICPA entries
  exist yet.
- Dependency-free registry tests cover valid profile unions, explicit profile
  selection, version rejection, profile shape/composition, and invalid skill
  audience combinations.
- `scripts/audit-global-skills` now requires `--profile dev|kicpa`. Its
  registry-v2 mode is deliberately read-only, and `--apply` fails closed until
  exact filesystem reconciliation replaces the old aggregate-agent logic.
- Core command references, the published `skills-cli` guidance, and the
  repo-local `sync` guidance now pass an explicit profile and do not advertise
  the disabled version 1 apply path. The full layout/install/prune rewrite
  remains phase 7 work after exact reconciliation exists.
- Development baseline evidence was captured read-only: Node 24.15.0, Skills
  CLI 1.3.13, GitHub CLI 2.96.0, authenticated GitHub `repo` access, a passing
  29-skill pre-migration audit, and path/type/content-hash inventory for the
  three known global roots. The recovery inventory is retained machine-locally
  at `.tmp/profile-aware-skill-sync-baseline-20260805.txt` (gitignored); no
  token or skill contents were recorded.
- Final slice validation passed: 9 dependency-free registry tests,
  `scripts/validate-skills`, JavaScript syntax checks, `git diff --check`,
  `bunx skills add ./skills --list`, and the real-machine read-only `dev`
  aggregate audit. The `kicpa` audit correctly reports the 21 development-only
  skills on this machine as drift. A bounded `code-review` pass found no
  remaining safe fixes or decisions after correcting the global-manual-record
  documentation.
- The 2026-08-06 exact-state slice added a fixture-only filesystem inventory,
  deterministic tree hashing, strict issue classification, and a non-executing
  operation planner in `scripts/lib/global-skill-state.js`. Existing roots and
  lock files are read only after their canonical paths are proven to stay
  inside an explicitly injected home; lock provenance is root-local and never
  authorizes a different managed root.
- Exact-state fixtures cover canonical and Claude copies, legacy duplicates,
  protected roots, symlinks, path escapes, missing/outdated/modified and
  ambiguous content, manager boundaries, and quarantine/restore plans. The
  combined dependency-free suite passes 22 tests. The model was never invoked
  against real machine skill roots, no real-home state was mutated, and the
  audit/apply/prune command remains unchanged.
- The 2026-08-06 reconciler slice replaced aggregate discovery with a strict
  exact-root command. Remote desired content is materialized in a temporary
  home; apply uses explicit Codex and Claude targets without a Pi copy; prune
  revalidates verified duplicates and canonical copies before moving them into
  a manifest-backed quarantine; restore is path-confined and refuses overwrite.
- Apply now adopts verified exact copies into a reconciler-owned state file and
  uses those root-local hashes to distinguish safe upstream updates from local
  modifications. Skills CLI v3 folder hashes are not treated as local-content
  evidence. Every apply, prune, restore, and manifest mutation validates the
  canonical nearest existing ancestor, including missing paths below symlinks.
- Dependency-free coverage now exercises both public profiles, remote command
  synthesis, verified-state adoption, v3 lock handling, apply preconditions,
  protected/manual state, symlink-ancestor escapes, prune races, crash-
  recoverable manifests, and restoration in temporary homes only. The real
  development-home audit remained read-only and truthfully reported protected
  state, verified Pi duplicates, and ambiguous copies without mutation.
- Final goal-bounded validation passes 49 dependency-free tests,
  `scripts/validate-skills`, JavaScript syntax checks, `git diff --check`, and
  `bunx skills add ./skills --list`. The final real-home `dev` audit was
  read-only and failed strictly on existing drift as intended. Bounded design
  and safety follow-up reviews report no remaining material findings after a
  partial replacement-failure recovery fixture and operator guidance were
  added.
- PR #6 review hardening materializes each desired remote skill once and
  installs that exact verified staged snapshot, quarantines verified old trees
  before update, hashes executable bits, binds prune and replacement approval
  to exact state, bounds Skills CLI subprocesses, and verifies publication by
  skill-tree equality across merge, squash, or rebase workflows. Fixture tests
  cover staged tampering, update restoration, moving replacement content, and
  candidate-set races.
- The 2026-08-13 registry v3 refactor replaced client-compatibility
  classification with the two actual installation destinations: `.agents` and
  `.claude`. Client names now appear only inside the Skills CLI adapter used to
  materialize those destinations.

## Accepted decisions

### Profile and registry model

- Use two explicit machine profiles: `dev` and `kicpa`.
- Compose profiles from per-skill audiences so skill lists are declared once:
  `dev = common + dev`, and `kicpa = common + kicpa`.
- Require `--profile dev` or `--profile kicpa` for reconciliation. Do not infer
  a profile from hostname or silently choose a default.
- The registry contract is version 3. Version 2 introduced profile-aware
  validation; version 3 replaces agent compatibility with installation targets.
- Every global recommendation must declare exactly one audience from
  `common`, `dev`, or `kicpa`.
- Project and catalog entries must not carry a machine audience.
- Use `recommendation.targets` to declare `.agents`, `.claude`, or both. Do not
  classify skills by client when reconciliation manages destinations.
- Do not recreate the former profile design that duplicated the same skill
  name under source/agent buckets.

The intended registry shape is:

```json
{
  "version": 3,
  "global": {
    "allowUnlistedSkills": false,
    "profiles": {
      "dev": { "audiences": ["common", "dev"] },
      "kicpa": { "audiences": ["common", "kicpa"] }
    }
  },
  "skills": [
    {
      "recommendation": {
        "targets": [".agents", ".claude"]
      }
    }
  ]
}
```

The implementation may adjust field placement to fit the existing registry,
but it must preserve these semantics and avoid duplicated profile lists.

### Installation layout

- Map the `.agents` target to the canonical `~/.agents/skills` managed root.
- Map the `.claude` target to `~/.claude/skills`.
- Do not maintain registry-managed duplicates in `~/.pi/agent/skills` or other
  legacy agent roots.
- Populate the `.agents` destination through the explicit Codex command target.
  Do not pass `-a pi`, because that recreates a Pi copy.
- Accept that a skill in the shared root is visible to both Codex and Pi.
  Independent Codex and Pi subsets are intentionally out of scope.
- Never use Skills CLI `--all` for these installs; it expands both skills and
  agents and can repopulate legacy roots.
- Verify Pi discovery on the KICPA machine before pruning its Pi-specific
  copies; do not assume its installed Pi distribution matches development.

### Desired-state behavior

- A read-only audit is the default and must exit nonzero for missing, outdated,
  misplaced, or unexpected managed entries.
- `--apply` installs or updates desired entries only. It does not delete.
- `--prune` is a separate, explicit operation. It moves only verified stale
  registry duplicates into a timestamped quarantine and reports how to restore
  them.
- Modified, ambiguous, or unknown entries fail the strict audit as
  `unclassified`; they are not auto-deleted.
- Protect system, vendor, cache, backup, remote-runtime, and editor-extension
  locations. Never remove an entire skills root merely because some children
  are stale.
- Compare actual entries in managed roots. Do not use the Skills CLI's
  aggregated agent labels as proof of exact placement.

### Public and private sources

- Keep the public repository as the profile and workflow control plane.
- Keep KICPA skill contents private unless there is an independent reason to
  publish them.
- Prefer direct authenticated Git installation from the private repository.
  Use GitHub CLI to establish Git credentials (`gh auth login`, then
  `gh auth setup-git`) and verify access with `git ls-remote` before invoking
  the Skills CLI.
- Do not add a `gh clone` plus raw-copy pipeline unless direct Git installation
  proves unreliable. A local clone may be used for development, not as the
  ongoing machine install source.
- It is acceptable for the public registry to reveal the private repository
  URL and skill names. Stop and choose a private registry overlay instead if
  that metadata must also be confidential.
- Add the private source to the live registry only after the repository and at
  least one real KICPA skill exist. Use `sjunepark/kicpa-skills` only as a
  placeholder until confirmed.
- Reserve
  `https://github.com/sjunepark/kicpa-skills/tree/main/skills` as the intended
  private catalog URL. It remains a placeholder and must not enter the live
  registry until the repository and a real KICPA skill are confirmed.
- Creating or changing the private repository is an external operation and is
  outside the first implementation slice.

## Proposed initial audience classification

Review this mapping at the start of implementation, then encode it once in the
registry.

### Common

- `brainstorming`
- `clarify`
- `distill-response`
- `end-state-review`
- `interview`
- `next-goal`
- `progress`
- `skills-cli`

### Development

- `address-pr-feedback`
- `agents-md-writer`
- `architecture-md-writer`
- `baton`
- `change-explainer`
- `code-review`
- `codebase-design`
- `context7-cli`
- `context7-mcp`
- `create-pr`
- `delegate-ui-to-claude`
- `develop-skills`
- `doc-comment-writer`
- `domain-modeling`
- `explore-repo`
- `harmonize-docs`
- `improve-codebase-architecture`
- `merge-branch`
- `project-status`
- `skill-cleaner`
- `teach`

### KICPA

No public-registry skills are assigned yet. Add private entries after their
source repository exists.

### Project-scoped and unchanged

- `clear-rust`
- `impeccable`
- `modern-go`
- `modern-rust`
- `release-please-release`
- `review-campaign`
- `ui-lab`
- `write-go-docs`

## Implementation plan

### 1. Freeze inputs and capture a safe baseline

- [x] Re-read this plan and the root `AGENTS.md` before changing code.
- [x] Confirm the proposed audience mapping.
- [x] Record the accepted privacy boundary: source URL and skill names may be
      public, while repository contents remain private. Reopen this only if the
      confidentiality requirement has changed.
- [x] Record `git status`, current tool versions, the current registry audit,
      and a read-only inventory of known global roots.
- [x] Export or otherwise record a recovery inventory of existing skill paths,
      symlink targets, and hashes without copying secret contents into this
      public repository.
- [x] Verify development-machine private Git access without printing tokens.
- [x] Leave KICPA-machine bootstrap and all pruning for the rollout phase.

Exit condition: the starting state is reproducible and unknown entries have a
recorded disposition of `preserve pending classification`.

### 2. Add the version 2 profile contract

Primary files:

- `skill-registry.json`
- `scripts/lib/skill-registry.js`

Work:

- [x] Add dependency-free fixtures or tests for valid and invalid version 2
      registries before migrating the live registry.
- [x] Validate supported profile names, audience names, nonempty profile
      composition, sorted/deduplicated values, and resolvable references.
- [x] Require exactly one audience on every global entry.
- [x] Reject audiences on project or catalog entries.
- [x] Add a resolver such as `globalSkillEntries(registry, profile)` that
      returns the union of the profile's audiences without duplicating skill
      declarations.
- [x] Require callers to provide a known profile; remove the implicit global
      baseline path.
- [x] Migrate all current registry entries using the reviewed classification.
- [x] Keep source managers, install modes, project conditions, and existing
      client compatibility unchanged unless a failing test demonstrates a
      necessary correction.

Exit condition: both profiles resolve deterministically, schema errors are
actionable, and every current entry is accounted for exactly once.

### 3. Model exact filesystem state separately from the CLI

Prefer a focused pure module, for example
`scripts/lib/global-skill-state.js`, rather than expanding the command into a
large collection of coupled branches.

- [x] Map each selected registry entry to its expected managed roots from its
      `.agents` and `.claude` targets.
- [x] Enumerate children only in explicit managed and known legacy roots.
- [x] Read lock-file provenance when available: source URL, ref, skill path,
      and recorded content hash.
- [x] Calculate current hashes in a stable way compatible with the lock format,
      or report provenance as unverifiable if compatibility cannot be proven.
- [x] Classify issues as at least `missing`, `outdated`, `misplaced`,
      `unexpected-managed`, `verified-legacy-duplicate`, `modified`,
      `unclassified`, and `protected`.
- [x] Detect symlinks explicitly and resolve targets read-only before deciding
      whether two entries are equivalent.
- [x] Generate an operation plan independently from executing it so dry-run,
      tests, apply, and prune share one decision path.
- [x] Ensure manual and workflow-managed registry entries keep their manager
      boundaries and are not silently converted to Skills CLI ownership.

Exit condition: fixture inventories produce a complete, deterministic report
without reading or mutating the real home directory.

### 4. Make the profile-aware global audit exact-state and strict

Primary command: `scripts/audit-global-skills`.

- [x] Require `--profile dev|kicpa` and provide useful help text.
- [x] Keep the default operation read-only and fail closed on `--apply` until
      exact-state reconciliation is implemented.
- [x] Report the selected audiences, expected skills by managed root, detected
      discrepancies, protected entries, and proposed operations.
- [x] Make strict discrepancies exit nonzero, including unexpected entries in
      managed roots and unclassified legacy collisions.
- [x] Update `--apply` to install/update only from remote sources and only into
      expected roots.
- [x] For the `.agents` destination, synthesize an explicit Codex command
      target; never synthesize a Pi command target.
- [x] For the `.claude` destination, synthesize the explicit Claude Code
      command target.
- [x] Preserve the rule that changed public skills are committed and pushed
      before reinstalling from the GitHub `skills/` subpath.
- [x] Add a separate `--prune` confirmation boundary that quarantines only
      verified legacy duplicates. Refuse to combine ambiguous cleanup with an
      ordinary apply.
- [x] Print the quarantine location and a concrete restore procedure.
- [x] Keep this as a repository maintenance script initially. Add a stable
      `bin/` wrapper only if a demonstrated cross-repository workflow needs it.

Exit condition: the command can explain and reconcile either profile while
all destructive behavior remains explicit and recoverable.

### 5. Support the private KICPA source

- [ ] Confirm or create the private GitHub repository separately, with
      `skills/<skill-name>/SKILL.md` as its distributable layout.
- [ ] On each machine, authenticate GitHub CLI and configure Git credentials.
- [ ] Verify the private remote with `git ls-remote <private-url> HEAD`.
- [ ] Verify discovery with
      `bunx skills add <private-skills-url> --list` before changing desired
      state.
- [ ] Add the external source and concrete KICPA skill entries to the registry,
      tagged with the `kicpa` audience.
- [ ] Ensure commands do not embed credentials, resolved tokens, or private
      file contents in logs, lock files committed to this public repository, or
      error output.
- [ ] Test a missing/expired credential failure and make the remediation point
      to GitHub authentication rather than suggesting public access.

Exit condition: an authenticated KICPA profile can discover and install the
private skill while an unauthenticated request cannot read its contents.

### 6. Inventory and migrate legacy roots

- [ ] Run strict read-only audits before any cleanup.
- [ ] Compare registry-named children in known legacy agent roots against the
      canonical managed copies using symlink and content evidence.
- [ ] Never scan every directory named `skills` as a deletion candidate.
- [ ] Explicitly protect Codex `.system`, vendor imports, caches, backups, Pi
      remote skills, editor extensions, and other runtime-owned trees.
- [ ] Classify `planning-doc-triad`, `find-docs`, and `gitlog-html` as managed,
      intentionally external, or removable before proceeding.
- [ ] Decide whether the root `skills-lock.json` is intentional project state.
      Remove it, replace its local absolute source, or document its owner based
      on evidence; do not leave a nonportable path unexplained.
- [ ] Smoke-test skill discovery in Codex, Claude Code, and Pi before pruning.
- [ ] Quarantine verified duplicates child-by-child, then repeat discovery and
      strict audit.
- [ ] Retain quarantine until both machines have completed a normal work cycle.

Exit condition: only desired managed copies remain in the designated roots;
all other entries are either protected, intentionally external, or explicitly
classified.

### 7. Update the operator documentation and skills

Update behavior and documentation together:

- [x] `docs/skill-registry.md`: version 2 schema, audience/profile semantics,
      managers, and validation errors.
- [x] `README.md`: concise two-machine workflow and pointers to detailed docs.
- [x] `docs/settings-sync.md`: machine bootstrap, authentication boundary, and
      what settings sync does and does not own.
- [x] `AGENTS.md`: only the durable current commands and invariants; do not turn
      it into a migration log.
- [x] `.agents/skills/sync/SKILL.md`: explicit profile, publish-before-install,
      public-source publication verification, and safe reconciliation order.
      Private-source authentication remains deferred to excluded phase 5.
- [x] `skills/skills-cli/SKILL.md`: strict profile workflow where generally
      useful.
- [x] `skills/skills-cli/references/cli.md`: verified CLI flags and shared-root
      behavior.
- [x] `skills/skills-cli/recipes/install-and-migrate.md`: exact install,
      audit, apply, prune, quarantine, and rollback commands.
- [x] Review any other install-management recipe referenced by the skill and
      update it only if the behavior changed.

Exit condition: a fresh session can bootstrap either machine without relying
on this plan's historical notes.

### 8. Validate the implementation

The repository has no root package manifest, CI workflow, or general test
suite. Keep new tests dependency-free and runnable directly with Node; do not
add package tooling solely for this feature.

Minimum automated cases:

- [x] Valid `dev` and `kicpa` profile resolution.
- [x] Missing, unknown, duplicate, and scope-invalid audiences.
- [x] Canonical `.agents` and `.claude` target placement.
- [x] No Pi-specific install command is generated.
- [x] Strict missing, extra, misplaced, outdated, modified, and unclassified
      states.
- [x] Protected-root exclusion.
- [x] Manual/workflow manager preservation.
- [x] Credential-free public remote source command synthesis and rejection of
      credential-bearing or unsupported source forms. Private KICPA
      discovery/authentication remains excluded.
- [x] Hash mismatch and unverifiable provenance handling.
- [x] First-run stale-copy quarantine and verified remote replacement using a
      temporary home fixture only.
- [x] Quarantine and restore plans using a temporary home fixture only.

Repository validation:

- [x] Run the new dependency-free test command.
- [x] Run `scripts/validate-skills` after the final review fixes.
- [x] Run `bunx skills add ./skills --list` after the final review fixes.
- [x] Run strict dry-run audits for both profiles against fixtures.
- [x] Run the development profile against the real machine read-only after the
      final review fixes.
- [x] Run `$code-review` with the diet lens, apply safe findings, and repeat
      validation.

Exit condition: tests never mutate the real home directory and all validation
commands pass.

### 9. Roll out in controlled order

Development machine:

- [ ] Finish implementation and read-only validation without pruning.
- [ ] Commit and push public repository changes.
- [ ] Reinstall changed public skills only from the published remote URL.
- [ ] Run `--profile dev` dry-run, then `--apply`, then all three client smoke
      tests.
- [ ] Review the prune plan, quarantine verified duplicates, and rerun the
      strict audit and smoke tests.

KICPA machine:

- [ ] Configure GitHub CLI authentication and Git credentials.
- [ ] Clone or update this control repository.
- [ ] Verify both public and private sources read-only.
- [ ] Run `--profile kicpa` dry-run, then `--apply`.
- [ ] Smoke-test Codex, Claude Code, and the installed Pi distribution.
- [ ] Review and execute the separate prune/quarantine step only after the
      shared-root Pi test passes.
- [ ] Rerun strict audit and retain rollback instructions locally.

Exit condition: both machines pass their profile audit and can invoke a sample
common skill plus a machine-specific skill in all intended clients.

## Guardrails and likely failure modes

- A strict audit and an automatic deleter are not the same feature. Keep
  detection strict and deletion conservative.
- Agent labels reported by `bunx skills list -g` may reflect shared discovery;
  inspect roots directly before declaring a duplicate.
- Private Git support depends on the underlying Git authentication environment.
  Never pass tokens in source URLs or command arguments.
- The KICPA Pi installation may differ from development. Its shared-root smoke
  test is a hard prerequisite for removing `~/.pi/agent/skills` copies.
- Lock data can be stale or incomplete. A lock entry alone does not authorize
  deletion, and a missing lock entry does not prove an installed skill is
  foreign.
- Preserve unrelated user changes and unmanaged skills throughout migration.
- Quarantine destinations must be outside active discovery roots, have an
  explicit manifest, and avoid overwriting an earlier backup.
- Do not create compatibility layers for older registry versions unless a
  concrete external consumer is found. Prefer one clear version 3 path.

## Completion criteria

- [x] `dev` resolves to exactly `common + dev`; `kicpa` resolves to exactly
      `common + kicpa`.
- [x] Project-scoped entries cannot enter global desired state.
- [ ] Every selected `.agents` target has one canonical copy in `~/.agents/skills`.
- [ ] Every selected `.claude` target has the intended copy in `~/.claude/skills`.
- [ ] No registry-managed copy remains in `~/.pi/agent/skills` or a legacy root
      after successful, verified migration.
- [x] Unknown and runtime-owned entries are classified or preserved rather
      than deleted.
- [ ] Strict audits pass on both machines.
- [ ] Private KICPA contents are unavailable without GitHub authorization.
- [x] Tests, validators, documentation, and operator skills describe the same
      workflow.
- [ ] Quarantine rollback has been tested or rehearsed and retained through a
      normal work cycle on both machines.

## Next action

No action remains within the completed goal. Phases 5, 6, and 9 require a new
authorization boundary: private KICPA support, real-machine apply/prune,
legacy-root migration, and rollout remain deliberately unstarted.
