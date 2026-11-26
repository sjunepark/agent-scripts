# System and Technical-Decision Lens

Determine whether the implementation delivers the intended outcome within the
system's constraints. Scope follows affected requirements and consequences,
not unrelated repository opportunities.

## Map delivery

Extract goals, non-goals, constraints, and success signals from governing
intent; expose conflicting sources. Classify relevant outcomes as `delivered`,
`partial`, `missing`, `deferred`, or `unverified`, and material work outside
intent as `extra`. Use `deferred` only when authoritative scope places it beyond
the claimed milestone. Delivery includes the callers, configuration, packaging,
deployment, docs, and upgrade or removal paths the result needs.

## Judge technical choices

Inspect choices introduced, replaced, or locked in when they create durable
contracts or meaningful data, security, operational, support, or reversal costs.
For a criticized choice, compare retaining the existing path and the strongest
credible alternative against actual requirements, including transition,
migration, operating, and exit costs. Verify time-sensitive claims with current
primary sources and label inference. Recommend replacement only for a concrete
net advantage. Weak rationale documentation is different from an unsound choice.
Documented decisions are evidence, not immunity; reopen them for changed
assumptions or observed consequences.

## Report

Report compact delivery coverage with evidence. Distinguish ambiguous intent,
absent implementation, and unavailable verification. Record significant
justified choices as keep-as-is and unexamined material boundaries as residual
risk. Do not produce a wishlist to fill the categories.
