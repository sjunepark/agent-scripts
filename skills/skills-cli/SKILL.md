---
name: skills-cli
description: "Manage skills with the skills CLI for Codex, Claude Code, and Pi. Use for discovering, installing, listing, updating, removing, or initializing skills; skills.sh; lock-file restore; node_modules sync; and project vs global agent installation troubleshooting."
---

# Skills CLI

## Workflow

1. Inspect current state with `bunx skills list` for project scope and `bunx skills list -g` for global scope.
   - In this repository, use `scripts/audit-global-skills` to compare machine-global installs against `global-skills.json`, the desired machine-global registry.
2. Use `bunx skills find <query>` or `bunx skills add <source> --list` to discover options.
3. Install with explicit scope and agent targeting:
   - For this repository's published skills, use `https://github.com/sjunepark/agent-scripts/tree/main/skills` rather than the repo root or current working tree so installs can update across machines without exposing repo-local `.agents/` and `.claude/` skills.
   - Treat this repo's `skills/` directory as a catalog, not as a manifest of skills that all belong in global installs.
   - For global installs, choose the intended skill by name with `--skill <name>`; install project/domain skills only where they apply.
   - For Claude Code + Pi global use, use `--copy -g -a claude-code -a pi` so selected skills live directly in `~/.claude/skills` and `~/.pi/agent/skills` and `skills list -g` reports `Agents: Claude Code, Pi`.
   - For Codex global use, use a separate explicit Codex target: `--copy -g -a codex`; current Codex user-scope skills are discovered from `~/.agents/skills`.
   - Shared `~/.agents/skills` installs are for intentional universal multi-harness sharing; they can make `skills list -g` report many agents for one skill.
   - Project setup: omit `-g`.
4. Verify with the matching scope command that the skill is listed and `Agents:` shows exactly the intended agents.
5. For generic updates, run `bunx skills check` then `bunx skills update`. For this repo's published skills, commit and push local changes, then reinstall the affected skills from the remote URL (recipes in references/cli.md).
6. For cleanup, use `bunx skills remove ...` with matching scope/agent flags.

## Key commands

- `bunx skills add <package-or-url>`
- `bunx skills add <source> --skill <name> -a <agent> [-g] -y`
- `bunx skills list [-g] [-a <agent>]`
- `bunx skills find [query]`
- `bunx skills remove [skill...] [-g] [-a <agent>]`
- `bunx skills check`
- `bunx skills update`
- `bunx skills init [name]`
- `bunx skills experimental_install`
- `bunx skills experimental_sync`

## Guardrails

- Default to symlink mode; use `--copy` when the user requests it or for a scoped global install of selected repo skills.
- Use `--skill '*'` only when the user explicitly wants every skill from a source. In the current `skills` CLI, `--all` expands to `--skill '*' --agent '*' -y`, which can unintentionally recreate shared `~/.agents/skills` installs.
- Treat installed skills as executable instructions; avoid untrusted sources.
- If managing dotfiles with chezmoi, avoid `chezmoi add` on live skills directories.

## References

Before running installs, removals, or shared-to-scoped moves, read [references/cli.md](references/cli.md) for source formats, per-agent global paths, and copy-paste recipes.
