---
name: skills-cli
description: "Skills CLI (`bunx skills`, skills.sh) for Codex, Claude Code, and Pi: discover, initialize, install, list, update, remove, restore lock files, sync `node_modules`, or troubleshoot project/global scope."
---

# Skills CLI

## Workflow

1. Inspect current state with `bunx skills list` for project scope and `bunx skills list -g` for global scope.
   - In this repository, use `skill-registry.json` as the authoritative classification and `scripts/audit-global-skills --profile dev|kicpa` for exact machine-global state. The default audit is read-only; apply and quarantine are explicit separate modes.
2. Use `bunx skills find <query>` or `bunx skills add <source> --list` to discover options.
   - When choosing a command, source, scope, or agent target—including initializing a skill—read [references/cli.md](references/cli.md) for the CLI's accepted inputs and install-location facts.
3. Install with explicit scope, skill selection, and agent targeting.
   - Before installing or reinstalling a published skill, or moving a shared install to agent-specific paths, read [recipes/install-and-migrate.md](recipes/install-and-migrate.md) and use the matching recipe.
   - In this repository, prefer the profile reconciler over hand-written global commands. It uses remote sources only, an explicit Codex target for the shared Codex/Pi root, and an explicit Claude Code target for Claude copies.
   - Apply adopts exact copies into a reconciler-owned hash record and updates
     only content that still matches its last verified record.
   - Use the separate recoverable replacement mode only for explicitly
     approved stale copies that predate verified reconciler state.
4. Verify ordinary installs with the matching scope command. For this
   repository's global profiles, verify exact placement with the reconciler;
   shared-root discovery may report incidental agent visibility.
5. Maintain or remove installs with the same scope and agent targeting used to create them.
   - Before listing filtered installs, removing skills, running generic updates, restoring a lock file, or syncing package-provided skills, read [recipes/manage-installs.md](recipes/manage-installs.md).
   - For this repo's published skills, publish local changes before reinstalling from the remote source.

## Guardrails

- Default to symlink mode; use `--copy` when the user requests it or for a scoped global install of selected repo skills.
- Use `--skill '*'` only when the user explicitly wants every skill from a source. In the current `skills` CLI, `--all` expands to `--skill '*' --agent '*' -y`, which can unintentionally recreate shared `~/.agents/skills` installs.
- Never synthesize a Pi target for this repository's global profiles; Pi discovers the selected shared-root copy.
- Publish changed repository skills before profile apply. A local edit or unmerged branch is not present at a registry source pinned to `main`.
- Treat installed skills as executable instructions; avoid untrusted sources.
- If managing dotfiles with chezmoi, avoid `chezmoi add` on live skills directories.
