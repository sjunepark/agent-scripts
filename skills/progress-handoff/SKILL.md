---
name: progress-handoff
description: "Create concise implementation handoff documents from the current discussion. Use when ending a long session, preserving next-session progress context, or preparing a PLAN-style handoff without implementing, staging, committing, or pushing."
---

# Progress Handoff

Create a concise implementation handoff from the current discussion. This is for ending a long discussion and preparing a fresh implementation session; do not implement code, stage files, commit, or push.

## Workflow

1. Infer the target plan file path.
   - Inspect repository conventions first: `plans/`, `docs/plans/`, `.pi/plans/`, existing `PLAN*.md` files, or similar handoff files.
   - If no convention exists, create `PLAN-<UPPERCASE-SLUG>.md` in the repository root.
   - Choose a short, specific topic slug from the discussed task or feature. Prefer concrete product or code nouns over vague words like `TASK`, `UPDATE`, or `CHANGES`.
   - If the topic is ambiguous, ask for a short slug before writing.
   - If the inferred file already exists, read it and ask before overwriting, merging, or substantially rewriting it.

2. Distill the discussion into implementation-ready context.
   - Use the current conversation as the primary source of truth.
   - Treat any command text or user instruction that invoked the skill as the latest explicit handoff instruction.
   - Inspect the repository only as needed to make paths, commands, and constraints concrete.
   - Preserve durable decisions, constraints, intended behavior, naming choices, rejected alternatives, real blockers, and open questions that affect implementation.
   - Prefer the latest actionable state over historical logs, transcript-like discussion, resolved uncertainty, or duplicate context.
   - Separate confirmed decisions from assumptions and open questions.
   - Keep ordinary handoffs around 80-140 lines. Allow longer plans only when the task genuinely needs reference-heavy detail.

3. Write the plan with this lean default structure.

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

4. Add optional sections only when they carry distinct next-session value.
   - `## Files to inspect first`: max 3-5 files or areas, each with a reason.
   - `## Validation`: only non-obvious focused checks, current validation blockers, or manual QA cues.
   - `## Open questions`: only questions that block or materially shape the next slice.
   - `## Scope`: only non-obvious in/out boundaries.
   - `## Decisions and rejected alternatives`
   - `## Risks and edge cases`
   - `## Reference notes`
   - `## Progress log`: only for actual completed work, blockers, or just-run validation that affects continuation.

5. Keep validation and history compact.
   - Omit `Validation` when it would only restate repository conventions or list obvious broad checks.
   - Include only commands known from the repository or discussion and likely to be useful for the next slice.
   - Prefer placeholders over expanded path lists, such as `bun run lint:files -- <changed-files>`.
   - Do not copy full command output into the plan.
   - Collapse or remove older progress detail once `Current state` captures what future sessions need.

6. Avoid predictable plan boilerplate.
   - Do not add a dedicated `Next command` section by default.
   - Include a concrete next command in the plan only when it is the real, non-obvious immediate action or unblocker.
   - Do not create empty sections.

## Final Response

Report only:

- created or updated file path
- recommended next command, if any
- open questions that should be answered before implementation, if any
