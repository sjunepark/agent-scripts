---
name: post-implementation-review-loop
description: Iteratively review already-implemented code using the required phase_checkpoint_compact Pi extension tool after each fix/validation pass; print visible phase reports, automatically apply straightforward Bucket I fixes, validate, and continue until no accepted Bucket I actions remain or the iteration limit is reached while Bucket II items receive recommended next actions.
---

# Post-Implementation Review Loop

Use this skill manually. Do not invoke it unless the user explicitly asks for a looped or iterative post-implementation review, or asks to fix straightforward findings from a post-implementation review automatically.

This is an implementation-and-review loop, not a one-pass advisory review. Review the completed code, verify findings, apply straightforward fixes, validate, and call `phase_checkpoint_compact` only after the fix/validation pass so Pi performs a real compacted handoff before the next review.

## Hard Requirement

This skill requires the Pi extension tool `phase_checkpoint_compact`.

If `phase_checkpoint_compact` is unavailable:

- Stop before starting the loop.
- Tell the user the `post-implementation-review-loop` skill requires the `phase-checkpoint-compact` Pi extension.
- Do not substitute `/compact`, soft checkpoints, summaries, or ordinary conversation handoffs.

## Contract

- Treat each review finding as advisory until verified against the real code path, nearby interfaces, tests, and docs.
- Do not blindly apply generated, previous review, or subagent review output.
- Reject weak, speculative, unrealistic, or overcomplicated recommendations.
- Prefer root-cause design, ownership, boundary, structure, or abstraction fixes over tactical bandages.
- Apply Bucket I findings automatically when they are straightforward, materially worthwhile, and safe within the user's requested scope.
- Do not automatically implement Bucket II findings unless the user explicitly approved that direction. For each Bucket II item, give a recommended next action that prefers improving overall code quality and design over temporary fixes unless the refactor cost, risk, or uncertainty is too high.
- If a review-triggered fix changes code, rerun focused tests or validation, call `phase_checkpoint_compact`, and rerun the review loop after the compacted handoff.
- Do not keep looping for optional polish, speculative improvements, or findings already deferred by the parent.
- Keep exactly one writer against the active worktree. Fresh reviewers are review-only; if a worker subagent applies fixes, the parent must not edit concurrently.
- Do not checkpoint after review-only or planning-only phases; move from review to verification to implementation in the same active context unless a real handoff safety issue requires stopping.
- Call `phase_checkpoint_compact` only after an implementation/fix and validation pass completes, with `nextPhase: "post-review"`.
- After calling `phase_checkpoint_compact`, do not continue substantial work; wait for the extension's next-phase prompt.
- Keep going until no accepted/actionable Bucket I findings remain or the iteration limit is reached.
- Use a default maximum of 5 review iterations unless the user gives a different number.
- Before the loop starts, create a Git baseline for the scoped pre-review changes so user edits and loop edits are separable in history.
- Maintain a cumulative loop ledger so the final response can brief the user on every Bucket I finding, Bucket II finding, rejected finding, fix applied, and validation result across all iterations.
- Stop when remaining work is Bucket II, out of scope, blocked by missing context, not worth changing now, or the iteration limit is reached.
- Do not push, install globally, or create non-baseline commits unless the user explicitly asks.

## Review Lens

Focus on issues that become visible only after the implementation touched real interfaces, control flow, state, tests, docs, or module boundaries.

Look for:

- Repeated branching, mapping, translation, or glue code that suggests the abstraction boundary is off.
- Modules knowing details they should not need to know.
- One concept forcing edits across too many files in a way likely to recur.
- Awkward ownership of data, side effects, orchestration, or validation.
- Responsibilities blurred enough that naming or call flow became hard to explain.
- Tests becoming hard to set up because dependencies or ownership are misplaced.
- Workarounds or special cases that will probably spread.
- Speculative structure: extra fields, hooks, wrappers, options, strategy objects, generic layers, or config surfaces current behavior barely uses.
- Under-structure: related concerns scattered or mixed so readers cannot follow one feature end to end.

Do not manufacture findings. If the code is structurally sound, say so.

## Optional Subagent-Assisted Review

`phase_checkpoint_compact` remains the only required extension for this skill. Subagents are a tunnel-vision control, not a hard dependency. If the `subagent` tool is unavailable, run the loop yourself.

When `subagent` is available and the change is non-trivial, use fresh-context parallel reviewers during `post-review` before finalizing Bucket I and Bucket II findings. Use this for broad diffs, risky changes, design-heavy work, UI/CLI behavior, security-sensitive code, or any change where a second independent read would reduce confirmation bias.

