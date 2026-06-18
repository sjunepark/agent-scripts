# Agent Scripts

Shared agent instructions, reusable skills, hooks, and migration notes for local
coding agents.

This repository treats `skills/` as the distributable source. Repo-local
`.agents/` and `.claude/` directories are working configuration for this repo,
not the published package layout.

## Layout

- `AGENTS.md`: maintenance instructions for this repository.
- `bin/`: stable user-facing commands intended to be on `PATH` or symlinked
  into `~/.local/bin`.
- `instructions/global-codex.md`: candidate personal Codex defaults for
  `~/.codex/AGENTS.md`.
- `plugins/`: repo-managed local Codex plugins.
- `.agents/plugins/marketplace.json`: repo-local Codex plugin marketplace.
- `skills/`: published reusable skills.
- `docs/`: migration and setup decisions.
- `scripts/`: repository maintenance scripts.
- `hooks/`: optional Git hooks.
- `skills.sh.json`: skills.sh grouping metadata.

Runtime state, auth files, sessions, logs, caches, and machine-local Codex or Pi
data do not belong in this repository.

## Shared Commands

Expose cross-repo commands from `bin/`, not `scripts/`. For a machine-local
setup, either add this repository's `bin/` directory to `PATH`:

```bash
export PATH="$HOME/IT/agent-scripts/bin:$PATH"
```

or keep `PATH` pointed at `~/.local/bin` and symlink selected commands:

```bash
mkdir -p ~/.local/bin
ln -s "$HOME/IT/agent-scripts/bin/codex-plan-loop" ~/.local/bin/codex-plan-loop
ln -s "$HOME/IT/agent-scripts/bin/codex-plan-log" ~/.local/bin/codex-plan-log
```

Then use the command from any git repository:

```bash
codex-plan-loop path/to/PLAN.md
codex-plan-log show latest
```

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

## Codex Plugins

Install the remote-backed plugin marketplace and the `chezmoi-sync` plugin:

```bash
codex plugin marketplace add https://github.com/sjunepark/agent-scripts.git --ref main
codex plugin add chezmoi-sync@personal
```

Use local plugin marketplace paths only for temporary development testing.
For ongoing machine setup, commit and push plugin changes first, then run
`codex plugin marketplace upgrade personal` and reinstall the affected plugin.

The `chezmoi-sync` startup hook only checks and reports. Use the bundled
review helper before mutating actions such as `chezmoi apply`, `chezmoi add`,
`chezmoi update`, commits, or pushes.

Use chezmoi for machine-level pointers and config templates, not for copying
live runtime directories such as `~/.codex`, `~/.pi`, or `~/.claude` wholesale.
See [docs/settings-sync.md](docs/settings-sync.md).
