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

Choose activation policy before writing the description or cases. Default to
explicit invocation. Treat implicit discovery as an opt-in that must satisfy all
of these conditions:

- the capability is expected to recur broadly within the locations where the
  skill will be installed;
- ordinary user wording identifies the capability reliably enough to keep both
  missed activations and false activations acceptably low;
- activating without an explicit request is safe and ordinarily useful; and
- the expected benefit justifies occupying persistent skill-catalog context.

Installation reach and activation policy are separate decisions. A globally
installed skill may remain manual-only, while a narrowly installed repository
skill may justify implicit discovery within that repository.

Record the portable activation intent in the description. Keep any
client-specific enforcement in that client's adapter metadata, following any
applicable client guidance routed from the skill entry point. Encode the chosen
value explicitly rather than relying on omission when a client may default
omitted policy to implicit discovery.

Then collect concrete requests:

- at least three positive examples that vary wording, specificity, and
  starting context;
- at least three near-miss examples that share vocabulary but require a
  different workflow;
- for a manual-only skill, positives that explicitly invoke it and in-scope but
  uninvoked requests labeled as near misses; or
- for an implicitly discoverable skill, implicit positives that describe the
  work without naming it.

For each example, record why the skill should or should not apply. Adjust the
scope or retain manual-only activation if reasonable reviewers cannot classify
the examples consistently.

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
   states both what the skill does and when it applies. For a manual-only skill,
   say `Explicit invocation only` rather than advertising implicit triggers.
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

## 7. Hand Off to Evaluation

Return to the entry point and run its evaluation workflow. That workflow owns
static validation, baseline comparison, repeated behavior trials, trigger
testing, and the pass decision; do not substitute an informal spot check or a
single candidate/baseline pair.

Revise the candidate only from observed failures, then start a new evaluation
round under the evaluator's versioning and frozen-criteria rules. Authoring is
complete when the candidate passes every applicable evaluation gate and any
remaining limitation is recorded as concrete evidence rather than a
speculative feature.
