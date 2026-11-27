# Catalog pruning audit — 2026-11-27

## Question

Which published skills still earn their place for Claude Fable 5.2, Claude
Opus 5.5, and GPT-5.x Codex, given that stronger models already do generic
explaining, clarifying, delegating, and reviewing well?

## Guidance read

- Agent Skills best practices: add only what the model lacks; scripts for
  fragile steps; test on each target model.
  (platform.claude.com/docs/en/agents-and-tools/agent-skills/best-practices)
- Prompting guides for Fable 5.2, Opus 5 and Opus 5.5: current models
  orchestrate subagents unprompted and tend to over-delegate; review standing
  delegate and verify instructions before migrating; soften emphasis.
  (platform.claude.com/docs/en/build-with-claude/prompt-engineering/)
- Claude Code skills and best-practice docs: every model-invocable skill
  description costs listing context each turn; procedures that must always
  hold belong in hooks. (code.claude.com/docs/en/skills, /best-practices)
- claude.dev posts on context engineering for Claude 5 models and on how the
  Claude Code team uses skills: keep local facts, gotchas, verification, and
  scripts; remove repeated or obvious instructions.

## Decisions

Retired, generic behavior or no install scope:

- `teach`: general explanation; only minor personal preferences.
- `clarify`: restated the global Collaboration rules; its one distinct point
  moved into `interview`.
- `brainstorming`, `distill-response`: native behavior; removed from the
  global baseline.
- `delegate-ui-to-claude`: a Codex-to-Claude UI handoff the user no longer
  needs while working mainly in Claude Code; Impeccable runs there directly.
- `progress`, `address-issues`, `next-goal`: retired at the user's request.
- `macos-storage-cleanup`, `release-please-release`, `ui-lab`: in no profile.

Merged, durable parts kept:

- `agents-md-writer` into `develop-skills/references/instruction-files.md`
  (Codex and Claude Code loading rules, rechecked against both guides). This
  changes activation: agents-md-writer was implicitly invocable, while
  develop-skills stays explicit-only, so rules-file troubleshooting now needs
  `develop-skills` by name. Accepted as part of the requested pruning.
- `update-base-branch` into `merge-branch/modes/update-base.md`.
- `write-go-docs` into `modern-go/guides/docs.md`.
- `clear-rust` into `modern-rust/guides/style.md`.

Trimmed:

- `code-review`: replaced the design and diet lenses with a compact list of
  review dimensions and complexity signals in the implementation lens; kept
  bucket policy, finding threshold, and the system lens. The `impeccable`
  registry entry moved from workflow to manual management with the
  `delegate-ui-to-claude` retirement.
- `harmonize-docs`: cut generic reconciliation prose; kept modes, the
  three-state document model, authority limits, and completion checks.
- `modern-rust`: condensed 1.72-1.87 into one decision-relevant reference.
- `delegate`: removed the Claude Code branch, which conflicted with the global
  Fable delegation rule; now Codex-only (`.agents` target).

Retained unchanged: `address-pr-feedback`, `cleanup-branches`,
`codex-cleanup`, `create-pr`, `explore-repo`, `pdf-to-markdown`, `sjskills`,
`windows-cleanup`.

## Evidence and limits

Static reading of every skill, the registry, and the official guidance above;
`scripts/validate-skills` and the Node registry suite pass. No paired
behavioral or trigger evaluations were run for the merged or trimmed skills,
and installed copies are unchanged until the branch is published and
`sjskills` reconciles each machine.
