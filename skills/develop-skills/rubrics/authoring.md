# Authoring Review Rubric

Use this rubric after drafting and again after behavior evaluation. Record each
answer as `Pass`, `Fail`, or `N/A`, and cite the file, line, or evaluation result
that supports it. `N/A` passes only when its rationale is explicit.

## Severity and decision

- **Blocker:** The skill cannot be selected as intended, followed, validated, or used
  portably; required behavior is unsafe or materially wrong.
- **Major:** A likely user task will trigger incorrectly, stall, skip a required
  result, or produce an unreliable outcome.
- **Minor:** Clarity, efficiency, or maintainability is reduced without making a
  representative task fail.
- **Suggestion:** A plausible improvement lacks evidence that it changes outcomes.

The authoring review passes with no blockers or majors. Minors may remain only
when documented with a reason. Treat suggestions as hypotheses for evaluation,
not mandatory edits.

## 1. Scope and purpose

- [ ] **Pass/Fail — Blocker:** Does the skill solve one coherent class of tasks
  rather than combining unrelated capabilities?
- [ ] **Pass/Fail — Major:** Is the durable, reusable behavior clear from the
  title, description, and opening section?
- [ ] **Pass/Fail — Major:** Would a one-off answer, ordinary reasoning, or an
  existing skill be insufficient for the stated need?
- [ ] **Pass/Fail — Major:** Are explicit boundaries or non-goals present where
  adjacent tasks could be confused with this skill?

## 2. Trigger description

- [ ] **Pass/Fail — Blocker:** Does the frontmatter description state both what
  the skill does and the concrete situations in which it should be used?
- [ ] **Pass/Fail — Major:** Is existing activation intent preserved unless a
  policy change was requested, with new skills defaulting to explicit invocation
  and implicit opt-ins supported by recurrence, matching, and usefulness evidence?
- [ ] **Pass/Fail — Major:** Is installation reach decided separately from
  activation policy, and does client adapter metadata enforce the selected
  policy when that capability exists?
- [ ] **Pass/Fail — Major:** Do the stated triggers match representative positive
  cases without claiming unrelated near-miss cases?
- [ ] **Pass/Fail — Major:** Does the description use observable task language
  rather than vague claims such as "helpful" or "powerful"?
- [ ] **Pass/Fail — Minor:** Has wording been tested against positive and
  near-miss prompts instead of optimized by intuition alone?

## 3. Entry-point quality

- [ ] **Pass/Fail — Blocker:** Is the entry file valid, self-contained enough to
  start the task, and free of required knowledge that is only implied?
- [ ] **Pass/Fail — Major:** Does the opening orient the reader to the outcome and
  the first decision or action within a short scan?
- [ ] **Pass/Fail — Major:** Are universal rules and required steps in the entry
  file rather than hidden in optional resources?
- [ ] **Pass/Fail — Minor:** Are examples concrete and representative without
  prescribing one user's answer as the only acceptable output?

## 4. Progressive disclosure and pointers

- [ ] **Pass/Fail — Major:** Is the entry file concise enough to load routinely,
  with conditional or detailed material moved to focused resources?
- [ ] **Pass/Fail — Blocker:** Is every required resource named by an exact,
  valid relative path and accompanied by when-to-read guidance?
- [ ] **Pass/Fail — Major:** Can each resource be reached directly from the entry
  file without following a chain of resource-to-resource pointers?
- [ ] **Pass/Fail — Minor:** Are long resources navigable? Use a contents list or
  split only when observed lookup friction justifies it; no line-count threshold
  is universally correct.

## 5. Calibrated specificity

- [ ] **Pass/Fail — Major:** Are fragile, safety-critical, or exact operations
  expressed as precise constraints, commands, schemas, or validation steps?
- [ ] **Pass/Fail — Major:** Do judgment-heavy tasks preserve room for adapting to
  evidence instead of forcing brittle, unearned rules?
- [ ] **Pass/Fail — Minor:** Are defaults stated where they reduce indecision, with
  alternatives introduced only for meaningful tradeoffs?
- [ ] **Pass/Fail — Minor:** Are important prohibitions paired with the desired
  behavior when that improves compliance? Positive phrasing is a testable
  heuristic, not a universal requirement.

