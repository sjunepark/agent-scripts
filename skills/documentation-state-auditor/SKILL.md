---
name: documentation-state-auditor
description: "Audit repository docs for confusion between implemented state, intended or planned state, and unresolved questions, then recommend documentation boundaries, structure, progressive disclosure, and writing rules."
---

# Documentation State Auditor

Use this skill when a repository's docs may blur:

- what exists today
- what the team intends to build
- what is still unresolved

The goal is not to enforce one documentation system. The goal is to make those kinds of truth legible with the lightest structure that fits the project, and to make the tree readable enough that readers can move from overview to detail without guesswork.

## Primary Questions

This skill should answer:

1. Can a new reader tell what is implemented now versus what is planned?
2. Which docs currently blur those boundaries?
3. Is the current organization helping or hurting clarity?
4. Does the directory tree itself reveal how the docs are organized and where a reader should start?
5. Can a reader move from broad overview to detailed docs progressively, only loading more detail when needed?
6. What documentation structure would fit this project with the least maintenance cost?
7. What concrete edits would most improve clarity right now?

## Core Principles

- Distinguish kinds of truth before choosing folder structure.
- Prefer structures that support progressive disclosure from overview to detail.
- Prefer visible prose and headings over hidden metadata when possible.
- Optimize for reader comprehension, not documentation purity.
- Let the methodology vary by project maturity, team size, and documentation volume.
- Prefer the smallest change that removes ambiguity.

## Kinds Of Truth

When reviewing docs, classify content into these buckets:

- `current`: factual present-state repository or runtime reality
- `target`: intended or accepted future-state design
- `question`: unresolved areas that should not be read as settled
- `history`: optional rationale or decision context, only when it still helps future readers

Do not assume every project needs a separate home for every bucket. Some projects need explicit directories. Others only need better section boundaries.

## Workflow

1. Inspect the repository shape.
- Start at the repo root.
- Read `AGENTS.md`, `ARCHITECTURE.md`, `README.md`, and the docs index if they exist.
- Load only the documentation files needed to understand the current documentation contract.
- List the files that appear to define architecture, workflows, design, or planning.
- Note whether the tree gives readers a legible path from overview docs into narrower topic or state-specific docs.

2. Reconstruct the actual documentation contract.
- Determine what the docs are trying to do today, even if the structure does not say it clearly.
- Identify whether the repository is pre-bootstrap, partially implemented, or mature.
- Note whether readers are likely to confuse target-state docs for implemented reality.

3. Audit for confusion signals.
- Look for present tense describing unbuilt systems.
- Look for future-state architecture mixed into current-state onboarding docs.
- Look for open questions embedded inside target docs.
- Look for directory names or file names that hide the distinction between current and planned.
- Look for trees where siblings mix overview docs, deep implementation notes, and planning artifacts without a visible organizing principle.
- Look for directories that do not tell a reader what gets more specific as they descend.
- Look for repeated rationale that does not help readers act.

4. Classify each important doc.
- For each important file, decide whether it is mainly `current`, `target`, `question`, or mixed.
- If mixed, identify whether the mix is acceptable or whether it needs to be split.

5. Recommend the lightest effective structure.
- Choose the simplest methodology that makes the distinctions clear.
- Possible outcomes include:
  - improved section conventions only
  - time-horizon directories such as `current/`, `target/`, and `questions/`
  - topic directories with explicit state sections inside each file
  - a hybrid, if the repository is large enough to justify it
- Recommend where overview or index documents should live so each branch of the tree makes sense before readers open leaf files.
- Explain why the chosen shape fits this specific repository.

6. Propose writing rules.
- Recommend tense and wording rules that keep the distinction visible in the body of the docs.
- Recommend where unresolved material should live.
- Recommend how much rationale should be retained.

7. End with concrete next actions.
- Name the smallest set of doc moves, renames, rewrites, or index changes that would materially reduce confusion.

## Evaluation Heuristics

Use these heuristics when judging documentation quality:

### Strong signals

- A new reader can answer "what exists now?" in one or two clicks.
- A new reader can answer "what are we trying to build?" separately.
- A new reader can understand the documentation structure by scanning the tree and the first doc at each branch.
- Parent docs give orientation, while child docs add detail rather than repeating or surprising.
- Unresolved items are clearly marked and isolated.
- Architecture docs do not quietly mix implemented behavior and aspirations.
- The top-level docs tell readers where to start.

