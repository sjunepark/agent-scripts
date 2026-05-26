---
name: learning-material-writer
description: "Create or update repo learning material under docs/learning for programmers new to the codebase. Use only when the user explicitly asks for learning material, onboarding docs, a codebase guide, repo handbook, or docs/learning content. Treat it as onboarding guidance, not the source of truth for behavior or decisions."
---

# Learning Material Writer

Write concise learning material that helps a programmer become fluent in an unfamiliar repository. Use the same teaching priorities as `teach`: build the mental model first, explain roles and boundaries before low-level details, use small snippets only when they genuinely clarify the point, and add compact ASCII diagrams only when they reduce explanation.

The difference is output shape: instead of replying with one lesson in chat, create a navigable document set inside `docs/learning`.

Learning material is for orientation and onboarding. It must not present itself as the repository's authoritative source for implementation truth, product decisions, or development policy.

## Defaults

- Default scope to the whole repository.
- Respect any user-provided focus areas, audiences, exclusions, or terminology.
- Assume the reader is new to this repo, but not new to programming.
- Write fast-reading docs: concise, concrete, and skimmable.
- Prefer stable design explanation over implementation trivia.
- Treat learning docs as secondary guidance that points readers toward more authoritative sources when needed.

## Documentation Status

- Make it explicit that `docs/learning` is for onboarding and reader understanding.
- Do not frame learning docs as the source of truth for implementation behavior, decision-making, or project policy.
- In `docs/learning/INDEX.md`, include a short note telling readers to defer to more authoritative sources when accuracy or current behavior matters.
- If the repository already has a suitable project-level surface such as `AGENTS.md`, `README.md`, `docs/INDEX.md`, or another docs index, update that surface when needed so the role of `docs/learning` is obvious there too.
- Prefer short notes in existing index, documentation-role, or contribution sections over adding a large new policy block.
- Do not create a brand-new top-level file just to carry this note unless the user asked for that.
- Point readers to the best available canonical sources for the repo, such as source code, tests, architecture docs, ADRs, API contracts, or product requirements.

## Workflow

1. Establish scope and current state.
- Read the repo root, nearby docs, and existing `docs/learning` if present.
- If the user did not narrow the task, cover the whole repo.
- Identify the main entry points, top-level directories, major modules, important flows, and architecture seams.

2. Build a teaching map before writing files.
- Decide what a new reader must understand first, second, and later.
- Group material by learning need, not by mirroring the source tree blindly.
- Prefer a small number of strong topics over a sprawling document dump.
- Make the planned folder names and file names read like a table of contents.

3. Create a clear hierarchy under `docs/learning`.
- Use `docs/learning/INDEX.md` as the root entry point.
- Use unnumbered, descriptive directories and filenames.
- Add `INDEX.md` inside directories that contain multiple child pages.
- Keep hierarchy shallow unless the codebase genuinely needs more depth.
- Cross-link parent, child, and related pages so readers can move through the material intentionally.

4. Teach in learning order.
- Start with what the system is for and how the major pieces fit together.
- Then explain runtime flows, module responsibilities, boundaries, and design decisions.
- Then cover deeper subsystem pages only where the repo has enough complexity to justify them.
- Reuse the `teach` style: purpose first, then main flow, then key mechanisms, then confusing parts.

5. Keep docs durable and readable.
- Prefer responsibilities, relationships, invariants, and tradeoffs.
- Avoid file-by-file catalogs unless a code map page specifically needs them.
- Avoid line-by-line walkthroughs.
- Avoid copying large source snippets into docs.
- If a rationale is inferred rather than explicit, label it as an inference.

## Required Coverage

The document set should normally cover these areas, adapted to the repo:

- big-picture system purpose and shape
- codebase structure and major directories
- major modules or subsystems and each one's responsibility
- how modules interact with each other
- important runtime or data flows
- boundary lines: ownership, APIs, persistence, external integrations, layering, or other seams
- key design decisions, tradeoffs, and constraints that help the structure make sense
- confusing or non-obvious areas that a newcomer should notice early

If some area does not meaningfully exist in the repo, omit it rather than forcing a template.

## Topology Rules

Treat the docs like a short book, not a dump of notes.

- The directory tree should read like a TOC.
- Folder names should describe topics, not internal chores.
- Put overview pages above detail pages.
- Split into subdirectories only when a topic has multiple natural child pages.
- Prefer topic groupings such as `system-overview`, `runtime-flows`, `module-boundaries`, `subsystems`, or `design-decisions` when they fit the repo.
- Do not force those exact names if the repo suggests clearer ones.

Good patterns:

```text
docs/learning/
  INDEX.md
  system-overview.md
  codebase-map.md
  runtime-flows/
    INDEX.md
    request-lifecycle.md
    background-jobs.md
  subsystems/
    INDEX.md
    auth.md
    billing.md
  design-decisions/
    INDEX.md
    state-management.md
```

```text
docs/learning/
  INDEX.md
  product-and-architecture.md
  how-data-moves/
    INDEX.md
    ingestion.md
    serving.md
  major-parts/
    INDEX.md
    api-layer.md
    storage-layer.md
```

Bad patterns:

- one page per source file
- directory names that only mirror `src/` without explaining why those parts matter
- deep nesting with thin content
- piles of isolated notes with no index page or reading path

## Writing Rules

- Optimize for fast understanding.
- Prefer short sections, flat bullets, and direct prose.
- Explain why a module exists before describing how it works.
- Name who owns what and where decisions happen.
- Explain interfaces and seams in plain language.
- Use snippets sparingly and only when a small example makes a boundary or flow easier to grasp.
- Use ASCII diagrams only when they shorten the explanation; do not use Mermaid.
- Keep tone instructional, calm, and matter-of-fact.
- Do not add exercises, quizzes, or checkpoints.
- Do not pad pages with generic onboarding advice.

## Page Design

Use a page shape that helps scanning. Adapt section names as needed, but most pages should answer some of these quickly:

- what this part is for
- where it sits in the system
- what it talks to
- what it owns
- what assumptions or invariants matter
- what design tradeoffs or confusing edges are worth knowing

Good page openings:
- one short paragraph that orients the reader
- a compact bullet list of responsibilities or relationships
- a tiny diagram if that is the shortest way to show structure

## Root Index Expectations

`docs/learning/INDEX.md` should usually include:

- what this learning set is for
- a short note that it is onboarding material, not the authoritative source of truth for implementation or decisions
- the recommended reading path
- a short map of the major topic pages and directories
- one-line descriptions for each linked page

Directory `INDEX.md` pages should do the same thing for their local section.

## Editing Existing Learning Docs

If `docs/learning` already exists:
- preserve useful structure and terminology unless it blocks clarity
- refactor weak hierarchy when needed so navigation becomes clearer
- merge redundant pages
- add missing index pages or cross-links
- add or tighten the onboarding-only / not-source-of-truth note if it is missing or vague
- keep filenames stable unless renaming materially improves the table-of-contents effect

## Non-Goals

Do not:
- write long-form tutorials or exercises
- mirror every implementation detail in the repo
- present learning docs as normative project policy or the decision record
- create architecture docs outside `docs/learning` unless the user asked for that too
- turn this into a code review or rewrite plan
- speculate about intent when the code does not support it

## Response Expectations

When doing the work:
- create or update files directly under `docs/learning`
- when a suitable repo-level documentation surface already exists, update it too if needed so the onboarding-only role of `docs/learning` is explicit
- briefly explain the chosen hierarchy before or after the edits
- call out any important gaps where code structure was too unclear to document confidently