Fresh context does not mean context-free. Give each reviewer a bounded review packet: objective, scope or diff target, relevant constraints, validation commands, and any user-approved direction. Do not include the implementer's rationale, rejected alternatives, or persuasive framing unless it is needed to understand a hard constraint.

If `subagent` is unavailable, do not describe the same-conversation pass as independent. Run a deliberate second pass yourself and treat it as weaker evidence.

If an external review command or model is noisy, prefer running it through a review-only subagent filter so the parent receives only accepted actionable findings, rejected findings with one-line reasons, and exact files/tests to rerun.

Reviewer rules:

- Use `context: "fresh"`, not forked context, unless the user explicitly asks otherwise.
- Reviewers must inspect repository instructions, relevant files, and the current diff directly from tools and commands.
- Reviewers must not edit files or run their own subagent workflows.
- Give each reviewer one distinct angle chosen from the actual change. Prefer three strong reviewers over many vague ones.
- Common angles: correctness/regressions, tests/validation, simplicity/maintainability.
- Add security/privacy, performance, docs/API contracts, user-flow behavior, accessibility, cleanup/deslop, or structural-boundary angles only when the change calls for them.
- Ask reviewers for concise, evidence-backed findings with file/line references or tight snippets and suggested fixes.

Parent synthesis rules:

- Treat reviewer output as leads, not verdicts.
- Verify every finding against the real code path, nearby interfaces, tests, docs, scope, and risk before acting.
- Reject noisy, speculative, unrealistic, overbroad, or overcomplicated findings even when multiple reviewers repeat them.
- Map verified findings into this skill's categories:
  - fixes worth doing now -> Bucket I
  - product, scope, architecture, compatibility, or broad refactor decisions -> Bucket II
  - optional polish -> deferred or kept as-is
  - weak or incorrect feedback -> rejected
- If reviewers surface an unapproved product, scope, or architecture decision, stop and ask the user before implementation.
- Preserve the normal phase rhythm. Subagent fanout belongs inside `post-review`; accepted fixes are verified in `impl-review`, implemented in `impl`, then compacted before the next `post-review`.

## Phase Loop

### Deterministic phase gate

Use `scripts/loop-decision` as the authoritative transition gate before stopping, asking the user whether to proceed, calling `phase_checkpoint_compact`, or entering the next phase. The script does not review or fix code; it converts the current phase state into one allowed transition.

Resolve the script path against this skill directory. From this repository root, run it with a JSON phase snapshot, for example:

```bash
skills/post-implementation-review-loop/scripts/loop-decision --pretty <<'JSON'
{"phase":"post-review","iteration":1,"limit":5,"bucketICandidates":2,"bucketII":0}
JSON
```

The output's `nextPhase` is binding. If it says `continue`, continue without asking for permission. If it says `stop`, provide the final briefing. If it says `checkpointRequired: true`, call `phase_checkpoint_compact` before the next `post-review`.

Minimum snapshot fields:

- Always include `phase`, `iteration`, and `limit`.
- For `post-review`, include `bucketICandidates` and `bucketII` counts or lists.
- For `impl-review`, include `acceptedBucketI` and `bucketII` counts or lists.
- For `impl`, include `appliedBucketI` count or list.
- Include boolean stop flags when relevant: `reviewOnly`, `scopeBlocked`, `validationBlocked`, `checkpointUnavailable`.

Transition table encoded by the gate:

| Phase | Condition | Next phase |
| --- | --- | --- |
| any | review-only, scope blocked, validation blocked, or checkpoint unavailable | final briefing |
| `post-review` | no Bucket I candidates | final briefing |
| `post-review` | iteration limit reached | final briefing |
| `post-review` | Bucket I candidates exist | `impl-review` |
| `impl-review` | no accepted/actionable Bucket I | final briefing |
| `impl-review` | iteration limit reached | final briefing |
| `impl-review` | accepted/actionable Bucket I exists | `impl` |
| `impl` | no Bucket I fixes were applied | final briefing |
| `impl` | Bucket I fixes were applied | `post-review` via `phase_checkpoint_compact` |

Treat user wording like "do not edit code in this phase" as phase-local unless the user explicitly says the whole run is review-only or asks to wait for approval before implementation. Producing Bucket I/Bucket II findings is not a stop condition by itself. Do not write "if you want me to proceed" during an active loop.

Active phase rhythm:

```text
review -> impl/fix -> compact -> review -> impl/fix -> compact -> ...
```

