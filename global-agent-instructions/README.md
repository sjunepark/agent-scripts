# Global Agent Instructions

The Markdown instruction files in this directory are regular, tracked source
files. Each agent's user-level configuration path is a symlink to its source:

| Agent | Symlink path | Source in this directory |
| --- | --- | --- |
| Codex | `~/.codex/AGENTS.md` | [global-codex.md](global-codex.md) |
| Claude Code | `~/.claude/CLAUDE.md` | [global-claude.md](global-claude.md) |
| Pi | `~/.pi/agent/AGENTS.md` | [global-pi.md](global-pi.md) |

Edit the source files here to update the content reached by those symlinks.
Do not assume an already-running session reloads changed instructions; verify
its loaded instructions before relying on the new behavior.

Keep these files focused on durable personal defaults. Keep this repository's
maintenance rules in its [root AGENTS.md](../AGENTS.md), and keep multi-step
procedures in skills. Separate agent files preserve tool-specific behavior.

For machine-specific pointer ownership and verification, see
[settings synchronization](../docs/settings-sync.md#global-agent-instructions).
