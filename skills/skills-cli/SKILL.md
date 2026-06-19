---
name: skills-cli
description: "Manage skills with the skills CLI for Codex, Claude Code, and Pi. Use for discovering, installing, listing, updating, removing, or initializing skills; skills.sh; lock-file restore; node_modules sync; and project vs global agent installation troubleshooting."
---

# Skills CLI

Use `bunx skills` commands to manage skills.

## Workflow

1. Inspect current state with `bunx skills list` for project scope and `bunx skills list -g` for global scope.
   - In this repository, use `scripts/audit-global-skills` to compare machine-global installs against `global-skills.json`.
2. Use `bunx skills find <query>` or `bunx skills add <source> --list` to discover options.
3. Install with explicit scope and agent targeting:
   - For this repository's published skills, use `https://github.com/sjunepark/agent-scripts/tree/main/skills` rather than the repo root or current working tree so installs can update across machines without exposing repo-local `.agents/` and `.claude/` skills.
   - Treat this repo's `skills/` directory as a catalog, not as a manifest of skills that all belong in global installs.
   - For global installs, choose the intended skill by name with `--skill <name>`; install project/domain skills only where they apply.
   - For Claude Code + Pi global use, use `--copy -g -a claude-code -a pi` so selected skills live directly in `~/.claude/skills` and `~/.pi/agent/skills` and `skills list -g` reports `Agents: Claude Code, Pi`.
   - For Codex global use, use a separate explicit Codex target: `--copy -g -a codex`; current Codex user-scope skills are discovered from `~/.agents/skills`.
   - Shared `~/.agents/skills` installs are for intentional universal multi-harness sharing; they can make `skills list -g` report many agents for one skill.
   - Project setup: omit `-g`.
4. Verify install with the matching scope command and test a prompt that should trigger the skill.
5. For generic updates, run `bunx skills check` then `bunx skills update`. For this repo's published skills, prefer reinstalling the affected skills from the remote URL after committing and pushing local skill changes when converting from shared `~/.agents/skills` installs to scoped agent installs.
6. For cleanup, use `bunx skills remove ...` with matching scope/agent flags.

## Key commands

- `bunx skills add <package-or-url>`
- `bunx skills add <source> --skill <name> -a <agent> [-g] -y`
- `bunx skills add <source> --all` only for intentional all-skills/all-agents installs
- `bunx skills add "$SOURCE" --skill "$SKILL_NAME" --copy -g -a claude-code -a pi -y`
- `bunx skills add "$SOURCE" --skill "$SKILL_NAME" --copy -g -a codex -y`
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
- Default to symlink mode unless the user requests `--copy` or the goal is a targeted machine-global install for selected repo skills, including Claude Code + Pi or Codex.
- `bunx skills list` without `-g` reports project-visible skills for the current directory; do not use it to answer what is globally installed.
- When installing this repository's skills for regular machine use, do not install from `.` or `./skills`, and do not use the repo root URL; use `https://github.com/sjunepark/agent-scripts/tree/main/skills`.
- If the user wants `skills list -g` to show only `Agents: Claude Code, Pi` for selected repo skills, do not leave those installs in `~/.agents/skills`; install them with `--skill <name> --copy -g -a claude-code -a pi -y`.
- If the user wants a selected repo skill available in Codex globally, install it to Codex with `--skill <name> --copy -g -a codex -y`; current CLI behavior writes Codex user-scope skills under `~/.agents/skills`.
- Use `--skill '*'` only when the user explicitly wants every skill from a source. In the current `skills` CLI, `--all` expands to `--skill '*' --agent '*' -y`, which can unintentionally recreate shared `~/.agents/skills` installs.
- Treat installed skills as executable instructions; avoid untrusted sources.
- If managing dotfiles with chezmoi, avoid `chezmoi add` on live skills directories.
- Treat `skills.sh.json` as display grouping metadata, not as the desired machine-global install registry.

## References

Read [references/cli.md](references/cli.md) for source formats, scopes, and practical command recipes.
