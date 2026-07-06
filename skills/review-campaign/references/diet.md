# Pass: diet

Unearned abstraction, speculative generality, and leftovers. The bar for keeping code is a current, concrete need — but the bar for *flagging removal* is verified deadness, because a false "unused" finding costs more than the weight it would save.

## Verify before flagging — mandatory

Before calling anything unused: grep imports *and* indirect references — templates, dynamic keys, config/registry tables, test fixtures, scripts, IPC/contract names. Check the repo's planning docs and TODO files (per the repo profile) before calling anything speculative — scaffolding with a written plan is not speculative, it is early.

## Check

- **Unearned abstraction.** Interfaces with one implementation and no second in sight, factories that build one thing, layers that only pass through, wrappers adding nothing over the callee. Inline them.
- **Speculative generality.** Options and parameters never passed, config flags never set, generics with one instantiation, "extensible" registries with one entry and no plan.
- **Dual paths past their transition.** Compatibility shims, old-and-new implementations both alive, dual-read/dual-write with the migration finished. Name which path stays and what removes the other.
- **Schema minimalism.** DB columns, JSON fields, contract properties not used by current behavior (trace each to a reader). Applies hardest in db schemas and shared contracts.
- **Leftovers.** Unused exports/files, commented-out blocks with a live replacement, stale TODO comments, debug scaffolding, dependencies used trivially or not at all (check against the repo's manifests and lockfiles).
- **Wrong-direction abstraction.** Clever indirection where simple redundancy would read better — small redundancy for stateless collaborators is acceptable. Flag the abstraction, not the duplication.

## Do not flag

- Redundancy that keeps modules independent or code obvious.
- Planned scaffolding (see verify step).
- Generality that already has two real users.
- Code that is merely long — length without one of the smells above is not weight.

## Severity and tier

Dual paths confusing live development: major. Unused schema fields: major in schemas and contracts (they constrain forever), minor elsewhere. Pass-through layers: minor. Removals look like opinions to someone — everything is triage except unused imports and byte-identical dead duplicates; deleting an exported symbol or a file is always triage with the usage-verification evidence attached.
