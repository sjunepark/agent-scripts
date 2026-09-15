# Goal: Check sjskills automatically without routine agent work

Status: active
Planning scope: ROADMAP.md

## Original contract

Goal contract
- Outcome: Deliver a tested check-only sjskills startup hook that runs automatically and reports actionable findings without routine agent work.
- Goal state: goals/sjskills-startup-check.md
- Included results and sources (semantic results define scope; paths supply detail):
  - Direct checks and concise reporting — plans/sjskills-startup-check.md, stages A–B.
  - Reliability and native acceptance — plans/sjskills-startup-check.md, stage C and acceptance matrix.
  - Packaged replacement and aligned documentation — plans/sjskills-startup-check.md, stage D.
- Complete when: Every included result achieves its cited outcome and applicable completion criteria within its named semantic boundary; repository-required validation and review pass; planning is truthful; Delivery finishes.
- Excluded: Real-machine installation/activation in stage E; automatic CLI/plugin updates or skill synchronization; routine agent maintenance on startup.
- Authority: Execute only included results and necessary supporting work; record anything else and ask before scope expansion or external actions not covered by this contract and Delivery.
- Resume: Initialize this contract with $progress goal mode before work; recover it before every resume, continuation, compaction, or handoff; stop if recovery fails.
- Delivery: PR delivery — use $progress's PR lifecycle and the fewest sequential reviewable PRs; finish each through $create-pr and $address-pr-feedback before starting the next, including the final implementation slice.

## Authorized amendments

_None._

## Execution status

### Completed included results

_None._

### Current in-scope result

Stages A–B: direct checks and concise reporting.

### Next in-scope action

Reproduce the registered hook in an isolated installation and pin the check-only behavior with failing tests.

### Evidence and blockers

- Clean starting checkout equals prepared commit `ce8b5047ce489f02ff5c1dfe79f0fc1926676205`; ancestor verification passed. All cited sources and required workflow skills are present.
- Native platform goal created and active in task `01a0a315-a2c5-7bb3-b740-5211ab1babb0`.
- PR integration branch: `codex/sjskills-startup-integration`, preserving the prepared commit. Repository permission is ADMIN; branch creation/push succeeded. Initialization and terminal metadata use this non-production integration branch.
- Candidate: one connected implementation PR for stages A–D. Classification: included. Contract basis: the three named results. Action: proceed after initialization is committed and pushed.
- Real-host installation, activation, plugin/CLI updates, and configured reconciliation remain excluded. Use isolated test homes only.
