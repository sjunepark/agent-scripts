---
name: bug-retro
description: "Analyze a bug, failure, regression, or failing test as an engineering postmortem. Use when identifying root cause, design smell, prevention refactors, fixes already applied, and remaining recommended work without editing files unless the user explicitly asks."
---

# Bug Retro

Analyze bugs as concise engineering retrospectives, not just fix summaries. Focus on cause, design signal, robust prevention, and what should happen next.

If the target is missing or ambiguous, infer it from the current conversation, failing test, error output, recent edits, or referenced files. Ask only when multiple plausible targets would materially change the analysis.

## Workflow

1. Establish the failure.
   - State the visible symptom, expected behavior, and broad root cause.
   - Explain the causal chain conceptually.
   - Name exact files, functions, logs, schemas, or tests only when they are essential evidence.
   - Distinguish confirmed facts from plausible hypotheses.

2. Classify the design signal.
   - Choose the best fit:
     - simple local mistake
     - weak validation or invariant gap
     - unclear ownership or boundary problem
     - weak type or schema model
     - state, lifecycle, concurrency, or ordering hazard
     - over-abstraction, under-abstraction, or duplicated logic
     - brittle integration or contract mismatch
   - Explain why the classification fits.
   - Note design issues that made the bug easy to introduce or hard to detect.

3. Recommend about three robustness improvements.
   - Prefer one immediate fix, one small hardening/refactor, and one larger design direction only if it is genuinely worth discussing.
   - Label each solution heading with exactly one status:
     - `[applied]` for fixes or refactors already made or applied during the response.
     - `[recommended now]` for next work that is not yet implemented.
     - `[deferred]` for worthwhile work that should not be done now.
     - `[discussion only]` for ideas needing product, domain, or maintainability decisions.
   - Explain what invariant or boundary each option strengthens, how it prevents the bug class, benefits, tradeoffs, and effort/risk/payoff.
   - Prefer making invalid states unrepresentable over defensive patches.

4. Keep the retro read-only unless edits are explicitly requested.
   - Do not edit files, stage, commit, or push when the user asks only for analysis, a retro, or recommendations.
   - If the user explicitly asks for fixes too, apply only obvious low-risk fixes within scope.
   - If tradeoffs require judgment, ask for a decision.

## Output

```markdown
## Cause

## Mistake vs. design signal

## Robustness improvements

### 1. [applied|recommended now|deferred|discussion only] <solution heading>

### 2. [applied|recommended now|deferred|discussion only] <solution heading>

### 3. [applied|recommended now|deferred|discussion only] <solution heading, optional if only two are warranted>

## Applied patches to keep, if any

## Remaining recommended work, if any

## Recommended next step

## Open questions, if any
```

Be direct and specific. Avoid vague advice like "add more tests" unless naming the exact behavior or invariant to protect.
