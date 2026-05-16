---
name: sync
description: Commit and push intended skill changes, then reinstall or update every published skill from the current repository onto the current machine with the skills CLI, using the repository's remote URL rather than a local path. Use whenever the user wants to sync, bootstrap, reinstall, or update all skills from this repo and explicitly wants the install to come from GitHub or another git remote, not `./skills` or the working tree.
---

# Sync

Use this skill to publish the intended local skill changes, then install all published skills from the current repository's remote `skills/` subpath onto the current machine.

Always use the skills CLI for installation. Do not install from `.` or `./skills`.

## Workflow

1. Confirm the repository remote URL.
- Run `git remote get-url origin`.
- If `origin` is missing, stop and tell the user you need a remote URL for this repo.
- For this repository, convert the remote into the GitHub `skills/` subpath: `https://github.com/sjunepark/custom-skills/tree/main/skills`.
- If the remote points somewhere unexpected, show it to the user before committing, pushing, or installing.

2. Inspect the working tree and decide what to publish.
- Run `git status --short` and inspect relevant diffs before committing.
- Commit only the changes that belong to this sync request. Do not sweep in unrelated dirty files.
- If the commit scope or message is unclear, ask the user before committing.
- If there are no relevant local changes, skip the commit and push steps and continue with the remote reinstall/update.

3. Commit the intended changes.
- Stage the intended files explicitly with `git add <paths>`.
- Create a clear commit message that explains the skill change, for example `docs: update sync skill workflow`.
- Do not commit generated install output, local-only caches, or unrelated work.

4. Push the commit so the remote contains the intended version.
- Run `git push` after committing.
- If the push fails, stop and report the failure. Do not install from a remote that lacks the intended changes.
- If no commit was needed because the working tree already matched the intended state, confirm the branch is up to date enough for the requested sync.

5. Inspect the remote skill source before installing.
- Run `bunx skills add "<skills-subpath-url>" --list`.
- Confirm the remote resolves and exposes only the published skills from `skills/`.

6. Reinstall or update every skill from the remote source onto this machine.
- For this repo's normal machine-global setup, target Claude Code + Pi with copy mode: `bunx skills add "<skills-subpath-url>" --skill '*' --copy -g -a claude-code -a pi -y`.
- Treat re-running this remote `skills add` command as the reinstall/update path for the current machine.
- Do not use `--all` for that setup. In the current `skills` CLI, `--all` expands to both `--skill '*'` and `--agent '*'`, which overrides the Claude Code + Pi agent restriction and recreates shared `~/.agents/skills` installs.
- Use shared `~/.agents/skills` installs only when the user explicitly wants universal multi-harness sharing.
- If the user asks for a project install instead, omit `-g` and only narrow agents if requested.

7. Verify the result.
- Run `bunx skills list` for project scope or `bunx skills list -g` for global scope.
- For the normal machine-global setup, confirm the listing shows `Agents: Claude Code, Pi`, not a broad shared-agent set.
- Report the commit hash when a commit was created, the pushed branch, and that the installed skills now come from the remote-backed source, not the local working tree.

## Guardrails

- Use the current repo's GitHub `skills/` subpath URL, not a filesystem path and not the repo root URL.
- Do not install from the remote URL until intended local changes have been committed and pushed.
- Do not commit unrelated working-tree changes just because they are present.
- Prefer the canonical published source `https://github.com/sjunepark/custom-skills/tree/main/skills` for this repo.
- Default to installing for Claude Code + Pi with `--skill '*' --copy -g -a claude-code -a pi -y` unless the user explicitly wants another target set.
- If the user wants `skills list -g` to stay scoped to Claude Code + Pi, do not leave this repo's published skills in `~/.agents/skills`.
- If the install command would overwrite existing skills, report that clearly in the summary.

## Command Pattern

```bash
SKILLS_URL="https://github.com/sjunepark/custom-skills/tree/main/skills"
git remote get-url origin
git status --short
git add <intended-skill-files>
git commit -m "docs: update skill workflow"
git push
bunx skills add "$SKILLS_URL" --list
bunx skills add "$SKILLS_URL" --skill '*' --copy -g -a claude-code -a pi -y
bunx skills list -g
```

If there are no relevant local changes, skip `git add`, `git commit`, and `git push`.

## Example Triggers

- "Sync all skills from this repo onto this machine."
- "Commit, push, and reinstall the repo skills from GitHub."
- "Install every skill in the current repo from GitHub, not the local path."
- "Update my machine's installed skills from the remote URL."
- "Reinstall the repo skills using the remote URL."
