# Goal: Automatic project and global skill-status notices

Status: active
Planning scope: ROADMAP.md

## Original contract

Goal contract

- Outcome: Implement automatic project and global skill-status notices in sjskills, with reliable drift reporting and bounded overhead.
- Goal state: goals/sjskills-status-notices.md
- Included results and sources (semantic results define scope; paths supply detail):
  - Complete status-notice feature, including cache lifecycle, CLI integration, approval compatibility, documentation, and acceptance coverage — /Users/sejunpark/IT/agent-scripts/plans/sjskills-status-notices.md.
- Complete when: Every included result achieves its cited outcome and applicable completion criteria within its named semantic boundary; repository-required validation and review pass; performance evidence supports the selected refresh behavior; planning is truthful; Delivery finishes.
- Excluded: Release, installation, and real-machine rollout of the feature.
- Authority: Execute only included results and necessary supporting work; resolve remaining decisions within that closed outcome using best judgment; record anything else and ask before scope expansion or external actions not covered by this contract and Delivery. Delegation explicitly includes notice coverage, refresh frequency, and acceptable cold-check delay; settle and document these without another approval request.
- Resume: Initialize this contract with $progress goal mode before work; recover it before every resume, continuation, compaction, or handoff; stop if recovery fails.
- Delivery: PR delivery — use $progress's PR lifecycle and the fewest sequential reviewable PRs; finish each through $create-pr and $address-pr-feedback before starting the next, including the final implementation slice.

## Authorized amendments

_None._

## Execution status

### Completed included results
_None._

### Current in-scope result
Complete status-notice feature in plans/sjskills-status-notices.md.

### Next in-scope action
Publish the single feature PR, finish native checks, and complete feedback and merge.

### Evidence and blockers
- Candidate: feature implementation and PR delivery. Classification: included. Contract basis: complete status-notice feature and Delivery. Action: proceed.
- Delivery base: main; repository admin access, no branch protection or rulesets. Initialization and terminal metadata may be committed directly.
- One PR covers the dependent implementation checkpoints. No release, installation, or real-machine reconciliation.
- Adopt all actionable drift notices, 24-hour refresh, 15-minute failure cooldown, and shared 30-second foreground refresh budget, supported by measured warm and cold performance.
- Candidate: bounded code review, documentation harmonization, and acceptance checks. Classification: necessary. Contract basis: repository-required review and complete documented feature. Action: proceed.
- Implementation and documentation are present; one bounded review found a pruning race, fixed with regression coverage. Native delivery checks remain pending.
- Source/option batching was selected after per-skill cold fetching exceeded the budget. Actual isolated cold checks completed both scopes in 7.35 s and 6.29 s with no remaining staging; warm 45-placement/360-file checks measured 11.4–18.1 ms.
- Local Go and race suites passed; repeated process tests exposed Darwin EPERM during killed-group disappearance. The fix verifies remaining live processes after signal errors; 20 repeated runs passed. Final Go race suite, vet, Node registry tests, skill validation, and Windows cross-compilation passed.
