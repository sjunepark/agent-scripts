---
name: sync
description: "Sync repo skills onto this machine — commit and push intended skill changes, then reinstall selected published skills with the skills CLI from the repository's remote skills/ URL. Use when the user wants to sync, bootstrap, or reinstall repo skills from GitHub or another git remote, not ./skills or the working tree."
---

# Sync

Use this skill to publish the intended local skill changes, then install selected published skills from the current repository's remote `skills/` subpath onto the current machine.

Always install with the skills CLI from the published source — the GitHub `skills/` subpath — never `.`, `./skills`, or any filesystem path.

## Workflow

1. Confirm the repository remote URL.
- Run `git remote get-url origin`.
- If `origin` is missing, stop and tell the user you need a remote URL for this repo.
- For this repository, convert the remote into the GitHub `skills/` subpath: `https://github.com/sjunepark/agent-scripts/tree/main/skills`.
- If the remote points somewhere unexpected, show it to the user before committing, pushing, or installing.

2. Inspect the working tree and decide what to publish.
- Run `git status --short` and inspect relevant diffs before committing.
- Commit only the changes that belong to this sync request. Do not sweep in unrelated dirty files.
- When `skill-registry.json` exists, treat it as the authoritative exact-state install policy. Select the machine profile explicitly and run `scripts/audit-global-skills --profile <dev|kicpa>` before changing installs. Do not remove or prune anything during this first pass.
- If the commit scope or message is unclear, ask the user before committing.
- If there are no relevant local changes, skip the commit and push steps and continue with the remote reinstall/update.

3. Commit the intended changes.
- Stage the intended files explicitly with `git add <paths>`.
- Do not commit generated install output or local-only caches.

4. Push the commit so the remote contains the intended version.
- Record the intended commit as `INTENDED_COMMIT=$(git rev-parse HEAD)`.
- Run `git push` after committing.
- If the push fails, stop and report the failure. Do not install from a remote that lacks the intended changes.
- If the registry source is pinned to another ref such as `main`, merge the
  intended commit through the normal review flow, then run `git fetch origin
  main` and `git merge-base --is-ancestor "$INTENDED_COMMIT" origin/main`.
  Stop if that check fails; a pushed feature branch is not published at the
  registry ref.

5. Inspect the remote skill source before installing.
- Run `bunx skills add "<skills-subpath-url>" --list`.
- Confirm the remote resolves and exposes only the published skills from `skills/`.

6. Reconcile the selected profile from remote sources.
- Treat this repo's `skills/` directory as a catalog, not as a global install manifest.
- When `skill-registry.json` exists, run `scripts/audit-global-skills --profile <dev|kicpa> --apply`. The reconciler installs or updates only unambiguous remote Skills CLI entries, maps Codex/Pi compatibility to the shared root through an explicit Codex target, maps Claude compatibility through an explicit Claude Code target, and never creates a Pi-specific copy.
- Apply adopts already-exact Skills CLI copies into its verified state file.
  Later updates are automatic only while the installed tree still matches that
  record; local edits remain blocked.
- If the read-only audit proposes first-run unverified replacements, show each
  candidate and its restore command. Run `--replace-unverified --yes` only
  after explicit approval; it quarantines the old tree before installing and
  verifying the remote copy.
- Do not run `--apply` until the intended commit is present at every registry source ref. A pushed feature branch does not update a registry source pinned to `main`.
- Manual, workflow, project, and catalog records remain outside synthesized apply operations.
- Without a registry, default to a machine-global all-agent install for selected skills by name in copy mode: `bunx skills add "<skills-subpath-url>" --skill <skill-name> --copy -g -a '*' -y`.
- Treat re-running this remote `skills add` command as the reinstall/update path for the current machine.
- Use `--skill '*'` only when the user explicitly asks to install every published repo skill. Do not use `--all`: it expands to both `--skill '*'` and `--agent '*'`, sweeping in the entire catalog instead of keeping the requested skill selection.
- If the user asks for a project install instead, omit `-g` and only narrow agents if requested.

7. Verify the result and keep cleanup separate.
- Run `bunx skills list` for project scope or `bunx skills list -g` for global scope.
- When the repo has `skill-registry.json`, rerun `scripts/audit-global-skills --profile <dev|kicpa>`. If it prints verified legacy duplicates, show the source candidates and restore procedure to the user. Run `--prune --yes` only after explicit approval of those candidates; prune allocates a fresh destination and revalidates each entry. Retain the generated manifest through a normal work cycle.
- For the default machine-global setup, confirm the selected skill resolves from `~/.agents/skills` and lists the supported agents broadly. For a user-requested narrow setup, confirm only those agents appear.
- Report the commit hash when a commit was created, the pushed branch, and that the installed skills now come from the remote-backed source, not the local working tree.
- If the install command overwrote existing skills, say so explicitly in the summary.

## Command Pattern

```bash
SKILLS_URL="https://github.com/sjunepark/agent-scripts/tree/main/skills"
PROFILE="dev"
git remote get-url origin
git status --short
scripts/audit-global-skills --profile "$PROFILE"
git add skills/merge-branch/SKILL.md
git commit -m "docs: update skill workflow"
INTENDED_COMMIT=$(git rev-parse HEAD)
git push
# After the reviewed change is merged to the registry's pinned main ref:
git fetch origin main
git merge-base --is-ancestor "$INTENDED_COMMIT" origin/main
bunx skills add "$SKILLS_URL" --list
scripts/audit-global-skills --profile "$PROFILE" --apply
scripts/audit-global-skills --profile "$PROFILE"
# Only for explicitly approved first-run stale-copy candidates:
scripts/audit-global-skills --profile "$PROFILE" --replace-unverified --yes
# After approving the printed duplicate source candidates:
scripts/audit-global-skills --profile "$PROFILE" --prune --yes
```
