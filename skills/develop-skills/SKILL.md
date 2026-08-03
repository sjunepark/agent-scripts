---
name: develop-skills
description: "Develop, consolidate, audit, and empirically evaluate portable Agent Skills. Use when the task is to create or revise a SKILL.md or bundled resource, merge overlapping skills, audit an existing skill's quality or portability, compare a skill candidate with a baseline, or test a trigger description before publishing or removing predecessors."
---

# Develop Skills

Build skills from observed tasks and failures, then keep only the guidance that
reliably improves future runs. Treat a skill as a reusable capability with one
coherent purpose, not as a storage place for accumulated advice.

## Route the work

- For a new skill or a revision, follow [Create or revise](workflows/create-or-revise.md).
- For overlapping or superseded skills, follow [Merge skills](workflows/merge.md).
- For static, behavior, or trigger testing, follow [Evaluate a skill](workflows/evaluate.md).
- During any authoring or review pass, apply the [Authoring rubric](rubrics/authoring.md).
- Whenever structure, metadata, resources, or dependencies may limit reuse, apply
  the [Portability contract](references/portability.md).

Read only the paths relevant to the requested operation, except that every
candidate must pass the authoring rubric and portability contract before it is
considered complete.

## Shared workflow

1. **Define the desired behavior.** Collect realistic examples of what should
   improve, what currently fails, and what should remain unchanged. Record the
   user's constraints and the completion evidence before editing.
2. **Inspect the effective skill.** Read the current entry point, every runtime
   resource it routes to, applicable repository instructions, validation rules,
   and representative outputs or failures. Do not design from a filename alone.
3. **Choose the smallest coherent scope.** Confirm that reusable procedural
   guidance is the right solution. Prefer a direct answer, ordinary project
   documentation, or executable enforcement when the need is one-off,
   explanatory, or mechanically checkable.
4. **Write trigger cases first.** Include clear positives, paraphrases, implicit
   requests, and near-miss negatives. The description must say both what the
   skill does and when it applies without claiming unrelated work.
5. **Author the minimum candidate.** Keep universally required rules and routing
   in `SKILL.md`. Move optional or detailed material into directly linked,
   purpose-named resources. Remove duplicated, obvious, obsolete, and no-op
   guidance.
6. **Validate in layers.** Check structure and links first, then compare candidate
   and baseline behavior in fresh isolated runs, then test trigger accuracy
   separately. Judge artifacts and decisions, not confident prose.
7. **Iterate from evidence.** Diagnose each failure as a scope, trigger,
   instruction, resource, environment, or evaluation problem. Make the smallest
   change that addresses the observed cause and rerun the affected cases.
8. **Complete the transition.** Publish, install, rename, or remove predecessors
   only after the replacement passes its defined checks. Preserve unrelated
   state and report anything that could not be verified.

## Authoring principles

- Give each instruction one owner and one purpose. Resolve conflicts instead of
  stacking alternatives or compatibility prose.
- State the desired action directly. Use exact prohibitions when violating a
  safety boundary, invariant, or fragile sequence would be costly.
- Calibrate specificity to fragility: leave judgment where several approaches
  work; provide exact steps, templates, or validation where consistency matters.
- Prefer one strong default over a menu of equal options. Document a branch only
  when evidence shows that the branch changes the correct procedure.
- Put unexpected failure modes beside the step they affect. Keep background,
  rationale, and lookup material out of the execution path unless needed to act.
- Use checklists for completeness. Enforce a strict sequence only when order is
  semantically required or evaluation shows premature completion.
- Make completion observable with concrete artifacts, assertions, or validation
  results. “Looks good” is not evidence.
- Treat stylistic formulas as hypotheses. Keep them only when trigger or behavior
  tests show a benefit without harmful false positives.

## Working boundaries

- Follow the requested operation. An audit does not authorize edits. Creating or
  revising a candidate does not authorize registration, installation,
  publication, or removal unless the user includes it.
- Treat source deletion, registry changes, and uninstalling an active copy as
  separate operations. Resolve the scope and evidence for each one.
- Do not encode assumptions about a particular model, client, invocation policy,
  or branded tool in portable runtime guidance. Express required capabilities and
  outcomes instead.
- Isolate optional interface metadata from runtime instructions. Do not make the
  skill depend on that metadata for correctness.
- Preserve licenses and attribution when reusing protected material. Prefer an
  original synthesis of principles over copying source wording.
- Use fresh isolated workers or sessions for independent trials when available;
  do not let the authoring conversation substitute for clean-context evidence.

## Completion report

Report the resulting scope, files changed, validation performed, evaluation
evidence, removals or migrations completed, and remaining uncertainty. Distinguish
measured results from recommendations and inferences.

Maintainers only: follow [Refresh upstream guidance](workflows/refresh-guidance.md) when explicitly requested.
