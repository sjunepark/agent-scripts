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

Inspect the fixed global baseline without changing managed roots:

```bash
bin/sjskills plan --global
```

The plan materializes remote expected content in isolated temporary storage,
then reads only the two managed skill roots and explicitly modeled migration
locations. Planning is read-only. Do not run `apply --global` against a real
home as repository validation; real-machine rollout requires a separately
reviewed [rollout plan](plans/sjskills-global-rollout.md) and explicit
authorization. That plan currently records an unresolved exact-content
approval-binding gap.

Enable the optional pre-commit hook:

```bash
git config core.hooksPath hooks
```

## Skill Installs

### Install the command

`bin/sjskills` is the stable source wrapper. It builds the Go command from the
checked-out repository into a temporary directory for each invocation, so Go
1.23 or newer must be available. Add this repository's `bin/` directory to
`PATH`, or create a one-time symlink from an absolute checkout path:

```bash
mkdir -p ~/.local/bin
ln -s /absolute/path/to/agent-scripts/bin/sjskills ~/.local/bin/sjskills
```

Update the command by fast-forwarding the checkout. No generated executable is
committed and no separate auto-updater is required.

Inspect the local source while developing:

```bash
bunx skills add ./skills --list
```

Install published skills from GitHub after committing and pushing. Treat
`skills/` as the available catalog, not as a list that must all be installed
globally.

Use `skill-registry.json` as the source of truth for whether a skill is global,
project-profile, workflow-managed, manual, or catalog-only, along with its
provenance and installation targets. See
[docs/skill-registry.md](docs/skill-registry.md) for the version 4 contract.

For a project, commit only the intent file and treat reconciled placements and
state as generated machine-local data:

```bash
sjskills init dev go
sjskills plan
sjskills apply
```

`sjskills.toml` may combine named profiles with direct third-party
declarations. In a project that adopts this ownership model, ignore
`.sjskills/`, `.agents/skills/`, and `.claude/skills/`; do not add those
patterns until any preexisting committed content has been reviewed and
migrated. Review every plan before apply. Unknown entries are preserved,
unmanaged desired paths and locally modified managed copies block, and removing
intent moves only unchanged trusted content into recoverable quarantine.

When apply prints a quarantine identifier, retain it until the replacement or
removal has completed a normal work cycle. Restore refuses to overwrite an
active destination:

```bash
sjskills restore <quarantine-id>
```

Global reconciliation uses the same transaction engine and one
machine-independent baseline:

```bash
sjskills plan --global
# Run only under a separately reviewed and explicitly authorized rollout:
sjskills apply --global
sjskills restore --global <quarantine-id>
```

The global state file is `~/.agents/.global-skill-state.json`; private locks,
journals, recovery data, and quarantine live under
`~/.agents/.sjskills-global/`. Former machine-profile placements and legacy
Pi copies are reported and preserved rather than automatically adopted or
removed. `scripts/audit-global-skills` is now only a read-only transition
wrapper for `sjskills plan --global`; its profile and mutation arguments are
retired.

## Codex Plugins

Install the remote-backed plugin marketplace, then install the plugins needed
on the machine:

```bash
codex plugin marketplace add https://github.com/sjunepark/agent-scripts.git --ref main
codex plugin add chezmoi-sync@personal
codex plugin add codex-pushover-notify@personal
```

`chezmoi-sync` checks and reviews chezmoi drift. `codex-pushover-notify` sends
turn-completion notifications and exposes Pushover MCP tools; it requires
machine-local Pushover credentials. See
[plugins/codex-pushover-notify/README.md](plugins/codex-pushover-notify/README.md).

Use local plugin marketplace paths only for temporary development testing.
For ongoing machine setup, commit and push plugin changes first, then run
`codex plugin marketplace upgrade personal` and reinstall the affected plugin.

The `chezmoi-sync` startup hook only checks and reports. Use the bundled
review helper before mutating actions such as `chezmoi apply`, `chezmoi add`,
`chezmoi update`, commits, or pushes.

Its current hook command expects this repository at
`$HOME/IT/agent-scripts`, requires executable plugin scripts, and invokes them
through `bash`; `chezmoi` must also be on the hook's `PATH`. If that exact
checkout is absent, the startup hook exits without reporting. Run the review
helper directly when using a different checkout layout.

Use chezmoi for machine-level pointers and config templates, not for copying
live runtime directories such as `~/.codex`, `~/.pi`, or `~/.claude` wholesale.
See [docs/settings-sync.md](docs/settings-sync.md).

Standalone personal lifecycle workarounds do not need a plugin. Install or
update the repository-owned hook module with `bin/install-codex-hooks`; inspect
drift with `bin/install-codex-hooks --check`. Repair boundaries, machine-state
ownership, and upstream removal checks are documented in
[docs/codex-lifecycle-workarounds.md](docs/codex-lifecycle-workarounds.md).
