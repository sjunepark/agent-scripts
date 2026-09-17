# Architecture and Code-Design Lens

Judge the target's fit with surrounding modules and consumers. Follow affected
contracts without turning a change review into a whole-codebase redesign.

Use relevant architecture decisions, domain language, callers, dependencies,
state owners, and tests to establish the affected boundaries. Compare before and
after for a diff; distinguish introduced effects from existing debt.

## Criteria

- **Cohesion:** behavior, state, and invariants that change together have a clear
  owner. Look for unrelated responsibilities or fragments whose coordination
  is the actual behavior.
- **Dependencies:** check cycles, cross-module reach-ins, deep imports, and callers
  coordinating another module's internals. Honor intended ownership and layering.
- **Interface depth:** a useful small surface hides meaningful knowledge.
  Removing a shallow abstraction merely removes a hop; removing a deep one
  pushes complexity into consumers. Ordering, failures, configuration, and
  performance assumptions may be part of the contract.
- **Seams and locality:** require concrete variation for ports or plug-in points.
  Ground likely changes in requirements, history, or an explicit roadmap; look
  for a single domain change requiring edits across unrelated layers.
- **Contract ownership:** validation, schemas, errors, and lifecycle rules should
  not drift across producers and consumers.
- **Testability:** deep mocks, internal assertions, and test-only public hooks can
  reveal misplaced seams; prefer observable behavior through real interfaces.

Trace representative paths and material failures across the affected boundaries.
A finding needs a concrete correctness, maintenance, testability, or change-locality
cost, affected consumers, the smallest viable alternative, and its tradeoff.
Taste, file size, and pattern unfamiliarity are insufficient. Preserve useful
explicitness and duplication that keeps modules independent.

Repository decisions outweigh generic patterns; reopen them when assumptions or
observed costs changed. Broader design changes belong in Bucket II under the
entry point's authority policy. Record missing consumers or context as residual
risk and meaningful justified shapes as keep-as-is.

Adapted from Matt Pocock's MIT-licensed
[codebase-design](https://github.com/mattpocock/skills/blob/9603c1cc8118d08bc1b3bf34cf714f62178dea3b/skills/engineering/codebase-design/SKILL.md),
[architecture](https://github.com/mattpocock/skills/blob/9603c1cc8118d08bc1b3bf34cf714f62178dea3b/skills/engineering/improve-codebase-architecture/SKILL.md),
and [code-review](https://github.com/mattpocock/skills/blob/9603c1cc8118d08bc1b3bf34cf714f62178dea3b/skills/engineering/code-review/SKILL.md)
guidance.
