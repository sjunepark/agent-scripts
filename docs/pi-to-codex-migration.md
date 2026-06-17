# Pi to Codex CLI Migration Tracker

## Purpose

Track the migration from the personal Pi TUI setup to Codex CLI.

This document is the working checklist. Migrate the easy, obvious, low-risk
pieces first. Leave the heavier Pi workflow controllers for a later design
discussion after Codex has been dogfooded with native features.

## Ground Rules

- Keep this repository as the source of truth for reusable agent assets.
- Keep Pi runtime state, auth files, sessions, logs, package caches, and paired
  device data out of this repository.
- Keep Codex runtime state, auth files, sessions, logs, caches, memories, and
  SQLite state out of this repository and out of chezmoi-managed dotfiles.
- Use chezmoi for machine-level pointers and stable non-secret settings, not
  for copying live agent runtime directories wholesale.
- Prefer Codex native features before porting Pi TypeScript extensions.
- Prefer skills over Codex custom prompts; custom prompts are deprecated.
- Keep workflow skills focused and reviewable.
- Use plugins, MCP, hooks, or an app-server driver only when a skill is not
  enough.

## Source Paths

- Pi runtime tree: `/Users/sejunpark/.pi`
- Pi migration notes: `/Users/sejunpark/.pi/CODEX-MIGRATION.md`
- Pi global instructions: `/Users/sejunpark/.pi/agent/AGENTS.md`
- Pi skills: `/Users/sejunpark/.pi/agent/skills`
- Pi personal extension source: `/Users/sejunpark/.pi/agent/git/github.com/sjunepark/pi-personal`
- Codex home: `/Users/sejunpark/.codex`
- Current reusable agent repo: `/Users/sejunpark/IT/agent-scripts`

## Current Baseline

- [x] Codex CLI is installed.
  - Observed version: `codex-cli 0.140.0`.
- [x] Codex already has a user config.
  - Path: `/Users/sejunpark/.codex/config.toml`
  - Current default model: `gpt-5.5`
  - Current reasoning effort: `medium`
- [x] Codex has bundled runtime plugins enabled.
  - `browser-use`, `documents`, `pdf`, `spreadsheets`, `presentations`
- [x] No user MCP servers are configured yet.
- [x] Existing `agent-scripts` repo is clean on `main`.
  - Remote: `https://github.com/sjunepark/agent-scripts.git`
- [x] Existing `agent-scripts/skills` already contains the 15 portable Pi
  skills found under `/Users/sejunpark/.pi/agent/skills`.
- [x] Chezmoi source repo inspected read-only.
  - Source: `/Users/sejunpark/.local/share/chezmoi`
  - Current status: `MM .zshrc`
  - Current risk: applying chezmoi would remove local `EDITOR` and `VISUAL`
    exports from `~/.zshrc` unless that diff is reconciled first.
- [x] Settings sync strategy decided.
  - This repo owns reusable agent assets.
  - Chezmoi owns machine-level pointers and stable non-secret settings.
  - Skill installs come from the GitHub `skills/` subpath with explicit agent
    targets.
  - `~/.agents/skills` stays a generated user-scope skill install location and
    must not be replaced by this repo.
  - Codex user-scope skills are currently installed under `~/.agents/skills`,
    so `bunx skills list -g` can report broad agent availability for Codex
    global skills.
  - Details: `docs/settings-sync.md`.
- [x] Codex global instructions are wired through chezmoi.
  - Live symlink:
    `/Users/sejunpark/.codex/AGENTS.md -> /Users/sejunpark/IT/agent-scripts/instructions/global-codex.md`
  - Chezmoi source commit: `aced3cb`.
- [x] Codex conservative live defaults applied.
  - `approval_policy = "on-request"`
  - `sandbox_mode = "workspace-write"`
  - `web_search = "cached"`
  - The full live `~/.codex/config.toml` is not managed by chezmoi.
- [ ] Decide whether the canonical local path should remain
  `/Users/sejunpark/IT/agent-scripts` or be exposed as `/Users/sejunpark/agent-scripts`.
  - `./agent-scripts` and `../agent-scripts` from `/Users/sejunpark/.pi` do not
    currently exist.

## Phase 1: Repository Shape

Goal: make `agent-scripts` look and behave like the canonical shared repo for
Codex-related instructions, skills, scripts, hooks, and docs.

- [x] Add or update `README.md`.
  - Explain that this repo owns shared agent instructions, reusable skills,
    portable scripts, hooks, and migration docs.
  - Document the global symlink/install strategy.
