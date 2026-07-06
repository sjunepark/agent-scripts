---
name: doc-comment-writer
description: "Add or improve doc comments in source files without obvious restatements. Use when the user asks to document code, explain modules, functions, or types for maintainers, or add a useful file-level overview."
---

# Doc Comment Writer

Add doc comments that help the next developer understand purpose, contract, constraints, and non-obvious decisions at the point of use. Prefer comments that save future readers from reconstructing intent; skip restatements — comments that merely paraphrase names, types, or straightforward control flow.

## Workflow

1. Establish scope and comment style.
- Infer the language's normal doc comment form and follow the file's existing style and comment density.
- Inspect nearby types, tests, callers, or sibling files only when needed to understand public behavior or an important invariant.

2. Decide what actually needs documentation.
- Consider a file-level comment first when a file has a real module-level responsibility, boundary, or usage pattern that is not obvious from the filename and exports alone.
- Prioritize the Good Targets and skip the Poor Targets listed below.
- Add comments where a future reader would otherwise need to inspect internal logic to answer "what is this for?" or "what must stay true?"
- If a symbol is already obvious from its name, signature, and surrounding code, leave it undocumented.
- If a file does not need new doc comments, say so instead of forcing low-value edits.
- Done when every exported or public symbol in scope, plus the file header, has an explicit document-or-skip decision.

3. Write for future readers, not for the current diff.
- Explain inputs, outputs, side effects, failure modes, lifecycle expectations, or invariants when they matter.
- Capture why a boundary exists, what a caller can rely on, and what would be easy to misuse.
- Mention tradeoffs or intentionally surprising behavior when that context will age well.
- Avoid describing step-by-step implementation details that will go stale after a refactor.

4. Keep comments lean and local.
- Keep file-level comments short and structural: explain the file's role, boundaries, or why related pieces live together.
- Prefer one strong doc comment over several weak inline comments.
- Keep wording tight; use full sentences only when they carry real information.
- Use inline comments only when a local invariant or subtle branch needs explanation beyond what a doc comment can carry.

5. Check the result against a high bar.
- Remove any restatement that survived — a comment that only echoes the name, type, or obvious return value.
- Remove generic filler such as "Helper function" or "Represents X" unless the next sentence adds real contract information.
- Ensure each new comment would still be useful if the reader never opened the function body.
- If you had to infer behavior from code rather than an explicit contract, phrase the comment conservatively; do not speculate about intent the code does not support.

## Good Targets

- File or module headers that explain responsibility, boundaries, integration points, or how to approach the file.
- Public functions whose names are clear but whose guarantees, side effects, or failure behavior are not.
- Types whose fields are simple but whose semantic meaning or lifecycle needs explanation.
- Internal helpers only when they encode a tricky invariant, protocol, or normalization rule.

## Poor Targets

- File headers that only restate the filename or say the file "contains utilities" without sharper guidance.
- Trivial getters, setters, and one-line wrappers with obvious names.
- Restatements of parameter names, type annotations, or control flow.
- Large doc blocks copied from implementation details.

## Response Expectations

- Keep the diff focused on documentation unless the user also asked for refactors.
- After editing, briefly report any file-level comments you added, what symbols you documented, what you intentionally left undocumented, and any places where the code itself remained too unclear to document confidently.
