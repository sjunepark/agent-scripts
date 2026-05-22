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

- Shared universal path used by many harnesses: `~/.agents/skills/`
- Claude Code global path: `~/.claude/skills/`
- Pi global path: `~/.pi/agent/skills/`
- Codex global path: `~/.codex/skills/`

If the same skill `name` exists in more than one discovered location, discovery can show multiple entries instead of merging them. Installs left in the shared `~/.agents/skills/` path can make `skills list -g` report many agents for one skill.

For this repository specifically:

- Use `https://github.com/sjunepark/custom-skills/tree/main/skills` as the normal install source.
- Prefer that GitHub subpath URL over `.` or `./skills` so installed skills can be updated across multiple machines without publishing repo-local `.agents/` or `.claude/` skills.
- Use `./skills` only for local validation or unpublished work.
- If you want to sync a just-edited skill using the GitHub `skills/` URL, commit and push first; otherwise the remote install will not contain the local changes.
- The normal machine-global setup for this repo is Claude Code + Pi only: use `--skill '*' --copy -g -a claude-code -a pi -y`.
- Do not use `--all` for that setup. In the current `skills` CLI, `--all` expands to `--skill '*' --agent '*' -y`, which overrides the Claude Code + Pi restriction and recreates shared `~/.agents/skills` installs.
- That copy-mode install writes directly to `~/.claude/skills/` and `~/.pi/agent/skills/` and keeps `skills list -g` reporting `Agents: Claude Code, Pi`.
- Do not leave this repo's published machine-global installs in `~/.agents/skills/` unless the user explicitly wants universal multi-harness sharing.
- If the user asks for a global install, add `-g`; that writes to user-level directories, not repo-local `.agents/` or `.claude/`.

## Practical recipes

### Inspect this repo as a remote skill source

```bash
bunx skills add https://github.com/sjunepark/custom-skills/tree/main/skills --list
```

### Add one published repo skill to Claude Code + Pi globally

Commit and push first if the skill was just edited locally.

```bash
git add skills/<skill-name>
git commit -m "update <skill-name>"
git push origin main
bunx skills add https://github.com/sjunepark/custom-skills/tree/main/skills --skill <skill-name> --copy -g -a claude-code -a pi -y
```

### Add every published repo skill to Claude Code + Pi globally

Commit and push first if the published repo skills were edited locally and those edits should be included in the reinstall.

```bash
bunx skills add https://github.com/sjunepark/custom-skills/tree/main/skills --skill '*' --copy -g -a claude-code -a pi -y
```

### Reset all published repo skills to Claude Code + Pi only

Use this when older installs in `~/.agents/skills/` or other global agent directories make `skills list -g` report many agents for one skill.

```bash
SKILLS_URL="https://github.com/sjunepark/custom-skills/tree/main/skills"
REPO_SKILLS=(
  agents-md-writer
  architecture-md-writer
  briefing
  change-explainer
  doc-comment-writer
  documentation-state-auditor
  pi-extension-intake
  skills-cli
  source-investigator
  structure-review
  summarize
  svelte
  sveltekit
  tailwind-semantic-theme
  teach
)

bunx skills remove "${REPO_SKILLS[@]}" -g -y
bunx skills add "$SKILLS_URL" --skill '*' --copy -g -a claude-code -a pi -y
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
