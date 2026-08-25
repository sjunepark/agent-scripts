---
name: sync
description: "Orient and operate this repository's skill publication and sjskills reconciliation workflow. Use for explaining catalog-to-published flow, publishing skill changes, project manifest reconciliation, or inspection and explicitly authorized rollout of the fixed global baseline. Do not use for plugin deployment or local catalog validation alone."
---

# Sync

Keep these layers distinct:

1. `skills/` is the editable, distributable catalog in this checkout.
2. The `agent-scripts` source in `skill-registry.json` identifies the published
   catalog consumed by reconciliation.
3. A project's committed `sjskills.toml` selects project profiles and direct
   third-party skills.
4. `global.baseline` in `skill-registry.json` is the one machine-independent
   global selection.

Local validation does not publish or install ongoing state. A pushed feature
branch is not published while the registry source points to `main`. Project
profiles such as `dev`, `go`, `rust`, and `kicpa` do not make skills global.

## Classify the request

- **Explain or inspect:** Read the relevant registry, manifest, and plans. Run
  `bin/sjskills plan` in a project or `bin/sjskills plan --global` for the fixed
  baseline. Do not mutate managed roots.
- **Validate local work:** Validate `./skills` or one local skill without
  installing it for ongoing use. This is not publication or reconciliation.
- **Reconcile a project:** Require an existing or explicitly requested
  `sjskills.toml`, review `bin/sjskills plan`, then apply only when the user
  asked to install, bootstrap, reconcile, or sync that project.
- **Publish and reconcile:** Validate intended catalog or registry changes,
  send them through the repository's normal commit, review, and merge flow,
  verify the published trees, then reconcile the requested project.
- **Inspect or roll out the global baseline:** Planning is read-only. Real-home
  apply or restore requires the separate evidence-bound procedure in
  `plans/sjskills-global-rollout.md` and explicit authorization for the exact
  reviewed plan.
- **Recover:** Restore only the quarantine identifier named by the user, after
  confirming that all destinations are absent.

Do not infer a project profile from installed copies. For a new manifest, use
only profiles the user selected or ask when that choice materially changes the
result. The global baseline has no profile selection or hostname inference.

## Authority boundaries

- Explanation, status, audit, and plan requests are read-only.
- Commit and push only when the user asked to publish or sync relevant local
  changes. Stage only in-scope files and preserve unrelated worktree state.
- Run project `apply` only for an explicitly requested reconciliation.
- Do not run global `apply` or restore from ordinary repository-validation or
  sync authority. Follow the reviewed rollout plan and require its stated
  authorization evidence.
- Restore only on explicit request. It refuses to overwrite active content.
- Never use a local checkout as the ongoing install source.

## Workflow

### 1. Establish source, scope, and current state

Run the smallest relevant read-only checks:

```bash
git remote get-url origin
git status --short
bin/sjskills plan
bin/sjskills plan --global
```

Use the project command only where `sjskills.toml` exists. Confirm that
`origin` corresponds to the repository source in `skill-registry.json`.
Planning may materialize remote content in temporary storage, but it does not
change managed roots. Report blocked placements, unsafe boundaries, malformed
provenance, and operational failures as such rather than calling every
non-success result drift.

`scripts/audit-global-skills` is only a compatibility wrapper for the read-only
global plan. Its former profile and mutation flags are retired.

### 2. Validate work that must be published

Inspect the relevant diff and validate the smallest affected surface:

```bash
bunx skills add "./skills/$SKILL_NAME" --list
scripts/validate-skills
```

Use `bunx skills add ./skills --list` when catalog-wide validation is warranted.
If the registry or reconciler changed, also run:

```bash
node --test scripts/lib/skill-registry.test.js scripts/audit-global-skills.test.js
go test ./...
```

These local checks prove that the checkout is usable as a source; they do not
publish it or authorize managed-root changes.

### 3. Publish before remote-backed reconciliation

Stage only intended files, commit with the repository's normal review flow,
and push. If the registry source is pinned to `main`, complete the merge before
continuing. Fetch that ref and verify each intended published skill tree:

```bash
git fetch origin main
git diff --quiet "$INTENDED_COMMIT:skills/$SKILL_NAME" \
  "origin/main:skills/$SKILL_NAME"
```

Set `SKILL_NAME` to each changed skill and `INTENDED_COMMIT` to the last
reviewed commit, refreshing both after later changes. Tree equality remains valid
after merge, squash, or rebase. Stop when a published tree differs; a feature
branch push alone does not make a `main`-backed source current.

Resolve and inspect the exact published catalog recorded in the registry:

```bash
SKILLS_URL=$(node -e 'const r=require("./skill-registry.json"); const u=r.sources?.["agent-scripts"]?.location; if (!u) process.exit(1); process.stdout.write(u)')
bunx skills add "$SKILLS_URL" --list
```

Do not treat the full published catalog as either a project's selected set or
the fixed global baseline.

### 4. Reconcile the requested scope

For a project, review the plan and then apply:

```bash
bin/sjskills plan
bin/sjskills apply
bin/sjskills plan
```

For the fixed global baseline, ordinary work stops after
`bin/sjskills plan --global`. Proceed to `apply --global` only through
`plans/sjskills-global-rollout.md`, using its exact commit, plan digest,
recheck, authorization, and recovery requirements.

`sjskills` uses only registry-approved remote sources for central declarations,
places supported targets in `.agents/skills` and `.claude/skills`, and creates
no Pi-specific copy. Do not substitute direct Skills CLI installs, `--all`,
manual root copying, or the retired audit-wrapper mutation flags.

Retain every quarantine identifier through a normal work cycle. Restore only
the requested identifier, and only while every modeled destination is absent:

```bash
bin/sjskills restore <quarantine-id>
bin/sjskills restore --global <quarantine-id>
```

The second command remains subject to the global rollout authorization
boundary.

## Completion

Report:

- project or fixed-global scope, and whether the operation was read-only or
  mutating;
- validation and published-ref evidence when publication occurred;
- the final plan result and any preserved unknown, legacy, or blocked state;
- every quarantine identifier retained or restored; and
- any global rollout or recovery action still awaiting authorization.

Do not claim exact state while the final plan still contains changes or blocks.
