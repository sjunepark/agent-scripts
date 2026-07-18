# Install and migrate skills

Use the recipe matching the requested source, target agents, and scope. For this repository, publish a changed skill before reinstalling it from GitHub so the remote contains the intended version.

## Inspect this repo as a remote skill source

```bash
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --list
```

## Add one published repo skill to Claude Code and Pi globally

```bash
SKILL_NAME="merge-branch"
git add "skills/$SKILL_NAME"
git commit -m "update $SKILL_NAME"
git push origin main
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill "$SKILL_NAME" --copy -g -a claude-code -a pi -y
```

## Add selected published repo skills to Claude Code and Pi globally

```bash
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill agents-md-writer --copy -g -a claude-code -a pi -y
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill change-explainer --copy -g -a claude-code -a pi -y
```

## Add one selected published repo skill to Codex globally

```bash
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill change-explainer --copy -g -a codex -y
```

## Add one external skill to Claude Code and Pi globally

Use the same copy-mode pattern for non-repo skills when `skills list -g` should report only `Claude Code, Pi` for the install.

```bash
bunx skills add vercel-labs/skills --skill find-skills --copy -g -a claude-code -a pi -y
```

## Consolidate retired progress skills into progress

Use this after the replacement commit is pushed so the remote catalog contains
`progress` and no longer publishes the retired skills.

```bash
SKILLS_URL="https://github.com/sjunepark/agent-scripts/tree/main/skills"

bunx skills remove progress-doc progress-handoff progress-run -g -y
bunx skills add "$SKILLS_URL" --skill progress --copy -g -a codex -y
bunx skills add "$SKILLS_URL" --skill progress --copy -g -a claude-code -a pi -y
bunx skills list -g
scripts/audit-global-skills
```

## Move selected shared repo skills to Claude Code and Pi only

Use this when older installs in `~/.agents/skills/` or other global agent directories make `skills list -g` report many agents for a skill that should be scoped to Claude Code and Pi.

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
