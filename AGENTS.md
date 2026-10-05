# AGENTS.md

## Scope
- This repository publishes personal agent skills, the `sjskills` reconciler
  that installs them, Codex plugins and hooks, and the user-level global agent
  instructions.
- Skills and global instructions are consumed by both Claude Code and Codex.
  Write and verify changes for both harnesses unless a file is explicitly
  harness-specific.
- `skills/` is the published catalog and the only skill source to edit or
  distribute. Publication makes a skill installable from the GitHub `skills/`
  subpath; it does not make it global.
- Repo-local `.agents/skills/` and `.claude/skills/` are git-ignored placements
  that `sjskills` generates from `sjskills.toml`. Edit `skills/<skill-name>/`
  instead of those copies.
- `plugins/` holds repo-managed Codex plugin source, and
  `.agents/plugins/marketplace.json` is its repo-local marketplace metadata.
- Keep shared instructions in this file. Add a nested `AGENTS.md` only when one
  subtree needs different rules.

## Global agent instructions
- On a machine set up from this repo, `~/.codex/AGENTS.md` is a symlink to
  `global-agent-instructions/global-codex.md` and `~/.claude/CLAUDE.md` is a
  symlink to `global-agent-instructions/global-claude.md` in this checkout.
  Those generated files are the live user-level instructions for every project:
  a rebuild, branch switch, or checkout here takes effect in new sessions
  before anything is committed.
- Edit only `global-agent-instructions/src/` (`template.md` for shared rules,
  `overlay-<harness>.md` for harness-specific material), then run
  `scripts/build-global-instructions` and commit sources with the regenerated
  files. Never hand-edit `global-agent-instructions/global-*.md` or write
  through the home-directory symlinks.
- Keep durable personal defaults there, this repository's maintenance rules
  here, and task-specific decisions, authority boundaries, and completion
  checks in skills.
- See `global-agent-instructions/README.md` for the slot format and
  `docs/settings-sync.md#global-agent-instructions` for pointer ownership.
  Verify a pointer with `readlink ~/.codex/AGENTS.md` or
  `readlink ~/.claude/CLAUDE.md` before changing it.

## Skill layout
- Store each skill in `skills/<skill-name>/` with `SKILL.md` as its entry point
  and OpenAI/Codex-facing metadata in `agents/openai.yaml`.
- When creating a new skill, start from
  `https://github.com/openai/skills/tree/main/skills/.system/skill-creator`.
- Keep `SKILL.md` frontmatter portable to the Agent Skills specification; put
  client-specific interface and invocation policy in that client's metadata
  file instead of adding custom top-level fields. The one exception is
  `disable-model-invocation: true`, which Claude Code reads only from
  frontmatter; use it only as
  `skills/develop-skills/guides/claude-invocation.md` directs. Use one-line
  scalar values and, when needed, a one-level string mapping under `metadata`.
  Quote metadata values and any scalar more complex than a simple word or
  phrase.
- Name bundled directories for what they contain: `workflows/` or `modes/`
  for alternate procedures, `rubrics/`, `lenses/`, or `checklists/` for
  evaluation criteria, `guides/` for topic-specific instruction, `recipes/`
  for operational examples, `cases/` for scenario-specific material,
  `templates/` for agent-read output shapes, and `references/` for lookup
  documentation. Keep copied output resources in `assets/`.
- Keep universally required steps and rules in `SKILL.md`. Link every bundled
  agent-read Markdown file directly from `SKILL.md` with an inline Markdown
  link, name every other runtime file such as a script or copied asset by its
  exact relative path, and state when to use it. Wrap inline-link destinations
  containing whitespace or parentheses in angle brackets. Do not use
  reference-style links for runtime pointers, and avoid link-like examples in
  fenced code because the validator scans inline-link syntax literally. Keep
  those resources one directory level from `SKILL.md` and avoid
  resource-to-resource routing. Interface metadata in `agents/` and test
  fixtures in `evals/` do not need runtime pointers.
- Keep bundled skill files self-contained: no symlinks and no links to paths
  outside the skill directory.
- `scripts/validate-skills` enforces the frontmatter subset, name and directory
  alignment, link targets, runtime pointers, and registry coverage.

## Skill install scope
- `skill-registry.json` is the authoritative classification and install policy
  for published repo skills and deliberately recommended external skills. See
  `docs/skill-registry.md` for its contract.
- Keep the fixed global baseline small and machine-independent.
- Projects select profiles or direct skills in their committed `sjskills.toml`;
  registry profile membership does not make a skill global.
- The registry's `.agents` and `.claude` targets map to `~/.agents/skills` and
  `~/.claude/skills`. The reconciler creates no Pi-specific copies.
- When checking which skills an agent loads, verify the intended installed
  subset, not every skill under `skills/`.

