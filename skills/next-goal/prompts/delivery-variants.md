# Delivery Variants

Read this file only after the user has confirmed a scope, `/goal` is warranted for that scope, and the readiness gate has passed.

## Output Order

By default, return exactly one unlabeled `text` fenced block containing only the body to enter after `/goal`. Do not repeat the earlier scope choices or put a boundary explanation, delivery rationale, prerequisite-mutation recap, validation or review status, copy instruction, label, alternate offer, or any other prose before or after the fence. The confirmed boundary is expressed by the contract, and the evidence-based delivery recommendation is expressed by its `Delivery` field.

When the user explicitly requests one named delivery variant, return exactly one unlabeled `text` fenced block for that variant even if it differs from the evidence-based recommendation. Do not add an explanation of the discrepancy; the user explicitly chose the emitted delivery mechanics.

When the user explicitly asks for both variants, return exactly two `text` fenced blocks and no other prose. Identify each variant through its `Delivery` field:

- **PR delivery** — require sequential PR creation, review, and merge.
- **No PR** — require completion without PR creation.

Assume each new session has none of this conversation. Emit a closed routing envelope whose only job is to carry scope, recovery, and delivery into that session. Let cited repository documents carry requirements, design detail, acceptance criteria, commands, and checklists.

Use as few words as the boundary permits and cap each fenced prompt at 220 words by default. Exceed that cap only when more included results or explicit exclusions are required to keep the boundary closed and unambiguous. Emit exactly one structured contract and keep the complete delivery lifecycle inside its `Delivery` field.

Use this field order:

```text
Goal contract
- Outcome:
- Goal state:
- Included results and sources (semantic results define scope; paths supply detail):
  - <semantic result> — <source path or paths>
- Complete when:
- Excluded:
- Authority: <use the applicable exact authority form below>
- Resume: Initialize this contract with $progress goal mode before work; recover it before every resume, continuation, compaction, or handoff; stop if recovery fails.
- Delivery: <variant-specific lifecycle below>
```

Keep the envelope tight:

- Write `Outcome` as one sentence.
- Name each included result with a short semantic label and the fewest authoritative source paths. Let the sources carry implementation and acceptance detail; applicable instructions and `$progress` supply routing documents.
- Default `Complete when` to: `Every included result achieves its cited outcome and applicable completion criteria within its named semantic boundary; repository-required validation and review pass; planning is truthful; Delivery finishes.` Add only a missing cross-cutting terminal condition needed for this goal.
- Populate `Excluded` from exactly two inputs: the immediate next out-of-scope milestone and exclusions stated directly by the user. Keep every later, unrelated, or merely plan-documented item implicit under `Authority`.
- Let applicable `AGENTS.md`, `$progress`, and the delivery skills supply standard execution behavior. Add execution text only for a missing permission or invariant required in the fresh session.

When both variants are requested, keep every contract field except `Delivery` textually identical.

## Authority

When planning passed the readiness gate without delegation, use:

`- Authority: Execute only included results and necessary supporting work; record anything else and ask before scope expansion or external authority.`

When the user delegated unresolved decisions at the readiness gate, use:

`- Authority: Execute only included results and necessary supporting work; resolve remaining decisions within that closed outcome using best judgment; record anything else and ask before scope expansion or external authority.`

## Delivery Recommendation

Treat PR creation as the point at which a completed change slice is sent for CodeRabbit review. Estimate the expected review surface from repository evidence, including the likely diff and breadth of affected behavior or contracts; size the implementation change, not the effort required to discover or produce it.

- Recommend **PR delivery** when the goal should produce one or more substantial, cohesive CodeRabbit review slices on its own. Review them during the goal so later work does not make a slice oversized.
- Recommend **No PR** when the goal's expected change is too small for a useful standalone CodeRabbit review slice. Preserve its coherent commits for aggregation with future related changes into a later PR and request CodeRabbit review then.

When the evidence is close, choose the option that yields the fewest substantial, cohesive review slices without allowing one to become oversized. The recommendation decides when CodeRabbit reviews the changes, not whether it eventually reviews them.

## PR Delivery

Use this `Delivery` field:

`- Delivery: PR delivery — use $progress's PR lifecycle and the fewest sequential reviewable PRs; finish each through $create-pr and $address-pr-feedback before starting the next, including the final implementation slice.`

## No PR

Use this `Delivery` field:

`- Delivery: No PR — use $progress's no-PR lifecycle, preserve coherent commits for later reviewed aggregation, and reserve PR creation and PR-only feedback workflows for that later delivery.`

## Final Check

Verify that the recommendation follows the expected review-surface rule and every emitted envelope:

- is actionable without the conversation;
- names a durable state path and invokes `$progress` for initialization, recovery, and fail-closed behavior;
- uses short semantic result labels while cited documents carry the detail;
- includes only boundary-relevant exclusions;
- states each invariant once;
- keeps the complete delivery lifecycle inside the persisted `Delivery` field;
- stays within 220 words unless boundary clarity requires more.

When both variants are emitted, also verify that their boundaries and contract fields match except for `Delivery`.

Finally verify that the complete response consists only of the requested `text` fenced prompt block or blocks.
