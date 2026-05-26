---
name: skills-cli
description: "Manage skills with the skills CLI for Codex, Claude Code, and Pi. Use for discovering, installing, listing, updating, removing, or initializing skills; skills.sh; lock-file restore; node_modules sync; and project vs global agent installation troubleshooting."
---

# Skills CLI

Use `bunx skills` commands to manage skills.

## Workflow

1. Inspect current state with `bunx skills list` for project scope and `bunx skills list -g` for global scope.
2. Use `bunx skills find <query>` or `bunx skills add <source> --list` to discover options.
3. Install with explicit scope and agent targeting:
   - For this repository's published skills, use `https://github.com/sjunepark/custom-skills/tree/main/skills` rather than the repo root or current working tree so installs can update across machines without exposing repo-local `.agents/` and `.claude/` skills.
   - For this repo's normal machine-global setup, use `--copy -g -a claude-code -a pi` so the skills live directly in `~/.claude/skills` and `~/.pi/agent/skills` and `skills list -g` reports `Agents: Claude Code, Pi`.
   - Shared `~/.agents/skills` installs are for intentional universal multi-harness sharing; they can make `skills list -g` report many agents for one skill.
   - Project setup: omit `-g`.
4. Verify install with the matching scope command and test a prompt that should trigger the skill.
5. For generic updates, run `bunx skills check` then `bunx skills update`. For this repo's published skills, prefer reinstall-from-remote when converting from shared `~/.agents/skills` installs to Claude Code + Pi only.
6. For cleanup, use `bunx skills remove ...` with matching scope/agent flags.

## Key commands

- `bunx skills add <package-or-url>`
- `bunx skills add <source> --skill <name> -a <agent> [-g] -y`
- `bunx skills add <source> --all`
- `bunx skills add <source> --skill '*' --copy -g -a claude-code -a pi -y`
- `bunx skills list [-g] [-a <agent>]`
- `bunx skills find [query]`
- `bunx skills remove [skill...] [-g] [-a <agent>]`
- `bunx skills check`
- `bunx skills update`
- `bunx skills init [name]`
- `bunx skills experimental_install`
- `bunx skills experimental_sync`

## Guardrails

- Prefer explicit `--agent` and `--skill` flags for reproducibility.
- Use `--yes` only for non-interactive runs.
- Default to symlink mode unless the user requests `--copy` or the goal is this repo's normal machine-global Claude Code + Pi setup.
- `bunx skills list` without `-g` reports project-visible skills for the current directory; do not use it to answer what is globally installed.
- When installing this repository's skills for regular machine use, do not install from `.` or `./skills`, and do not use the repo root URL; use `https://github.com/sjunepark/custom-skills/tree/main/skills`.
- If the user wants `skills list -g` to show only `Agents: Claude Code, Pi` for this repo's published skills, do not leave those installs in `~/.agents/skills`; install them with `--skill '*' --copy -g -a claude-code -a pi -y`.
- Do not use `--all` for that setup. In the current `skills` CLI, `--all` expands to `--skill '*' --agent '*' -y`, which overrides the Claude Code + Pi restriction and recreates shared `~/.agents/skills` installs.
- Treat installed skills as executable instructions; avoid untrusted sources.
- If managing dotfiles with chezmoi, avoid `chezmoi add` on live skills directories.

## References

Read [references/cli.md](references/cli.md) for source formats, scopes, and practical command recipes.