### Risk signals

- The same file alternates between current implementation facts and future design without headings.
- The repository says it is early-stage, but architecture docs read as if the system already exists.
- "Open questions" docs contain settled decisions.
- The folder layout implies stable categories, but the writing inside ignores them.
- The tree forces readers to open several files before they can tell how docs are organized.
- Directory depth increases, but scope does not narrow predictably.
- Metadata or frontmatter carries the truth distinction, but the visible prose does not.

## Recommended Methodologies

Do not prescribe these mechanically. Pick one based on the repository.

### Section-first

Use when:

- the repository has few docs
- splitting files would add more friction than clarity
- a topic naturally contains both current and target information

Approach:

- keep topic docs
- add explicit sections such as `Current State`, `Target State`, and `Open Questions`
- ensure the file opens with enough orientation that a reader can decide whether to keep reading or move deeper elsewhere

Trade-offs:

- low directory churn
- easy to adopt incrementally
- depends on disciplined writing inside each file

### Time-horizon-first

Use when:

- readers are repeatedly confusing present state with planned state
- the repo is pre-bootstrap or early-stage
- there are already enough docs to justify structural separation

Approach:

- use directories such as `docs/current/`, `docs/target/`, and `docs/questions/`
- add short index or overview docs where readers need help choosing the right branch

Trade-offs:

- very clear for new readers
- can split one topic across several files
- usually the best default for planning-heavy repositories

### Topic-first with explicit status inside docs

Use when:

- the project is mature
- readers mostly navigate by subsystem rather than by timeline
- implementation and design need to live close together by topic

Approach:

- organize by topic such as `architecture/`, `database/`, and `workflows/`
- require visible section boundaries in mixed docs
- keep parent topic docs lightweight and directional, with deeper files carrying the narrower detail

Trade-offs:

- scales well with system complexity
- weaker by default at separating current and planned states
- needs stronger writing discipline

## Writing Rules

Recommend rules like these when they fit:

- Use present tense only for implemented reality.
- Use `will`, `should`, `target`, or equivalent language for intended state.
- Keep unresolved items in a clearly marked place rather than buried in target docs.
- At each directory level, make the scope narrower and more specific than the parent.
- Put orientation near branch points and detail in leaf docs.
- Keep rationale short and local unless the project truly benefits from decision-history documents.
- Prefer visible headings over hidden metadata for state distinctions.

## Output Shape

Use this shape unless the user asks for something else:

### Overall Assessment
- State whether the repository has low, medium, or high confusion risk.

### What Readers Will Misread
- Name the specific files or patterns likely to confuse new readers.

### Best-Fit Methodology
- Recommend the documentation methodology that best fits this repository.
- Explain what it gains and what it gives up.

### Reader Path
- Explain whether the current tree supports progressive disclosure.
- Recommend the smallest changes that make the navigation path legible.

### Concrete Changes
- List the smallest edits that would materially improve clarity.

### Writing Rules
- Give 3 to 6 practical rules the team can keep following.

## Communication Rules

- Prioritize clarity over doctrine.
- Avoid acting like one documentation pattern is universally correct.
- Tie every recommendation to the repository's current size, maturity, and risk of confusion.
- Treat progressive disclosure as a reader-comprehension tool, not a mandate for more folders.
- If the current structure is already good enough, say so directly and suggest only minimal refinements.
- If a proposed structure adds maintenance cost, call that out explicitly.

## Non-Goals

Do not:

- turn the review into a generic style-guide lecture
- insist on ADRs or frontmatter status fields by default
- recommend heavyweight process without evidence the project needs it
- confuse "more folders" with "better boundaries"
- assume progressive disclosure always means deeper nesting

## Example Triggers

- "Check whether these docs confuse current state and future state."
- "Audit this repository's docs organization."
- "Do these docs make it obvious what exists now?"
- "Should this project use `current/` and `target/` directories?"
- "Find places where design docs read like implemented reality."
- "Does this documentation tree support progressive disclosure?"
- "Can a reader understand this docs structure just by looking at the tree?"
