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

Set assertions and stopping criteria before scoring. Expand after failures,
candidate changes, unstable outcomes, or an unresolved risk; passing checks do
not need repetition for its own sake.

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

For subjective judgments, use concrete acceptable/unacceptable anchors and blind
review where it matters. Alternate or randomize condition order and preserve
failed runs. Keep holdouts out of tuning; if a holdout exposes a defect, fix it
and use a new unexposed case for a fresh holdout claim.

Record the request, assertions, baseline/candidate versions, outputs, observed
failures, model, reasoning setting, and harness when available. Distinguish
confirmed settings from requested ones, simulations from live operations, and
unavailable metrics from measured values. Small samples support case-level
findings, not general reliability claims.

Accept changed behavior when required assertions pass without a material
regression on the covered tasks. For a removal decision, every retained source
capability needs evidence or an explicit decision to drop it. Keep the baseline
when a proposed complication has no demonstrated benefit.

## Selection checks

Use explicit positives, uninvoked in-scope requests, and likely near misses.
Uninvoked requests are negatives for manual-only skills and positives when
implicit discovery is intended. Include ambiguous or unrelated cases when they
could reveal a real routing error.

Freeze expected selections before trials. Test in fresh contexts using the
client's actual discovery path when possible. A prompted classification of a
description is a simulation, not proof that an installed client selects it.
Record missed positives and false activations by case type; repeat when outcomes
vary. If discovery is unchanged, verify consistency without claiming new measured
trigger evidence.

## Finish

Fix evidenced defects and rerun affected checks. Report what passed, failed, or
could not be tested, with concise reusable cases and raw results when trials were
run. Missing tools need not prevent finishing safe source edits, but unresolved
critical checks prevent claims of validated adoption or safe predecessor removal.
Complete publication or installation only within the already-authorized scope.
