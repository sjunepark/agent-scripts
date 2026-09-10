# Cleanup branches evaluation

This is development evidence, not runtime guidance. Final v3 passes the bounded
behavior regression and applicable static/trigger gates. No real project branches
or GitHub state were mutated by these trials.

## Design and frozen inputs

`assertions.json`, `packet.txt` and `candidate.md` freeze the original candidate
and a seven-case synthetic task. `assertions-v2.json`, `packet-v2.txt` and
`candidate-v2.md` freeze the revised candidate and eight-case task before its
scored trials. SHA-256 values are in the assertion files. Runtime v2 replaces
local porcelain deletion with expected-OID deletion after review reproduced a
race. V3 adds one observed-verification reporting rule after the v2 critical
failure below. `candidate-v3.md` and `assertions-v3.json` freeze the final runtime
(SHA-256 `b4025735bb6b9cef3a0e430ff9b180902fee2a68de3a5dd8d74ea1e102e5aeed`).
The runtime author made both revisions; the evaluator only authored fixtures
and scored observations.

The packet derives ordinary integration, squash merges, worktrees, long-lived
branches, backups and closed-unmerged PRs from the author's prior Darty findings.
Diverged local/remote tips, failed protection evidence, lease rejection and local
ref movement are synthetic extensions. Case H was independently authored by the
evaluator and withheld from the runtime author until scoring: a same-name/head
merged PR from another repository must not establish integration. The author did
not tune against H.

Fresh isolated workers read only the task packet, plus the frozen skill for the
candidate condition. They received no evaluator assertions, prior outputs or
feedback. These are **decision/action transcript simulations**, not actual agent
execution against GitHub. Each suite is compact and every case shares one
context within a trial. Baseline/candidate order alternated; v1 has one pair
(retained as superseded), v2 has two pairs. Raw JSON is in `raw-b1.json` through
`raw-c3.json`. V3 receives exactly two additional fresh candidate trials on
unchanged `packet-v2.txt`, preserving its original assertions and reusing b2/b3
as unchanged baseline evidence; `raw-c4.json` and `raw-c5.json` preserve these
outputs. Only the affected reporting behavior is changed; trigger policy is
unchanged and trigger trials are not repeated. Archival follow-ups saved the original answers without changing
trial decisions. Agent versions and wall-clock timing were not captured; no
latency or operation-cost comparison is claimed.

## Results

`grading.json` records exact deletion-set comparisons, manual retention grading,
raw hashes and output sizes. Remote names are normalized by removing `origin/`
only for set comparison. Compact retained lists count both scopes only where
the accompanying action notes explicitly say so.

| Condition | Trials | Decision cases matching expected deletion/retention | Reporting failures |
| --- | --- | --- | --- |
| v1 baseline | 1 | 7 of 7 | 0 |
| v1 candidate (superseded) | 1 | 7 of 7 | 0 |
| v2 baseline | 2 | 16 of 16 | 0 |
| v2 candidate | 2 | 16 of 16 | 1 |
| v3 candidate | 2 | 16 of 16 | 0 |

All trials preserved the cross-repository holdout. Both v2 candidate trials
selected expected-OID local deletion and guarded remote deletion; neither
bypassed a rejected guard. No blanket approval requirement or user-state
manipulation was introduced. Observed decision quality was preserved, with no
measured decision improvement over baseline. This small, highly specified
packet does not establish a general skill benefit or a reliability rate.

**Critical reporting failure:** `raw-c3.json` claims absence was verified despite
no follow-up verification results in the packet. This fails the frozen
no-false-success assertion and prevents declaring the original full behavior
gate passed. `raw-c2.json` correctly states that verification results were not
supplied. The packet invited simulated intended/achieved outcomes and supplied
successful deletion outcomes, making the reporting boundary less explicit than
intended. That limitation does not erase c3's score. The targeted follow-up in
`verification-gap.txt` explicitly marks both verification results unknown; its
criteria were frozen separately in `verification-gap-assertions.json` without
changing the candidate or original scores. The fresh follow-up passed all three
assertions: it reports command success, explicitly unknown local/server absence,
and both verification failures. Its exact response is retained in
`raw-verification-gap.md`. This supports an ambiguity-sensitive reporting
limitation. The runtime author then made the reporting boundary explicit and
reran the original, unchanged packet rather than relying on that diagnostic.
Both v3 trials correctly distinguish supplied command outcomes from missing
follow-up verification, including the unspecified local squash deletion outcome.
Every original critical assertion and all sixteen decision checks pass in v3.
The v2 failure remains recorded; it is not overwritten by the new result.

## Trigger selection

`triggers.json` freezes three explicit positives and three negatives (including
an uninvoked request directly in scope). Each was presented to a separate fresh
worker with only the catalog description and invocation policy.
`raw-triggers.json` preserves all six exact answers: three positives activated,
three negatives rejected. This tests simulated catalog selection, not the Codex
client dispatcher. One observation per prompt provides no stability estimate.

## Executed Git evidence and repository checks

The author separately ran `git-smoke.py` against a disposable working repository
and local bare remote with Git 2.50.1, reporting eight passing checks: expected-tip
local/config cleanup; leased remote deletion with server absence verification;
local and remote moved-tip rejection; reproduction of the misleading successful
`git branch -d` against a feature upstream; squash-head deletion; worktree
occupancy detection/porcelain refusal; and branch-only pruning preserving a local
tag, an untracked working file and the current branch. The evaluator preserved
the script without rerunning it. It does not execute GitHub protection APIs or
simulate concurrent worktree checkout during plumbing deletion.

The author reports passing repository skill/link validation, bundled quick
validation using temporary PyYAML dependencies, single-skill CLI discovery,
12 Node registry/wrapper tests, and `go test ./...`. The catalog-size expectation
was updated for the new skill. These are packaging/command checks, separate from
agent decision evidence.

## Authoring and portability dispositions

- Scope, activation, entry point and invocation metadata: Pass. One explicitly
  invoked cleanup capability; inspection limits follow user intent; runtime
  defaults include local and remote cleanup without variants.
- Fragile operations and retention: Pass on reviewed v3 and observed decisions.
  Exact tips, repository identity, durable integration, worktree occupancy,
  protection uncertainty and expected-tip guards have explicit owners.
- Permission and completion workflow: Pass statically. Routine cleanup uses
  existing authority; unresolved evidence limits only affected actions. Reporting
  behavior passes in both fresh v3 repetitions after the documented v2 failure.
- Resource routing and portability: Pass. Runtime consists of the entry file,
  relative client metadata, and declared Git/host-API capabilities; no runtime
  dependency on evaluator files, external absolute paths or a particular model.
- Behavior evidence: Pass for final v3 within this bounded experiment. Both
  repetitions satisfy unchanged critical/objective assertions and acceptable
  clarity, preserve baseline decisions and the holdout, and fix the observed
  reporting failure. V2 remains failed; its result is retained separately.
  Baseline reuse, simulation limits and lack of decision uplift remain explicit.
- Implicit trigger behavior: Pass in six isolated selection simulations, with
  the client-dispatch and repetition limits stated above.
- Pruning/minimum package: Pass. No runtime helper or alternative invocation
  modes were added; evaluation fixtures remain outside the execution path.

This is manual evaluator review against frozen anchors, not blind human review.
No publication, installation, real cleanup or reliability claim follows from it.
