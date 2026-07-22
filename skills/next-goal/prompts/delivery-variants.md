# Delivery Variants

Read this file only after establishing that `/goal` is warranted.

## Output Order

Return:

1. **Recommended boundary** — a short scope statement and evidence-based rationale.
2. **Delivery recommendation** — **PR delivery** or **No PR**, with a short rationale based on the expected size of the implementation change and its CodeRabbit review surface.
3. **Copy-paste `/goal` prompts** — exactly two labeled fenced blocks, each containing only the body to enter after `/goal`:
   - **PR delivery** — require sequential PR creation, review, and merge.
   - **No PR** — require completion without PR creation.

Assume each new session has none of this conversation. Make each prompt self-contained as a routing instruction, not a detailed specification. Put exclusions inside each fenced prompt; a summary outside the prompt cannot carry authorization into the fresh session. Point to existing repository documents for implementation detail without copying their contents.

Use one compact structured contract plus concise execution and delivery paragraphs. Boundary clarity takes precedence over a paragraph-count target. Include:

- the complete goal contract specified by `SKILL.md`, with concrete values and every included result listed as its own sub-bullet;
- an explicit `$progress` invocation in goal mode to initialize the named goal-state file from the exact contract text, or recover and verify that file, before implementation. Require the same invocation and recovery at every resumed turn, automatic continuation, compaction recovery, or handoff before candidate work is selected. Do not summarize or broaden the contract while persisting it. Keep its immutable scope separate from mutable status, and keep the project queue truthful without deriving authorization from it;
- a transition check before every new plan, work branch, PR, review program, or independently reviewable outcome. Proceed only when the candidate is explicitly included or is the smallest bounded work necessary for an included completion condition. Record useful outside work without executing it; ask before ambiguous or materially expansive work;
- a direction to mark the goal complete and return immediately after its completion condition is satisfied, before inspecting or starting the next project item;
- a fail-closed direction: if `$progress` goal mode or the exact goal contract cannot be recovered, stop and request direction instead of substituting `ROADMAP`, `Current`, or another project queue;
- a direction to follow applicable `AGENTS.md`, preserve unrelated changes, and perform repository-required validation and review;
- a requirement to commit incrementally as meaningful, passing units finish; do not defer all commits until goal completion or collapse the run into one final commit. Make each message state what changed and why, including non-obvious decisions needed later.

Use this field order inside both prompt bodies:

```text
Goal contract
- Outcome:
- Goal state:
- Sources (implementation detail, not scope authority):
- Included results:
- Complete when:
- Explicitly excluded:
- Derived work:
- Expansion authority: explicit user instruction only
- Resume invariant: invoke $progress in goal mode and recover Goal state before selecting work after every resume, automatic continuation, compaction, or handoff
- Delivery: PR delivery | No PR
```

Keep every field except `Delivery` textually identical across the two variants. Express included results and completion semantically so plan renames or decomposition cannot silently change membership.

## Delivery Recommendation

Treat PR creation as the point at which a completed change slice is sent for CodeRabbit review. Estimate the expected review surface from repository evidence, including the likely diff and breadth of affected behavior or contracts; size the implementation change, not the effort required to discover or produce it.

- Recommend **PR delivery** when the goal should produce one or more substantial, cohesive CodeRabbit review slices on its own. Review them during the goal so later work does not make a slice oversized.
- Recommend **No PR** when the goal's expected change is too small for a useful standalone CodeRabbit review slice. Preserve its coherent commits for aggregation with future related changes into a later PR and request CodeRabbit review then.

When the evidence is close, choose the option that yields the fewest substantial, cohesive review slices without allowing one to become oversized. The recommendation decides when CodeRabbit reviews the changes, not whether it eventually reviews them.

## PR Delivery

Also require:

- the fewest substantial, cohesive PRs that keep each CodeRabbit review manageable. Combine small or tightly coupled slices instead of making one PR per plan item, but do not let the whole large goal accumulate into one oversized PR;
- a non-production integration branch: use the repository's established branch such as `dev` only after verifying that the goal's initial and terminal metadata-only commits can be committed and pushed there directly; otherwise create and push a temporary non-production integration branch before work begins. Invoke `$progress` in goal mode there to initialize, commit, and push the goal-state file before creating the first work branch. Create each working branch from the current integration branch and target the PR back to it. Do not mutate `main` or open or merge PRs targeting it. Production promotion is outside the goal; leave the integration branch intact for that later workflow;
- a per-PR completion loop: open the PR, request CodeRabbit review for that completed slice, invoke `$address-pr-feedback`, wait for active Codex and CodeRabbit reviews, address all feedback and follow-ups, resolve every review thread, and confirm required validation and checks. Merge only then, update the integration branch, and create the next working branch from that merged state. Keep commits incremental within each PR. Do not create stacked or dependent PRs, start downstream work before the current PR merges, or leave an unreviewed final tail.
- durable terminal bookkeeping: after the final included PR merges, use `$progress` to mark the goal complete, then commit and push only that terminal goal state and truthful project-planning metadata directly on the preflighted integration branch. Treat this metadata-only commit as bookkeeping rather than an implementation tail; never include code or expanded scope. If policy changed and now forbids it, stop and ask instead of opening an unplanned PR or leaving false state.

## No PR

Require invoking `$progress` in goal mode in the selected working branch to initialize and commit goal state before implementation, then completing the selected boundary through implementation, validation, repository-required non-PR review, and incremental commits. Require the terminal goal and project-planning update to be committed before return. State that the completed changes remain available to aggregate with future related work into a substantial later PR and CodeRabbit review. Explicitly prohibit opening or creating a PR and invoking PR-only feedback workflows during this goal.

## Final Check

Leave implementation steps, file inventories, detailed design guidance, test matrices, commands, git details beyond the selected variant's required delivery lifecycle, and status recaps to the goal-running agent and cited repository documents. The compact scope contract, transition invariant, and recovery rule are required routing information, not plan duplication. Include another omitted detail only when the plans do not contain it and the goal would otherwise be ambiguous or unsafe. Do not invent a token budget or turn uncertain choices into instructions.

Verify that the recommendation follows the expected review-surface rule and that both prompts are actionable without the conversation, invoke `$progress` in goal mode, name the same durable state path, carry the same closed scope contract, distinguish project order from authorization, fail closed on lost goal state, and complete before consulting excluded work.
