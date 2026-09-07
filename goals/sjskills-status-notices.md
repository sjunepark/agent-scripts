# Goal: Automatic project and global skill-status notices

Status: complete
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
Automatic project and global status notices, complete-snapshot cache lifecycle,
CLI integration, approval compatibility, documentation, and acceptance coverage.
Delivered through [PR #19](https://github.com/sjunepark/agent-scripts/pull/19),
merged as `57dd2f1` on 2026-09-07 with individual commits preserved.

### Current in-scope result
None — all included results delivered.

### Next in-scope action
None — goal complete.

### Evidence and blockers
- Selected all actionable drift notices, 24-hour refresh, 15-minute failure cooldown, and one shared 30-second foreground refresh budget.
- Source/option batching completed both real remote cold scopes in 7.35 s and 6.29 s with no remaining staging. Warm checks measured 11.4–18.1 ms across 45 placements and 360 files.
- Local Go/race suites, vet, Node registry tests, skill validation, repeated process-cleanup tests, and release unit tests passed. Bounded initial and follow-up reviews and documentation harmonization completed.
- Final [native validation](https://github.com/sjunepark/agent-scripts/actions/runs/34108846117) passed Linux source/build checks, macOS Intel/ARM and Windows Go suites, and temporary installer/consumer preservation checks.
- Codex review completed without findings. CodeRabbit confirmed four fixes and withdrew its batching finding; all threads are resolved and validation is recorded on the PR.
- Native checks required a platform-specific golden test correction and a narrow Windows installer backup-path fix. Both are reviewed and validated; no release or real-machine installation occurred.
- No blockers remain. Release, installation, and real-machine rollout remain excluded and unstarted. Terminal bookkeeping is limited to goal and project-planning metadata on main.
