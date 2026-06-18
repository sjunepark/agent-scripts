---
name: post-review-loop
description: "Post-implementation review loop for code changes. Use when the user asks for a post-review loop, post-implementation review, after-the-fact code review, review-and-fix pass, review-only pass, or an equivalent to Pi's post-review-loop extension. Supports one-shot review, persistent start/continue/status/report loops, strict review finding quality gates, Bucket I safe fixes, Bucket II decision items, validation, and final reporting."
---

# Post-Review Loop

Run a post-implementation review workflow. Keep Codex's built-in `/review` discipline for finding quality, then add the loop mechanics that `/review` does not provide: safe fixes, decision tracking, validation, and resumable state.

Codex cannot own hidden extension state or inject follow-up prompts. For persistent loops, keep state in a visible ledger. For explicit one-shot work that completes in the same turn, use conversation-local state.

## Modes

Infer the mode from the user's wording:

- `loop` or no mode: create or reuse a persistent ledger, run the phase loop, apply safe fixes, validate, and stop when no accepted/actionable Bucket I findings remain or the iteration limit is reached. Default limit: 5 iterations.
- `oneshot`: review once, optionally apply safe fixes, validate, and report without creating a ledger unless the work cannot finish in one turn.
- `review-only`: review once without editing files.
- `start`: create a persistent ledger and run the default loop.
- `continue` or `resume`: read the accepted ledger and run the next phase.
- `status`: report lifecycle, phase, unresolved buckets, validation, and next action without changing files.
- `report` or `stop`: render the final/current report and stop after the current phase.
- `cancel`: stop immediately and leave a short cancellation note if a ledger exists.

Default scope is `uncommitted changes`. If the user names a branch, commit range, feature, file list, or implementation target, use that scope instead.

## Ledger

Use `reviews/post-review-loop.md` at the target repo root unless the repo already has a better review-state convention. Use a ledger for `loop`, `start`, `continue`, `resume`, `status`, `report`, or any work that cannot finish in one turn.

Ledger shape:

```markdown
# Post-Review Loop

- ledger-version: 3
- repo: <target repo root or repository identity>
- lifecycle: active | paused | finalizing | complete | cancelled
- scope: <review target>
- phase: post-review | impl-review | impl | final-report
- iteration: <n>/<limit>
- review-only: true | false
- baseline: <HEAD or scope notes>
- scope-fingerprint: <short stable identity for the reviewed diff/scope>
- updated: <ISO-8601 timestamp>

## What Was Reviewed
<target-oriented briefing>

## Bucket I - Safe In-Scope Fixes
| status | priority | title | design signal | evidence | fix | validation |

## Bucket II - Needs Decision
| status | priority | title | design signal | recommendation | tradeoffs |

## Keep As-Is
| title | reason |

## Code Changes Applied
| title | files | issue addressed | validation |

## Validation
| command | result | notes |

## Phase Log
- <iteration> <phase>: <summary>; gate: <decision>
```

Treat the ledger as authoritative. If it is not in the ledger, it did not happen. Keep it concise; do not turn it into a transcript.

## Ledger Freshness

Before reading any existing ledger body, prove it belongs to the requested work:

1. Determine the target repo root, requested mode, requested scope, current `HEAD`, and current scope fingerprint. For default `uncommitted changes`, base the fingerprint on `HEAD`, `git status --short`, `git diff --name-only`, `git diff --staged --name-only`, and relevant untracked paths.
2. If `reviews/post-review-loop.md` exists, read only the title and metadata block through `baseline`, `scope-fingerprint`, and `updated`. Do not read findings, phase logs, or old report sections until the ledger is accepted as fresh.
3. Reuse the ledger only when repo, scope, lifecycle, and fingerprint clearly match the current request, or when the user explicitly asked to `continue`, `resume`, `status`, or `report` for that ledger and the header does not contradict the request.
4. Treat completed, cancelled, mismatched, missing-metadata, or legacy ledgers as stale for a new `loop`, default loop, or `start`.
5. Supersede stale ledgers by moving them to `reviews/archive/post-review-loop-<timestamp>.md`, then create a fresh ledger. Do not delete stale ledgers unless the user explicitly asks. Mention the archive path once in the final or status output.
6. If the user explicitly asked to `continue`, `resume`, `status`, or `report` and the only ledger is stale or ambiguous, stop with a concise mismatch summary instead of importing old findings into the new context.

Refresh `scope-fingerprint` and `updated` after every phase. Keep the fingerprint short; it is an identity check, not a transcript.

## Review Contract

Review after the implementation exists. Inspect real files and diffs before deciding:

- For `uncommitted changes`, inspect `git status --short`, `git diff`, `git diff --staged`, and relevant untracked files.
- For a branch, commit range, or named scope, inspect that target plus nearby interfaces, tests, and docs.
- Distinguish issues introduced by the change from pre-existing debt merely exposed by it.
- Prefer no finding over a weak finding. A clean review is valid.

### Finding Quality Gate

Do not record a Bucket I or Bucket II item until it passes this gate:

