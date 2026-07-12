---
name: doc-comment-writer
description: "Add or improve durable, maintainer-facing doc comments without restatements or brittle details. Use for in-place source documentation of files, modules, functions, types, contracts, invariants, or non-obvious behavior."
---

# Doc Comment Writer

Add durable doc comments that help future maintainers understand purpose, contract, constraints, and non-obvious decisions without reconstructing them from implementation. A restatement merely paraphrases names, types, or straightforward control flow; omit it.

## Workflow

1. Establish scope and local conventions.
   - Follow the user's stated scope. Infer the language's normal doc comment form and the file's existing style and comment density.
   - Inspect nearby types, tests, callers, or sibling files only when needed to understand public behavior or an important invariant.
   - Continue once the files and symbols in scope and their applicable comment syntax and style are clear.

2. Make an explicit document-or-skip decision.
   - Consider a file-level comment when the file has a responsibility, boundary, or usage pattern that is not obvious from its name and exports.
   - Add comments where a future reader would otherwise need to inspect internal logic to answer "what is this for?" or "what must stay true?"
   - Leave symbols undocumented when their names, signatures, types, and context already answer those questions. If the file needs no new comments, say so instead of forcing edits.
   - Continue once every exported or public symbol in scope, plus any relevant file header, has an explicit document-or-skip decision.

3. Write and validate durable comments.
   - Apply every rule in Comment Standard to each new or revised comment.
   - Use inline comments only when a local invariant or subtle branch needs explanation beyond what a doc comment can carry.
   - Run the narrowest relevant formatter, parser, lint, documentation, or example check. If none is available, inspect the changed comment syntax and report the limitation.
   - Complete once every changed comment meets the standard, uses valid local syntax, and relevant validation passes or its limitation is reported.

## Comment Standard

- Explain guarantees, side effects, failure modes, lifecycle expectations, invariants, boundaries, misuse risks, or durable tradeoffs when they matter.
- Keep file-level comments short and structural. Prefer one strong doc comment over several weak comments.
- Keep code, types, and configuration as the source of truth for change-prone facts such as numeric thresholds, timeouts, limits, defaults, constant values, member lists, filenames, and paths. Describe the semantic rule instead of copying literals or lists into prose; name the owning symbol only when the reader needs it. If an unexplained literal needs a name, report the code clarity problem rather than documenting around it.
- Treat external references as exceptional. Do not cite temporary or weakly maintained artifacts such as plans, progress files, handoff notes, temporary decision logs, or issue and pull-request threads. Put durable rationale in the comment itself; cite a document only when it is an authoritative, maintained specification or contract with a stable location.
- Avoid time-relative wording and step-by-step implementation details that will become stale after routine changes.
- Phrase behavior inferred from code conservatively; do not invent intent or guarantees the implementation does not support.
- Remove generic filler and restatements. Each comment must remain useful without opening the documented symbol's implementation.

## Targeting Guide

Good targets include file or module boundaries, public guarantees and failure behavior, type semantics and lifecycle, and internal helpers that enforce a tricky invariant, protocol, or normalization rule.

Skip file headers that merely restate the filename, trivial accessors and wrappers, parameter or type narration, straightforward control flow, and large blocks copied from implementation details.

## Response Expectations

- Keep the diff focused on documentation unless the user also asked for refactors.
- After editing, briefly report the files or symbols documented, notable skip decisions, validation, and any code that remained too unclear to document confidently.