Concretely, `post-review` and `impl-review` run together as the review/planning side of an iteration, then accepted Bucket I work is implemented in `impl`, validated, and compacted before the next `post-review`.

If the user already provided an accepted implementation plan, the loop may start at `impl`. Otherwise, start at `post-review` for an already-completed change.

### Git baseline before loop

Before the first active phase, create or record a baseline so pre-loop edits and review-loop edits are easy to separate.

1. Run `git status --short` and identify the user-requested review scope.
2. If there are uncommitted changes in scope, stage only those scoped files and commit them before reviewing.
3. Use this commit subject:

   ```text
   wip(post-review-loop): baseline before review
   ```

4. Include this marker in the commit body:

   ```text
   Post-review-loop baseline

   Captures the pre-loop implementation state before automated post-review fixes.
   Loop edits should appear after this commit.
   ```

5. If the working tree is already clean, record the current `HEAD` as the baseline instead of creating an empty commit.
6. If unrelated dirty files exist and the scoped files cannot be separated safely, stop and ask the user which files belong in the baseline.
7. Do not commit loop-generated fixes by default. Leave them after the baseline commit so the user can inspect `git diff HEAD` or ask for a separate follow-up commit.

Treat invoking this loop skill as approval to create the baseline WIP commit for scoped pre-review changes. It is not approval to push or to commit the loop's later fixes.

### Iteration limit

- Default maximum: 5 review iterations.
- If the user provides an iteration count, use that count instead.
- Count one iteration each time a `post-review` phase completes against the current implementation state.
- Count the initial `post-review` as iteration 1.
- If the loop starts at `impl`, the first `post-review` after that implementation is iteration 1.
- Stop instead of starting another active phase when the completed review iteration count reaches the limit and more Bucket I work would otherwise remain.
- Include the current count and limit in every visible phase report and checkpoint handoff.

### Cumulative loop ledger

Maintain a concise ledger across iterations and compacted handoffs. Update it after each visible phase report and after each fix/validation checkpoint so the final briefing has the full loop history.

Track:

- Baseline commit hash or baseline `HEAD`, plus which scoped files were included.
- Bucket I findings discovered, accepted, rejected, applied, or still remaining.
- Bucket II findings, recommended actions, and whether the user approved any direction.
- Rejected or kept-as-is findings and why they were rejected.
- Code fixes/refactors applied, including the files touched and the root-cause issue each fix addressed.
- Validation commands and results for each implementation pass.

Keep the ledger concise. Preserve facts and outcomes, not long reasoning traces.

### Phase meanings

1. `post-review`
   - Re-read the current diff and relevant files.
   - For non-trivial changes, consider fresh-context subagent review fanout before finalizing findings.
   - Separate pre-existing debt, issues introduced by the change, and issues merely exposed by the change.
   - Synthesize review output into fixes worth doing now, decisions needing user approval, optional/deferred items, and rejected feedback.
   - Produce Bucket I and Bucket II findings.
   - Do not edit code.
   - Print a visible phase report, then run `scripts/loop-decision` with `phase: "post-review"`, the current `iteration` and `limit`, `bucketICandidates`, `bucketII`, and any active stop flags.
   - Obey the gate output: continue directly to `impl-review` only when it returns `nextPhase: "impl-review"`; otherwise stop with the final report and do not checkpoint into a no-op phase.

2. `impl-review`
   - Verify each post-review finding against the actual code path, adjacent files, tests, scope, and risk.
   - Accept, reject, downgrade, or move findings between buckets.
   - Convert accepted Bucket I findings into a concrete implementation plan.
   - Keep Bucket II as decisions and attach a recommended action to each.
   - Do not edit code unless the item is already accepted as Bucket I and implementation has begun.
   - Print a visible phase report, then run `scripts/loop-decision` with `phase: "impl-review"`, the current `iteration` and `limit`, `acceptedBucketI`, `bucketII`, and any active stop flags.
   - Obey the gate output: continue directly to `impl` only when it returns `nextPhase: "impl"`; otherwise stop with the final report and list any unimplemented Bucket I items.

3. `impl`
   - Implement only accepted Bucket I actions or the user's explicitly approved Bucket II direction.
   - Implement the root-cause change, not just the smallest quieting patch.
   - Keep changes tight and reviewable.
   - If the apparent fix grows into a technical decision, move it to Bucket II and stop instead of guessing.
   - Rerun focused validation after changes.
   - Update the cumulative ledger with the applied fixes, validation, rejected items, Bucket II items, and any remaining Bucket I items.
   - Print a visible phase report, then run `scripts/loop-decision` with `phase: "impl"`, the current `iteration` and `limit`, `appliedBucketI`, and any active stop flags.
   - Obey the gate output: when it returns `checkpointRequired: true`, call `phase_checkpoint_compact` with `phaseCompleted: "impl"` and `nextPhase: "post-review"` so the next review can confirm whether Bucket I is clean or the iteration limit has been reached; otherwise stop with the final report.