- [x] Add `scripts/validate-skills`.
  - Validate every `skills/*/SKILL.md`.
  - Check YAML front matter starts at the first line.
  - Require non-empty `name` and quoted `description`.
  - Catch duplicate skill names.
- [x] Add `hooks/pre-commit`.
  - Run `scripts/validate-skills`.
  - Enable with `git config core.hooksPath hooks` if desired.
- [x] Add or update `skills.sh.json`.
  - Group the existing skills for discovery.
  - Keep groupings small and practical.
- [x] Decide whether to add `tools.md`.
  - Use only if there is a real local tool catalog worth documenting.
  - Decision: do not add it yet; there is no separate local tool catalog beyond
    the skills CLI and optional hook documented in `README.md`.

## Phase 2: Codex Global Wiring

Goal: make Codex load shared instructions and skills from this repository.

- [x] Decide skill discovery path.
  - Use explicit published installs from
    `https://github.com/sjunepark/agent-scripts/tree/main/skills`.
  - Keep the existing Claude Code + Pi setup:
    `bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill '*' --copy -g -a claude-code -a pi -y`.
  - Add Codex separately when ready:
    `bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill '*' --copy -g -a codex -y`.
  - Do not replace `~/.agents/skills`; it currently contains `context7-mcp`,
    `impeccable`, and `skill-cleaner`, and Codex user-scope skills are
    installed there.
- [x] Install published repo skills for Codex global use.
  - Commit pushed: `e51f164`.
  - Remote source validated:
    `https://github.com/sjunepark/agent-scripts/tree/main/skills`.
  - Installed command:
    `bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --skill '*' --copy -g -a codex -y`.
  - Observed location: `~/.agents/skills`.
- [x] Wire global Codex instructions.
  - Candidate:
    `~/.codex/AGENTS.md -> /Users/sejunpark/IT/agent-scripts/instructions/global-codex.md`
  - If keeping repo-maintenance instructions in `AGENTS.md`, create a separate
    shared global instruction file first.
  - Shared file exists; machine-level symlink is applied and tracked by
    chezmoi.
- [x] Verify Codex instruction loading.
  - Run from a normal repo:
    `codex --ask-for-approval never exec "Summarize the current instructions."`
  - Confirm Codex reports the user-level instruction context and any repo-local
    `AGENTS.md`.
  - Verification used:
    `codex --strict-config --ask-for-approval never exec "Summarize the current instructions and list the instruction files you loaded."`
  - Codex reported the user-level instruction context and this repo's
    `AGENTS.md`.
- [ ] Verify Codex skill loading.
  - Start Codex and run `/skills`.
  - Confirm the expected skills appear.
  - Test one explicit mention, for example `$manual-branch-integrator`.
  - CLI verification completed with `bunx skills list -g -a codex`; TUI
    `/skills` check remains pending.

## Phase 3: Config Migration

Goal: carry over only stable personal defaults from Pi settings.

- [x] Map Pi `defaultModel`.
  - Pi: `gpt-5.5`
  - Codex: already set to `gpt-5.5`.
- [x] Decide whether to map Pi `defaultThinkingLevel`.
  - Pi: `high`
  - Codex currently: `model_reasoning_effort = "medium"`
  - Change to `high` only if that is the desired default cost/latency tradeoff.
  - Decision: keep Codex at `medium` for now.
- [x] Decide sandbox and approval defaults.
  - Recommended interactive default: `sandbox_mode = "workspace-write"` with
    `approval_policy = "on-request"`.
  - Use `danger-full-access` only for externally sandboxed or explicitly trusted
    sessions.
  - Applied live to `/Users/sejunpark/.codex/config.toml`.
- [x] Decide web search mode.
  - Codex default is cached web search.
  - Use live search only when current external facts matter.
  - Applied live: `web_search = "cached"`.
- [ ] Preserve trusted projects deliberately.
  - Existing trusted entries include `/Users/sejunpark/IT/creo`,
    `/Users/sejunpark/IT/codex`, `/Users/sejunpark/IT/pi-personal`, and
    `/Users/sejunpark/.pi`.
  - Add other Pi trusted paths only when they should load project `.codex/`
    config and hooks.
- [ ] Do not migrate Pi UI-only keys.
  - Leave behind `lastChangelogVersion`, Pi theme, tree filter mode,
    `hideThinkingBlock`, and Pi package wiring.

## Phase 4: Global Instructions

Goal: migrate useful personal behavior defaults without dumping all Pi text
unfiltered into Codex.

