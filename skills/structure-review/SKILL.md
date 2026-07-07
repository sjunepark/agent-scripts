---
name: structure-review
description: "Review code structure for unnecessary complexity and insufficient organization. Use when asked whether code is overengineered (speculative abstractions, extra DB columns/fields/helpers, should be simplified) or under-organized (too flat, mixes responsibilities, should be reorganized)."
---

# Structure Review

Review code with two questions in mind: does each piece of structure earn its keep right now, and is there missing structure that would make the code easier to navigate and understand?

Over-structure (speculative abstractions, premature genericity) hurts because it obscures the real flow. Under-structure (flat directories, mixed responsibilities, everything in one file or one level) hurts because it forces readers to scan unrelated code to find what they need.

## Workflow

1. Anchor the review in the real code.
- Read the changed files, nearby interfaces, and affected tests or docs.
- Distinguish between:
  - necessary structure serving current behavior
  - readability-oriented structure that pays for itself now
  - speculative or low-value structure added mostly for hypothetical future use
- Done when every changed file is read and classified into one of these categories.

1. Look for over-structure signals.
- Extra DB columns, JSON fields, DTO properties, config keys, or hooks that current behavior barely touches.
- Helpers or wrappers that add another jump without improving naming, reuse, invariants, or boundaries.
- Callers still needing to know internal details the abstraction claimed to hide.
- Extension points, strategy objects, plugin hooks, or flags with no concrete second use in sight.
- Mapping objects or translation layers that exist purely to preserve genericity.
- Multiple layers of indirection around a concept that still has only one concrete path.
- Options, enums, or mode switches where only one case is valid in practice.
- State or persistence added before a concrete consumer, producer, or invariant needs it.
- Module splits or directory nesting that increase hunting rather than understanding.
- The current feature became harder to follow because the design optimized for hypothetical later work.

Done when every changed file and directory has been checked against each signal.

1. Look for under-structure signals.
- A directory has 10+ files spanning three or more unrelated concerns with no subdirectory grouping.
- A module's public surface mixes capabilities a caller must mentally filter to find the one they need.
- Files that always change together live far apart in the tree while unrelated files are adjacent.
- The filesystem tree does not communicate the project's major concerns; a new reader cannot predict where to look.
- A single-level directory could be split into two or three subdirectories that each name a clear responsibility.
- Coding agents tend to default to flat layouts; treat a flat structure as a smell worth investigating, not as automatically correct.

Done when every changed file and directory has been checked against each signal.

1. Check whether the structure earns its keep.
Ask what the extra layer buys now:
- clearer names or call sites
- a stronger type boundary
- a sharper ownership boundary
- reduced repeated branching or glue code
- simpler invariants
- isolation of ugly external details
- clearer file or directory-level comprehension

If the answer is weak or hypothetical, treat it as a simplification candidate.

1. Keep the change bar high.
- Do not recommend removal just because something is abstract.
- Do not recommend nesting just because a directory has many files.
- Do not punish code for being explicit.
- Do not treat readability-oriented structure as waste.
- Recommend simplification only when the extra structure adds recurring maintenance cost, obscures the real flow, or mainly exists for futures the code does not currently need.
- Recommend reorganization only when the flat layout forces readers to scan unrelated code regularly or when mixed responsibilities create real confusion about where a concern lives.

Signs to keep the current structure:

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

- Default to the smallest useful recommendation.
- If the code is already appropriately lean, say so plainly.
- If a concern is plausible but weakly supported, label it as a watch item rather than a recommendation.
- Do not manufacture simplification opportunities just to make the review feel useful.
