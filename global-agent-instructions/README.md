# Global Agent Instructions

Each agent's user-level instruction path is a symlink to a generated file in
this directory:

| Agent | Symlink path | Generated file | Harness overlay |
| --- | --- | --- | --- |
| Codex | `~/.codex/AGENTS.md` | [global-codex.md](global-codex.md) | [overlay-codex.md](src/overlay-codex.md) |
| Claude Code | `~/.claude/CLAUDE.md` | [global-claude.md](global-claude.md) | [overlay-claude.md](src/overlay-claude.md) |

Edit [the shared template](src/template.md) for rules every agent follows, and
an overlay only for harness-specific material such as tool names, skill
invocation syntax, or machine paths. Then run `scripts/build-global-instructions`
and commit the sources with the regenerated files. `scripts/validate-skills`
fails when a generated file is stale or an overlay does not define exactly the
template's slots.

The template uses `{{slot}}` placeholders. Each overlay defines every slot as a
`<!-- slot: name -->` block; a placeholder alone on a line takes a multi-line
block and disappears when that block is empty. Overlays are named
`overlay-<harness>.md` so Claude Code does not load one as a nested `CLAUDE.md`
on a case-insensitive filesystem.

Do not assume an already-running session reloads changed instructions; verify
its loaded instructions before relying on the new behavior.

Keep these files focused on durable personal defaults. Keep this repository's
maintenance rules in its [root AGENTS.md](../AGENTS.md), and keep multi-step
procedures in skills.

For machine-specific pointer ownership and verification, see
[settings synchronization](../docs/settings-sync.md#global-agent-instructions).
