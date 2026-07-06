---
name: architecture-md-writer
description: "Create, update, review, or split ARCHITECTURE.md files covering component maps, runtime flow, code maps, and invariants. Use when architecture docs are missing, stale, too detailed, or need nested subsystem docs linked from the root."
---

# ARCHITECTURE.md Writer

Write architecture docs that give contributors a fast mental model of system shape, starting points, and non-negotiable constraints. Prefer stable structure over implementation trivia, and prefer editing existing `ARCHITECTURE.md` files over creating new ones unless the scope is wrong.

## Workflow

1. Determine scope and document topology.
- Inspect whether the task is for the repo root, a subsystem, or both.
- Treat the root `ARCHITECTURE.md` as the system map, not the full encyclopedia.
- Add nested `ARCHITECTURE.md` files only when a subtree has enough independent structure, lifecycle, or constraints to justify its own page.
- When adding a nested doc, update the parent doc to link to it and explain its scope in one line.
- Read [references/progressive-disclosure.md](references/progressive-disclosure.md) when deciding where to split content.

2. Build context from stable code structure.
- Read existing docs only to recover boundaries, terminology, and decision history.
- Inspect entry points, package boundaries, long-lived directories, interfaces, background jobs, storage seams, and external integrations.
- Trace one or two important flows end to end so the document reflects how the system actually moves.
- Capture responsibilities, interactions, and invariants before drafting prose.

3. Draft the smallest useful document.
- Explain what the system or module is for, what major parts exist, and how they interact.
- Prefer concrete file and directory starting points over exhaustive inventories.
- Use a small ASCII diagram only when it shortens explanation; do not use Mermaid.
- Choose section names that fit the system instead of forcing a rigid template.
- Read [references/section-patterns.md](references/section-patterns.md) when choosing sections or shaping a component map.

4. Keep the document high-level and durable.
- Omit class-by-class or function-by-function catalogs.
- Omit setup steps, coding standards, runbooks, and tutorials unless a design constraint depends on them.
- Link ADRs for rationale and tradeoff history instead of embedding long decision logs.
- If an area is unclear, say less and point to the best start-here files rather than speculating.

5. Review for contributor usefulness.
- Ensure a new engineer can answer: what are the major parts, where should I start tracing behavior, and what must not be broken?
- Remove statements that would go stale after an ordinary refactor.
- When updating an existing doc, check every remaining claim against the current code; fix or delete any that no longer hold.
- Verify links, file paths, and parent-child architecture doc references.
- Preserve repository terminology so readers can jump from document to code without translation.

## Writing Rules

- Prefer short paragraphs and flat bullets.
- Write from outside in: system shape first, then the most important flows and files.
- Call out `start here` files explicitly.
- Name invariants as enforceable rules, not vague values.
- When a module doc exists, state how it fits into the parent architecture and link back upward.
- Treat missing ADRs as a fact to note briefly, not a reason to invent rationale.
- Favor searchable headings and concrete paths over narrative filler.