## `phase_checkpoint_compact` Usage

Call `phase_checkpoint_compact` only after an implementation/fix pass is complete and focused validation has been recorded. Do not call it after review-only or planning-only phases.

Populate the tool fields as follows:

- `phaseCompleted`: `impl`.
- `nextPhase`: `post-review`.
- `goal`: the user's review/fix objective.
- `scope`: the reviewed diff, files, commit range, or user-provided scope.
- `changedFiles`: repo-relative paths touched or relevant to the loop.
- `validation`: commands run and whether each passed, failed, or was skipped. Each `result` value must be exactly `passed`, `failed`, or `skipped`; never use values like `not run`, `pending`, `unknown`, or `n/a`.
- `bucketIApplied`: Bucket I fixes already applied.
- `bucketIRemaining`: accepted/actionable Bucket I items still to implement or verify.
- `bucketII`: discussion items, each with options when useful and a recommended action.
- `rejectedOrKeptAsIs`: meaningful findings rejected after verification and why.
- `handoffSummary`: concise next-phase handoff; include only what the next phase needs.

Handoff summaries must preserve:

- current goal and scope
- baseline commit hash or baseline `HEAD`
- phase completed (`impl`) and next phase (`post-review`)
- changed files and important seams
- iteration count and iteration limit
- validation commands and results using only `passed`, `failed`, or `skipped`
- accepted Bucket I actions already applied
- remaining Bucket I items, if any
- cumulative Bucket I findings and how each was resolved or left remaining
- Bucket II items with recommended actions and tradeoffs
- rejected or kept-as-is findings and why
- fixes/refactors applied and the files they touched
- the next phase's first concrete action

Do not preserve long reasoning traces, stale alternatives, raw tool output dumps, or rejected findings that no longer matter.

## Bucket Rules

### Bucket I — Straightforward / Recommended

Put a finding here only when all are true:

- The evidence is concrete in the implemented code.
- The improvement is materially worthwhile for correctness, ownership, boundaries, organization, or future change cost.
- The fix path is clear enough to implement without a user decision.
- The change is inside the user's requested scope.

For loop execution, Bucket I means: verify it, implement it, validate it, then review again.

Avoid Bucket I for tiny helper extractions, one-off naming polish, isolated dedupe, or logging niceties unless they reflect a broader boundary or ownership issue.

### Bucket II — Worth Discussing

Put a finding here when any are true:

- Multiple credible designs exist.
- The correct choice depends on product, domain, rollout, compatibility, or risk tolerance.
- The refactor is broad enough that user approval matters.
- The evidence is real but not yet strong enough to justify immediate change.
- The issue may be pre-existing debt outside the requested scope.

For loop execution, Bucket II means: do not auto-implement. Explain the tradeoff and recommend how to act.

When recommending an option, prefer the path that improves the codebase's design, ownership, boundaries, and long-term maintainability. Recommend a temporary/local fix only when the broader refactor cost is disproportionate, the evidence is not strong enough, the change would introduce meaningful new risk, or the decision depends on product/domain constraints.

Recommended actions:

- `Discuss before changing`: user/domain decision needed.
- `Defer`: real concern, but not worth changing until the next related feature or until stronger evidence justifies the refactor.
- `Keep as-is for now`: current shape is acceptable; watch for repetition.
- `Prototype separately`: validate a broader refactor before committing when the design improvement is promising but risk or uncertainty is meaningful.
- `Implement next if approved`: best design-improving path is clear enough, but requires explicit approval because cost/risk/scope is meaningful.

State the reason for the recommendation, not just the available options. If recommending a temporary fix, explicitly say why the stronger design improvement is not worth doing now.

## Validation Rules

- Identify focused tests, typechecks, formatters, or validation commands from the repository's existing workflow.
- For `phase_checkpoint_compact`, every validation item must use one of exactly three result strings: `passed`, `failed`, or `skipped`.
- No-edit review and planning phases do not trigger checkpoint payloads; record skipped validation in the visible phase report or ledger only when it helps explain that validation was not applicable.
- After every code change, rerun focused checks that cover the touched area.
- If validation fails because of the loop's changes, fix or report before moving phases.
- If validation fails for unrelated/pre-existing reasons, record it clearly in `validation` and continue only if the next phase can still reason safely.
- Do not invent validation commands when the repo has an established workflow.

