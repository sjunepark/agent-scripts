---
name: end-state-review
description: "Review a plan or in-progress change from its intended final state, separating real compatibility and migration obligations from historical residue. Explicit invocation only; use after planning discussions, requirement changes, prototypes, or rollout completion when the user wants a coherent end-state proposal before implementation."
---

# End-State Review

Reconstruct the target as if the settled requirements had been known from the
start. Remove history-shaped complexity from the proposal without erasing real
contracts, migrations, or rollout obligations.

Default to review-only. After explicit approval, update planning artifacts
only. Do not modify implementation in this workflow; leave code changes to a
separate implementation request.

## Resolve the Review Target

1. Review the user-named plan, roadmap item, specification, change, diff, or
   migration. When no artifact is named, infer the smallest plausible target
   from the current discussion and repository planning state. Ask when
   materially different targets remain plausible.
2. Read the target completely. Inspect governing requirements, direct user
   decisions, relevant implementation, callers, tests, history, and external
   contracts only far enough to test the proposed final shape.
3. Keep direct user intent, planning claims, implementation evidence, and
   external obligations distinct. A proposal does not prove an obligation, and
   absence from the plan does not prove that a deployed contract is dead.
4. Bound the review to complexity introduced, retained, or locked in by the
   target. Mention unrelated debt only when it invalidates the proposal; do not
   absorb it into the review.

## State the End-State Contract

Describe in a few sentences what should be true when the work is finished:

- the user-visible outcome;
- the domain rules and ownership boundaries that must survive;
- the public, persisted-data, migration, security, or rollout obligations that
  constrain the design;
- material non-goals.

Ask only when a missing answer could materially change that contract. Do not
choose a cleaner architecture by silently changing product behavior.

## Separate Residue from Obligations

Inspect modes, flags, wrappers, aliases, fallbacks, duplicated policy, parallel
flows, temporary schema, transitional names, and workaround plan steps. For
each consequential candidate, record its evidence and classify it as:

- **Remove** — no current consumer, contract, or migration need remains.
- **Consolidate** — the behavior is real but its rules or ownership are
  duplicated.
- **Retain temporarily** — a current rollout or migration requires it; name the
  exit condition.
- **Retain permanently** — a domain or external contract justifies the cost.
- **Unresolved** — evidence or a user decision is still required.

Search real callers before recommending deletion, but do not treat an empty
repository search as proof when use may be dynamic, reflective, configured,
persisted, or external. Require stronger evidence in proportion to the cost of
being wrong.

## Reconstruct the Coherent Shape

Design from the end-state contract rather than replaying the sequence of
earlier decisions.

- Prefer one clear product path over mode flags or parallel implementations.
- Give each shared rule, such as permissions, routing, URL state, feature
  gating, or command naming, one evident owner.
- Integrate with an existing sound boundary before adding a side channel,
  adapter, or new abstraction.
- Split a surface only when state, lifecycle, layout, control, or domain
  ownership creates a durable boundary.
- Prefer product-intent names over names that preserve implementation history.
- Do not invent a framework for one current use or preserve machinery for an
  imagined future.
- Keep staged rollout and backward compatibility when evidence shows they are
  live obligations.
- Keep the proposed rework no broader than necessary to make the target
  coherent.

## Reconcile the Target

For a planning target, identify obsolete steps and propose the smallest set of
replacements needed to express the coherent end state. Preserve approved
scope, non-goals, acceptance conditions, and sequencing that reflects real
dependencies rather than historical discussion order.

For an implementation target, describe the desired code shape, likely
deletions, and affected consumers. Stop at recommendations even when a fix
looks safe. Code mutation requires a separate request.

Map each end-state outcome and retained obligation to validation. Include
tests for deleted assumptions when navigation, permissions, persistence,
external consumers, or rollout behavior could regress.

## Apply the Planning Gate

Keep the repository unchanged during the initial review. Treat an explicit
instruction such as `apply these plan changes` or `update the plan` as approval
to edit only the reviewed planning artifacts. Acknowledgement, continued
discussion, or agreement with the analysis is not edit approval.

When applying an approved planning revision:

1. Replace superseded proposal text instead of appending a conversation log.
2. Record the end-state outcome, retained obligations and exit conditions,
   material non-goals, validation, and the next implementation action. Link to
   an existing contract source instead of duplicating requirements that it
   owns.
3. Preserve the planning system's ownership and queue invariants.
4. Re-read the touched files and verify that no old step contradicts the
   approved final shape.
5. Stop without starting implementation or changing code.

## Report

Return only sections that carry information:

- **End state** — the reconstructed final contract.
- **Residue and obligations** — candidates, evidence, and disposition.
- **Proposed shape** — the coherent design and material tradeoffs.
- **Plan amendments** — replacements needed or approved edits made.
- **Verification** — behavior and deleted assumptions to prove.
- **Decisions required** — unresolved load-bearing choices.
- **Scope boundary** — relevant work deliberately left outside the review.

A defensible conclusion that the target is already coherent is valid. Do not
manufacture debt or claim that any review can guarantee zero technical debt.