## Codex plugins
- Store each plugin in `plugins/<plugin-name>/` with its manifest at
  `.codex-plugin/plugin.json`. Keep plugin skills under the plugin's own
  `skills/` directory, and keep plugins skillless unless agent-facing
  instructions are worth the persistent context.
- A plugin that also ships for Claude Code adds `.claude-plugin/plugin.json`
  beside its Codex manifest and an entry in `.claude-plugin/marketplace.json`,
  sharing one `hooks/hooks.json`. Validate with `claude plugin validate`.
- Keep plugin lifecycle hooks read-only unless the user explicitly asks for a
  mutating hook. The `chezmoi-sync` startup hook must only check and report.
- Validate a plugin with
  `python3 ~/.codex/skills/.system/plugin-creator/scripts/validate_plugin.py plugins/<plugin-name>`.
  If it reports missing `yaml`, run it from a temporary virtualenv with
  `PyYAML` installed.
- After changing plugin metadata, skills, hooks, or scripts for ongoing use:
  update the plugin's cachebuster, validate, commit and push, then run
  `codex plugin marketplace upgrade personal` and
  `codex plugin add <plugin-name>@personal`. `docs/settings-sync.md` has the
  full sequence; update it in the same change when plugin behavior changes.
- The marketplace for ongoing machine use is the remote:
  `codex plugin marketplace add https://github.com/sjunepark/agent-scripts.git --ref main`.
  Point it at a local working tree only for explicitly requested temporary
  development testing.

## Commands
- `bin/` holds stable user-facing commands and is the only directory intended
  for `PATH` or symlinking into `~/.local/bin`. Name its commands without
  extensions.
- `scripts/` holds repository maintenance helpers. Add a `bin/` wrapper only
  when a command is meant to be used across repositories.
- `scripts/audit-global-skills` is only a read-only transition wrapper for
  `bin/sjskills plan --global`; its profile and mutation interfaces are
  retired.

## Validation
- Run `scripts/validate-skills` after changing skills, the registry, or global
  instructions. `git config core.hooksPath hooks` enables it as a pre-commit
  hook.
- Run the Node tests with `node --test scripts/lib/skill-registry.test.js
  scripts/lib/global-instructions.test.js scripts/audit-global-skills.test.js
  scripts/sjskills-hook.test.js`.
- Run `go vet ./...` and `go test ./...` after changing `cmd/` or `internal/`,
  and `python3 scripts/release_test.py` after changing release tooling.
- Pull-request checks run on Linux only; checks targeting `main` and releases
  also build and test the native targets. See `docs/sjskills-releases.md`.
- Inspect the skills this repo exposes with `bunx skills add ./skills --list`,
  or one skill with `bunx skills add ./skills/<skill-name> --list`.
  `bunx skills add . ...` does not work for this repo; do not document it.
- `bunx skills list` shows project-visible skills for the current directory;
  `bunx skills list -g` shows user-level global installs.

## Installing and reconciling
- Install for ongoing use only from the remote GitHub `skills/` subpath, never
  from `.`, `./skills`, or another working tree. Local-path installs are for
  validation, unpublished work, or explicitly requested temporary testing.
- Commit and push a skill change before reinstalling or reconciling from the
  remote. Before reconciling, fetch the registry's remote ref and verify each
  intended skill tree matches the published tree, including after squash or
  rebase.
- Use `bin/sjskills plan`, `apply`, and `restore <quarantine-id>` in a project
  that commits `sjskills.toml`, and `bin/sjskills plan --global` for read-only
  global inspection.
- Do not run `sjskills apply --global` or global restore against a real home
  as repository validation. A user-requested `sjskills` sync authorizes the
  configured scopes under `skills/sjskills/SKILL.md`; the agent reviews the
  plans and may use `--yes` without another approval turn. Global apply still
  requires the reviewed JSON artifact through `--approved-plan` and its digest
  through `--approved-plan-sha256`. Preserve provenance, current-tree, and
  filesystem checks; those flags bind evidence rather than expanding scope.
- Byte equality does not grant ownership. A first-run copy without trusted
  reconciler provenance is not an ordinary update; resolve unmanaged desired
  paths explicitly, because `sjskills` has no force-adopt or force-replace
  interface.
- Restore project or global quarantines only with the identifier `sjskills`
  reported; restoration refuses to overwrite an active path. State and
  quarantine locations are in `docs/skill-registry.md`.
- Use `-g` with `bunx skills` only for a task specifically about a global
  install. Do not use `--all` for scoped installs: it expands to both
  `--skill '*'` and `--agent '*'`, which overrides the intended agent
  restriction and recreates shared `~/.agents/skills` installs.

## Editing expectations
- Prefer editing an existing skill in place over adding new top-level
  conventions.
- When a skill's behavior changes, update `SKILL.md` and any referenced files
  in the same change.
- For catalog-wide revisions, use the explicitly requested skill-development
  workflow, account for each skill, and retain compliant skills without
  cosmetic edits. Preserve activation policies unless their migration is
  requested.
