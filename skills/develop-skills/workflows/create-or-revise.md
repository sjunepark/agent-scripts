# Create or Revise a Skill

Use this workflow to turn observed task friction into one focused, portable
skill. For a revision, begin with the current skill and preserve behavior that
still earns its place.

## 1. Collect Task Evidence

Start with a real task, representative artifact, failure report, or repeated
request. Record:

- the outcome the user needed;
- the inputs and environment available at the start;
- where an unaided attempt became inconsistent, slow, error-prone, or
  incomplete;
- the non-obvious decisions, reusable resources, or validation steps that
  changed the result; and
- what a successful artifact or response looked like.

Do not begin by drafting generic advice. If no real example is available,
construct the smallest realistic scenario and label its assumptions so they can
later be replaced with evidence.

For a revision, inspect recent uses and failures before editing. Distinguish a
problem in the skill from a problem in its inputs, surrounding instructions,
or unavailable capabilities.

## 2. Decide Whether a Skill Is Warranted

Create or retain a skill when the evidence shows a reusable procedure,
specialized judgment, fragile sequence, recurring tool interaction, or useful
bundled resource. Prefer ordinary project documentation when the material is
only reference information for humans. Handle a one-off answer directly rather
than preserving it as a skill.

State one sentence for each of these before continuing:

- **Reusable outcome:** what becomes reliably easier or better.
- **Distinctive help:** what the skill contributes beyond ordinary reasoning.
- **Expected reuse:** which future tasks share the same need.

Stop if these statements are vague or describe unrelated outcomes. Gather more
evidence, narrow the proposal, or split genuinely independent concerns into
separate skills.

## 3. Frame the Skill Contract

Define the boundary before choosing files:

1. Name the starting state and intended end state.
2. List required inputs and capabilities.
3. Identify permitted changes and any actions that require explicit authority.
4. State material exclusions and nearby tasks that belong elsewhere.
5. Write observable completion criteria.

Keep one coherent workflow under one skill. A skill may have branches when
they share the same outcome, vocabulary, and validation boundary; split them
when they can be selected, executed, and evaluated independently.

## 4. Build the Trigger Set

Collect concrete requests before writing the description:

- at least three positive examples that vary wording, specificity, and
  starting context;
- at least three near-miss examples that share vocabulary but require a
  different workflow; and
- any implicit request that should activate the skill even without naming it.

For each example, record why the skill should or should not apply. Adjust the
scope if reasonable reviewers cannot classify the examples consistently.

## 5. Plan the Minimum Package

Assign each needed piece of content to the smallest suitable location:

- Put universal decisions, routing, and completion rules in `SKILL.md`.
- Put a substantial alternate procedure in a directly linked workflow file.
- Put lookup material in a directly linked reference file.
- Use a script only when deterministic execution or repeated mechanics justify
  maintaining code.
- Use a template or asset only when the output benefits from an exact reusable
  starting point.

Name every planned resource and state exactly when the skill will load or use
it. Remove resources that merely repeat `SKILL.md`, preserve background
reading, or are not reachable from the entry point. Keep navigation shallow so
the executor never has to discover a directory by exploration.

## 6. Author the Guidance

Write the smallest instructions that reproduce the successful decisions from
the task evidence:

1. Use valid portable frontmatter with a stable name and a description that
   states both what the skill does and when it applies.
2. Lead with the workflow and defaults. Use imperative steps, explicit branch
   conditions, and observable stopping criteria.
3. Match specificity to fragility: leave room for judgment where several
   approaches work, but give exact checks or sequences where small deviations
   cause failure.
4. Prefer positive directions that name the desired action. Reserve strict
   prohibitions for safety, irreversible effects, or repeatedly observed
   failure modes.
5. Place surprising constraints and recovery guidance beside the step they
   affect.
6. Include examples only when they resolve ambiguity that the procedure cannot
   express more directly.

Delete generic background knowledge, exhaustive menus, motivational prose,
conversation history, and duplicated source documentation. Replace repeated
explanation with a default, decision rule, checklist, or reusable resource.

For a revision, make the smallest change supported by the failure evidence.
Avoid preserving obsolete behavior through compatibility branches unless an
active consumer requires them.

## 7. Run Static Validation

Before behavior testing, use the repository's validator and inspect the
package directly. Confirm that:

- frontmatter parses and the directory name matches the skill name;
- every runtime resource is named or directly linked from `SKILL.md` with a
  condition for use;
- all referenced paths exist, remain inside the skill, and contain no circular
  routing;
- scripts, templates, and examples are syntactically valid where applicable;
  and
- the package contains no unused, duplicated, or unexpectedly large material.

Fix structural failures before interpreting behavior results.

## 8. Evaluate Behavior in Fresh Contexts

Select a small task set covering a straightforward case, a boundary case, and
the most failure-prone case from the evidence. Run each task twice in separate,
fresh contexts under the same conditions: once without the candidate skill and
once with it.

Before running, define the essential acceptance criteria from the skill
contract. After seeing the first pair, add only discriminating assertions that
capture observable quality without changing the intended outcome to favor one
run. Compare:

- completion of the requested outcome;
- correctness and completeness of the artifact or response;
- unnecessary steps, rework, or unsupported assumptions;
- adherence to authority and safety boundaries; and
- errors, recovery quality, and useful resource selection.

Keep outputs labeled but ask an independent reviewer to judge quality without
knowing which path produced each one when practical. Treat timing or length as
secondary evidence, not a substitute for correctness.

Trace each failure to the smallest plausible cause: missing instruction,
ambiguous branch, excess prescription, poor resource routing, unsuitable
scope, or inadequate capability. Change one cause at a time and rerun the
affected pairs in fresh contexts. Remove instructions that do not improve the
result.

## 9. Evaluate Triggering Separately

Test the description against the saved positive and near-miss examples without
loading the full skill. Repeat borderline examples across fresh contexts.
Revise the description or scope until positives activate reliably and near
misses remain outside the boundary. Do not compensate for an unclear scope with
a long list of keywords.

## 10. Finish

Consider the skill ready only when static validation passes, the candidate
improves or preserves behavior on the representative tasks, no regression is
unexplained, and trigger tests support the intended boundary. Record remaining
limitations as concrete follow-up evidence rather than speculative features.
