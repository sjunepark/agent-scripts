# Evaluate a Skill

Use this workflow to determine whether a new, revised, or merged skill is structurally valid, improves task performance, and activates for the right requests. Treat static validation, behavior evaluation, and trigger evaluation as separate gates. Passing one gate does not imply passing another.

## Prepare the Experiment

1. Freeze the candidate version before collecting scored results.
2. Choose a baseline:
   - the previous skill version for a revision;
   - each source skill for a merge; or
   - the same task without the skill for a new capability.
3. Build cases from realistic requests and representative artifacts. Include:
   - ordinary success cases;
   - difficult or failure-prone cases;
   - explicit positive trigger cases for every skill;
   - for manual-only skills, in-scope but uninvoked requests expected not to
     activate;
   - for implicitly discoverable skills, positive cases that imply the need
     without naming the skill; and
   - near-miss negatives that share vocabulary but belong outside the scope.
4. Reserve at least one case as a holdout when the decision is consequential or
   the case set is large enough to support one. Do not tune against its result.
5. Give baseline and candidate runs the same request, artifacts, available tools, constraints, and resource limits.
6. Start every trial in a fresh, isolated context. Do not carry explanations, outputs, or evaluator feedback between trials.
7. Record the candidate and baseline versions, case identifier, condition, trial number, output, observable work log, elapsed time, and failures. Record input/output size, operations, and retries when available.

Use synthetic cases only to cover gaps that real examples cannot exercise. Mark them as synthetic.

## Gate 1: Static Validation

Run the project's skill validator when one exists, then inspect the package directly. Verify that:

- required metadata parses and matches the skill directory;
- the description states both capability and when to use it;
- the entry file is concise enough to load reliably;
- every required resource is named directly, exists, and has a stated use condition;
- paths are relative and remain inside the skill package;
- unused, duplicated, or stale resources are absent;
- scripts and examples are safe, internally consistent, and executable where applicable; and
- runtime guidance does not depend on an undocumented client, interface, or environment.

Stop here if static validation fails. Structural failures make behavior results ambiguous.

## Gate 2: Behavior Evaluation

### 1. Run an observation case

Run one representative case against the baseline and candidate. Treat this pair as discovery, not as scored evidence. Inspect the artifacts and work logs to learn which outcomes distinguish competent behavior from plausible-looking failure.

### 2. Freeze assertions and grading

After the first observation, write assertions before running the remaining cases. This permits evidence-informed criteria without grading later outputs post hoc.

Derive the criteria from the skill contract and observed failure modes. Where
applicable, cover outcome completeness and correctness, authority and safety
boundaries, error recovery, appropriate resource selection, and unnecessary
work or unsupported assumptions.

Separate the criteria into:

- **Critical assertions:** facts or invariants that must hold in every acceptable output.
- **Objective assertions:** machine-checkable properties such as required files, schema validity, exact values, successful commands, or absence of forbidden content.
- **Subjective criteria:** qualities such as clarity, judgment, maintainability, or fidelity. Give each criterion anchored descriptions for low, acceptable, and excellent results.
- **Efficiency measures:** elapsed time, operations, retries, input/output size, and unnecessary work. Use only measures observable under both conditions.

State acceptance thresholds and the relative importance of criteria now. Derive
them from task criticality, baseline performance, sample size, and measurement
resolution; do not invent precision that the evidence cannot support. With a
small sample, report counts and uncertainty instead of precise percentages. Do
not change criteria after seeing scored results; start a new evaluation round if
they must change.

### 3. Run paired trials

For every scored case:

1. Run the baseline and candidate from separate fresh contexts.
2. Randomize or alternate their order to reduce ordering effects.
3. Repeat each condition at least three times when outputs or execution paths can vary.
4. Preserve all outputs, including failed and unusually slow trials.
5. Label outputs with neutral identifiers before subjective review.

Use more trials when results are unstable or the decision is consequential.
Scale the experiment to the decision: a low-risk diagnostic may need only an
observation pair, while adoption or removal needs representative scored cases.
A single successful demonstration is not evidence of reliability.

