# AGENTS.md

## Scope
- This repository stores custom local skills for agentic coding tools and is meant to be consumed with `bunx skills`.
- Treat `skills/` as the distributable source for this repository.
- Treat `plugins/` as repo-managed local Codex plugin source.
- Treat repo-local `.agents/` and `.claude/` as skills/config used while working in this repository, not as the source to distribute or globally install from.
- Treat `.agents/plugins/marketplace.json` as repo-local Codex marketplace metadata, not as a global install target.
- Keep shared instructions at the repo root. Add a nested `AGENTS.md` only when one skill subtree needs different rules.

## Skill layout
- Store each skill in `skills/<skill-name>/`.
- Treat `skills/` as a published catalog, not as a list of skills that should
  all be installed globally.
- Do not treat `.agents/skills/` or `.claude/skills/` as the canonical distribution layout for this repo.
- When creating a new skill, start from `https://github.com/openai/skills/tree/main/skills/.system/skill-creator`.
- Keep `SKILL.md` as the entry point for each skill.
- Keep OpenAI/Codex-facing metadata in `agents/openai.yaml`.
- Keep `SKILL.md` frontmatter portable to the Agent Skills specification; put
  client-specific interface and invocation policy in that client's metadata
  file instead of adding custom top-level fields. Use one-line scalar values
  and, when needed, a one-level string mapping under `metadata` so the local
  dependency-free validator can parse the frontmatter strictly. Quote metadata
  values and any scalar more complex than a simple word or phrase.
- Name bundled directories for what they contain: `workflows/` or `modes/`
  for alternate procedures, `rubrics/`, `lenses/`, or `checklists/` for
  evaluation criteria, `guides/` for topic-specific instruction, `recipes/`
  for operational examples, `cases/` for scenario-specific material,
  `templates/` for agent-read output shapes, and `references/` for lookup
  documentation. Keep copied output resources in `assets/`.
- Keep universally required steps and rules in `SKILL.md`. Link every bundled
  agent-read Markdown file directly from `SKILL.md` with an inline Markdown
  link, name every other runtime file such as a script or copied asset by its
  exact relative path, and state when to use it; a directory name alone does
  not expose its contents. Wrap inline-link destinations containing whitespace
  or parentheses in angle brackets. Do not use reference-style links for
  runtime pointers, and avoid link-like examples in fenced code because the
  dependency-free validator scans inline-link syntax literally. Keep those
  resources one directory level from `SKILL.md` and avoid resource-to-resource
  routing. Interface metadata in `agents/` and test fixtures in `evals/` do not
  need runtime pointers.
- Keep bundled skill files self-contained; do not use symlinks or links to
  absolute paths outside the skill directory.
- Keep the directory name and the `name:` field in `SKILL.md` aligned.

## Skill install scope
- Treat `skill-registry.json` as the authoritative classification and install
  policy for all published repo skills and deliberately recommended external
  skills. See `docs/skill-registry.md` for its contract.
- Published means available to install from the GitHub `skills/` subpath; it
  does not imply global installation.
- Keep the fixed global baseline small and machine-independent.
- Select project profiles or direct skills in the project's committed
  `sjskills.toml`; registry profile membership does not make a skill global.
- When checking whether Codex loads skills, verify the intended installed
  subset, not every skill present under this repo's `skills/`.

## Codex plugin layout
- Store repo-local Codex plugins in `plugins/<plugin-name>/`.
- Keep each plugin manifest at `plugins/<plugin-name>/.codex-plugin/plugin.json`.
- Keep plugin skills under the plugin's `skills/` directory, not the published
  root `skills/` catalog.
- Keep plugin lifecycle hooks read-only unless the user explicitly asks for a
  mutating hook. The `chezmoi-sync` startup hook must only check and report.
- After changing plugin metadata, skills, or hooks, update the Codex cachebuster
  and reinstall the plugin from the configured repo marketplace before testing.

## Command layout
- Store stable user-facing commands in `bin/`.
- Treat `bin/` as the only directory intended to be added to `PATH` or symlinked
  into `~/.local/bin`.
- Store repository maintenance helpers in `scripts/`.
- Do not put one-off maintenance helpers in `bin/`; add a stable wrapper there
  only when the command is meant to be used across repositories.
- Prefer exact command names without extensions for `bin/` commands.

## Working commands
- When explicitly invoked, use `$progress` to organize, orient to, continue,
  or hand off repo-local plans and tasks.
- Use `$code-review` after implementation to run one bounded review pass. It
  applies the implementation, system, design, and diet lenses proportionately,
  applies obvious safe fixes, and validates.
- Inspect project-visible skills for the current working directory with `bunx skills list`.
- `bunx skills list` is for understanding what this repo exposes locally in the current directory; it is not the command to verify machine-wide installs.
- Use `bunx skills list -g` to inspect user-level global installs.
- Use `skill-registry.json` as the desired skill registry.
- Use `bin/sjskills plan --global` for read-only global inspection.
- Use `bin/sjskills plan`, `apply`, and `restore <quarantine-id>` for a
  project that commits `sjskills.toml`.
