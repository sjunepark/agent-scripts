---
name: post-review-loop
description: "Post-implementation review loop for code changes. Use when the user asks for a post-review loop, post-implementation review, after-the-fact code review, review-and-fix pass, review-only pass, or an equivalent to Pi's post-review-loop extension. Supports one-shot review, persistent start/continue/status/report loops, Bucket I safe fixes, Bucket II decision items, validation, and final reporting."
---

# Post-Review Loop

Emulate Pi's post-review-loop as an explicit Codex workflow. Codex cannot own hidden extension state or inject follow-up prompts, so persist loop state in the conversation for one-shot work and in a visible ledger for multi-turn or resumed loops.

## Modes

Interpret the user's requested mode from natural language:

- `oneshot` or no mode: review the implementation once, optionally apply safe fixes, validate, and report.
- `review-only`: review once without editing files.
- `start`: create a persistent loop ledger and run the first phase.
- `continue` or `resume`: read the ledger and run the next phase.
- `status`: report current lifecycle, phase, unresolved buckets, validation, and next action without changing files.
- `report` or `stop`: render the final/current report. Stop after the current phase if already working.
- `cancel`: stop immediately and leave a short cancellation note if a ledger exists.

Default scope is `uncommitted changes`. If the user names a branch, commit range, feature, file list, or implementation target, use that scope instead.

## Ledger

Use an explicit ledger only for `start`, `continue`, `resume`, `status`, `report`, or work that cannot finish in one turn. Prefer `reviews/post-review-loop.md` at the target repo root unless the repo already has a better review-state convention.

Ledger shape:

```markdown
# Post-Review Loop

- lifecycle: active | paused | finalizing | complete | cancelled
- scope: <review target>
- phase: post-review | impl-review | impl | final-report
- iteration: <n>/<limit>
- review-only: true | false
- baseline: <HEAD or scope notes>

## What Was Reviewed
<target-oriented briefing>

## Bucket I - Safe In-Scope Fixes
| status | title | design signal | files | fix | validation |

## Bucket II - Needs Decision
| status | title | design signal | recommendation | tradeoffs |

## Code Changes Applied
| title | files | issue addressed | validation |

## Validation
| command | result | notes |

## Phase Log
- <iteration> <phase>: <summary>; gate: <decision>
```

Treat the ledger as authoritative. If it is not in the ledger, it did not happen. Keep it concise; do not turn it into a transcript.

## Review Lens

Review after the implementation exists. Focus on issues that became visible through real interfaces, state, control flow, tests, docs, or module boundaries.

Inspect real files and diffs before deciding:

- For `uncommitted changes`, inspect `git status --short`, `git diff`, `git diff --staged`, and relevant untracked files.
- For a branch, commit range, or named scope, inspect that target and nearby interfaces, tests, and docs.
- Distinguish issues introduced by the change from pre-existing debt merely exposed by it.
- A clean review is valid. Do not invent findings.

Classify meaningful findings with exactly one design signal:

- `simple local mistake`
- `weak validation or invariant gap`
- `unclear ownership / boundary problem`
- `weak type or schema model`
- `state, lifecycle, concurrency, or ordering hazard`
- `over-abstraction, under-abstraction, or duplicated logic`
- `brittle integration or contract mismatch`

Prefer root-cause findings over polish. Reject speculative preferences, broad future-proofing, tiny helper extraction, naming polish, isolated dedupe, or logging niceties unless they reveal a broader boundary, invariant, ownership, schema, lifecycle, abstraction, or integration problem.

## Buckets

Bucket I is the auto-fix track. Use it only when the issue is concrete, safe, in scope, materially useful, and root-cause-fixable now. Apply Bucket I fixes automatically unless the user requested `review-only`.

Bucket II is the decision track. Use it when a real improvement needs product, domain, architecture, rollout, compatibility, churn, or risk judgment. Do not implement Bucket II without explicit user approval.

Keep As-Is is for meaningful concerns rejected after inspection because evidence is weak, current structure earns its keep, or the change would be overcorrection.

When a safe Bucket I fix exposes a plausible broader redesign, apply the safe fix when appropriate and record the broader option as Bucket II with options, tradeoffs, recommendation, and why approval is needed. Do not manufacture a broader option when the focused fix is clearly sufficient.

## Phase Loop

Persistent loops use this phase order:

```text
post-review -> impl-review -> impl -> post-review -> ... -> final-report
```

Phase rules:

- `post-review`: review the current diff and nearby architecture. Do not edit code. Record Bucket I candidates and Bucket II decisions.
- `impl-review`: re-verify each actionable Bucket I item against real code paths and tests. Do not edit code. Mark safe items `accepted`; reject, downgrade, or move uncertain items to Bucket II.
- `impl`: implement accepted Bucket I fixes tightly. Mark fixed items `applied`, record code changes, and validate.

Gate after each phase:

- In `review-only`, stop after the review report.
- If scope/context is insufficient, stop with `scope or context needed`.
- If validation is blocking safe continuation, stop with `validation failure remains`.
- After `post-review`, continue only when Bucket I candidates exist and the iteration limit is not reached. If only Bucket II remains, stop for decision.
- After `impl-review`, continue only when accepted Bucket I work exists and the iteration limit is not reached.
- After `impl`, continue to another `post-review` only if at least one Bucket I fix was applied.
- Stop cleanly when no accepted/actionable Bucket I findings remain.

Use a default iteration limit of 5 for persistent loops unless the user gives another limit.

## Git And Validation

Do not push. Do not create commits unless the user explicitly asked this skill to manage a persistent loop with commits or separately asked for commits.

When committing is requested:

- At loop start, record the starting `HEAD`. If reviewing dirty work, stage and commit only changes relevant to the review target; use partial hunks when needed; leave unrelated work untouched.
- Use ordinary project commit messages. Do not mention checkpointing, the loop, automation, or internal ids.
- At the end, commit only loop-applied edits. If relevant edits are already committed or no loop edits exist, do not create a duplicate commit.

Run focused validation from the repo's existing workflow after edits. If no validation is run, say why. If validation fails, fix failures caused by the loop or clearly report unrelated/pre-existing failures.

## Output

For `oneshot`, `review-only`, and final reports, output these sections:

```markdown
### What I Reviewed
<1-3 short sentences. Distinguish files reviewed from files edited when useful.>

### Applied / Resolved
1. [applied] <Bucket I item with design signal, root cause, fix, files, and validation>
```

If no automatic changes were applied, write `No automatic changes were applied.` For `review-only`, list concrete Bucket I findings as `[recommended but not applied]` and say review-only mode is why no edit was made.

```markdown

### Needs Decision / Bucket II
1. [recommended now|deferred|discussion only] <decision, options, recommendation, tradeoffs, why permission is needed>
```

If no unresolved Bucket II decisions remain, write `No unresolved Bucket II decisions remain.`

```markdown

### Keep As-Is
<Meaningful rejected findings and why, or omit tiny non-findings.>

### Validation
<commands and results, or why skipped>

### Verdict
<one of: No meaningful improvement identified; Applied straightforward design improvements; Applied improvements; decision needed for remaining tradeoff; Decision needed before refactor; Validation failure remains>
```

For `status`, keep the response shorter: lifecycle, phase, iteration, Bucket I applied/actionable counts, Bucket II unresolved count, latest validation outcome, and next action.
