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
- If the commit scope or message is unclear, ask the user before committing.
- If there are no relevant local changes, skip the commit and push steps and continue with the remote reinstall/update.

3. Commit the intended changes.
- Stage the intended files explicitly with `git add <paths>`.
- Do not commit generated install output or local-only caches.

4. Push the commit so the remote contains the intended version.
- Run `git push` after committing.
- If the push fails, stop and report the failure. Do not install from a remote that lacks the intended changes.
- If no commit was needed because the working tree already matched the intended state, confirm the remote branch already contains HEAD (`git status` reports up to date with `origin/<branch>`).

5. Inspect the remote skill source before installing.
- Run `bunx skills add "<skills-subpath-url>" --list`.
- Confirm the remote resolves and exposes only the published skills from `skills/`.

6. Reinstall or update the intended selected skills from the remote source onto this machine.
- Treat this repo's `skills/` directory as a catalog, not as a global install manifest.
- Default: machine-global Claude Code + Pi install, targeting selected skills by name in copy mode: `bunx skills add "<skills-subpath-url>" --skill <skill-name> --copy -g -a claude-code -a pi -y`. Use another target set only when the user explicitly asks.
- Treat re-running this remote `skills add` command as the reinstall/update path for the current machine.
- Use `--skill '*'` only when the user explicitly asks to install every published repo skill. Do not use `--all` for scoped setup. In the current `skills` CLI, `--all` expands to both `--skill '*'` and `--agent '*'`, which overrides the Claude Code + Pi agent restriction and recreates shared `~/.agents/skills` installs.
- Use shared `~/.agents/skills` installs only when the user explicitly wants universal multi-harness sharing.
- If the user asks for a project install instead, omit `-g` and only narrow agents if requested.

7. Verify the result.
- Run `bunx skills list` for project scope or `bunx skills list -g` for global scope.
- For a scoped machine-global setup, confirm the selected skill listing shows `Agents: Claude Code, Pi`, not a broad shared-agent set.
- Report the commit hash when a commit was created, the pushed branch, and that the installed skills now come from the remote-backed source, not the local working tree.
- If the install command overwrote existing skills, say so explicitly in the summary.

## Command Pattern

```bash
SKILLS_URL="https://github.com/sjunepark/agent-scripts/tree/main/skills"
SKILL_NAME="merge-branch"
git remote get-url origin
git status --short
git add skills/merge-branch/SKILL.md
git commit -m "docs: update skill workflow"
git push
bunx skills add "$SKILLS_URL" --list
bunx skills add "$SKILLS_URL" --skill "$SKILL_NAME" --copy -g -a claude-code -a pi -y
bunx skills list -g
```
