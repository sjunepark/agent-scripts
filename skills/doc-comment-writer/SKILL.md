---
name: doc-comment-writer
description: "Add or improve doc comments in source files without obvious restatements. Use when the user asks to document code, add doc comments, explain modules, functions, or types for maintainers, add a useful file-level overview, or make a file understandable without tracing internals."
---

# Doc Comment Writer

Add doc comments that help the next developer understand purpose, contract, constraints, and non-obvious decisions at the point of use. Prefer comments that save future readers from reconstructing intent; skip comments that merely paraphrase names, types, or straightforward control flow.

## Workflow

1. Establish scope and comment style.
- Read the mentioned files first.
- Infer the language's normal doc comment form and follow the file's existing style.
- Inspect nearby types, tests, callers, or sibling files only when needed to understand public behavior or an important invariant.

2. Decide what actually needs documentation.
- Consider a file-level comment first when a file has a real module-level responsibility, boundary, or usage pattern that is not obvious from the filename and exports alone.
- Prioritize exported APIs, public modules, extension points, non-obvious helpers, and code with important invariants or tradeoffs.
- Add comments where a future reader would otherwise need to inspect internal logic to answer "what is this for?" or "what must stay true?"
- If a symbol is already obvious from its name, signature, and surrounding code, leave it undocumented.
- If a file does not need new doc comments, say so instead of forcing low-value edits.

3. Write for future readers, not for the current diff.
- Explain what the symbol does at a useful level of abstraction.
- Explain inputs, outputs, side effects, failure modes, lifecycle expectations, or invariants when they matter.
- Capture why a boundary exists, what a caller can rely on, and what would be easy to misuse.
- Mention tradeoffs or intentionally surprising behavior when that context will age well.
- Avoid describing step-by-step implementation details that will go stale after a refactor.

4. Keep comments lean and local.
- Keep file-level comments short and structural: explain the file's role, boundaries, or why related pieces live together.
- Prefer one strong doc comment over several weak inline comments.
- Keep wording tight; use full sentences only when they carry real information.
- Preserve existing formatting and keep edits close to the symbols they describe.
- Use inline comments only when a local invariant or subtle branch needs explanation beyond what a doc comment can carry.

5. Check the result against a high bar.
- Remove comments that only restate the name, type, or obvious return value.
- Remove generic filler such as "Helper function" or "Represents X" unless the next sentence adds real contract information.
- Ensure each new comment would still be useful if the reader never opened the function body.
- If you had to infer behavior from code rather than an explicit contract, phrase the comment conservatively.

## Writing Rules

- Prefer purpose, contract, invariants, and caller-relevant behavior.
- Use file-level comments only when they add context the exports and naming do not already provide.
- Omit obvious information the code already makes clear.
- Do not narrate internal control flow.
- Do not speculate about intent when the code does not support it.
- Do not document every symbol for symmetry; document the ones that benefit from it.
- Match the repository's tone and comment density.

## Good Targets

- File or module headers that explain responsibility, boundaries, integration points, or how to approach the file.
- Public functions whose names are clear but whose guarantees, side effects, or failure behavior are not.
- Types whose fields are simple but whose semantic meaning or lifecycle needs explanation.
- Internal helpers only when they encode a tricky invariant, protocol, or normalization rule.

## Poor Targets

- File headers that only restate the filename or say the file "contains utilities" without sharper guidance.
- Trivial getters, setters, and one-line wrappers with obvious names.
- Comments that duplicate parameter names or type annotations.
- Comments that only describe how a loop or conditional works.
- Large doc blocks copied from implementation details.

## Response Expectations

- Edit the mentioned files directly when the user asked for code changes.
- Keep the diff focused on documentation unless the user also asked for refactors.
- After editing, briefly report any file-level comments you added, what symbols you documented, what you intentionally left undocumented, and any places where the code itself remained too unclear to document confidently.