## Stop Conditions

Stop when any of these are true:

- A `post-review` phase finds no Bucket I candidates.
- An `impl-review` phase finds no accepted/actionable Bucket I items.
- The loop reaches the maximum review iteration count, defaulting to 5 unless the user provided another number.
- Remaining work is Bucket II and needs explicit user approval.
- A validation failure remains and cannot be safely resolved inside scope.
- Scope or context is missing.
- The required `phase_checkpoint_compact` tool is unavailable.

Do not run extra review cycles just to get nicer wording after a clean pass.

## Output

During active phases, print a visible phase report at each review/planning decision point and before the post-implementation `phase_checkpoint_compact` call. Do not hide review findings only inside tool-call arguments or compacted handoffs.

Use this active phase report structure:

### Phase Report — <phase>

- Iteration: `<current>/<limit>`
- Next step: `<impl-review>`, `<impl>`, `<post-review via checkpoint>`, or `<final briefing>`

#### Bucket I

List Bucket I findings, accepted actions, or applied fixes. If empty, say `No Bucket I items.`

#### Bucket II

List discussion items with `Recommended action: ...`. If empty, say `No Bucket II items.`

#### Rejected / Kept As-Is

List meaningful rejected findings. If empty, say `None.`

#### Validation

List commands and results. Use only `passed`, `failed`, or `skipped` as status words.

Only the `impl` phase report is followed by `phase_checkpoint_compact`, using the same facts in the tool payload. Review-only and planning-only reports continue directly to the next in-context phase or stop with the final briefing.

When the loop stops, provide a final briefing. Do not stop with only a checkpoint handoff or terse verdict.

Final briefing structure:

### Loop Briefing

- Baseline commit hash or baseline `HEAD`, and whether a WIP baseline commit was created.
- Iterations and phases run, including the maximum allowed iterations.
- Files changed by the loop.
- Validation commands run.
- Final clean condition: no accepted/actionable Bucket I findings, or why the loop stopped.
- Confirmation that you inspected the final diff yourself and either ran or confirmed focused validation.
- If stopped by iteration limit, list any remaining Bucket I items that were not implemented.

### Bucket I — Findings and Fixes

Number each Bucket I finding discovered across the loop and include:

- what the implementation revealed
- whether it was accepted, applied, rejected, or left remaining
- the root-cause fix/refactor applied, if any
- files changed by the fix
- why a smaller patch would have been a bandage, if relevant
- validation evidence

If empty, say: `No Bucket I findings were found.`

### Bucket II — Findings and Recommendations

Number each Bucket II finding discovered across the loop and include:

- what the implementation revealed
- the design or quality weakness it implies
- main options or decision points
- `Recommended action: ...`
- tradeoffs, risks, or uncertainty
- whether it was left for user decision, deferred, kept as-is for now, or implemented after explicit approval

If empty, say: `No Bucket II findings were found.`

### Code Changes Applied

Summarize each fix/refactor applied by the loop:

- files changed
- behavior/design issue addressed
- why this was the right scope of change
- validation evidence
- how to inspect it relative to the baseline, for example `git diff <baseline>..HEAD` if committed or `git diff HEAD` if left uncommitted after the baseline

If empty, say: `No code changes were applied by the loop.`

### Rejected / Kept As-Is

Call out meaningful findings rejected after verification, plus why no change is recommended.

### Verdict

End with one of:

- `Loop clean: no accepted/actionable Bucket I findings remain`
- `Loop stopped: Bucket II decision needed`
- `Loop stopped: iteration limit reached`
- `Loop stopped: validation failure remains`
- `Loop stopped: scope or context needed`
- `Loop stopped: phase_checkpoint_compact unavailable`

## Snippet Rules

- When citing existing code, prefer embedded snippets over bare file paths and line numbers.
- Put the source file path on the first line of each snippet as a comment.
- Keep snippets tight and evidence-focused.
- Use file and line references only when snippets would add noise.

## Communication Rules

- Be direct and specific.
- Do not make the user inspect the editor just to understand the review.
- Keep Bucket I concise because those items should already be resolved by the loop.
- Put the user's attention on Bucket II decisions and your recommended action for each.
- Surface real tradeoffs without padding the review.
- Do not recommend broad rewrites unless they clearly improve the bug class or design seam.
- If there are no meaningful design flaws or worthwhile improvements, say so plainly.
