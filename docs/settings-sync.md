# Settings Sync Strategy

## Decision

Use this repo as the source for reusable agent assets, and use chezmoi for
selected machine-level pointers and non-secret local config. Do not put live
runtime state, credentials, histories, caches, logs, or SQLite state into
either this repo or public skill installs.

## Ownership Boundaries

| Layer | Owner | Examples | Sync method |
| --- | --- | --- | --- |
| Reusable agent assets | `agent-scripts` | `skills/`, `bin/`, docs, validation scripts, optional hooks | Git commit and push |
| Repo-local Codex plugins | `agent-scripts` | `plugins/`, `.agents/plugins/marketplace.json` | Local Codex marketplace install |
| Published skill installs | `bunx skills` from the GitHub `skills/` subpath | Selected Claude Code, Pi, Codex skill copies | Re-run explicit install commands |
| Machine pointers | chezmoi or harness-specific setup | symlinks or scripts that point tools at this repo | Chezmoi source repo or explicit local setup |
| Stable personal defaults | chezmoi, with care | selected model, reasoning effort, sandbox default | Template or idempotent update script |
| Runtime state | local machine only | auth, sessions, logs, caches, memories, SQLite files | Do not sync |

## Current Chezmoi Status

Read-only inspection on 2026-07-30 found:

- Chezmoi source: `/Users/sejunpark/.local/share/chezmoi`.
- Source repo: clean on `main`, tracking `origin/main` with `0` ahead and
  `0` behind.
- `chezmoi status`: clean.
- `~/.codex/AGENTS.md` is a symlink to
  `/Users/sejunpark/IT/agent-scripts/global-agent-instructions/global-codex.md`, maintained
  outside chezmoi.
- `~/.codex/config.toml` and `~/.agents` are not currently managed by chezmoi.
- `~/.claude/CLAUDE.md` is managed as a symlink to
  `/Users/sejunpark/IT/agent-scripts/global-agent-instructions/global-claude.md`.
- `~/.pi/agent/AGENTS.md` is a symlink to
  `/Users/sejunpark/IT/agent-scripts/global-agent-instructions/global-pi.md`, maintained
  outside chezmoi; `~/.pi/agent/extensions` is managed by chezmoi.

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

Install or verify the repo marketplace on this machine from the remote:

```bash
codex plugin marketplace add https://github.com/sjunepark/agent-scripts.git --ref main
codex plugin add chezmoi-sync@personal
codex plugin list --marketplace personal --json
```

Do not leave this repo's marketplace pointed at `/Users/sejunpark/IT/agent-scripts`
or another local working tree for normal machine use. Local marketplace paths
are only for temporary development testing.

After editing plugin metadata or hooks for ongoing use, commit and push first,
then refresh the remote-backed marketplace snapshot and start a new thread:

```bash
python3 ~/.codex/skills/.system/plugin-creator/scripts/update_plugin_cachebuster.py \
  /Users/sejunpark/IT/agent-scripts/plugins/chezmoi-sync
git add plugins/chezmoi-sync
git commit -m "Update chezmoi sync plugin"
git push origin main
codex plugin marketplace upgrade personal
codex plugin add chezmoi-sync@personal
```

Run the plugin validation before committing any plugin change. Script-only
edits still need the commit, push, marketplace upgrade, and reinstall flow
before they affect remote-backed installs.

Do not manage `~/.codex/plugins/cache`, plugin trust records, or installed
plugin state through chezmoi. Recreate those with `codex plugin marketplace
add`, `codex plugin add`, `/plugins`, and `/hooks` on each machine.

Machine-specific lifecycle repairs use the standalone hook module in
`codex-hooks/` instead of a plugin. This repository owns the canonical manifest
and scripts; `bin/install-codex-hooks` copies the scripts and merges the owned
registrations into `~/.codex` without replacing unrelated hooks. Codex remains
the owner of hook trust state. The repair boundaries and upstream removal tests
are in
[codex-lifecycle-workarounds.md](codex-lifecycle-workarounds.md).

## Skills

Keep `~/.agents/skills` as a generated user-scope skill install location, not
as a chezmoi-managed directory. Codex discovers user skills there, and other
harnesses may also report skills from that shared location. It also holds
external/manual skills classified in `skill-registry.json`, so do not replace
it with a symlink to this repo.

Treat this repo's `skills/` directory as the published catalog. A skill being
published here means it can be installed from the GitHub `skills/` subpath; it
does not mean it belongs in every global agent install. Use
`skill-registry.json` for the authoritative scope, provenance, installation targets,
and installation manager; install project recommendations only when their
`when` condition matches.

