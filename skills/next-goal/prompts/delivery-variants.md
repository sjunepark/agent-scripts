# Delivery Variants

Read this file only after establishing that `/goal` is warranted.

## Output Order

Return:

1. **Recommended boundary** — a short scope statement and evidence-based rationale.
2. **Delivery recommendation** — **PR delivery** or **No PR**, with a short rationale based on the expected size of the implementation change and its CodeRabbit review surface.
3. **Copy-paste `/goal` prompts** — exactly two labeled fenced blocks, each containing only the body to enter after `/goal`:
   - **PR delivery** — require sequential PR creation, review, and merge.
   - **No PR** — require completion without PR creation.
4. **Next excluded area** — one short statement when it clarifies the stopping boundary.

Assume each new session has none of this conversation. Make each prompt self-contained as a routing instruction, not a detailed specification. Point to existing repository documents for detailed guidance; do not summarize or duplicate their contents.

Usually write one to three short paragraphs per variant containing only:

- the outcome and stopping boundary;
- the few authoritative plan documents to follow;
- verified state or a completion condition only when needed to disambiguate the plans;
- a direction to follow applicable `AGENTS.md`, preserve unrelated changes, keep plans current, and perform repository-required validation and review;
- a requirement to commit incrementally as meaningful, passing units finish; do not defer all commits until goal completion or collapse the run into one final commit. Make each message state what changed and why, including non-obvious decisions needed later.

## Delivery Recommendation

Treat PR creation as the point at which a completed change slice is sent for CodeRabbit review. Estimate the expected review surface from repository evidence, including the likely diff and breadth of affected behavior or contracts; size the implementation change, not the effort required to discover or produce it.

- Recommend **PR delivery** when the goal should produce one or more substantial, cohesive CodeRabbit review slices on its own. Review them during the goal so later work does not make a slice oversized.
- Recommend **No PR** when the goal's expected change is too small for a useful standalone CodeRabbit review slice. Preserve its coherent commits for aggregation with future related changes into a later PR and request CodeRabbit review then.

When the evidence is close, choose the option that yields the fewest substantial, cohesive review slices without allowing one to become oversized. The recommendation decides when CodeRabbit reviews the changes, not whether it eventually reviews them.

## PR Delivery

Also require:

- the fewest substantial, cohesive PRs that keep each CodeRabbit review manageable. Combine small or tightly coupled slices instead of making one PR per plan item, but do not let the whole large goal accumulate into one oversized PR;
- a non-production integration branch: use the repository's established branch such as `dev`, or create a temporary integration branch when none exists. For each PR, create a fresh working branch from the current integration branch and target the PR back to it. Do not mutate `main` or open or merge PRs targeting it. Production promotion is outside the goal; leave the integration branch intact for that later workflow;
- a per-PR completion loop: open the PR, request CodeRabbit review for that completed slice, invoke `$address-pr-feedback`, wait for active Codex and CodeRabbit reviews, address all feedback and follow-ups, resolve every review thread, and confirm required validation and checks. Merge only then, update the integration branch, and create the next working branch from that merged state. Keep commits incremental within each PR. Do not create stacked or dependent PRs, start downstream work before the current PR merges, or leave an unreviewed final tail.

## No PR

Require completing the selected boundary through implementation, validation, repository-required non-PR review, and incremental commits. State that the completed changes remain available to aggregate with future related work into a substantial later PR and CodeRabbit review. Explicitly prohibit opening or creating a PR and invoking PR-only feedback workflows during this goal.

## Final Check

Leave implementation steps, file inventories, design guidance, invariants, test matrices, commands, git details beyond the selected variant's required delivery lifecycle, and status recaps to the goal-running agent and cited repository documents. Include an omitted detail only when the plans do not contain it and the goal would otherwise be ambiguous or unsafe. Do not invent a token budget or turn uncertain choices into instructions.

Verify that the recommendation follows the expected review-surface rule and that both prompts are actionable without the conversation and preserve exactly the same goal boundary and completion contract.
