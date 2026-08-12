# skills CLI reference

## Sources accepted by `add`

- GitHub shorthand: `owner/repo`
- GitHub URL: `https://github.com/owner/repo`
- Git subpath URL: `https://github.com/owner/repo/tree/main/skills/my-skill`
- Other git remotes (including GitLab)
- Local path: `./my-local-skills`

## Installation scope

- Project (default): installs into project agent skill paths
- Global (`-g`): installs into user-level agent paths

`bunx skills list` shows the project-visible skills for the current working directory.
Use it to inspect what the current repo exposes locally.
Use `bunx skills list -g` to inspect machine-wide global installs.

For the harnesses used most often in this repo:

- Claude Code global path: `~/.claude/skills/`
- Pi global path: `~/.pi/agent/skills/`
- Current raw Codex global copy target: `~/.agents/skills/`
- Universal shared user-scope path: `~/.agents/skills/`

This repository's profile reconciler deliberately invokes the explicit Codex
target and defensively scopes the subprocess `CODEX_HOME` to `~/.agents`. That
pins the current shared destination for Codex/Pi discovery without passing a Pi
target, even if the caller has customized Codex's environment.

If the same skill `name` exists in more than one discovered location, discovery can show multiple entries instead of merging them.

For this repository specifically:

- Use `https://github.com/sjunepark/agent-scripts/tree/main/skills` for published installs, not the repository root or current working tree, so updates can flow across machines without exposing repo-local `.agents/` and `.claude/` skills.
- Use `./skills` only for local validation or unpublished work.
- Treat `skills/` as a catalog. Select the intended global skill with `--skill <name>` rather than installing the whole catalog by default.
- If you want to sync a just-edited skill using the GitHub `skills/` URL, commit and push first; otherwise the remote install will not contain the local changes.
- Use `skill-registry.json` for authoritative scope, audience, profile, source, installation-target, and installation-manager decisions. Use `scripts/audit-global-skills --profile <dev|kicpa>` to compare exact managed-root state with one explicitly selected profile. Its default mode is read-only, `--apply` materializes each remote skill once and reuses the verified staged snapshot, and the exact printed `--prune <sha256:digest> --yes` command separately quarantines the reviewed verified legacy duplicates; manual and workflow-managed entries stay with their recorded manager.
- The reconciler accepts only credential-free public HTTPS URLs or GitHub
  shorthand. It rejects URL credentials, query strings, fragments, local
  paths, and `npm:` sources for global profile reconciliation.
- A stale pre-reconciler copy is not silently adopted or overwritten. The
  separate printed `--replace-unverified <sha256:digest> --yes` boundary
  preserves the exact reviewed set in a manifest-backed quarantine before
  verified staged install.
- Install project recommendations only when their registry `when` condition matches unless the user explicitly changes the desired scope.
- The global profile declares `.agents` for `~/.agents/skills/` copies and
  `.claude` for `~/.claude/skills/` copies.

## Command lookup

- Add a source: `bunx skills add <package-or-url>`
- Add one selected skill: `bunx skills add <source> --skill <name> -a <agent> [-g] -y`
- List installs: `bunx skills list [-g] [-a <agent>]`
- Find skills: `bunx skills find [query]`
- Remove installs: `bunx skills remove [skill...] [-g] [-a <agent>]`
- Check for updates: `bunx skills check`
- Apply updates: `bunx skills update`
- Initialize a skill: `bunx skills init [name]`
- Restore from the project lock file: `bunx skills experimental_install`
- Sync `node_modules`-provided skills: `bunx skills experimental_sync`
