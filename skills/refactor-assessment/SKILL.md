---
name: refactor-assessment
description: "Assess whether an applied issue, bug fix, review finding, diff, PR, file, or proposed fix should stay local or become a broader refactor. Use for design triage focused on clearer boundaries, invariants, ownership, and simplicity."
---

# Refactor Assessment

Pressure-test whether a local fix reveals a broader refactor opportunity. Start skeptical: recommend broader refactoring only when evidence shows it will make the design simpler, clearer, safer, or easier to maintain.

If the target is missing or ambiguous, infer it from the conversation, recent edits, review findings, failing tests, or referenced files. Ask only if multiple plausible targets would materially change the recommendation.

## Workflow

1. Summarize the related issue or fix.
   - Name relevant files, functions, components, schemas, or tests when useful.
   - Separate confirmed facts from assumptions.

2. Classify the design signal.
   - Choose the best fit:
     - simple local mistake
     - weak validation or invariant gap
     - unclear ownership or boundary problem
     - weak type or schema model
     - state, lifecycle, concurrency, or ordering hazard
     - over-abstraction, under-abstraction, or duplicated logic
     - brittle integration or contract mismatch
   - If the issue is only a local mistake, say so and avoid inventing a refactor.

3. Decide whether a broader refactor is justified.
   - Prefer one best recommendation. Include a second only if clearly distinct and worth considering now.
   - Good recommendations may delete, inline, collapse, rename, move ownership, or tighten a type/schema.
   - Do not equate broader refactoring with adding helpers, adapters, services, interfaces, frameworks, compatibility layers, speculative extension points, or broad rewrites.
   - Prefer making the current design simpler and more explicit.
   - If no broader refactor is justified, say `No broader refactor recommended` and explain why.

4. Scope the recommendation when a broader refactor is justified.
   - Explain what boundary, invariant, or ownership concern improves.
   - Name what code would move, collapse, rename, tighten, inline, or delete.
   - Explain what gets simpler, why it beats keeping only the local fix, why it is not over-abstraction, expected benefit, risk, effort, and validation.
   - Do not apply code changes unless the user explicitly asks for implementation. Mention tiny obvious cleanups separately if relevant.

## Output

```markdown
## Related issue or fix

## Design signal

## Refactor recommendation

Recommended now | Defer | Decision needed | No broader refactor recommended

### Recommendation, if a broader refactor is justified

### Why this is not over-abstraction, if recommending a refactor

### Scope and affected files, if recommending a refactor

### Benefit, risk, and effort, if recommending a refactor

### Validation, if recommending a refactor

## Alternative considered, if any

## Recommended next step
```

Keep the goal narrow: design triage for a clearer system, not more architecture.
