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
- Put supplemental docs in `references/` and add only files that are actually referenced from `SKILL.md`.
- Keep the directory name and the `name:` field in `SKILL.md` aligned.

## Skill install scope
- Published means available to install from the GitHub `skills/` subpath; it
  does not imply global installation.
- Keep global installs to the small baseline that is useful in most repos for a
  given agent.
- Install domain/project skills selectively where they apply. For example,
  `svelte`, `sveltekit`, and `ui-lab` belong in frontend/Svelte projects, not
  in a generic Codex global install.
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

## Working commands
- Inspect project-visible skills for the current working directory with `bunx skills list`.
- `bunx skills list` is for understanding what this repo exposes locally in the current directory; it is not the command to verify machine-wide installs.
- Use `bunx skills list -g` to inspect user-level global installs.
- Validate this repo as a local source with `bunx skills add ./skills --list`.
- Validate one skill directly with `bunx skills add ./skills/<skill-name> --list`.
- Validate published skill metadata and local links with `scripts/validate-skills`.
- Validate one Codex plugin with `python3 ~/.codex/skills/.system/plugin-creator/scripts/validate_plugin.py plugins/<plugin-name>`.
- If that plugin validator reports missing `yaml`, run it from a temporary
  virtualenv with `PyYAML` installed.
- Install the repo-local Codex marketplace with `codex plugin marketplace add /Users/sejunpark/IT/agent-scripts`.
- Install or reinstall the local `chezmoi-sync` plugin with `codex plugin add chezmoi-sync@personal`.
- Enable the optional Git hook with `git config core.hooksPath hooks`; it runs `scripts/validate-skills`.
- For installs on individual machines, use the GitHub `skills/` subpath so updates can flow across machines without publishing repo-local `.agents/` and `.claude/` skills.
- If a skill change should be synced or reinstalled from the remote URL, commit and push that change first, then run the remote-URL `bunx skills add ...` command. Do not reinstall from the remote before the relevant commit is published.
- To install one published repo skill for Claude Code + Pi global use, set `SKILL_NAME=change-explainer` and run: `bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill "$SKILL_NAME" --copy -g -a claude-code -a pi -y`.
- To install one published repo skill for Codex global use, set `SKILL_NAME=change-explainer` and use an explicit Codex target: `bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill "$SKILL_NAME" --copy -g -a codex -y`; current Codex user-scope discovery uses `~/.agents/skills`.
- Do not use `--all` for scoped installs; in the current `skills` CLI it expands to both `--skill '*'` and `--agent '*'`, which can override the intended agent restriction and recreate shared `~/.agents/skills` installs.
- Do not leave this repo's published machine-global installs under `~/.agents/skills` unless the user explicitly wants Codex user-scope/global sharing; that shared path makes `bunx skills list -g` report many agents.
- Do not install this repo's skills from the current working tree, `.` or `./skills`, when the goal is to install them for ongoing use on a machine.
- Use local-path installs only for local validation or unpublished work.
- Use `-g` only when the task is specifically about a global install. Global installs write to user-level directories such as `~/.claude/skills`, `~/.pi/agent/skills`, or the shared `~/.agents/skills` depending on agent and install mode.
- Do not document `bunx skills add . ...` for this repo unless that path is made to work; `./skills` is the local validation path that currently works.
- When converting shared `~/.agents/skills` installs to scoped Claude Code + Pi installs, remove only the affected skill names globally first, then reinstall the intended selected skills from the GitHub `skills/` subpath in copy mode. Keep the concrete command sequence in `skills/skills-cli/references/cli.md`.

## Editing expectations
- Prefer editing an existing skill in place over adding new top-level conventions.
- When a skill's behavior changes, update `SKILL.md` and any referenced files in the same change.
- When a plugin's behavior changes, update its manifest, bundled skill, hooks,
  scripts, and `docs/settings-sync.md` together when those docs are affected.
- When the goal is to sync that changed skill onto a machine, tell the user to commit and push first so the GitHub `skills/` URL can be used for the install.
- Keep skill instructions concise, executable, and tool-facing.
- Prefer exact commands and concrete paths over generic guidance.

## Current repo facts
- There is no package manifest, CI workflow, formatter config, or automated test suite at the repo root today.
- There is a repository-local skill validation script at `scripts/validate-skills`.
- Do not add build, lint, or test instructions to this file unless those workflows are added to the repository.
