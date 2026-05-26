---
name: structure-review
description: "Review code structure for unnecessary complexity and insufficient organization. Use when asked whether code is overengineered, too abstract, too flat, poorly organized/configurable/future-proof, has extra DB columns/fields/helpers, mixes responsibilities, feels heavier than needed, or should be simplified or reorganized while preserving readability."
---

# Structure Review

Review code with two questions in mind: does each piece of structure earn its keep right now, and is there missing structure that would make the code easier to navigate and understand?

This skill reviews in both directions. Over-structure (speculative abstractions, premature genericity) hurts because it obscures the real flow. Under-structure (flat directories, mixed responsibilities, everything in one file or one level) hurts because it forces readers to scan unrelated code to find what they need.

Do not flatten or remove structure that materially improves readability, naming, local reasoning, or ownership boundaries. A single-use helper, nested directory, or small composed object can be worth keeping when it makes the code easier to understand.

## Workflow

1. Anchor the review in the real code.
- Read the changed files, nearby interfaces, and affected tests or docs.
- Distinguish between:
  - necessary structure serving current behavior
  - readability-oriented structure that pays for itself now
  - speculative or low-value structure added mostly for hypothetical future use

2. Look for over-structure signals.
- Extra DB columns, JSON fields, DTO properties, or config keys that current behavior does not meaningfully use.
- Helpers or wrappers that mostly forward calls without improving naming, boundaries, or reuse.
- Extension points, strategy objects, plugin hooks, or flags added for imagined future cases.
- Generic abstractions that force repeated mapping or translation without reducing real duplication.
- Multiple layers of indirection around a concept that still has only one concrete path.
- Options, enums, or mode switches where only one case is valid in practice.
- State or persistence added before there is a concrete consumer, producer, or invariant that needs it.
- Module splits or directory nesting that increase hunting rather than understanding.

3. Look for under-structure signals.
- A directory mixes files with clearly different responsibilities (e.g., data access, UI, validation, and orchestration all at one level with no grouping).
- A module or file handles multiple unrelated capabilities that a reader must mentally separate.
- Flat file lists grow long enough that scanning them requires domain knowledge to find a given concern.
- Related files that change together are scattered across sibling directories rather than co-located.
- A single-level directory could be split into two or three subdirectories that each name a clear responsibility, making the structure self-documenting.
- Coding agents tend to default to flat layouts; treat a flat structure as a smell worth investigating, not as automatically correct.

4. Check whether the structure earns its cost.
Ask what the extra layer buys now:
- clearer names or call sites
- a stronger type boundary
- a sharper ownership boundary
- reduced repeated branching or glue code
- simpler invariants
- isolation of ugly external details
- clearer file or directory-level comprehension

If the answer is weak or hypothetical, treat it as a simplification candidate.

5. Keep the change bar high.
- Do not recommend removal just because something is abstract.
- Do not recommend nesting just because a directory has many files.
- Do not punish code for being explicit.
- Do not treat readability-oriented structure as waste.
- Recommend simplification only when the extra structure adds recurring maintenance cost, obscures the real flow, or mainly exists for futures the code does not currently need.
- Recommend reorganization only when the flat layout forces readers to scan unrelated code regularly or when mixed responsibilities create real confusion about where a concern lives.

## Review Standard

Treat these as strong signals that code may be too heavy:

- A new field, column, option, or hook was added for future flexibility, but current behavior barely touches it.
- A helper or abstraction adds another jump without improving naming, reuse, invariants, or boundaries.
- Callers still need to know internal details the abstraction claimed to hide.
- One concept now requires extra mapping objects or translation layers purely to preserve genericity.
- A single implementation is wrapped in extension machinery that has no concrete second use in sight.
- The current feature became harder to follow because the design optimized for hypothetical later work instead of today's path.

Treat these as strong signals that code may be too flat or disorganized:

- A directory has 10+ files spanning three or more unrelated concerns with no subdirectory grouping.
- A module's public surface mixes capabilities a caller must mentally filter to find the one they need.
- Files that always change together live far apart in the tree while unrelated files are adjacent.
- The filesystem tree does not communicate the project's major concerns; a new reader cannot predict where to look.

Treat these as signs to keep the current structure:

- A single-use helper materially improves the signature, name, or readability of the call site.
- A small object or composed value groups related data in a way that makes invariants easier to see.
- Nested modules or directories help a reader predict where responsibility lives.
- A wrapper isolates awkward dependency details or side effects behind a clearer seam.
- A little duplication is cheaper than a shared abstraction, but the current structure is still locally clear.
- A flat directory has few files and each file's name clearly signals its purpose; adding subdirectories would just add clicks.

## Output

Use this structure when reporting:

### Findings
- List only concrete structural issues supported by the code, whether over-structure or under-structure.
- For each issue, explain why the current shape does not appear to earn its cost or why missing structure hurts navigation.
- Support each issue with a short embedded snippet or directory listing when the code itself is central to the point.

### Keep As-Is
- Call out choices that may look heavy or flat at first glance but are worth keeping.
- Say what benefit they provide now.

### Restructure Candidates
- Include only changes that clear the bar above.
- For each one, state:
  - the smallest reasonable change (simplify, reorganize, split, or merge)
  - what it would remove, collapse, or regroup
  - what would be gained
  - what could be lost
  - why now is or is not the right time

### Open Questions
- List missing constraints or future requirements that could justify the current shape.

### Verdict
- End with one of:
  - `Structure is appropriate`
  - `Small restructure is justified`
  - `Broader reorganization may help, but only after clarifying constraints`

## Snippet Rules

- Prefer embedded snippets over editor-style line references.
- Put the source file path on the first line of each snippet as a comment.
- Keep snippets tight: signatures, schema definitions, wrappers, config shapes, and translation layers are usually enough.

## Communication Rules

- Be direct and specific.
- Default to the smallest useful recommendation.
- Separate readability wins from speculative complexity; they are not the same thing.
- If the code is already appropriately lean, say so plainly.
- If a concern is plausible but weakly supported, label it as a watch item rather than a recommendation.
- Do not manufacture simplification opportunities just to make the review feel useful.

## Example Triggers

- "Is this overengineered?"
- "Do we actually need these extra fields and helpers?"
- "Can you review this diff for speculative code or unnecessary abstractions?"
- "I want this implementation to stay lean. What should we remove, if anything?"
- "Check whether this new schema and helper layer is earning its keep."
- "This directory is getting hard to navigate. Should we reorganize?"
- "Is this too flat? Should we add subdirectories?"
- "Review the file/module organization for mixed responsibilities."
