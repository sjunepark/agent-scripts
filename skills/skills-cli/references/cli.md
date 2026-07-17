# skills CLI reference

## Sources accepted by `add`

- GitHub shorthand: `owner/repo`
- GitHub URL: `https://github.com/owner/repo`
- Git subpath URL: `https://github.com/owner/repo/tree/main/skills/my-skill`
- Other git remotes (including GitLab)
- Local path: `./my-local-skills`

## Installation scope

- Project (default): installs into project agent skill paths
- Global (`-g`): installs into user-level agent paths

`bunx skills list` shows the project-visible skills for the current working directory.
Use it to inspect what the current repo exposes locally.
Use `bunx skills list -g` to inspect machine-wide global installs.

For the harnesses used most often in this repo:

- Claude Code global path: `~/.claude/skills/`
- Pi global path: `~/.pi/agent/skills/`
- Codex user-scope path (shared with many other harnesses): `~/.agents/skills/`
- Claude Code + Pi copy-mode installs write directly to `~/.claude/skills/` and `~/.pi/agent/skills/` and keep `skills list -g` reporting `Agents: Claude Code, Pi` for those selected skills.
- The Codex copy-mode install currently writes directly to `~/.agents/skills/`.

If the same skill `name` exists in more than one discovered location, discovery can show multiple entries instead of merging them.

For this repository specifically:

- Use `https://github.com/sjunepark/agent-scripts/tree/main/skills` for published installs, not the repository root or current working tree, so updates can flow across machines without exposing repo-local `.agents/` and `.claude/` skills.
- Use `./skills` only for local validation or unpublished work.
- Treat `skills/` as a catalog. Select the intended global skill with `--skill <name>` rather than installing the whole catalog by default.
- If you want to sync a just-edited skill using the GitHub `skills/` URL, commit and push first; otherwise the remote install will not contain the local changes.
- Use `scripts/audit-global-skills` to compare live `bunx skills list -g --json` output against `global-skills.json`. Use `scripts/audit-global-skills --apply` only to reinstall missing managed entries; audit-only/manual entries still need their own source handled separately.
- Install domain/project skills, such as `svelte`, `sveltekit`, and `ui-lab`, only in matching projects unless the user explicitly wants them globally.
- Use shared `~/.agents/skills/` installs only for intentional multi-harness sharing; they can make `skills list -g` report many agents for one skill.

## Command lookup

- Add a source: `bunx skills add <package-or-url>`
- Add one selected skill: `bunx skills add <source> --skill <name> -a <agent> [-g] -y`
- List installs: `bunx skills list [-g] [-a <agent>]`
- Find skills: `bunx skills find [query]`
- Remove installs: `bunx skills remove [skill...] [-g] [-a <agent>]`
- Check for updates: `bunx skills check`
- Apply updates: `bunx skills update`
- Initialize a skill: `bunx skills init [name]`
- Restore from the project lock file: `bunx skills experimental_install`
- Sync `node_modules`-provided skills: `bunx skills experimental_sync`