### 4. Grade the evidence

- Evaluate objective assertions mechanically where practical.
- Evaluate subjective criteria against the frozen anchors, not against general impressions.
- Use blind human review when quality depends materially on taste, domain judgment, or user trust. Present outputs in randomized order without revealing the condition.
- Report each critical failure separately; do not hide it inside an average score.
- Compare paired results per case, then summarize pass rate, score difference, variance, failure frequency, and efficiency difference.
- Inspect whether gains occur only in cases that closely resemble the authored examples.

### 5. Decide the behavior gate

The candidate passes only when it:

- satisfies every predeclared critical assertion;
- meets the objective and subjective acceptance thresholds;
- improves or preserves behavior across the representative case set;
- introduces no material regression on any holdout; and
- has acceptable cost and variance for its benefit.

Prefer the simpler baseline when the measured difference is negligible or inconsistent.

## Gate 3: Trigger Evaluation

Evaluate activation independently from task quality. A strong workflow is still
a poor skill if it violates its selected activation policy.

1. Create a balanced set of prompts:
   - explicit positives that invoke the skill directly;
   - in-scope prompts without the skill name, labeled positive only for an
     implicitly discoverable skill and negative for a manual-only skill;
   - near-miss negatives with overlapping terms;
   - unrelated negatives; and
   - ambiguous cases with an explicitly documented expected outcome.
2. Freeze the expected activate/do-not-activate label for each prompt.
3. Present each prompt in a fresh context without extra steering.
4. Repeat each prompt when selection can vary.
5. Record activation as a binary outcome and calculate:
   - positive activation rate;
   - missed-positive rate;
   - negative rejection rate; and
   - false-activation rate.
6. Inspect errors by case type instead of relying only on aggregate rates.

Set thresholds before scoring. Explicit positives should normally activate
reliably. A manual-only skill must reject uninvoked prompts even when they match
its subject. An implicitly discoverable skill must justify its catalog cost by
reliably accepting intended implicit prompts while rejecting near misses. If
behavior is already sound, revise only the description or adapter policy and
rerun this gate; do not distort the workflow to compensate for poor selection.

## Diagnose Failures

| Symptom | Likely cause | Next change |
| --- | --- | --- |
| Static checks fail | Invalid structure or broken resource routing | Fix packaging before evaluation |
| Baseline and candidate both fail | Task, tools, or assertions may be unrealistic | Recheck the case and required capability |
| Candidate fails a known invariant | Guidance is missing, ambiguous, or too weak | Add the smallest explicit rule or validation step |
| Candidate succeeds only on copied examples | Guidance is overfit | Replace examples with a general decision rule and test the holdout |
| Results vary widely | Instructions allow unstable choices or the case is underspecified | Clarify the decision point, then add trials |
| Candidate improves quality but costs much more | Workflow is redundant or overly detailed | Remove low-value steps and remeasure |
| Subjective scores conflict | Rubric anchors are vague or reviewer context differs | Tighten anchors and repeat blind review |
| Positive prompts do not activate | Description omits user language or use conditions | Add concrete capability and trigger phrases |
| Near-miss negatives activate | Scope boundary is too broad | Add a concise exclusion or narrow the capability statement |
| Manual-only skill activates implicitly | Adapter policy is absent or inconsistent | Disable implicit invocation in client metadata and retest |

Change one meaningful variable at a time when diagnosing a failure. Start a new round with a new version identifier, fresh contexts, and frozen criteria.

## Stop Criteria

Ship or adopt the candidate only after all three gates pass. Stop and revise when a failure has a plausible corrective change. Stop and retain the baseline when repeated rounds show no material, reliable benefit, or when the added complexity outweighs the gain. Gather more trials rather than choosing a winner when variance is larger than the observed difference.

Preserve the cases, frozen assertions, raw results, and decision summary so later revisions can be compared against the same evidence. Keep holdout cases unexposed until the final evaluation round.
