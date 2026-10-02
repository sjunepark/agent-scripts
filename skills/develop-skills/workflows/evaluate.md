# Evaluate a Skill

Establish what the change needs to prove, then collect evidence for that claim.
Structural validity, task behavior, and discovery are separate questions.

## Select the checks

- **Wording only:** run static validation and focused review when decisions and
  triggers remain unchanged.
- **Changed runtime decisions:** compare the old and new instructions on a
  realistic affected task and a relevant boundary, using existing cases where
  they fit.
- **Changed discovery:** test positive and near-miss requests separately from
  task quality. Preserve the selected invocation policy.
- **Broad redesign, new capability, or predecessor retirement:** cover distinct
  capabilities and failure modes against the baseline. Use holdouts and repeated
  trials when adoption risk or observed variability makes them informative.

Draw cases from observed session failures and reported defects first, then
hand-written tasks, and use synthetic cases last. Note in one line what makes
each case hard; a case the baseline handles easily cannot show improvement.

Set assertions and stopping criteria before scoring. Expand after failures,
candidate changes, unstable outcomes, or an unresolved risk; passing checks do
not need repetition for its own sake.

Check headroom before claiming improvement. When the baseline already passes
nearly every assertion, the suite is a regression guard: add harder cases, or
limit the claim to preserved behavior, cost, or size.

## Structural checks

Run the repository validator when available. Verify metadata and names, direct
resource pointers and their use conditions, self-contained package paths,
declared prerequisites, scripts affected by the change, and licensing obligations.
Fix structural failures before using behavior results to justify adoption.

## Behavioral comparisons

Use the prior version for revisions, each predecessor for merges, or an unaided
attempt for a new capability. Preserve the baseline before editing. Give each
condition the same request, raw artifacts, tools, and limits in separate fresh
contexts. Avoid supplying the desired answer or authoring rationale to the
executing worker. An isolated worker or session is appropriate when available;
if isolation is unavailable, report the limitation.

Judge artifacts and observable decisions. Assertions should cover required
results, important prohibitions, and the failure that motivated the change.
Use an observation case to develop assertions only when needed, and do not score
that same case as independent confirmation. Freeze criteria before scored trials;
changes to criteria or candidate begin a new round.

For subjective judgments, use concrete acceptable/unacceptable anchors. Grade
outputs unlabeled in a fresh context that did not author the candidate whenever
an isolated worker is available, and report unblinded grading as a limitation.
When scored trials back an adoption or improvement claim, grade judgment-based
assertions twice; treat disagreement between the passes as an unreliable
assertion to tighten, not as a result. Alternate or randomize
condition order and preserve failed runs. Keep holdouts out of tuning; if a holdout exposes a defect, fix it
and use a new unexposed case for a fresh holdout claim.

Record the request, assertions, baseline/candidate versions, outputs, observed
failures, model, reasoning setting, and harness when available. Distinguish
confirmed settings from requested ones, simulations from live operations, and
unavailable metrics from measured values. Small samples support case-level
findings, not general reliability claims.

Accept changed behavior when required assertions pass without a material
regression on the covered tasks. A difference of one case between single runs
is within noise: repeat both conditions before attributing it to the candidate.
For a removal decision, every retained source
capability needs evidence or an explicit decision to drop it. Keep the baseline
when a proposed complication has no demonstrated benefit.

## Selection checks

Use explicit positives, uninvoked in-scope requests, and likely near misses.
Uninvoked requests are negatives for manual-only skills and positives when
implicit discovery is intended. Include ambiguous or unrelated cases when they
could reveal a real routing error.

Freeze expected selections before trials. Test in fresh contexts using each
target client's actual discovery path when possible; clients enforce invocation
policy through different controls, so evidence from one does not cover another. A prompted classification of a
description is a simulation, not proof that an installed client selects it.
Record missed positives and false activations by case type; repeat when outcomes
vary. If discovery is unchanged, verify consistency without claiming new measured
trigger evidence.

## Finish

Fix evidenced defects and rerun affected checks. Report what passed, failed, or
could not be tested. When trials were run, keep the reusable cases and the raw
per-case results beside the report, not only its narrative. Missing tools need not prevent finishing safe source edits, but unresolved
critical checks prevent claims of validated adoption or safe predecessor removal.
Complete publication or installation only within the already-authorized scope.
