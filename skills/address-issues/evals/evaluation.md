# Address issues candidate evaluation — 2026-09-09

## Contract and evidence

Reusable outcome: deliver a selected GitHub issue through verified resolution,
or coordinate the same delivery over a confirmed queue of separate app tasks.
Distinctive help: exact task dispatch settings, confirmation and sequencing,
pending-identity recovery, evidence-based completion, and dependency deferral.
Expected reuse: future issue implementation requests in GitHub repositories.

The user selected merged PR plus verified resolution as the completion boundary,
and continuing independent issues when another is blocked. Intake used the
actual open issues cpaikr/ytm#45, #42, and #27 as examples, read through `gh`.
No issue in that repository was implemented or changed during skill authoring.

The current app tool schemas establish `create_thread`, `list_projects`,
`list_threads`, `wait_threads`, `read_thread`, and `send_message_to_thread`,
including `gpt-6-astra` and `thinking: medium`. OpenAI's
[voice coordination documentation](https://learn.chatgpt.com/docs/features/voice#delegate-and-coordinate-work)
describes related delegation and result collection. Public searches did not
document the exact app tool names. This capability claim is based on exposed
tool schemas, not an assertion that Astra introduced cross-task communication.

## Frozen experiment

Runtime candidate SHA-256 values, frozen before scored trials:

| File | SHA-256 |
| --- | --- |
| SKILL.md | 3cc47cfd2afb4bc14b413ed8f28d9a15c230db4758842ab06c8370970361dd43 |
| workflows/multiple.md | 47beb9a6affe1e0df3e0e9ca19c6211e1f1ec39976d47ec2d4add93ca0d37d48 |
| agents/openai.yaml | ddb08b0d75c868970027c4a83166aed5d0f8625cd2cc3b5d9e8516cc1f063edd |

`cases.json` freezes the requests, state, critical assertions, and stopping
criteria. The new-capability baseline receives the same request and fixtures
without the skill. Both conditions use fresh, no-history workers and return
proposed actions; neither may execute external operations. Condition order
alternates. The pending-merge case repeats once because premature advancement
is consequential. A separate worker prepared `holdout.json`; its contents were
withheld from the author until the frozen candidate's final trials completed.

All critical candidate assertions must pass, with no material holdout regression.
Reports must distinguish proposed operations from execution and missing proof
from success. Clarity is acceptable when the outcome, next action, and reason
are explicit; all candidate outputs met that standard. No taste-dependent
human-quality claim or statistical reliability threshold is made.

## Behavior results

Raw outputs and observed timestamps are in `results.json`.

| Case | Baseline | Candidate | Observed result |
| --- | --- | --- | --- |
| Independent consumer proof | Pass | Pass | Both preserve unrelated work and block merge/resolution without the required Windows proof. |
| Unconfirmed queue | Pass | Pass | Both request explicit membership/order confirmation and create no task. |
| Turn ended before merge, trial 1 | Pass | Pass | Both return to the existing task and withhold the next issue. |
| Turn ended before merge, trial 2 | Pass | Pass | Both again wait for actual delivery evidence. |
| Blocked issue with dependent and independent successors | Fail | Pass | Baseline dispatches the independent issue but omits fetching/starting from current main. Candidate records the blocker/dependent and supplies current-base instructions and a full delivery contract. |
| Pending creation recovery holdout | Pass | Pass | Both reconcile explicit pending/final identity, verify scope, monitor the existing task, and keep the next issue queued. |

Candidate: 6/6 trials passed every critical assertion. Baseline: 5/6. No runtime
candidate edits were made after freezing or in response to the holdout. Most
baseline decisions were already sound; measured improvement is limited to the
explicit fresh-base handoff in this small sample.

The evaluation harness initially retained completed workers and reached its
worker limit before the next trial could start. Completed outputs were retained,
workers were closed, and only the unstarted trials ran afterward. This was a
harness resource failure, not an issue-task or candidate failure. Model token
usage was unavailable; timestamps measure observed trial duration, including
coordination overhead, rather than pure inference latency.

## Invocation boundary

`trigger-cases.json` contains the frozen labels; `trigger-results.json` retains
all responses. Each prompt was classified in a separate fresh context using
only the skill description and adapter metadata. All 3 explicit positives and
5 negatives matched their labels. Negatives cover uninvoked issue work,
PR-feedback-only work, triage, authoring, and a quoted non-executing example.
This is an offline selection test; actual client discovery was not tested by
installing the skill.

## Review and structural validation

An independent bounded `code-review` pass applied implementation, system,
design, and diet lenses plus the authoring and portability rubrics. It found
no actionable blockers or majors. No source edits were made by the reviewer.

| Rubric area | Disposition and evidence |
| --- | --- |
| Scope and purpose | Pass: one issue-delivery lifecycle with two routes; SKILL.md excludes triage, authoring, and PR-feedback-only execution. |
| Trigger description | Pass: explicit invocation in description and false adapter policy; 8 fresh classifications above. |
| Entry point | Pass: inputs, requirements, authority, delivery, and dispositions are directly available in SKILL.md. |
| Resource pointers | Pass: the multiple-mode workflow has a direct conditional relative link; agents/evals contain metadata/evidence only. |
| Specificity | Pass: exact app model/effort and identity constraints; implementation and technical decisions remain adaptable. |
| Workflow and completion | Pass: confirmation reuses prior authority, verified merge/acceptance/closure gates, and blocked-child handoff exercised above. |
| Robustness | Pass: pending creation holdout, stale completion trials, data preservation and missing consumer proof case. |
| Portability | Pass within declared scope: GitHub CLI required; multi-issue mode explicitly requires Codex app/Astra. No hidden host paths, external symlinks, copied code, or optional runtime scripts. Licensing/script subchecks are N/A because neither is bundled. |
| Evaluation evidence | Pass for synthetic decisions: paired fresh contexts, frozen criteria, holdout, repeat, retained raw outputs, separate trigger tests. Live execution remains unmeasured. |
| Pruning and quality | Pass: one shared lifecycle, one substantial mode resource, no runtime script or duplicated lookup manual. |

Checks:

- `bunx skills add ./skills/address-issues --list`: discovered exactly the skill.
- Standalone skill-creator validator: passed with ephemeral PyYAML through `uv`.
  Its older schema rejects the portable compatibility field, so prerequisites
  were moved to the body before freezing; they remain explicit.
- Repository skill/resource checks report no errors in the candidate itself.
  Full `scripts/validate-skills` passes after registration in the dev profile
  and synchronization of the CLI-embedded registry.
- JSON parsing, runtime hash comparison, and `git diff --check`: passed.

## Delivery status and limits

Candidate authoring, bounded review, and synthetic evaluations are complete.
The README routes to the skill. On 2026-09-10 the user authorized dev-profile
registration, commit/push, and configured skill sync. The canonical and embedded
registries now select the skill for the dev profile under their default targets.
Publication and reconciliation status are tracked in the repository PROGRESS.md.

Live app task creation, cross-task messages, GitHub merges, review services,
and consumer verification were not exercised. The trials establish observed
decision behavior, not end-to-end delivery reliability or host availability.
