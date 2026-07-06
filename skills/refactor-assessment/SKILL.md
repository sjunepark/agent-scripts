---
name: refactor-assessment
description: "Assess whether an applied or proposed fix, review finding, diff, PR, or file should stay local or become a broader refactor. Use for design triage."
---

# Refactor Assessment

Pressure-test whether a local fix reveals a broader refactor opportunity. Start skeptical: recommend broader refactoring only when evidence shows it leaves the design net-simpler.

If the target is ambiguous, infer it; ask only if multiple plausible targets would materially change the recommendation.

## Workflow

1. Summarize the related issue or fix.
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
   - If the issue is only a local mistake, say so.

3. Decide whether a broader refactor is justified.
   - Prefer one best recommendation. Include a second only if clearly distinct and worth considering now.
   - Good recommendations may delete, inline, collapse, rename, move ownership, or tighten a type/schema.
   - Do not equate broader refactoring with adding helpers, adapters, services, interfaces, frameworks, compatibility layers, speculative extension points, or broad rewrites.
   - Pick one verdict: `Recommended now`; `Defer` (justified but not worth doing now); `Decision needed` (justified but hinges on a user tradeoff); `No broader refactor recommended` (say so and explain why).

4. Scope the recommendation when a broader refactor is justified.
   - Explain what boundary, invariant, or ownership concern improves.
   - Name the affected code and what happens to it.
   - Explain why it beats keeping only the local fix.
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
