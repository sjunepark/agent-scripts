---
name: skills-cli
description: "Skills CLI (`bunx skills`, skills.sh) for Codex, Claude Code, and Pi: discover, initialize, install, list, update, remove, restore lock files, sync `node_modules`, or troubleshoot project/global scope."
---

# Skills CLI

## Workflow

1. Inspect current state with `bunx skills list` for project scope and `bunx skills list -g` for global scope.
   - In this repository, use `skill-registry.json` as the authoritative classification, `sjskills plan` for committed project intent, and `sjskills plan --global` for exact fixed-baseline state.
2. Use `bunx skills find <query>` or `bunx skills add <source> --list` to discover options.
   - When choosing a command, source, scope, or agent target—including initializing a skill—read [references/cli.md](references/cli.md) for the CLI's accepted inputs and install-location facts.
3. Install with explicit scope, skill selection, and agent targeting.
   - Before installing or reinstalling a published skill, or moving a shared install to agent-specific paths, read [recipes/install-and-migrate.md](recipes/install-and-migrate.md) and use the matching recipe.
   - In this repository, prefer `sjskills` over hand-written commands for both committed project intent and the fixed global baseline. Its private adapter materializes with explicit Codex and Claude Code targets; `sjskills` owns verified placement, state, quarantine, and restore.
   - Byte equality alone does not grant reconciler ownership. Managed updates
     and removals proceed only while current content matches trusted state.
4. Verify ordinary installs with the matching scope command. For this
   repository's managed state, verify exact placement with `sjskills plan` for
   a project or `sjskills plan --global` for the fixed global baseline;
   shared-root discovery may report incidental agent visibility.
5. Maintain or remove installs with the same scope and agent targeting used to create them.
   - Before listing filtered installs, removing skills, running generic updates, restoring a lock file, or syncing package-provided skills, read [recipes/manage-installs.md](recipes/manage-installs.md).
   - For this repo's published skills, publish local changes before reinstalling from the remote source.

## Guardrails

- Default to symlink mode; use `--copy` when the user requests it or for a scoped global install of selected repo skills.
- Use `--skill '*'` only when the user explicitly wants every skill from a source. In the current `skills` CLI, `--all` expands to `--skill '*' --agent '*' -y`, which can unintentionally recreate shared `~/.agents/skills` installs.
- Never synthesize a Pi target for this repository's fixed global baseline; Pi discovers the selected shared-root copy.
- Publish changed repository skills before reconciliation. A local edit or unmerged branch is not present at a registry source pinned to `main`.
- Do not run `sjskills apply --global` or global restore against a real home without a separately reviewed rollout plan and explicit authorization.
- Treat installed skills as executable instructions; avoid untrusted sources.
- If managing dotfiles with chezmoi, avoid `chezmoi add` on live skills directories.
