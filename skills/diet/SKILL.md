---
name: diet
description: "Manual standalone lens for unearned code weight; prefer code-review as the automatic review entry point."
disable-model-invocation: true
---

# Diet

Review code for **unearned weight**.

The core question is:

> What in this code does not earn its maintenance cost today?

This skill is not a general architecture or folder-layout review. It is a simplification review for code that feels heavier than the problem requires: too many helpers, wrappers, schema fields, columns, options, generic layers, or appended paths that solve the task by adding more machinery instead of integrating with the existing design. Do not drift into folder, directory, or module reorganization unless that is central to the weight problem or the user explicitly asked for it.

Do not assume code came from an LLM or agent. If the user mentions vibe-coded or agent-written changes, treat that only as context for the review lens, not as evidence.

Do not force deletions. Explicit, readable, well-bounded code is often worth keeping even when it is not minimal.

The goal is not fewer lines. The goal is fewer obligations: fewer concepts, states, branches, compatibility promises, and surfaces to preserve.

## Review Posture

Treat these principles as guides, not dogma:
- **YAGNI**: do not keep machinery only for hypothetical future cases.
- **AHA / rule of three**: avoid hasty abstractions; duplicate twice before extracting the wrong shared layer.
- **Duplication can be cheaper than the wrong abstraction.**
- **Prefer integration over appending**: solve the problem in the existing design when possible instead of adding side channels, one-off adapters, or parallel flows.
- **Keep readability-oriented explicitness**: a small helper, wrapper, or object can be worth its cost when it clarifies naming, boundaries, or invariants.

## Judgment Guardrails

- First separate **essential complexity** from **accidental complexity**. Essential complexity comes from the domain, external contracts, data lifecycle, failure modes, security, performance, or user workflow. Accidental complexity comes from implementation machinery that current behavior does not require.
- Before recommending removal, identify what would stop working, what tests or contracts would need to change, and whether the behavior is public, persisted, migrated, or used by another subsystem. If local evidence cannot answer that, put the item in Bucket II.
- Treat compatibility layers, migration fields, fallbacks, aliases, and dual-read or dual-write paths as justified only when there is a current rollout, external contract, or documented migration window. Otherwise, they are usually Bucket II candidates.
- Prefer simplification in this order — the simplification ladder: delete unused or speculative code; inline pass-through wrappers; make generic code specific to the current use; merge appended paths into the existing design; split diverging cases instead of adding flags or modes; redesign only when local simplification would leave the underlying weight intact.
- A longer direct implementation can be leaner than a shorter abstraction when it creates fewer concepts, jumps, configuration paths, and preserved states.

## Workflow

1. Anchor the review in the real code.
- Read the changed files, nearby interfaces, and affected tests or docs when relevant.
- Distinguish between:
  - code that is truly serving current behavior
  - code that mainly exists for hypothetical future flexibility
  - code that was appended as a workaround instead of integrated cleanly

1. Ask what each extra piece buys **now**.
- For every helper, wrapper, field, column, option, mode, schema entry, or abstraction layer, ask:
  - what current behavior depends on this?
  - what concrete maintenance cost does it remove?
  - which rung of the simplification ladder applies, and what code shape remains after it?

1. Check the change against every strong signal in Review Standard, and note each signal that fires.

1. Keep the bar high.
- Do not recommend broad redesign when a narrow simplification is enough.
- Do not manufacture findings. A valid conclusion is that the code is already lean enough.

## Review Standard

Treat these as strong signals that code may deserve a diet:
- A field, column, option, config key, or hook exists mostly for future flexibility, and current behavior barely touches it.
- A helper, wrapper, or abstraction adds another jump without improving naming, ownership, invariants, or real reuse.
- A generic abstraction, extension point, or strategy surface has only one real path and no concrete second use.
- Callers still need to know internal details the abstraction claimed to hide.
- A supposedly shared helper now branches by flag, type, mode, or conditionals to serve diverging cases.
- One concept requires repeated mapping or translation layers mostly to preserve a generic design.
- The fix was appended alongside the existing design instead of integrated into it, and that workaround is likely to spread.
- The code became harder to follow because it optimized for hypothetical later work instead of today's path.

Treat these as weak signals that often do **not** justify action by themselves:
- The code is a little repetitive but still clear.
- A helper is single-use but materially improves the call site or name.
- A wrapper isolates an awkward dependency or side effect behind a better seam.
- A small object groups related data in a way that makes invariants easier to see.
- The code is explicit rather than clever.
- A bit of duplication is cheaper than a shared abstraction.

## Output

Use this structure when reporting:

### Bucket I — Obvious / High-Confidence
- Put findings here when there is a likely simplification path that does not need much discussion.
- Keep this concise.
- For each item, include:
  - what seems overweight
  - why it does not appear to earn its cost today
  - the smallest reasonable simplification direction and resulting code shape
  - the main tradeoff only if it matters

### Bucket II — Worth Discussing
- Put findings here when there are real tradeoffs, unclear constraints, multiple plausible directions, or the concern is plausible but weakly supported — do not overstate confidence.
- For each item, include:
  - what seems overweight or bandaged on
  - what decision or missing constraint makes it uncertain
  - the main options and tradeoffs

### Keep As-Is
- Call out choices that may look heavy at first glance but are paying for themselves now.
- Say plainly when explicitness, duplication, or a small abstraction is the right choice.

### Verdict
- End with one of:
  - `Lean enough as-is`
  - `Recommended targeted simplifications`
  - `Recommended changes include design decisions`
  - `Broader redesign may be warranted after clarifying constraints`
- If Bucket I or Bucket II is empty, say so explicitly.

## Snippet Rules

- Prefer small embedded snippets over editor-style line references when code evidence matters.
- Put the source file path on the first line of each snippet as a comment.
- Keep snippets tight: helper signatures, schema shapes, wrapper layers, flags, mode switches, and appended branches are usually enough.
