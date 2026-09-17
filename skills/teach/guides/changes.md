# Teaching Changes

Use for code or document diffs, commits, ranges, PR patches, and version
comparisons. Assume ordinary programming and Git knowledge but no project or
domain knowledge unless the learner says otherwise.

Apply the entry point's selection and redaction rules. Explain directly in the
conversation, keep the work read-only, and create no explanation artifact unless
requested. Describe excluded sensitive files only by category. Stay descriptive;
raise a correctness, regression, security, contract, or design concern only when
it changes what the reader should understand.

## Explain the change in context

Establish what the project or document does and where the changed area fits.
Read only the nearest overview, entry point, surrounding section, or architecture
context needed for that baseline. Show a realistic scenario that makes the old
behavior, the new behavior, and the practical consequence clear before explaining
the mechanisms.

Organize the lesson around the few ideas that explain the difference. Group a
large change by behavior rather than file, and give supporting tests, mechanical
edits, and migration detail only their necessary weight. Define project terms
before using them, introduce roles before identifiers, and attach source
locations to the claims they support.

For a tiny change, the baseline, example, difference, and consequence can fit in
a short answer. For a large change, lead with a brief takeaway, develop the
concrete story, and explain mechanisms and maintenance consequences progressively.
Shorten the primer for an expert while preserving old and new contracts.

Use a table, compact ASCII diagram, or narrowly scoped excerpt when it clarifies
a consequential relationship; follow the entry point's snippet and diagram
routes as needed. The reader should leave able to say what this area does,
what happened before, what happens now, why it matters, and which constraints
matter for future changes.
