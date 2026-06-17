---
name: handoff-plan
description: "Create concise implementation handoff plans from the current discussion. Use when ending a long conversation, preserving next-session implementation context, or preparing a PLAN-style document without implementing, staging, committing, or pushing."
---

# Handoff Plan

Create a compact implementation handoff that a fresh session can execute without replaying the old conversation. Preserve durable decisions and immediate next work; drop transcript history and generic boilerplate.

## Workflow

1. Infer the target plan path.
   - Check for repository conventions such as `plans/`, `docs/plans/`, `.pi/plans/`, or existing `PLAN*.md` files.
   - If no convention exists, create `PLAN-<UPPERCASE-SLUG>.md` at the repo root.
   - Choose a short, specific slug from the task or feature. Prefer product or code nouns over vague words like `TASK`, `UPDATE`, or `CHANGES`.
   - If the file already exists, read it and ask before overwriting, merging, or substantially rewriting it.

2. Distill the current discussion into implementation-ready context.
   - Use the conversation as the primary source of truth.
   - Inspect the repository only as needed to make paths, commands, and constraints concrete.
   - Preserve decisions, constraints, intended behavior, naming choices, rejected alternatives, real blockers, and open questions that affect implementation.
   - Prefer latest actionable state over historical logs, resolved uncertainty, duplicate context, or transcript-like notes.
   - Separate confirmed decisions from assumptions and open questions.
   - Keep ordinary handoffs around 80-140 lines. Allow longer plans only when reference-heavy context is genuinely needed.

3. Write a lean default plan.

```markdown
# <Feature / task title>

## Purpose

Short paragraph explaining what should change and why.

## Current state

- Confirmed decisions and constraints that matter for implementation.
- Relevant completed work or repository evidence, only when needed.
- Open blockers or assumptions, if any.

## Next implementation slice

- [ ] Small, reviewable step 1
- [ ] Small, reviewable step 2
- [ ] Small, reviewable step 3
```

4. Add optional sections only when they carry distinct value.
   - `## Files to inspect first`: max 3-5 files or areas, each with a reason.
   - `## Validation`: only non-obvious focused checks, current validation blockers, or manual QA cues.
   - `## Open questions`: only questions that block or materially shape the next slice.
   - `## Scope`: only non-obvious in/out boundaries.
   - `## Decisions and rejected alternatives`
   - `## Risks and edge cases`
   - `## Reference notes`
   - `## Progress log`: only for actual completed work, blockers, or just-run validation that affects continuation.

5. Keep validation compact.
   - Omit `Validation` when it would only restate repository conventions.
   - Include only known commands likely to be useful for the next slice.
   - Prefer placeholders over long file lists, such as `bun run lint:files -- <changed-files>`.
   - Use at most three categories when needed: fast check, before merge, and manual QA.
   - Do not list broad `check`, `test`, `lint`, or `build` commands unless repository convention or task risk makes them specifically useful.
   - Record previous validation only when it was just run and materially affects the next session.

6. Avoid predictable plan boilerplate.
   - Do not add a dedicated `Next command` section by default.
   - Put a concrete next command in the plan only when it is a real, non-obvious immediate action or unblocker.
   - Do not create empty sections.

## Final Response

Report only:

- created or updated file path
- recommended next command, if any
- open questions that should be answered before implementation, if any