## 6. Workflow and completion

- [ ] **Pass/Fail — Major:** Do permission gates account for prior authorization
  and block only dependent actions, while preserving deliberate hard boundaries?
- [ ] **Pass/Fail — Major:** Are questions limited to unresolved consequential
  decisions, with the responsible instruction identifiable when work pauses?
- [ ] **Pass/Fail — Major:** Does the workflow identify inputs, key decisions,
  actions, validation, and the completion condition?
- [ ] **Pass/Fail — Blocker:** Can a reader determine when the task is genuinely
  complete rather than merely drafted or attempted?
- [ ] **Pass/Fail — Major:** Are irreversible or high-impact actions preceded by
  the necessary checks and explicit scope resolution?
- [ ] **Pass/Fail — Minor:** Is the workflow represented in the lightest form that
  works? Prefer a checklist unless evaluation shows that enforced sequencing is
  needed to prevent skipped or premature steps.
- [ ] **Pass/Fail — Major:** Are delegation, validation, and output requirements
  proportional to the task, with a clear stopping condition for repeated work?

## 7. Robustness and gotchas

- [ ] **Pass/Fail — Major:** Are known failure modes handled at the decision point
  where they occur, with a recovery or escalation path?
- [ ] **Pass/Fail — Major:** Are assumptions, prerequisites, and unsupported cases
  explicit enough to prevent silent misuse?
- [ ] **Pass/Fail — Major:** Does the skill preserve user data and unrelated work,
  especially around destructive or state-changing operations?
- [ ] **Pass/Fail — Minor:** Are edge cases included because evidence or domain
  risk warrants them, rather than as speculative completeness?

## 8. Resources and portability

- [ ] **Pass/Fail — Blocker:** Do all bundled runtime files stay within the skill
  directory and avoid machine-specific absolute paths?
- [ ] **Pass/Fail — Blocker:** Does general guidance avoid hidden model or host
  dependencies, while declaring any named tool or platform intrinsic to the job?
- [ ] **Pass/Fail — Major:** Does each script, reference, template, or asset have a
  clear runtime purpose and an exact pointer from the entry file when needed?
- [ ] **Pass/Fail — Major:** Are scripts deterministic where practical, explicit
  about inputs and outputs, and safe to rerun?
- [ ] **Pass/Fail — Minor:** Are optional resources omitted when the entry file can
  express the same behavior more clearly and cheaply?

## 9. Evaluation evidence

- [ ] **Pass/Fail — Major:** Have changed runtime decisions been exercised in
  fresh, isolated contexts with validation scope set from the change's risk?
- [ ] **Pass/Fail — Major:** Do assertions test observable outcomes rather than
  only stylistic resemblance to an expected response?
- [ ] **Pass/Fail — Major:** Where revising or merging, is there baseline evidence
  showing that the candidate preserves or improves important behavior?
- [ ] **Pass/Fail — Major:** When discovery changes, are trigger tests separate
  from behavior tests with positive, negative, and ambiguous cases? For unchanged
  discovery, is the consistency check and lack of new trigger evidence explicit?
- [ ] **Pass/Fail — Minor:** Have variable outcomes been repeated enough to reveal
  instability, and has human review covered qualities automation cannot judge?
- [ ] **Pass/Fail — Minor:** Are numeric thresholds justified by risk, baseline
  performance, sample size, and measurement resolution rather than arbitrary
  precision?

## 10. Pruning and final quality

- [ ] **Pass/Fail — Major:** Does every instruction change a decision or action,
  supply necessary knowledge, or define an observable completion condition?
- [ ] **Pass/Fail — Major:** Have duplicated guidance, background exposition,
  generic advice, and material already supplied by the task been removed?
- [ ] **Pass/Fail — Minor:** Are specialized terms introduced only when they make
  recurring reasoning shorter or more precise?
- [ ] **Pass/Fail — Minor:** If distinctive leading words or labels are used, did
  evaluation show that they improve triggering or execution? Treat them as an
  optional tested technique, not a requirement.
- [ ] **Pass/Fail — Minor:** Did a final read remove contradictions, stale paths,
  orphaned resources, and instructions unsupported by evaluation evidence?
