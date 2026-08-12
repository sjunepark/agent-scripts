---
name: sync
description: "Orient and operate this repository's skill publication and machine-reconciliation workflow. Use when the user asks how repo skills move from local edits to published remote sources, asks to audit, bootstrap, sync, or reinstall a dev or kicpa machine profile, or asks to handle findings from that profile audit. Do not use for plugin deployment or local catalog validation alone."
---

# Sync

Move intended skill state through three distinct layers without treating them as interchangeable:

1. `skills/` is the editable, distributable catalog in this checkout.
2. The registry's remote source ref is the published catalog.
3. One explicitly selected `dev` or `kicpa` profile is the desired exact state on a machine.

Local validation proves that a checkout is usable as a source; it does not publish or install it. A pushed feature branch is not published when the registry points to `main`. A published catalog contains more skills than either global profile.

## Choose the operation

Classify the request before changing anything:

- **Explain or inspect:** Explain the three layers, inspect relevant files, and run only the read-only profile audit when a profile is known. Do not commit, push, apply, replace, prune, or restore.
- **Validate local work:** Validate `./skills` or one `./skills/<name>` without installing it for ongoing use. This is not a sync.
- **Reconcile published state:** When the requested profile is already published, run the read-only audit, apply ordinary safe changes, and audit again. Do not create a commit merely to reconcile a machine.
- **Publish and reconcile:** Validate intended local changes, send them through the repository's normal review and merge flow, verify the published trees, then reconcile the selected profile.
- **Recover:** Restore only the manifest the user identifies, after checking that every destination is absent.

Choose `dev` for the public `common + dev` baseline and `kicpa` for the public `common + kicpa` baseline. Any explicit `dev` or `kicpa` mention in the request selects that profile; do not ask the user to repeat it. Never infer a profile from whatever happens to be installed. Ask the user to choose only when a mutating request names neither profile.

## Authority boundaries

- An explanation, status check, or audit request is read-only.
- Commit and push only when the user asked to publish or sync relevant local changes. Stage only in-scope files and preserve unrelated working-tree state.
- Run `--apply` only when the user asked to bootstrap, install, reconcile, or sync the selected machine profile.
- Ordinary apply never authorizes first-run replacement or legacy pruning. Show the exact candidates and printed restore procedure, then wait for explicit approval before copying either digest-bound command.
- Restore only on explicit request. It refuses to overwrite an active path; move an active replacement aside only with the user's authority.
- Never use a local checkout as the ongoing machine install source.

## Workflow

### 1. Establish source, profile, and current state

Run:

```bash
git remote get-url origin
git status --short
scripts/audit-global-skills --profile "$PROFILE"
```

Confirm that `origin` corresponds to the source recorded in `skill-registry.json`. A completed report ending in `Strict result: fail` means drift. An error without that strict report is an operational failure even though both cases exit nonzero; report materialization, network, or validation errors instead of relabeling them as drift. The first audit is always read-only.

If there are no relevant local changes and the requested state is already published, skip the validation and publication steps but still verify the registry's remote catalog before reconciliation.

### 2. Validate changes that must be published

Inspect relevant diffs and validate the smallest affected surface:

```bash
bunx skills add ./skills --list
scripts/validate-skills
```

Use `bunx skills add ./skills/<skill-name> --list` for a focused skill check. If the registry or reconciler changed, also run:

```bash
node --test scripts/lib/skill-registry.test.js scripts/lib/global-skill-state.test.js scripts/audit-global-skills.test.js
```

Do not use `.`, `./skills`, or the working tree for an ongoing machine install.

### 3. Publish before machine reconciliation

Stage only intended files, create a reviewable commit, and push it:

```bash
git add <intended-paths>
git commit
git push
```

If the registry source is pinned to `main`, complete the normal review and merge flow before continuing. After the last review-driven change is pushed, record the final PR head as `INTENDED_COMMIT`; refresh it after every later change. Once merged, fetch the pinned ref and verify each changed published skill by tree equality:

```bash
INTENDED_COMMIT=$(git rev-parse HEAD)
git fetch origin main
git diff --quiet "$INTENDED_COMMIT:skills/$SKILL_NAME" "origin/main:skills/$SKILL_NAME"
```

This check remains valid after merge, squash, or rebase. Stop if any intended skill tree differs. A feature-branch push alone does not authorize installing from a `main` source.

### 4. Verify the remote catalog

Resolve this repository's published catalog source from the registry and inspect that exact URL before apply:

```bash
SKILLS_URL=$(node -e 'const r=require("./skill-registry.json"); const url=r.sources?.["agent-scripts"]?.location; if (!url) process.exit(1); process.stdout.write(url)')
bunx skills add "$SKILLS_URL" --list
```

Confirm that the remote exposes the published catalog. Do not interpret the full catalog as the selected machine profile.

### 5. Reconcile the selected profile

Run ordinary apply, then audit the resulting exact state:

```bash
scripts/audit-global-skills --profile "$PROFILE" --apply
scripts/audit-global-skills --profile "$PROFILE"
```

Apply uses only registry-approved remote sources. It installs selected Codex/Pi-compatible skills through the shared Codex target, selected Claude-compatible skills through the Claude target, and creates no Pi-specific copy. It adopts already-exact copies, updates only verified unchanged copies, quarantines the prior verified tree before an update, and blocks locally modified or unverified stale copies.

Do not hand-build global commands or use `--all` when this registry-backed reconciler applies. Project, workflow, catalog, and manual records remain with their declared policy or manager.

### 6. Handle gated follow-up separately

If the audit reports an unverified stale copy, present the candidates, digest, quarantine behavior, and restore command. Only after explicit approval, copy the exact command printed by that audit:

```bash
scripts/audit-global-skills --profile "$PROFILE" --replace-unverified 'sha256:PRINTED_DIGEST' --yes
```

Audit again. A changed candidate set or verified replacement snapshot invalidates the digest.

If the audit reports verified legacy duplicates, present the candidates, digest, quarantine behavior, and restore command. Only after separate explicit approval, copy the exact printed command:

```bash
scripts/audit-global-skills --profile "$PROFILE" --prune 'sha256:PRINTED_DIGEST' --yes
```

Audit again and retain the generated manifest through a normal work cycle. Prune moves reviewed duplicates into quarantine; it does not delete them. A changed candidate set invalidates the digest.

For an explicitly requested rollback, use only the manifest-specific command produced by the reconciler:

```bash
scripts/audit-global-skills --restore '/absolute/path/to/manifest.json' --home '/absolute/home' --yes
```

Inspect every modeled destination first. Restoration must stop rather than overwrite an existing destination.

## Completion

Finish with:

- the selected profile and whether the operation was read-only or mutating;
- the validation results and published ref verified, when publication occurred;
- the final audit result, including any remaining drift;
- every quarantine manifest or restore command created; and
- any replacement or prune plan left pending approval.

Do not claim exact state while the final audit still reports drift. Do not call a blocked local modification repaired merely because ordinary apply completed.
