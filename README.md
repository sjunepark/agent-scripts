# Agent Scripts

Shared agent instructions, reusable skills, hooks, and migration notes for local
coding agents.

This repository treats `skills/` as the distributable source. Repo-local
`.agents/` and `.claude/` directories are working configuration for this repo,
not the published package layout.

Bundled directory names describe their contents; `references/` has no special
loading behavior. Progressive disclosure comes from explicit, conditional
links in each `SKILL.md` to focused resources such as workflows, guides,
rubrics, recipes, or factual references.

## Layout

- `AGENTS.md`: maintenance instructions for this repository.
- `bin/` when present: stable user-facing commands intended to be on `PATH` or
  symlinked into `~/.local/bin`.
- `global-agent-instructions/global-codex.md`,
  `global-agent-instructions/global-claude.md`, and
  `global-agent-instructions/global-pi.md`: harness-specific personal defaults
  for `~/.codex/AGENTS.md`, `~/.claude/CLAUDE.md`, and `~/.pi/agent/AGENTS.md`.
- `plugins/`: repo-managed local Codex plugins.
- `.agents/plugins/marketplace.json`: repo-local Codex plugin marketplace.
- `codex-hooks/`: canonical standalone Codex hook definitions and scripts.
- `bin/install-codex-hooks`: idempotent installer for those hooks.
- `skills/`: published reusable skills.
- `docs/`: migration and setup decisions.
- `skill-registry.json`: authoritative classification and installation policy
  for published and deliberately recommended external skills.
- `scripts/`: repository maintenance scripts.
- `hooks/`: optional Git hooks.

Runtime state, auth files, sessions, logs, caches, and machine-local Codex or Pi
data do not belong in this repository.

## Shared Commands

Expose stable cross-repo commands from `bin/`, not `scripts/`. Prefer skills for
agent workflows that do not need a stable executable.

`bin/op-agent` provides non-interactive 1Password CLI access for any agent
harness. It keeps the service-account token in the host's secret store rather
than agent configuration. See [docs/1password.md](docs/1password.md) for setup
and migration from the former `op-codex` wrapper.

Use `$progress` when explicitly invoked to organize, orient to, brief or review,
continue, or hand off repo-local plans and tasks. Use `$code-review` for a
bounded review pass that applies only obvious safe fixes.

## Validation

Validate the published skills before committing:

```bash
scripts/validate-skills
```

This checks the repository's strict, dependency-free frontmatter subset, local
links from `SKILL.md`, direct `SKILL.md` pointers for every bundled runtime
file, and complete classification of the published catalog in
`skill-registry.json`; `agents/` metadata and `evals/` fixtures are excluded
from runtime-pointer checks. Runtime Markdown pointers use inline links; wrap
destinations containing whitespace or parentheses in angle brackets.

Audit this machine's global skills against the desired registry:

```bash
PROFILE="dev"
scripts/audit-global-skills --profile "$PROFILE"
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

Use `skill-registry.json` as the source of truth for whether a skill is global,
project-specific, workflow-managed, or catalog-only, along with its provenance
and target agents. See [docs/skill-registry.md](docs/skill-registry.md) for the
schema contract. Use `scripts/audit-global-skills --profile dev|kicpa` to
report drift in the selected global profile. The registry-v2 audit is
read-only until exact filesystem reconciliation is implemented.

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

For project-scoped recommendations, install only when the registry's `when`
condition matches the target repository.

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

Standalone personal lifecycle workarounds do not need a plugin. Install or
update the repository-owned hook module with `bin/install-codex-hooks`; inspect
drift with `bin/install-codex-hooks --check`. Repair boundaries, machine-state
ownership, and upstream removal checks are documented in
[docs/codex-lifecycle-workarounds.md](docs/codex-lifecycle-workarounds.md).