- [x] Review `/Users/sejunpark/.pi/agent/AGENTS.md`.
- [x] Move durable personal defaults into shared `instructions/global-codex.md`.
  - Concise response defaults.
  - Context-building habits.
  - Refactor and code-quality defaults.
  - Change-management safety rules.
  - Frontend design preferences that should apply globally.
- [ ] Keep repo-specific maintenance rules in this repo's `AGENTS.md`.
  - Do not mix skill repository maintenance instructions with global Codex
    behavior if they have different scopes.
- [x] Verify the final shared instructions stay small enough for global use.
  - Avoid long generic advice.
  - Prefer rules that are frequently useful or prevent recurring mistakes.

## Phase 5: Existing Skills

Goal: make the already-portable skills reliable in Codex.

- [x] Confirm these skills already exist in `skills/`:
  - `agents-md-writer`
  - `architecture-md-writer`
  - `briefing`
  - `change-explainer`
  - `diet`
  - `doc-comment-writer`
  - `manual-branch-integrator`
  - `review-campaign`
  - `skills-cli`
  - `source-investigator`
  - `structure-review`
  - `svelte`
  - `sveltekit`
  - `teach`
  - `ui-lab`
- [x] Validate front matter for every skill.
- [ ] Tighten long descriptions where needed.
  - Descriptions should be routing triggers, not full workflow summaries.
- [x] Verify every referenced file exists.
  - `references/`
  - `scripts/`
  - `agents/openai.yaml`
- [ ] Decide what to do with `evals/`.
  - Keep only if a real evaluator still uses them.
- [ ] Add `agent-network` only if Codex will have equivalent remote-pi mesh
  tools.
  - Otherwise leave it Pi-only.

## Phase 6: Prompt Template Migration

Goal: convert Pi slash-style prompts into Codex skills or fold them into
existing skills.

- [ ] `merge.md`
  - Target: fold into or replace with `manual-branch-integrator`.
  - Rewrite Pi positional placeholders such as `$1` and `${@:2}` into skill
    instructions that work with normal Codex prompts.
- [ ] `merge-review.md`
  - Target: add an audit mode to `manual-branch-integrator` or create a focused
    `merge-auditor` skill.
- [ ] `progress-check.md`
  - Target: new read-only progress-doc skill.
- [ ] `progress-handoff.md`
  - Target: new handoff-plan skill.
- [ ] `progress-run.md`
  - Target: new plan-runner skill that updates the plan but does not commit.
- [ ] `progress-run-full.md`
  - Target: separate explicit long-run plan executor.
  - Keep WIP commit behavior opt-in and obvious.
- [ ] `progress-commit.md`
  - Target: separate commit-oriented skill or leave as an explicit custom
    prompt if it is too command-like.
- [ ] `bug-retro.md`
  - Target: new bug-retro skill or fold into `diet`/`structure-review` only if
    that does not blur the workflow.
- [ ] `broader-refactor.md`
  - Target: new refactor-assessment skill or fold into `diet`.
- [ ] Archive prompt ticket notes under docs only if they are still useful.
  - Candidate: `prompts/tickets/2026-06-16-concise-progress-handoff.md`

## Phase 7: Native Codex Replacement Checks

Goal: dogfood Codex native features before porting Pi controllers.

- [ ] Use Codex `/goal` for one long task.
  - Compare against Pi `extensions/goal/**`.
  - Note gaps in token-budget behavior, status display, continuation, and
    compaction.
- [ ] Use Codex `/review` on a working tree.
  - Compare against simple post-implementation review needs.
- [ ] Use Codex subagents on a real parallel investigation.
  - Compare against Pi `extensions/subagent.ts`.
- [ ] Use Codex `/compact` during a long session.
  - Note whether task state survives well enough.
- [ ] Use Codex Browser Use or in-app browser for a real UI check.
  - Compare against Pi `agent-browser.ts`.
- [ ] Use `/status`, `/statusline`, `/fast`, and `/model`.
  - Confirm whether Pi status widgets and fast-mode patches are obsolete.

## Phase 8: Secret and State Cleanup

Goal: avoid moving private runtime state into reusable Codex assets.

- [ ] Do not copy `/Users/sejunpark/.pi/agent/auth.json`.
  - Re-auth with Codex instead.
- [ ] Do not copy Pi sessions or run history.
  - `/Users/sejunpark/.pi/agent/sessions`
  - `/Users/sejunpark/.pi/agent/run-history.jsonl`
