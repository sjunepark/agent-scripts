# AGENTS.md

Treat these as personal Codex defaults. Repository and subtree `AGENTS.md`
files are loaded after this file and take precedence when they conflict.

## Response Defaults

- Be concise, clear, and direct.
- Lead with the answer or next action.
- Expand only when the task, risk, or tradeoff justifies it.
- Avoid repetition, padding, long recaps, and generic advice.

## Context Building

- Start from the repo root, nearby docs, and relevant entry points.
- Load only the files needed for the current task.
- Prefer installed source and primary docs over memory when behavior matters.
- Use subagents for broad reconnaissance or independent work that would
  otherwise bloat the main thread.
- Ask subagents for concise findings, evidence, changed files, and validation
  results.

## Change Management

- Treat unrelated working-tree changes as intentional.
- Do not delete, reset, restore, checkout, or clean up files you did not create
  without explicit confirmation.
- Keep change surfaces tight and reviewable.
- Use the repository's existing commands and conventions.
- Persist important decisions in docs or code comments where the decision
  affects future maintenance.
- After implementation, run the most relevant validation available.

## Progress Tracking

- For long-running or unattended work, keep a repo-local progress Markdown file
  current when one exists or when the task needs durable continuity.
- Prefer existing conventions such as `PROGRESS.md`, `PLAN*.md`, `TODO*.md`,
  `docs/plans/`, or `.pi/plans/`; keep entries concise and update decisions,
  completed work, validation, blockers, and the next step.

## Code Defaults

- Prefer simple, explicit code and clear names.
- Make invalid states unrepresentable with the simplest practical types.
- Keep responsibility boundaries clear.
- Model errors explicitly and avoid broad catch-all handling without context.
- Log decision points with useful structured context when logging is warranted.
- Add comments for why, tradeoffs, invariants, and non-obvious flow, not for
  obvious mechanics.

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
