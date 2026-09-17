# Create or Revise a Skill

Start from the requested outcome and representative tasks, artifacts, or failures.
For a revision, inspect the current entry point and resources governing the
affected decisions. Preserve useful behavior; investigate surrounding instructions
or missing capabilities before blaming the skill.

## Choose the contract

Identify the reusable help the skill provides beyond ordinary reasoning: a
non-obvious decision, domain constraint, fragile operation, or useful resource.
A one-off answer needs no package. When evidence is missing, use a realistic
scenario and label its assumptions rather than demanding a discovery interview.

Define the intended result, scope, prerequisites, and observable completion.
Keep one coherent capability; add branches only when they change how that
capability is delivered.

Preserve the activation policy required by the entry point. Use concise task
language in the description; add an exclusion only to prevent likely misrouting.
For changed discovery, retain representative positive and near-miss requests,
including uninvoked negatives for manual-only skills. Keep client enforcement in
adapter metadata.

## Write the smallest useful package

Keep shared constraints and routing in `SKILL.md`. A short skill may be entirely
self-contained. Split substantial conditional guidance into directly linked,
purpose-named resources and say when to read them. Add scripts or assets only
when repeatable mechanics or output requirements justify maintaining them.

Describe outcomes and decision criteria where judgment works. Use exact commands,
checks, or ordering where deviation threatens correctness, data, permissions, or
a fragile operation. Remove duplicated instructions, generic tutorials, stale
workarounds, and rules that merely restate the task or host behavior.

Treat the instruction stack as context with a cost: descriptions should make
selection easy, and using one route should not force unrelated reading. Keep
shared collaboration defaults at their existing owner.

## Finish the revision

Apply the entry point's authoring review and relevant evaluation checks. Exercise
changed decisions and important boundaries; do not demand a full experiment for
a wording-only fix. Address evidenced failures, report unavailable checks, and
complete any already-authorized transition. Do not claim measured improvement
from a shorter file or a persuasive explanation alone.
