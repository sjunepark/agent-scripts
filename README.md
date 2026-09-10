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
- [global-agent-instructions/](global-agent-instructions/README.md): personal
  instruction sources symlinked from each agent's user-level configuration.
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

Use [$cleanup-branches](skills/cleanup-branches/SKILL.md) to clean completed
local and remote Git branches. It decides routine deletions from repository and
PR evidence, preserves active or retained work, and asks only about consequential
unresolved cases. It is included in the `dev` project profile.

Use [$address-issues](skills/address-issues/SKILL.md) for a specified GitHub
issue or a confirmed issue queue, through reviewed PR merge and verified
resolution. Multiple mode requires Codex app task tools and runs one
Astra/medium task at a time; blocked issues defer dependents while independent
issues continue in the confirmed order.

Use [$next-goal](skills/next-goal/SKILL.md) to select a substantial goal and
return its prompt. `$next-goal prepare` completes and reviews planning, commits
the relevant preparation, and returns the prompt. `$next-goal spawn luna`
also creates a new Codex task from the prepared state and verifies native goal
startup. `spawn astra` selects Astra/medium; `spawn luna` selects Luna/max;
an optional reasoning argument overrides the preset, as in `spawn astra high`.
Bare `spawn` uses the app's configured task defaults. Preparation preserves
unrelated work; spawning requires the Codex app task tools and native goal
support in the destination task.

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
then reads the selected scope's managed roots and explicitly modeled migration
locations. Automatic notices also inspect the other configured scope as described
below. Planning never changes managed roots. Do not run `apply --global` against
a real home as repository validation. A configured sync request authorizes real-machine
reconciliation; the agent reviews and retains the evidence using the
[global procedure](skills/sjskills/references/global-rollout.md). Global apply
requires the exact reviewed JSON plan artifact and its SHA-256; it fails before
mutation when either the artifact digest or a fresh plan recheck differs.

Enable the optional pre-commit hook:

```bash
git config core.hooksPath hooks
```

## Skill Installs

### Install the command

Standalone releases are available for macOS (Intel and Apple silicon) and
Windows x64. See [installation and release details](docs/sjskills-releases.md)
for downloads and the verified installer contract.

For development, `bin/sjskills` builds the checked-out Go command into a temporary
directory on each invocation. It requires Go 1.23 or newer; call it explicitly
when testing checkout changes. If a native executable is installed, put its
directory before this checkout's `bin/` on PATH to avoid rebuilding during
everyday use. Fast-forwarding the checkout updates the development wrapper's
source; an installed native executable needs a separate rebuild or update.

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
migrated. Review every plan before apply. Sync removes undeclared skills from
these roots into recoverable quarantine, including unknown and locally modified
copies. Previously managed desired copies with local edits are quarantined and
replaced with verified published content. Unmanaged desired copies, source
mismatches, and unverifiable entries still block apply. See the
[reconciliation contract](docs/skill-registry.md#ownership-and-reconciliation).

Run `sjskills` (or `sjskills status`) for CLI version, project, and global status.
The CLI check compares the running version against the latest stable published
`sjskills` release and links to available updates. It reports absent releases or
unavailable evidence explicitly and never installs an update.
It shows the nearest configured project root, or setup guidance when no manifest
exists, without creating anything. Both scopes visibly report no drift, findings,
or unavailable evidence. Review findings with `sjskills plan` or
`sjskills plan --global` before changing anything.

Status reports go to stdout and exit 0 even for drift, missing setup, stale
sources, or unavailable inspection; they are advisory, not a synchronization gate.
`--json` emits one `operation: "status"` envelope with setup metadata and full
`advisories`, plus independent version evidence in `cliAdvisory`. Argument errors exit 64; cancellation exits nonzero. Successful
`init`, `profiles`, `plan`, `apply`, and `restore` retain incidental notices on
stderr (or the advisory fields in JSON), with silence for fresh scopes without
findings and fresh CLI comparisons without updates or problems.

Local drift is checked each time. Matching skill selections share upstream hashes
across projects, so changing directories does not repeat the fetch. Upstream
hashes and CLI release metadata refresh daily within one shared budget; a cold
or expired check can add up to 30 seconds plus
bounded process cleanup. Failed refreshes use
clearly labeled stale evidence when available and wait 15 minutes before retrying.
Use `--no-status-check` to skip status inspection, fetching, and cache writes;
the status command reports that checks are disabled. Other commands retain their
own verification and network requirements. Help, exact version requests, invalid
invocations, and unsuccessful commands skip these checks. See the
[status evidence contract](docs/skill-registry.md#automatic-status-evidence)
for JSON, cache, and approval boundaries.

When apply prints a quarantine identifier, retain it until the replacement or
removal has completed a normal work cycle. Restore refuses to overwrite an
active destination:

```bash
sjskills restore <quarantine-id>
```

Global reconciliation uses the same transaction engine and one
machine-independent baseline. A request to `$sjskills` to sync configured state
covers that baseline and the current project's committed manifest when present;
explicit scope or inspection limits take precedence. The agent reviews each
plan and can apply with `--yes` without another approval turn. This is skill
routing, not a new CLI `sync` subcommand.

Use the same verified executable through global plan and apply:

```bash
sjskills --json plan --global > plan.json
plan_sha256=$(shasum -a 256 plan.json | awk '{print $1}')
# Review the plan, then apply within the requested sync scope:
sjskills apply --global \
  --approved-plan plan.json \
  --approved-plan-sha256 "$plan_sha256"
sjskills restore --global <quarantine-id>
```

The two approval flags are mandatory for global apply and unavailable for
project apply. They bind execution to the reviewed artifact bytes and to a
fresh, complete global plan built from one retained verified materialization
session. The sync request supplies authority; the agent prepares and reviews
the evidence. See the [global procedure](skills/sjskills/references/global-rollout.md)
for executable verification and recovery.

The global state file is `~/.agents/.global-skill-state.json`; private locks,
journals, recovery data, and quarantine live under
`~/.agents/.sjskills-global/`. Both scopes strictly reconcile their
`.agents/skills` and `.claude/skills` roots. Built-in skills, plugin caches,
and legacy Pi copies remain outside that boundary.
`scripts/audit-global-skills` is now only a read-only transition
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
