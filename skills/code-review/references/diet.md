# Diet Lens

Use this lens for **unearned weight**: code that does not earn its maintenance
cost today.

The goal is not fewer lines. The goal is fewer obligations: fewer concepts,
states, branches, compatibility promises, and surfaces to preserve.

This is not a general architecture or folder-layout review. It is a
simplification lens for code that feels heavier than the problem requires: too
many helpers, wrappers, schema fields, columns, options, generic layers, or
appended paths that solve the task by adding machinery instead of integrating
with the existing design.

Do not assume code came from an LLM or agent. Judge the code on its merits. Do
not force deletions. Explicit, readable, well-bounded code is often worth
keeping even when it is not minimal.

## Review Posture

- **YAGNI**: do not keep machinery only for hypothetical future cases.
- **AHA / rule of three**: avoid hasty abstractions; duplicate twice before
  extracting the wrong shared layer.
- **Duplication can be cheaper than the wrong abstraction** when a shared helper
  now needs flags, modes, or branching to serve diverging cases.
- **Prefer integration over appending**: solve the problem in the existing
  design when possible instead of adding side channels, one-off adapters, or
  parallel flows.
- **Keep readability-oriented explicitness**: a small helper, wrapper, or
  object can be worth its cost when it clarifies naming, boundaries, or
  invariants.

## Judgment Guardrails

- First separate essential complexity from accidental complexity. Essential
  complexity comes from the domain, external contracts, data lifecycle, failure
  modes, security, performance, or user workflow. Accidental complexity comes
  from implementation machinery that current behavior does not require.
- Before recommending removal, identify what would stop working, what tests or
  contracts would need to change, and whether the behavior is public, persisted,
  migrated, or used by another subsystem. If local evidence cannot answer that,
  put the item in Bucket II.
- Treat compatibility layers, migration fields, fallbacks, aliases, and
  dual-read or dual-write paths as justified only when there is a current
  rollout, external contract, or documented migration window. Otherwise, they
  are usually Bucket II candidates.
- Prefer simplification in this order: delete unused or speculative code; inline
  pass-through wrappers; make generic code specific to the current use; split
  diverging cases instead of adding flags or modes; redesign only when local
  simplification would leave the underlying weight intact.
- A longer direct implementation can be leaner than a shorter abstraction when
  it creates fewer concepts, jumps, configuration paths, and preserved states.

## Workflow

1. Anchor the lens in the real code.
   - Read the changed files, nearby interfaces, and affected tests or docs when
     relevant.
   - Distinguish code serving current behavior, code mainly serving hypothetical
     flexibility, and code appended as a workaround instead of integrated
     cleanly.

2. Ask what each extra piece buys now.
   - For every helper, wrapper, field, column, option, mode, schema entry, or
     abstraction layer, ask:
     - what current behavior depends on this?
     - what concrete maintenance cost does it remove?
     - would the code be clearer if this were inlined, deleted, or made more
       specific?
     - what replacement shape would remain after simplification: deleted,
       inlined, made specific, merged into an existing path, or split into
       explicit paths?

3. Look for weight-gain patterns.
   - Extra JSON fields, DTO properties, DB columns, or config keys that current
     behavior barely uses.
   - Helpers or wrappers that mostly forward calls or rename things without
     improving boundaries, reuse, or invariants.
   - Shared helpers that now need booleans, enums, `mode` parameters, or
     branching to cover multiple cases.
   - Generic abstractions with only one real path.
   - Translation or mapping layers added mainly to preserve genericity.
   - Extension points, hooks, strategy objects, or configuration surfaces with
     no concrete second use.
   - Appended code paths, special cases, or side channels that solve one feature
     without cleaning up the underlying path.
   - Data models shaped for imagined futures rather than today's behavior.

4. Keep the bar high.
   - Do not call something waste just because it is abstract.
   - Do not punish code for being explicit.
   - Do not recommend broad redesign when a narrow simplification is enough.
   - Do not manufacture findings. A valid conclusion is that the code is already
     lean enough.

## Review Standard

Strong signals that code may deserve a diet:

- A field, column, option, or hook exists mostly for future flexibility, and
  current behavior barely touches it.
- A helper, wrapper, or abstraction adds another jump without improving naming,
  ownership, invariants, or real reuse.
- Callers still need to know internal details the abstraction claimed to hide.
- A supposedly shared helper now branches by flag, type, mode, or conditionals
  to serve diverging cases.
- One concept requires repeated mapping or translation layers mostly to preserve
  a generic design.
- The fix was appended alongside the existing design instead of integrated into
  it, and that workaround is likely to spread.
- The code became harder to follow because it optimized for hypothetical later
  work instead of today's path.

Weak signals that often do not justify action by themselves:

- The code is a little repetitive but still clear.
- A helper is single-use but materially improves the call site or name.
- A wrapper isolates an awkward dependency or side effect behind a better path.
- A small object groups related data in a way that makes invariants easier to
  see.
- The code is explicit rather than clever.
- A bit of duplication is cheaper than a shared abstraction.

## Output Contribution

Add diet findings to the parent code-review buckets:

- Bucket I: only high-confidence simplifications that are narrow, local, and
  safe under the current edit policy. Include what is overweight, why it does
  not appear to earn its cost today, and the resulting code shape.
- Bucket II: simplifications with real tradeoffs, unclear constraints, external
  contracts, rollout concerns, or multiple plausible directions.
- Keep As-Is: choices that may look heavy but pay for themselves now.
