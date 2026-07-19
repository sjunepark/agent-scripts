---
name: code-review
description: "Review code or completed implementation changes for correctness, intent fit, regressions, validation gaps, system and architecture fit, material technical and dependency choices, maintainability, and unearned complexity. Use for code review, roadmap or specification implementation audit, architecture or technical-choice critique, and simplification review. Do not use merely because code is being edited."
---

# Code Review

## Workflow

1. Establish scope and edit policy.
   - If the user names files, a directory, a module or subsystem, a branch, a
     commit, a PR patch, or a diff, review that target.
   - When the user asks to review implementation against a roadmap,
     specification, issue, ADR, or migration plan, treat that document as the
     source of intent and resolve the implementation it governs. Do not reduce
     the target to the document alone.
   - For an initiative review, establish four scope axes: governing intent,
     claimed milestone or outcome, implementation range, and edit policy.
     Inspect along requirement, consumer, and risk edges as needed, while
     keeping findings bounded to those axes.
   - If no target is named and the repository has uncommitted changes, default
     to the current uncommitted change set.
   - If no target is named and no review target can be inferred, ask for the
     target.
   - Honor `review-only`, `do not edit`, and equivalent instructions.
   - Distinguish issues introduced by the reviewed change from pre-existing
     debt.
   - Continue once the target and edit policy are explicit.

2. Run the implementation gate when reviewing completed work, current
   uncommitted changes, or any code review where correctness, regressions, or
   validation matter. Read
   [lenses/implementation-review.md](lenses/implementation-review.md).
   - Continue once material issues, obvious safe fixes, validation gaps, and
     remaining decision points have been checked.

3. Add the system lens when the user asks for a broad, holistic, strategic, or
   end-to-end review. Also add it when the target implements a roadmap,
   specification, ADR, or migration; spans a substantial initiative; or
   introduces, replaces, or commits the system to a material language,
   runtime, framework, direct dependency, protocol, storage, service, build,
   deployment, or infrastructure choice. Read
   [lenses/system.md](lenses/system.md).
   - Follow the named intent and its consequences across affected areas without
     turning a bounded review into an unsolicited whole-repository audit.
   - Continue once every relevant stated outcome, constraint, material choice,
     and applicable cross-cutting consequence has a disposition, and every
     criticized choice has an evidenced alternative and tradeoff.

4. Add the design lens when the user asks about architecture, code design,
   module responsibilities, interfaces, seams, dependency direction,
   coupling, cohesion, or testability. Also add it when the review target
   creates or reshapes a module or public contract, changes dependency
   direction, or spreads one behavior across ownership boundaries. Read
   [lenses/design.md](lenses/design.md).
   - Keep the implementation gate alone for a routine local edit that does not
     alter system shape.
   - Continue once the target's responsibilities, interfaces, dependencies,
     affected consumers, and relevant architecture decisions have been checked.

5. Add the diet lens when the user asks about overengineering, simplification,
   deletion, trimming, wrappers, helpers, schema/config surface, compatibility
   paths, modes, flags, generic layers, or bolted-on design. Also add it when
   the review target itself introduces those surfaces. Read
   [lenses/diet.md](lenses/diet.md).
   - Do not load the diet lens for a routine small edit with no complexity
     signal unless the user asked for that lens.
   - Continue once unearned complexity candidates are either reported or
     explicitly kept as-is.

6. Report in buckets.
   - `Delivery Coverage`: only for roadmap, specification, or initiative
     reviews; use the system lens's outcome classifications and evidence.
   - `Bucket I - Safe Fixes`: safe issues found, marked `applied` or `proposed`
     according to edit policy, with changed or affected files and validation
     evidence, or `none`.
   - `Bucket II - Needs Decision`: real issues or simplifications that need
     user judgment, with the decision needed and the main tradeoff, or `none`.
   - `Keep As-Is`: meaningful concerns inspected and intentionally rejected,
     when useful.
   - `Validation`: commands run and results, or skipped/blocked reason.
   - `Residual Risk`: only when validation, scope, or context leaves a material
     gap.

## Review Rules

- Prefer no finding over a weak finding.
- Record only discrete, actionable issues with affected paths, trigger
  scenarios, and likely maintainer action. When implementation is absent, cite
  the governing requirement, consumer, or system surface instead of inventing
  an affected path.
- Record each issue once under the lens that owns its remedy: implementation
  for concrete behavior, requirement fulfillment, and validation; system for
  initiative outcome coverage, non-goals, readiness, and durable technical
  decisions; design for system shape and contracts; and diet for
  simplification.
- Apply obvious safe Bucket I fixes by default. When the edit policy forbids
  edits, report them as `proposed`.
- Do not implement Bucket II without explicit user approval.
- Keep edits tightly scoped to the reviewed target.
- Do not stage, commit, or push unless explicitly asked.
