# End-State Planning

Use this reference automatically while creating or materially revising a plan.
The aim is to write the coherent target that would have been chosen if settled
requirements had been known from the start, without erasing real contracts,
migrations, or rollout obligations.

## Establish the target contract

State what should be true when the work is finished:

- the user-visible or operational outcome;
- the domain rules and ownership boundaries that must survive;
- public, persisted-data, migration, security, compatibility, or rollout
  obligations that constrain the design; and
- material non-goals.

Keep direct user decisions, existing plan claims, implementation evidence, and
external obligations distinct. Do not choose a cleaner architecture by silently
changing product behavior. Ask only when a missing answer materially changes
the target contract and the active goal does not delegate that decision.

## Classify history-shaped complexity

Inspect consequential modes, flags, wrappers, aliases, fallbacks, duplicated
policy, parallel flows, temporary schema, transitional names, and workaround
steps. Give each one an evidence-backed disposition:

- **Remove:** no current consumer, contract, persisted-data need, or migration
  obligation remains.
- **Consolidate:** the behavior is real but its rules or ownership are
  duplicated.
- **Retain temporarily:** a live compatibility, rollout, or migration obligation
  requires it; record the observable exit condition.
- **Retain permanently:** a domain or external contract justifies the ongoing
  complexity.
- **Unresolved:** evidence or an authorized decision is still required.

Search real callers before planning deletion, but do not treat an empty source
search as proof when use may be dynamic, reflective, configured, persisted, or
external. Require stronger evidence in proportion to the cost of being wrong.

## Write the coherent shape

Prefer one clear product path, one owner for each shared rule, and integration
with an existing sound seam. Split a surface only when state, lifecycle,
layout, control, or domain ownership creates a durable distinction. Use
product-intent names rather than names that preserve discussion or prototype
history. Do not add machinery for an imagined future.

Retain staged rollout and backward compatibility only when evidence shows a live
obligation. Preserve approved scope, non-goals, acceptance conditions, and real
dependency ordering. Replace superseded proposal text instead of appending a
decision transcript.

Map the target outcome, every retained obligation, each temporary exit
condition, and every deleted assumption to validation. The plan is coherent
only when future implementation can tell what survives, what disappears, what
must be proved, and which unresolved decision still blocks it.
