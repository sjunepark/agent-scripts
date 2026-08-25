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

- Reviewed-plan binding contract: global apply now requires one exact reviewed
  JSON artifact and its approved SHA-256, recomputes the full global plan from
  a retained materialization session, and binds that session fingerprint at
  the mutation boundary.
- Fail-closed apply/recheck behavior and regression coverage: missing or forged
  approval, artifact substitution, moved expected content, post-binding session
  mutation, and locked warning/evidence drift all abort without placement.

### Current in-scope result

Operator guidance, validation, and delivery.

### Next in-scope action

Deliver the one implementation PR through feedback and merge, then record
terminal planning metadata.

### Evidence and blockers

- Candidate: implement the reviewed-plan binding contract and its regression coverage in one reviewable slice. Classification: included. Contract basis: the first two included results and repository-required validation. Action: proceed.
- `dev` is the repository-directed non-production integration branch. It was fast-forwarded locally to `1ad4e86`, and a dry-run confirmed direct creation/push permission before goal initialization. The implementation will use a separate work branch and target `dev`.
- No real-home mutation is authorized; validation must use isolated temporary homes or read-only commands.
- The independent code-review pass found three approval-boundary issues: an
  internally stale binding, warning/evidence-only locked drift, and Windows
  PowerShell BOM output. The implementation now fingerprints the complete live
  session at apply, compares global warnings/evidence during locked replans,
  and writes no-BOM UTF-8 artifacts on Windows.
- Focused `go test -count=1 ./internal/sjskills ./cmd/sjskills` and
  `git diff --check` pass after those review fixes. All apply tests use isolated
  temporary homes; no real-home global apply or restore has run.
- Full validation passes: `go test -count=1 ./...`, race-enabled command and
  internal tests, `go vet ./...`, registry/reconciler Node tests,
  `scripts/validate-skills`, `bunx skills add ./skills --list`,
  `bunx skills list`, and `git diff --check`.
- PR #15 received a Codex 👍 with no findings. CodeRabbit identified one valid
  Windows PowerShell 5.1 UTF-8 decoding gap in the operator recipe; every native
  plan capture now uses an explicitly UTF-8 `ProcessStartInfo` stream and a
  no-BOM artifact write. Follow-up validation and delivery are in progress.
