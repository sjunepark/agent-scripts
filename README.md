# Agent Scripts

Shared agent instructions, reusable skills, hooks, and migration notes for local
coding agents.

This repository treats `skills/` as the distributable source. Repo-local
`.agents/` and `.claude/` directories are working configuration for this repo,
not the published package layout.

## Layout

- `AGENTS.md`: maintenance instructions for this repository.
- `instructions/global-codex.md`: candidate personal Codex defaults for
  `~/.codex/AGENTS.md`.
- `skills/`: published reusable skills.
- `docs/`: migration and setup decisions.
- `scripts/`: repository maintenance scripts.
- `hooks/`: optional Git hooks.
- `skills.sh.json`: skills.sh grouping metadata.

Runtime state, auth files, sessions, logs, caches, and machine-local Codex or Pi
data do not belong in this repository.

## Validation

Validate the published skills before committing:

```bash
scripts/validate-skills
```

Enable the optional pre-commit hook:

```bash
git config core.hooksPath hooks
```

## Skill Installs

Inspect the local source while developing:

```bash
bunx skills add ./skills --list
```

Install published skills from GitHub after committing and pushing. Treat
`skills/` as the available catalog, not as a list that must all be installed
globally.

For a selected skill:

```bash
SKILL_NAME="change-explainer"
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill "$SKILL_NAME" --copy -g -a claude-code -a pi -y
```

For a selected Codex global skill, use an explicit Codex target:

```bash
SKILL_NAME="change-explainer"
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill "$SKILL_NAME" --copy -g -a codex -y
```

Install domain skills such as `svelte`, `sveltekit`, and `ui-lab` only in
projects where they are relevant.

Use chezmoi for machine-level pointers and config templates, not for copying
live runtime directories such as `~/.codex`, `~/.pi`, or `~/.claude` wholesale.
See [docs/settings-sync.md](docs/settings-sync.md).
