# Settings Sync Strategy

## Decision

Use this repo as the source for reusable agent assets, and use chezmoi for
machine-level pointers and non-secret local config. Do not put live runtime
state, credentials, histories, caches, logs, or SQLite state into either this
repo or public skill installs.

## Ownership Boundaries

| Layer | Owner | Examples | Sync method |
| --- | --- | --- | --- |
| Reusable agent assets | `agent-scripts` | `skills/`, docs, validation scripts, optional hooks | Git commit and push |
| Repo-local Codex plugins | `agent-scripts` | `plugins/`, `.agents/plugins/marketplace.json` | Local Codex marketplace install |
| Published skill installs | `bunx skills` from the GitHub `skills/` subpath | Selected Claude Code, Pi, Codex skill copies | Re-run explicit install commands |
| Machine pointers | chezmoi | symlinks or scripts that point tools at this repo | Chezmoi source repo |
| Stable personal defaults | chezmoi, with care | selected model, reasoning effort, sandbox default | Template or idempotent update script |
| Runtime state | local machine only | auth, sessions, logs, caches, memories, SQLite files | Do not sync |

## Current Chezmoi Status

Read-only inspection on 2026-06-17 found:

- Chezmoi source: `/Users/sejunpark/.local/share/chezmoi`.
- Source repo: clean on `main`, tracking `origin/main` with `0` ahead and
  `0` behind.
- `chezmoi status`: clean.
- `~/.codex/AGENTS.md` is managed by chezmoi as a symlink to
  `/Users/sejunpark/IT/agent-scripts/instructions/global-codex.md`.
- `~/.codex/config.toml` and `~/.agents` are not currently managed by chezmoi.
- `~/.claude/CLAUDE.md` is managed as a symlink to
  `/Users/sejunpark/.pi/agent/AGENTS.md`.
- `~/.pi/agent/AGENTS.md` and `~/.pi/agent/extensions` are managed.

## Codex Settings

Do not manage the entire live `~/.codex` directory. It contains mutable runtime
files such as auth, history, logs, goals, memories, model caches, plugin state,
and SQLite databases.

For `~/.codex/config.toml`, avoid blindly replacing the file because Codex also
writes plugin, marketplace, trust, and UI state there. The safe path is:

1. Keep an audited list of desired stable keys.
2. Apply those keys with a small idempotent chezmoi script or a carefully
   reviewed template.
3. Preserve Codex-managed sections unless intentionally resetting the machine.

Stable candidates:

```toml
model = "gpt-5.5"
model_reasoning_effort = "medium"
approval_policy = "on-request"
sandbox_mode = "workspace-write"
web_search = "cached"
```

Machine-local or Codex-managed sections to preserve:

- `[projects.*]` trust entries.
- `[marketplaces.*]` local cache paths.
- `[plugins.*]` installed plugin state.
- UI onboarding and keymap state unless intentionally standardized.

## Codex Plugins

Keep local Codex plugin source in this repo under `plugins/<plugin-name>/`.
Keep the repo marketplace at `.agents/plugins/marketplace.json`; it is
repository metadata, not machine runtime state.

The `chezmoi-sync` plugin is the local workflow for checking and reviewing
chezmoi drift from Codex:

- `plugins/chezmoi-sync/hooks/hooks.json` registers a `SessionStart` hook.
- `plugins/chezmoi-sync/scripts/chezmoi-check.sh` is read-only and exits zero;
  it reports only.
- `plugins/chezmoi-sync/scripts/chezmoi-review.sh` is the explicit review
  helper for status and optional diff/fetch.

The plugin intentionally does not bundle Codex skills. Keep mutating actions
explicit: review `chezmoi-review.sh --diff`, choose `chezmoi add`,
`chezmoi apply`, or `chezmoi update`, then re-run the review.

Install or verify the repo marketplace on this machine:

```bash
codex plugin marketplace add /Users/sejunpark/IT/agent-scripts
codex plugin add chezmoi-sync@personal
codex plugin list --marketplace personal --json
```

After editing plugin metadata or hooks, refresh Codex's installed cache and
start a new thread:

```bash
python3 ~/.codex/skills/.system/plugin-creator/scripts/update_plugin_cachebuster.py \
  /Users/sejunpark/IT/agent-scripts/plugins/chezmoi-sync
codex plugin add chezmoi-sync@personal
```

Script-only edits to `plugins/chezmoi-sync/scripts/chezmoi-check.sh` are picked
up by the installed hook because the hook command calls this repo path. Still
run the plugin validation before committing any plugin change.

Do not manage `~/.codex/plugins/cache`, plugin trust records, or installed
plugin state through chezmoi. Recreate those with `codex plugin marketplace
add`, `codex plugin add`, `/plugins`, and `/hooks` on each machine.

## Skills

Keep `~/.agents/skills` as a generated user-scope skill install location, not
as a chezmoi-managed directory. Codex discovers user skills there, and other
harnesses may also report skills from that shared location. It currently also
holds non-repo skills (`context7-mcp`, `impeccable`, and `skill-cleaner`), so do
not replace it with a symlink to this repo.

Treat this repo's `skills/` directory as the published catalog. A skill being
published here means it can be installed from the GitHub `skills/` subpath; it
does not mean it belongs in every global agent install. Keep global installs to
the skills that are broadly useful, and install domain skills such as `svelte`,
`sveltekit`, and `ui-lab` only in matching projects.

Use explicit skill and agent targets:

```bash
# Selected setup for Claude Code + Pi
SKILL_NAME="change-explainer"
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill "$SKILL_NAME" --copy -g -a claude-code -a pi -y

# Selected Codex setup; current CLI behavior writes Codex user skills under ~/.agents/skills
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill "$SKILL_NAME" --copy -g -a codex -y
```

Chezmoi can run those commands as bootstrap/update scripts after this repo is
cloned, but it should not `chezmoi add` the generated skill copies.

## Instructions

Keep repo-maintenance rules in this repository's root `AGENTS.md`. For global
personal Codex defaults, create a separate shared instruction file and let
chezmoi install a pointer to it, for example:

```text
~/.codex/AGENTS.md -> /Users/sejunpark/IT/agent-scripts/instructions/global-codex.md
```

That file should contain durable personal defaults only. Do not mix it with
this repo's maintenance-specific rules.

## Bootstrap Order For A New Machine

1. Install Codex, Claude Code, Pi, Bun, Git, and chezmoi.
2. Apply chezmoi only after resolving any pending managed-file diffs.
3. Clone or update this repo.
4. Run `scripts/validate-skills`.
5. Register the repo Codex marketplace and install local repo plugins.
6. Install published skills from the GitHub `skills/` subpath with explicit
   agent targets.
7. Apply or verify Codex stable config keys.
8. Re-authenticate tools locally; do not copy auth files from another machine.