- The issue is discrete, actionable, and materially affects correctness, maintainability, security, performance, compatibility, lifecycle behavior, or test reliability.
- The issue was introduced by the reviewed change, or the change makes a pre-existing invariant violation newly reachable.
- The affected path, caller, state transition, schema contract, permission boundary, or user-visible behavior is concrete. Do not rely on speculative breakage.
- The scenario, input, environment, or ordering needed to trigger the issue is explicit.
- The likely author would fix it if shown the evidence.
- The evidence can be anchored to a narrow changed location when possible, with nearby supporting code only as needed.

If the concern fails this gate, either omit it or record it under Keep As-Is when the rejection itself is useful.

### Finding Fields

For each meaningful finding, capture:

- `priority`: `P0`, `P1`, `P2`, or `P3`.
- `design signal`: exactly one of the signals below.
- `evidence`: changed line or narrow file/line location, plus the concrete affected path.
- `scenario`: when the issue occurs.
- `impact`: why it matters.
- `action track`: Bucket I, Bucket II, or Keep As-Is.

Use priorities this way:

- `P0`: blocks release, operations, major usage, or data/security integrity.
- `P1`: urgent and should be addressed in the next cycle.
- `P2`: normal issue worth fixing.
- `P3`: low severity but still materially useful.

Classify each meaningful finding with exactly one design signal:

- `simple local mistake`
- `weak validation or invariant gap`
- `unclear ownership / boundary problem`
- `weak type or schema model`
- `state, lifecycle, concurrency, or ordering hazard`
- `over-abstraction, under-abstraction, or duplicated logic`
- `brittle integration or contract mismatch`

Prefer root-cause findings over polish. Reject broad future-proofing, tiny helper extraction, naming polish, isolated dedupe, logging niceties, or stylistic preferences unless they reveal a real boundary, invariant, ownership, schema, lifecycle, abstraction, or integration problem.

## Buckets

Use Bucket I for automatic fixes only when the finding is concrete, safe, in scope, materially useful, and root-cause-fixable now. Apply Bucket I fixes automatically unless the user requested `review-only`.

Use Bucket II when a real improvement needs product, domain, architecture, rollout, compatibility, churn, or risk judgment. Do not implement Bucket II without explicit user approval.

Use Keep As-Is for meaningful concerns rejected after inspection because evidence is weak, the current structure earns its keep, or the change would be overcorrection.

If a safe Bucket I fix exposes a plausible broader redesign, apply the safe fix when appropriate and record the broader option as Bucket II with options, tradeoffs, recommendation, and why approval is needed. Do not manufacture a broader redesign when the focused fix is sufficient.

## Phase Loop

Persistent loops use this order:

```text
post-review -> impl-review -> impl -> post-review -> ... -> final-report
```

Phase rules:

- `post-review`: review the current diff and nearby architecture. Do not edit code. Record Bucket I candidates, Bucket II decisions, and useful Keep As-Is entries.
- `impl-review`: re-verify each actionable Bucket I item against real code paths and tests. Do not edit code. Mark safe items `accepted`; reject, downgrade, or move uncertain items to Bucket II.
- `impl`: implement accepted Bucket I fixes tightly. Mark fixed items `applied`, record code changes, and validate.

Gate after each phase:

- In `review-only`, stop after the review report.
- If scope/context is insufficient, stop with `scope or context needed`.
- If validation blocks safe continuation, stop with `validation failure remains`.
- After `post-review`, continue only when Bucket I candidates exist and the iteration limit is not reached. If only Bucket II remains, stop for decision.
- After `impl-review`, continue only when accepted Bucket I work exists and the iteration limit is not reached.
- After `impl`, continue to another `post-review` only if at least one Bucket I fix was applied.
- Stop cleanly when no accepted/actionable Bucket I findings remain.

Use a default iteration limit of 5 unless the user gives another limit.

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
1. [applied] <Bucket I item with priority, design signal, root cause, fix, files, and validation>
```

If no automatic changes were applied, write `No automatic changes were applied.` For `review-only`, list concrete Bucket I findings as `[recommended but not applied]` and say review-only mode is why no edit was made.

```markdown

### Needs Decision / Bucket II
1. [recommended now|deferred|discussion only] <decision with priority, design signal, options, recommendation, tradeoffs, and why permission is needed>
```

If no unresolved Bucket II decisions remain, write `No unresolved Bucket II decisions remain.`

```markdown

### Keep As-Is
<Meaningful rejected findings and why, or omit tiny non-findings.>

### Verdict
<one of: No meaningful improvement identified; Applied straightforward design improvements; Applied improvements; decision needed for remaining tradeoff; Decision needed before refactor; Validation failure remains>
```

Include `### Validation` before `### Verdict` only when validation was run, failed, skipped for a meaningful reason, or affects the verdict. Otherwise fold validation into the relevant applied item or omit it. In review-only or no-edit reports, do not add a validation section only to say no commands were run.

For `status`, keep the response shorter: lifecycle, phase, iteration, Bucket I applied/actionable counts, Bucket II unresolved count, latest validation outcome, and next action.
