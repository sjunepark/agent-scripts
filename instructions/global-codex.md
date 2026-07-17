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
- Prefer preserving individual commits when merging pull requests; do not
  squash by default.
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

<!-- context7 -->
Use the `ctx7` CLI to fetch current documentation whenever the user asks about a library, framework, SDK, API, CLI tool, or cloud service — even well-known ones like React, Next.js, Prisma, Express, Tailwind, Django, or Spring Boot. This includes API syntax, configuration, version migration, library-specific debugging, setup instructions, and CLI tool usage. Use even when you think you know the answer — your training data may not reflect recent changes. Prefer this over web search for library docs.

Do not use for: refactoring, writing scripts from scratch, debugging business logic, code review, or general programming concepts.

## Steps

1. Resolve library: `npx ctx7@latest library <name> "<user's question>"` — use the official library name with proper punctuation (e.g., "Next.js" not "nextjs", "Customer.io" not "customerio", "Three.js" not "threejs")
2. Pick the best match (ID format: `/org/project`) by: exact name match, description relevance, code snippet count, source reputation (High/Medium preferred), and benchmark score (higher is better). If results don't look right, try alternate names or queries (e.g., "next.js" not "nextjs", or rephrase the question)
3. Fetch docs: `npx ctx7@latest docs <libraryId> "<user's question>"` — run a separate `docs` command per distinct concept if the question spans multiple topics, unless it's about how they interact
4. Answer using the fetched documentation

You MUST call `library` first to get a valid ID unless the user provides one directly in `/org/project` format. Use the user's full question as the query — specific and detailed queries return better results than vague single words, but keep each query to a single concept unless the question is about how concepts interact; combined multi-topic queries dilute ranking and return shallow results for each topic. Do not run more than 3 commands per question. Do not include sensitive information (API keys, passwords, credentials) in queries.

For version-specific docs, use `/org/project/version` from the `library` output (e.g., `/vercel/next.js/v14.3.0`).

If a command fails with a quota error, inform the user and suggest `npx ctx7@latest login` or setting `CONTEXT7_API_KEY` env var for higher limits. Do not silently fall back to training data.
Run Context7 CLI requests outside Codex's default sandbox. If a Context7 CLI command fails with DNS or network errors such as ENOTFOUND, host resolution failures, or fetch failed, rerun it outside the sandbox instead of retrying inside the sandbox.
<!-- context7 -->
