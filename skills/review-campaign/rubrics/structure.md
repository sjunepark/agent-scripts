# Structure Pass Rubric

Judge shape, not lines. Outputs beyond findings: the area's `## Map` (responsibilities, dependency direction, public surface), keep-as-is verdicts so later passes don't re-litigate, and row-split proposals when the area is too big for one detail session.

## Check

- **One job per module.** Each directory and file should have a statable single responsibility. Flag files that accumulate jobs (service + config + fixtures + operator commands at one level) and name the split.
- **Placement incentives.** Would new code of the kinds the repo actually grows (check the profile's growth domains, TODOs, plans) have one obvious home? Technology-bucket layouts (`utils/`, `helpers/`, framework-named dirs) give weak incentives; domain layouts give strong ones. Young, growing directories are worth reorganizing early; stable ones are not.
- **Dependency direction.** Domain must not import UI or process glue; `shared/` must not become a dumping ground that imports from its consumers; no cycles. Check the actual imports, not the intended layering.
- **Encapsulation.** Compare what a module exports against what consumers use. Flag internals leaked across a boundary, barrel files re-exporting implementation details, and cross-boundary reach-ins (deep relative imports past a module's public surface).
- **Boundary contracts.** IPC and HTTP surfaces: contract, validation guard, and types should sit together with a single source of truth; both sides derive from it rather than re-declare.
- **Cohesion outliers.** God files, parallel structures that must be edited in sync, and sibling modules that duplicate a concern (note duplication across areas for the seam rows rather than re-reporting it per area).

## Anchors

- Honor keep-as-is verdicts recorded in the repo profile, prior Maps, and the repo's existing review docs — verify they still hold rather than re-deriving them.
- Flat-but-predictable layouts can be fine when naming carries the pairing. Predictability is the bar, not depth.

## Do not flag

- Stylistic depth or grouping preferences without a concrete cost (misplacement, sync-edit burden, leaked internals).
- Speculative reorganization of stable code that nothing new will join.
- Small redundancy that keeps modules independent — prefer it to clever sharing.

## Severity and tier

Misplaced responsibility that is actively attracting wrong code: major. Leaked boundary internals: major. Cosmetic grouping: nit. Structural moves are never `auto` — every reorg is triage, proposed as an ordered move list (`git mv` paths, import updates, doc updates, validation commands).
