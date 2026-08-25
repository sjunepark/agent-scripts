# Goal: Bind global apply to reviewed expected-content evidence

Status: active
Planning scope: ROADMAP.md

## Original contract

Goal contract

- Outcome: Make the `sjskills` global rollout approval-safe by binding apply to the exact reviewed expected-content evidence while leaving real homes untouched.
- Goal state: goals/sjskills-global-rollout-approval-binding.md
- Included results and sources (semantic results define scope; paths supply detail):
  - Reviewed-plan binding contract — plans/sjskills-global-rollout.md; cmd/sjskills/main.go; internal/sjskills
  - Fail-closed apply/recheck behavior and regression coverage — cmd/sjskills/*\_test.go; internal/sjskills/*\_test.go
  - Operator guidance, validation, and delivery — README.md; docs/skill-registry.md; plans/sjskills-global-rollout.md; AGENTS.md
- Complete when: Every included result achieves its cited outcome and applicable completion criteria within its named semantic boundary; repository-required validation and review pass; planning is truthful; no real-home mutation occurs; Delivery finishes.
- Excluded: Per-machine approval and real-home `apply --global`, restore, migration cleanup, or rollout execution.
- Authority: Execute only included results and necessary supporting work; resolve remaining decisions within that closed outcome using best judgment; record anything else and ask before scope expansion or external authority.
- Resume: Initialize this contract with $progress goal mode before work; recover it before every resume, continuation, compaction, or handoff; stop if recovery fails.
- Delivery: PR delivery — use $progress's PR lifecycle and the fewest sequential reviewable PRs; finish each through $create-pr and $address-pr-feedback before starting the next, including the final implementation slice.

## Authorized amendments

_None._

## Execution status

### Completed included results

_None._

### Current in-scope result

Reviewed-plan binding contract.

### Next in-scope action

Inspect the existing global plan and apply boundaries, then define the smallest exact-content evidence contract that apply can verify fail-closed.

### Evidence and blockers

- Candidate: implement the reviewed-plan binding contract and its regression coverage in one reviewable slice. Classification: included. Contract basis: the first two included results and repository-required validation. Action: proceed.
- `dev` is the repository-directed non-production integration branch. It was fast-forwarded locally to `1ad4e86`, and a dry-run confirmed direct creation/push permission before goal initialization. The implementation will use a separate work branch and target `dev`.
- No real-home mutation is authorized; validation must use isolated temporary homes or read-only commands.