- Treat `scripts/audit-global-skills` only as a read-only transition wrapper
  for `bin/sjskills plan --global`; its profile and mutation interfaces are
  retired.
- Run the dependency-free registry and reconciler tests with `node --test
  scripts/lib/skill-registry.test.js scripts/audit-global-skills.test.js`.
- Validate this repo as a local source with `bunx skills add ./skills --list`.
- Validate one skill directly with `bunx skills add ./skills/<skill-name> --list`.
- Validate published skill metadata and local links with `scripts/validate-skills`.
- Validate one Codex plugin with `python3 ~/.codex/skills/.system/plugin-creator/scripts/validate_plugin.py plugins/<plugin-name>`.
- If that plugin validator reports missing `yaml`, run it from a temporary
  virtualenv with `PyYAML` installed.
- Install this repo's Codex marketplace for ongoing machine use from the remote:
  `codex plugin marketplace add https://github.com/sjunepark/agent-scripts.git --ref main`.
- Do not leave this repo's Codex plugin marketplace pointed at
  `/Users/sejunpark/IT/agent-scripts` or another local working tree unless the
  user explicitly asks for a temporary local development install.
- After changing repo-managed plugins for ongoing use, commit and push first,
  then run `codex plugin marketplace upgrade personal` and reinstall the
  affected plugin with `codex plugin add <plugin-name>@personal`.
- Install or reinstall `chezmoi-sync` with `codex plugin add chezmoi-sync@personal`.
- Enable the optional Git hook with `git config core.hooksPath hooks`; it runs `scripts/validate-skills`.
- For installs on individual machines, use remote GitHub sources so updates can
  flow across machines without depending on the current working tree.
- For skills, use the GitHub `skills/` subpath so installs do not publish
  repo-local `.agents/` and `.claude/` skills.
- If a skill change should be synced or reinstalled from the remote URL, commit and push that change first, then run the remote-URL `bunx skills add ...` command. Do not reinstall from the remote before the relevant commit is published.
- Reconcile only after intended public skill changes are committed, pushed,
  and present at the registry's remote source; fetch that ref and verify each
  intended skill tree matches the published tree, including after squash or
  rebase.
- A first-run copy without trusted reconciler provenance is not an ordinary
  update even when its bytes match. Resolve unmanaged desired paths explicitly;
  `sjskills` has no force-adopt or force-replace interface.
- Do not use `--all` for scoped installs; in the current `skills` CLI it expands to both `--skill '*'` and `--agent '*'`, which can override the intended agent restriction and recreate shared `~/.agents/skills` installs.
- The registry declares selected installation targets as `.agents` and
  `.claude`. The reconciler places them in `~/.agents/skills` and
  `~/.claude/skills`, respectively; it does not create Pi-specific copies.
- Global apply records verified local tree hashes in
  `~/.agents/.global-skill-state.json`. Byte equality alone does not grant
  ownership; later updates proceed only while installed content still matches
  trusted reconciler state.
- Global locks, transaction journals, recovery data, and quarantine live under
  `~/.agents/.sjskills-global/`. Restore uses
  `sjskills restore --global <quarantine-id>` and refuses to overwrite.
- Do not install this repo's skills from the current working tree, `.` or `./skills`, when the goal is to install them for ongoing use on a machine.
- Use local-path skill or plugin installs only for local validation,
  unpublished work, or explicitly requested temporary development testing.
- Use `-g` only when the task is specifically about a global install. Global installs write to user-level directories such as `~/.claude/skills`, `~/.pi/agent/skills`, or the shared `~/.agents/skills` depending on agent and install mode.
- Do not document `bunx skills add . ...` for this repo unless that path is made to work; `./skills` is the local validation path that currently works.
- Do not run `sjskills apply --global` or global restore against a real home
  as repository validation. Real-machine rollout requires a separate reviewed,
  evidence-bound plan and explicit authorization. Global apply also requires
  the reviewed JSON artifact through `--approved-plan` and its approved digest
  through `--approved-plan-sha256`; those flags bind evidence but do not grant
  authorization.
- Restore project or global quarantines only with the identifier reported by
  `sjskills`; restoration refuses to overwrite an active path.

## Editing expectations
- Prefer editing an existing skill in place over adding new top-level conventions.
- When a skill's behavior changes, update `SKILL.md` and any referenced files in the same change.
- When a plugin's behavior changes, update its manifest, hooks, scripts, and
  `docs/settings-sync.md` together when those docs are affected. Keep plugins
  skillless unless agent-facing instructions are worth the persistent context.
- When the goal is to sync that changed skill onto a machine, tell the user to commit and push first so the GitHub `skills/` URL can be used for the install.
- Keep skill instructions concise, executable, and tool-facing.
- Prefer exact commands and concrete paths over generic guidance.

## Current repo facts
- There is no package manifest, CI workflow, or formatter config at the repo root today.
- Dependency-free Node tests cover the registry and read-only audit transition
  wrapper; Go tests cover project and global reconciliation.
- There is a repository-local skill validation script at `scripts/validate-skills`.
- Do not add build or lint instructions to this file unless those workflows are added to the repository.
