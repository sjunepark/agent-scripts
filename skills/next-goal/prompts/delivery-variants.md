# Delivery Variants

Read this file only after establishing that `/goal` is warranted.

## Output Order

By default, return:

1. **Recommended boundary** — a short scope statement and evidence-based rationale.
2. **Delivery recommendation** — **PR delivery** or **No PR**, with a short rationale based on the expected size of the implementation change and its CodeRabbit review surface.
3. **Copy-paste `/goal` prompt** — one labeled fenced block for the recommended delivery, containing only the body to enter after `/goal`.
4. One sentence offering the alternate delivery prompt on request.

When the user explicitly requests one named delivery variant, return that prompt even if it differs from the recommendation; keep the recommendation evidence-based, label the prompt as requested, and omit the default offer of another variant.

When the user explicitly asks for both variants, return two labeled fenced blocks:

- **PR delivery** — require sequential PR creation, review, and merge.
- **No PR** — require completion without PR creation.

Assume each new session has none of this conversation. Make each prompt self-contained as a routing instruction, not a detailed specification. Put exclusions inside each fenced prompt; a summary outside the prompt cannot carry authorization into the fresh session. Point to existing repository documents for implementation detail without copying their contents.

Use one compact structured contract followed by no more than two short execution paragraphs: one common paragraph and one delivery-specific paragraph. Target 300–500 words inside the recommended fenced prompt. Exceed that soft range only when shortening would make scope ambiguous or unsafe.

Use this field order:

```text
Goal contract
- Outcome:
- Goal state:
- Included results and sources (semantic results define scope; paths supply detail):
  - <semantic result> — <source path or paths>
- Complete when:
- Excluded:
- Authority: Only included results and the smallest bounded work necessary for their completion; record other work without executing it; expansion requires explicit user instruction.
- Resume: Invoke $progress in goal mode to initialize this exact contract before initial work and recover it before every resume, automatic continuation, compaction recovery, or handoff; fail closed if recovery fails.
- Delivery: PR delivery | No PR
```

Pair each semantic result with its implementation source instead of repeating parallel `Sources` and `Included results` lists. Every child bullet must name an authorized semantic result. Attach a cross-cutting source to the relevant result bullets when needed; omit routing-only documents already discovered through applicable instructions and `$progress`. Never turn project order, a roadmap, or `AGENTS.md` into an included result. Express completion semantically so plan renames or decomposition cannot silently change membership. Keep it delivery-neutral by requiring the selected `Delivery` lifecycle to finish instead of copying PR or no-PR mechanics into `Complete when`.

After the contract, use one common execution sentence: follow applicable `AGENTS.md` and relevant implementation skills, keep planning truthful, preserve unrelated changes, perform repository-required validation and review, and commit meaningful passing units incrementally. Do not repeat `$progress`, scope, recovery, or terminal mechanics already owned by the contract and `$progress`.

State each invariant once. Delegate standard PR creation and feedback mechanics to `$create-pr` and `$address-pr-feedback`; do not reproduce their workflows. When both variants are requested, keep every contract field except `Delivery` textually identical.

## Delivery Recommendation

Treat PR creation as the point at which a completed change slice is sent for CodeRabbit review. Estimate the expected review surface from repository evidence, including the likely diff and breadth of affected behavior or contracts; size the implementation change, not the effort required to discover or produce it.

- Recommend **PR delivery** when the goal should produce one or more substantial, cohesive CodeRabbit review slices on its own. Review them during the goal so later work does not make a slice oversized.
- Recommend **No PR** when the goal's expected change is too small for a useful standalone CodeRabbit review slice. Preserve its coherent commits for aggregation with future related changes into a later PR and request CodeRabbit review then.

When the evidence is close, choose the option that yields the fewest substantial, cohesive review slices without allowing one to become oversized. The recommendation decides when CodeRabbit reviews the changes, not whether it eventually reviews them.

## PR Delivery

Also require:

- use `$progress`'s PR-delivery lifecycle for integration-branch handling, goal-state persistence, and terminal bookkeeping instead of restating that lifecycle;
- use the fewest substantial sequential PRs that keep CodeRabbit review manageable;
- use `$create-pr` for each completed slice and `$address-pr-feedback` through review completion before merging and starting the next slice; prohibit stacked PRs, downstream work before the current merge, and an unreviewed implementation tail.

## No PR

Require `$progress`'s no-PR delivery lifecycle and repository-required non-PR review. Preserve the result for later aggregation into a substantial reviewed PR. Prohibit PR creation and PR-only feedback workflows during this goal.

## Final Check

Leave implementation steps, file inventories, detailed design guidance, test matrices, commands, routine git mechanics, and status recaps to the goal-running agent, invoked skills, and cited repository documents. Include an omitted detail only when the plans do not contain it and the goal would otherwise be ambiguous or unsafe. Do not invent a token budget or turn uncertain choices into instructions.

Verify that the recommendation follows the expected review-surface rule and that every emitted prompt is actionable without the conversation, invokes `$progress` in goal mode, names a durable state path, carries the closed scope contract, fails closed on lost goal state, and stays near the soft length target. When both variants are emitted, verify that their boundaries and contracts match.
