# Prepare/spawn evaluation results

The candidate passes the bounded behavior evaluation: 6 of 6 paired trials, compared with 3 of 6 for the baseline, plus 3 of 3 focused argument/availability checks. These are observed counts from a small synthetic sample, not a general reliability estimate.

## Behavior evidence

| Case | Baseline | Candidate | Observed difference |
|---|---|---|---|
| P0: prompt-only | Pass | Pass | One closed, copy-ready prompt; no mutation or launch |
| P1: prepare with unrelated staging | Fail | Pass | Candidate writes the detailed plan, records one bounded review, commits only preparation, and returns its prompt; baseline leaves the sparse plan untouched |
| P2: Luna spawn, first trial | Fail | Pass | Baseline omits max reasoning; candidate passes exact model/reasoning, prepared branch, native-goal startup instructions and checks |
| P2: fresh repeat | Fail | Pass | Baseline again omits max and also omits prepared startingState; candidate preserves both |
| P3: Astra/high startup failure | Pass | Pass | Both create once and report missing child native-goal capability without claiming an active goal |
| H1: pending setup holdout | Pass | Pass | Both omit unrequested settings, retain clientThreadId, avoid ready-ID-only calls, and report pending without duplicate creation |

The candidate has no frozen critical-assertion failures. The baseline repeat P2 has a critical prepared-state failure; its other failures concern missing preparation and the Luna preset. Candidate focused checks reject unknown `nova` and unsupported `luna ultra` before mutation; absent task/preparation tools produce a stated limitation with no invented success.

P1 uses real disposable Git repositories. Candidate commit `4d0fda11d54a45a45d6a392939b1c7bdc2b9fb6e` changes only `plans/import.md`. The unrelated `notes/personal.md` remains the sole staged change with its original content and blob `6326793`; baseline and candidate cached diffs match. The evaluator and parent independently inspected the actual commit, tracked paths, plan contents, working-tree diff, and staged note. Planning/review skill invocations use supplied synthetic contracts; their full installed workflows were not exercised.

## Trial integrity

The baseline runtime was copied from the recorded pre-change HEAD; candidate runtime hashes were frozen before its scored trials, with the progress goal-contract companion hashed for integration provenance. Every behavior trial ran in a fresh isolated worker. P2 was repeated once per condition as predeclared. Most first pairs dispatched baseline before candidate while authoring completed; the repaired P2 first pair dispatched candidate before baseline and its repeat reversed that order.

Initial P2 inputs had a dirty flag without diff evidence and omitted the child `update_goal` capability. The candidate correctly declined an unverifiable handoff. These three initial outputs are preserved as unscored diagnostics. The evaluator repaired the environment to a verified, clean, already-reviewed preparation commit on a non-main branch and supplied the complete child capability catalog, then reran both conditions with unchanged assertions. H1 received the same environment-only repair before its first run. No runtime instructions changed in response.

The holdout request, pending responses, and assertions stayed hidden from the author until both results were graded. Its case was not used for tuning.

## Trigger gate

Baseline and candidate each match all 6 frozen labels: three explicit positives activate; a matching but uninvoked preparation request, a quoted invocation explanation, and unrelated process-spawn debugging do not activate. Twelve distinct fresh workers received only their request and the corresponding description/adapter policy, with labels withheld and condition order alternated. No false activation or missed explicit invocation was observed. Each prompt was tested once per condition; these classifications do not test the production skill loader. Exact responses and worker identities are preserved in the result artifact.

## Limits and reproducibility

Task APIs and native-goal results are canned responses interpreted through recorded action traces; no actual task, native goal, remote ref transfer, publish, install, or implementation was performed. The traces verify the agent's decisions and exact creation arguments, not live Codex integration, actual destination settings, or child execution reliability. Canned startup responses do not establish completion of the later goal or PR lifecycle.

P1 is the only live repository mutation trial. Fixtures contain plans rather than an implementation codebase, so freshness and API compatibility checks are bounded by supplied evidence. Non-Git targets, remote hosts, mixed changes within one planning file, real asynchronous identity resolution, and model availability drift were not exercised. The frozen scope did not expand after passing results.

Byte sizes and recorded task creation counts are retained. Comparable token, wall-time, pricing, and latency measures were unavailable; no efficiency improvement is claimed. Raw outputs and task API records remain in `prepare-spawn-results.json`; fixtures, assertions, and protocol are in `prepare-spawn-inputs.json`, with cases discoverable through `evals.json` ids 20–33. Temporary path prefixes in evidence are replaced with `<trial-root>`; the required evidence is preserved in the package rather than depending on temporary files.

Static package validation and local CLI discovery passed before candidate trials; the parent's bounded code review reported no blocker or major issue. After adding evaluation evidence, `scripts/validate-skills` again passed for 31 skills and the registry. JSON parsing, preservation of original eval ids 1–19, contiguous new ids 20–33, all case-file references, and equality with every frozen runtime hash also passed.
