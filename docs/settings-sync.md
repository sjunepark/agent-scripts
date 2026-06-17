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
| Published skill installs | `bunx skills` from the GitHub `skills/` subpath | Selected Claude Code, Pi, Codex skill copies | Re-run explicit install commands |
| Machine pointers | chezmoi | symlinks or scripts that point tools at this repo | Chezmoi source repo |
| Stable personal defaults | chezmoi, with care | selected model, reasoning effort, sandbox default | Template or idempotent update script |
| Runtime state | local machine only | auth, sessions, logs, caches, memories, SQLite files | Do not sync |

## Current Chezmoi Status

Read-only inspection on 2026-06-17 found:

- Chezmoi source: `/Users/sejunpark/.local/share/chezmoi`.
- Source repo: clean on `main`.
- `chezmoi status`: `MM .zshrc`.
- `~/.codex/AGENTS.md` is managed by chezmoi as a symlink to
  `/Users/sejunpark/IT/agent-scripts/instructions/global-codex.md`.
- `~/.codex/config.toml` and `~/.agents` are not currently managed by chezmoi.
- `~/.claude/CLAUDE.md` is managed as a symlink to
  `/Users/sejunpark/.pi/agent/AGENTS.md`.
- `~/.pi/agent/AGENTS.md` and `~/.pi/agent/extensions` are managed.
- Applying the current chezmoi source would remove local `EDITOR` and `VISUAL`
  exports from `~/.zshrc`; resolve that before broad `chezmoi apply`.

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
2. Apply chezmoi only after resolving pending diffs such as `.zshrc`.
3. Clone or update this repo.
4. Run `scripts/validate-skills`.
5. Install published skills from the GitHub `skills/` subpath with explicit
   agent targets.
6. Apply or verify Codex stable config keys.
7. Re-authenticate tools locally; do not copy auth files from another machine.