- [ ] Do not copy debug logs or package caches.
  - `/Users/sejunpark/.pi/agent/pi-debug.log`
  - `/Users/sejunpark/.pi/agent/npm/node_modules`
- [ ] Do not copy remote-pi runtime state.
  - `/Users/sejunpark/.pi/remote/peers.json`
  - sockets, locks, relay state
- [ ] Do not copy service state files.
  - Exa usage and API keys
  - Pushover state
- [ ] Consider tightening permissions or archiving Pi transcripts separately.
  - The session audit found large, world-readable transcripts with sensitive
    prompt and tool context.

## Harder Migrations To Discuss Later

These are intentionally not part of the first easy migration pass.

### `post-review-loop`

- Current Pi behavior:
  - Deterministic multi-phase review loop.
  - Structured tools for phase results and final commit results.
  - Bucket I and Bucket II ledger.
  - Git fingerprints, validation cache, selective commits, final reports.
- Possible Codex target:
  - One-shot review skill for the simple path.
  - Codex plugin plus MCP or app-server driver for the deterministic stateful
    loop.
- Discussion needed:
  - How much deterministic gating must survive?
  - Should state live in repo files, plugin state, or an external driver?
  - Should commits remain agent-authored or driver-mediated?

### `auto-review-loop`

- Current Pi behavior:
  - Selects review slices.
  - Records `reviews/auto-review/` ledger.
  - Starts `post-review-loop` after safe fixes.
- Possible Codex target:
  - `review-campaign` skill for a simpler human-steered workflow.
  - External app-server driver if autonomous slice selection remains important.
- Discussion needed:
  - Whether the existing `review-campaign` skill is enough.
  - Whether autonomous continuation is worth the complexity.

### `auto-dev-loop`

- Current Pi behavior:
  - Picks one task from plan docs.
  - Implements it.
  - Runs post-review.
  - Pauses for Bucket II decisions.
  - Compacts between tasks.
- Possible Codex target:
  - External app-server driver around Codex threads.
  - Keep task ledgers in normal repo files.
- Discussion needed:
  - Cost and runaway-loop controls.
  - Phone/remote answer flow.
  - When to start new threads versus continue existing ones.

### Custom Agent-Authored Compaction

- Current Pi behavior:
  - `compact_conversation` tool lets the agent author a dense replacement
    context.
  - Threshold guard injects compaction advisories at 40/50/60/70 percent.
- Possible Codex target:
  - Native `/compact` first.
  - Custom compact prompt or hooks only if native summaries lose critical state.
- Discussion needed:
  - Whether native Codex compaction preserves enough workflow state.
  - Whether a hook can enforce the desired behavior without fighting Codex.

### Browser Tooling

- Current Pi behavior:
  - `agent-browser.ts` wraps an external `agent-browser` CLI.
- Possible Codex target:
  - Built-in Browser Use/in-app browser first.
  - MCP wrapper only if the external CLI is still required.
- Discussion needed:
  - Which browser profile/session behavior matters.
  - Whether screenshots and DOM snapshots need the same exact shape.

### Notifications

- Current Pi behavior:
  - Pushover and cmux completion notifications.
  - Suppression logic for workflow continuations.
- Possible Codex target:
  - `Stop` hooks, plugin hook bundle, or external app-server observer.
- Discussion needed:
  - Which events should notify.
  - Where secrets should live.
  - How to suppress noise from multi-step automation.

### Remote Pi / Agent Mesh

- Current Pi behavior:
  - `remote-pi` peer mesh and `agent-network` skill.
- Possible Codex target:
  - MCP server only if equivalent tools are available or worth building.
- Discussion needed:
  - Whether this remains Pi-only.
  - Security model for relay and message contents.

## Completion Criteria For Easy Migration

The easy migration pass is done when:

- [x] `agent-scripts` has README, validation script, optional hook, and this
  tracker.
- [ ] Codex loads shared global instructions from `agent-scripts`.
- [ ] Codex sees the portable skills from `agent-scripts`.
- [ ] High-value Pi prompts have been converted to skills or intentionally
  deferred.
- [ ] Codex config has the desired model/reasoning/sandbox defaults.
- [ ] No Pi secrets, transcripts, caches, or runtime state were copied into this
  repository.
- [ ] Native Codex feature checks have notes for `/goal`, `/review`, subagents,
  compaction, browser, and status/fast mode.

## Next Action

Start with Phase 1 and Phase 2. They give the most migration value with the
least risk and create the foundation for the prompt-to-skill work.
