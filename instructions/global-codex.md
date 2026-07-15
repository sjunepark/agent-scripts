# AGENTS.md

Treat these as personal Codex defaults. Repository and subtree `AGENTS.md`
files are loaded after this file and take precedence when they conflict.

## Response Defaults

- Be concise, clear, and direct.
- Lead with the answer or next action.
- Expand only when the task, risk, or tradeoff justifies it.
- Avoid repetition, padding, long recaps, and generic advice.

## Subagents

- Use subagents for broad reconnaissance or independent work that would
  otherwise bloat the main thread.
- Ask subagents for concise findings, evidence, changed files, and validation
  results.

## Browser Interaction

- Use browser tools only when the task requires an actual browser: interacting
  with rendered UI, using an existing authenticated session, or visually
  verifying browser behavior.
- Do not open a browser merely because a request includes a URL. Prefer the
  relevant CLI, API, or connector; for GitHub repositories, issues, pull
  requests, Actions runs, checks, and logs, use `gh` by default.
- Use `agent-browser` by default for browser interaction, automation,
  extraction, screenshots, and testing that do not require the user's existing
  Chrome state.
- Use `browser-use` only when the task requires the user's running Chrome,
  including its open tabs, profile, cookies, authenticated sessions, or
  extensions, or direct user participation. Warn before actions may open,
  activate, or navigate tabs or shift browser focus.
- Use `chrome:control-chrome` only when the user explicitly requests it and it
  is available in the current session. Do not select it automatically.

## Documentation Defaults

- Treat Markdown as agent-loaded context: keep files short, current, and
  task-scoped.
- Use progressive disclosure for long docs: keep parent files to routing,
  invariants, and high-signal summaries; move details into focused child docs
  and link them.
- Before expanding a large Markdown file, delete stale or duplicate material,
  then split by topic or ownership instead of appending.
- For progress, plan, and review docs, update the current state in place:
  keep latest decisions, validation, blockers, and next action; compress or
  archive prior run notes instead of appending session logs.
- Prefer concise summaries plus paths to source files over pasted transcripts,
  logs, or broad architecture dumps.

## Change Management

- Treat unrelated working-tree changes as intentional.
- Do not delete, reset, restore, checkout, or clean up files you did not create
  without explicit confirmation.
- Persist important decisions in docs or code comments where the decision
  affects future maintenance.
- Prefer enforcing recurring agent mistakes with types, schemas, lint rules,
  tests, or validation scripts before adding more prose to AGENTS.md.
- After finishing a reviewable implementation or editing slice, run
  `$code-review`.
- Use subagents for that review when the change touches shared behavior,
  cross-module contracts, user-facing flows, security, data migration, or a
  nontrivial refactor.
- Use CodeRabbit as a review option when the user asks for it. It is expensive
  (10 reviews/hour), so do not invoke it freely or for routine review passes.
- For bug fixes, start by reproducing the bug in an E2E setting as closely
  aligned with the end-user experience as practical, so the fix addresses the
  real problem.
- For bug fixes, collect structured runtime evidence such as logs, traces,
  error payloads, or reproduction output before speculating about the fix.

## Progress Tracking

- For long-running or unattended work, keep a repo-local progress Markdown file
  current when one exists or when the task needs durable continuity.
- Prefer existing conventions such as `PROGRESS.md`, `PLAN*.md`, `TODO*.md`,
  `docs/plans/`, or `.pi/plans/`; keep entries concise and update decisions,
  completed work, validation, blockers, and the next step.

## Code Defaults

- Make invalid states unrepresentable with the simplest practical types.
- Model errors explicitly and avoid broad catch-all handling without context.
- Log decision points with useful structured context when logging is warranted.
- Add comments for why, tradeoffs, invariants, and non-obvious flow, not for
  obvious mechanics.
- When making technical decisions, do not give much weight to development cost.
  Prefer quality, simplicity, robustness, scalability, and long-term
  maintainability.

## Refactoring Defaults

- Refactor before extending when existing structure fights the change.
- Prefer one clear path over compatibility layers unless staged rollout is
  required.
- Add dependencies only when they remove durable complexity the project should
  not own.
- Avoid speculative schemas, future-proof fields, and clever abstractions.

## Frontend Defaults

- Let typography, spacing, alignment, and contrast carry the hierarchy before
  adding containers.
- Avoid box-heavy UI and nested rounded panels unless grouping materially
  improves comprehension.
- Keep controls complete, responsive, and usable across desktop and mobile.
