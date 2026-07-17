---
name: skills-cli
description: "Manage skills with the skills CLI for Codex, Claude Code, and Pi. Use for discovering, installing, listing, updating, removing, or initializing skills; skills.sh; lock-file restore; node_modules sync; and project vs global agent installation troubleshooting."
---

# Skills CLI

## Workflow

1. Inspect current state with `bunx skills list` for project scope and `bunx skills list -g` for global scope.
   - In this repository, use `scripts/audit-global-skills` to compare machine-global installs against `global-skills.json`, the desired machine-global registry.
2. Use `bunx skills find <query>` or `bunx skills add <source> --list` to discover options.
   - When choosing a command, source, scope, or agent target—including initializing a skill—read [references/cli.md](references/cli.md) for the CLI's accepted inputs and install-location facts.
3. Install with explicit scope, skill selection, and agent targeting.
   - Before installing or reinstalling a published skill, or moving a shared install to agent-specific paths, read [recipes/install-and-migrate.md](recipes/install-and-migrate.md) and use the matching recipe.
4. Verify with the matching scope command that the skill is listed and `Agents:` shows exactly the intended agents.
5. Maintain or remove installs with the same scope and agent targeting used to create them.
   - Before listing filtered installs, removing skills, running generic updates, restoring a lock file, or syncing package-provided skills, read [recipes/manage-installs.md](recipes/manage-installs.md).
   - For this repo's published skills, publish local changes before reinstalling from the remote source.

## Guardrails

- Default to symlink mode; use `--copy` when the user requests it or for a scoped global install of selected repo skills.
- Use `--skill '*'` only when the user explicitly wants every skill from a source. In the current `skills` CLI, `--all` expands to `--skill '*' --agent '*' -y`, which can unintentionally recreate shared `~/.agents/skills` installs.
- Treat installed skills as executable instructions; avoid untrusted sources.
- If managing dotfiles with chezmoi, avoid `chezmoi add` on live skills directories.