Run `bin/sjskills plan --global` from this repo to inspect exact managed-root
state against the one fixed baseline. `sjskills`, not chezmoi, owns
reconciliation. Machine profiles and host inference are retired; manual and
workflow entries remain with their recorded manager.

The `delegate-ui-to-claude` orchestration skill is intentionally installed only
for Codex. Impeccable is not a machine-global prerequisite: when the skill
delegates work in a UI repository, it provisions Claude's project-scoped copy
with `npx --yes impeccable@latest install -y --providers=claude --scope=project`
if needed and refreshes an existing copy with the matching `update` command.
Using `@latest` is intentional: delegation should use the current published
Impeccable workflow rather than a pinned or long-lived local CLI.
The intended harness exposure stays under `.claude/`, including the Impeccable
skill tree and Claude hook settings. Codex-facing copies such as
`.agents/skills/impeccable`, user-level Impeccable skills under `~/.agents` or
`~/.codex`, and Impeccable entries in `.codex/hooks.json` are conflicts; the
skill reports them and asks the user to remove them rather than deleting them
automatically. Shared hook files should retain their unrelated entries.

Greenfield work, redesigns, missing product context, and other consequential
design choices use resumable approval phases. Claude returns questions or
options to Codex, Codex relays them to the user, and implementation resumes in
the same Claude session after approval. Claude remains read-only during these
phases; Codex may persist an approved PM-owned contract such as `PRODUCT.md`
when Impeccable needs it to derive the next options. Scoped work that inherits
an established product and visual world can use a one-shot run; so can work
for which the user explicitly authorizes unattended design decisions.

Do not reproduce the reconciler's placement work in chezmoi. It maps the
registry's `.agents` target to `~/.agents/skills` and `.claude` to
`~/.claude/skills`, and creates no Pi-specific copy. Read-only inspection is:

```bash
bin/sjskills plan --global
```

Global apply uses trusted provenance only; it does not adopt or replace a
preexisting desired-path tree merely because the bytes happen to match.
Verified global updates preserve prior content in manifest-backed quarantine
under `~/.agents/.sjskills-global/`. Former-profile and other non-baseline
placements remain report-only in v1. Restore uses the reported identifier and
refuses to overwrite:

```bash
sjskills restore --global <quarantine-id>
```

Do not put `sjskills apply --global` or global restore into chezmoi bootstrap.
Those real-home mutations require a separate reviewed, evidence-bound rollout
plan and explicit authorization. Chezmoi may run the read-only global plan
after this repo is cloned, but it must not own or copy the generated skill
roots. `scripts/audit-global-skills` is only a read-only transition wrapper
for that plan; its former profile and mutation arguments are retired.

## Global Agent Instructions

Keep repo-maintenance rules in this repository's root `AGENTS.md`. Keep global
personal defaults in separate harness-specific files even when most guidance
is shared, so tool-specific behavior does not leak between agents. Point each
harness at its file:

```text
~/.codex/AGENTS.md -> /Users/sejunpark/IT/agent-scripts/global-agent-instructions/global-codex.md
~/.claude/CLAUDE.md -> /Users/sejunpark/IT/agent-scripts/global-agent-instructions/global-claude.md
~/.pi/agent/AGENTS.md -> /Users/sejunpark/IT/agent-scripts/global-agent-instructions/global-pi.md
```

These files should contain durable personal defaults only. Keep multi-step
procedures in skills and route to them with a concise harness-specific rule.
Do not mix personal defaults with this repo's maintenance-specific rules.
Chezmoi currently owns only the Claude pointer; Codex and Pi pointers are
maintained outside chezmoi.

## Bootstrap Order For A New Machine

1. Install Codex, Claude Code, Pi, Node.js, Bun, Git, chezmoi, and the 1Password CLI.
2. Apply chezmoi only after resolving any pending managed-file diffs.
3. Clone or update this repo.
4. Provision `op-agent` using [the 1Password host setup](1password.md).
5. Run `scripts/validate-skills`.
6. Register the repo Codex marketplace and install local repo plugins.
7. Select `dev` or `kicpa`, run the exact-state audit, and use its `--apply`
   mode only after the registry's public remote ref contains the intended
   changes; compare each intended skill tree with that remote ref. Keep pruning
   as a separately reviewed step.
8. Apply or verify Codex stable config keys.
9. Re-authenticate other tools locally; do not copy auth files from another
   machine.
