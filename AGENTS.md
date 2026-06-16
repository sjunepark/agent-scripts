# AGENTS.md

## Scope
- This repository stores custom local skills for agentic coding tools and is meant to be consumed with `bunx skills`.
- Treat `skills/` as the distributable source for this repository.
- Treat repo-local `.agents/` and `.claude/` as skills/config used while working in this repository, not as the source to distribute or globally install from.
- Keep shared instructions at the repo root. Add a nested `AGENTS.md` only when one skill subtree needs different rules.

## Skill layout
- Store each skill in `skills/<skill-name>/`.
- Do not treat `.agents/skills/` or `.claude/skills/` as the canonical distribution layout for this repo.
- When creating a new skill, start from `https://github.com/openai/skills/tree/main/skills/.system/skill-creator`.
- Keep `SKILL.md` as the entry point for each skill.
- Keep OpenAI/Codex-facing metadata in `agents/openai.yaml`.
- Put supplemental docs in `references/` and add only files that are actually referenced from `SKILL.md`.
- Keep the directory name and the `name:` field in `SKILL.md` aligned.

## Working commands
- Inspect project-visible skills for the current working directory with `bunx skills list`.
- `bunx skills list` is for understanding what this repo exposes locally in the current directory; it is not the command to verify machine-wide installs.
- Use `bunx skills list -g` to inspect user-level global installs.
- Validate this repo as a local source with `bunx skills add ./skills --list`.
- Validate one skill directly with `bunx skills add ./skills/<skill-name> --list`.
- For installs on individual machines, use the GitHub `skills/` subpath so updates can flow across machines without publishing repo-local `.agents/` and `.claude/` skills.
- If a skill change should be synced or reinstalled from the remote URL, commit and push that change first, then run the remote-URL `bunx skills add ...` command. Do not reinstall from the remote before the relevant commit is published.
- For this repo's normal machine-global setup, install directly to Claude Code + Pi with copy mode so `bunx skills list -g` shows `Agents: Claude Code, Pi`: `bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill '*' --copy -g -a claude-code -a pi -y`.
- Do not use `--all` for that setup; in the current `skills` CLI it expands to both `--skill '*'` and `--agent '*'`, which overrides the Claude Code + Pi agent restriction and recreates shared `~/.agents/skills` installs.
- To install one published repo skill for the same setup, use: `bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill <skill-name> --copy -g -a claude-code -a pi -y`.
- Do not leave this repo's published machine-global installs under `~/.agents/skills` unless the user explicitly wants universal multi-harness sharing; that shared path makes `bunx skills list -g` report many agents.
- Do not install this repo's skills from the current working tree, `.` or `./skills`, when the goal is to install them for ongoing use on a machine.
- Use local-path installs only for local validation or unpublished work.
- Use `-g` only when the task is specifically about a global install. Global installs write to user-level directories such as `~/.claude/skills`, `~/.pi/agent/skills`, `~/.codex/skills`, or the shared `~/.agents/skills` depending on agent and install mode.
- Do not document `bunx skills add . ...` for this repo unless that path is made to work; `./skills` is the local validation path that currently works.
- When converting this repo's published skills from shared `~/.agents/skills` installs to Claude Code + Pi only, remove the published repo skill names globally first, then reinstall from the GitHub `skills/` subpath in copy mode. Keep the concrete command sequence in `skills/skills-cli/references/cli.md`.

## Editing expectations
- Prefer editing an existing skill in place over adding new top-level conventions.
- When a skill's behavior changes, update `SKILL.md` and any referenced files in the same change.
- When the goal is to sync that changed skill onto a machine, tell the user to commit and push first so the GitHub `skills/` URL can be used for the install.
- Keep skill instructions concise, executable, and tool-facing.
- Prefer exact commands and concrete paths over generic guidance.

## Current repo facts
- There is no package manifest, CI workflow, formatter config, or automated test suite at the repo root today.
- Do not add build, lint, or test instructions to this file unless those workflows are added to the repository.
