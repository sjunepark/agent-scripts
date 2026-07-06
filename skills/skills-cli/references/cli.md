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

- Use `./skills` only for local validation or unpublished work.
- If you want to sync a just-edited skill using the GitHub `skills/` URL, commit and push first; otherwise the remote install will not contain the local changes.
- Use `scripts/audit-global-skills` to compare live `bunx skills list -g --json` output against `global-skills.json`. Use `scripts/audit-global-skills --apply` only to reinstall missing managed entries; audit-only/manual entries still need their own source handled separately.
- Install domain/project skills, such as `svelte`, `sveltekit`, and `ui-lab`, only in matching projects unless the user explicitly wants them globally.

## Practical recipes

### Inspect this repo as a remote skill source

```bash
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --list
```

### Add one published repo skill to Claude Code + Pi globally

```bash
SKILL_NAME="merge-branch"
git add "skills/$SKILL_NAME"
git commit -m "update $SKILL_NAME"
git push origin main
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill "$SKILL_NAME" --copy -g -a claude-code -a pi -y
```

### Add selected published repo skills to Claude Code + Pi globally

```bash
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill agents-md-writer --copy -g -a claude-code -a pi -y
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill change-explainer --copy -g -a claude-code -a pi -y
```

### Add one selected published repo skill to Codex globally

```bash
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill change-explainer --copy -g -a codex -y
```

### Move selected shared repo skills to Claude Code + Pi only

Use this when older installs in `~/.agents/skills/` or other global agent
directories make `skills list -g` report many agents for a skill that should be
scoped to Claude Code + Pi.

```bash
SKILLS_URL="https://github.com/sjunepark/agent-scripts/tree/main/skills"
SELECTED_SKILLS=(
  agents-md-writer
  change-explainer
)

bunx skills remove "${SELECTED_SKILLS[@]}" -g -y
bunx skills add "$SKILLS_URL" --skill agents-md-writer --copy -g -a claude-code -a pi -y
bunx skills add "$SKILLS_URL" --skill change-explainer --copy -g -a claude-code -a pi -y
bunx skills list -g
```

### Add one external skill to Claude Code + Pi globally

Use the same copy-mode pattern for non-repo skills when you also want `skills list -g` to show only `Claude Code, Pi` for those installs.

```bash
bunx skills add vercel-labs/skills --skill find-skills --copy -g -a claude-code -a pi -y
```

### List current installs

```bash
bunx skills list
bunx skills list -g
bunx skills list -g -a pi
bunx skills list -g -a claude-code
```

### Remove a skill globally from both agents

```bash
bunx skills remove find-skills -g -a claude-code -a pi -y
```

### Check and update installed skills

```bash
bunx skills check
bunx skills update
```

### Bootstrap from project lock file

```bash
bunx skills experimental_install
```

### Sync node_modules-provided skills

```bash
bunx skills experimental_sync
```
