# Authenticated private-source delivery

PR #23 review follow-up: confirmed comments 3979633085 (commit pins), 3979633091 (unsupported GitHub page URLs), and 3979633101 (manual/workflow validation). Fixed with regression coverage; PR checks own follow-up native validation and the merge gate.

- PR #23 combines local skill-authoring work, origin main, and the PR #22
  recovery history. Public `kicpa` membership is restored; recovered private
  registrations now belong to `kicpa-private`.
- Access validation/resolution, controlled Git/gh staging, evidence separation,
  and compatible public-plan loading are implemented. The
  [private-source plan](plans/sjskills-private-github-sources.md) owns the design
  and acceptance evidence. Original recovery stash remains
  `038118178cce4759e3acebf854f19b36723488fc`.
- Full Go/race suites, vet, Node tests, release tests, and skill validation
  passed on macOS. Bounded independent review found no implementation defect;
  its requested native auth cancellation coverage and token-login acceptance
  were added and passed. Operator documentation and the plugin bundle agree.
- Implementation `84ba2e7` passed source checks and hosted native acceptance on
  macOS Intel/ARM and Windows in
  [run 34482396368](https://github.com/sjunepark/agent-scripts/actions/runs/34482396368).
  [PR #23](https://github.com/sjunepark/agent-scripts/pull/23) owns the final merge
  status and final check results. CodeRabbit skipped its
  requested review because automatic reviews are disabled.
- Published as immutable v1.3.0 after tagged source and all supported native
  release checks passed. The [release record](plans/sjskills-release.md) owns
  publication evidence. Installed binaries and managed skills were not changed.

# sjskills startup-check delivery

- Stages A–D are delivered through [PR #24](https://github.com/sjunepark/agent-scripts/pull/24), merged as `751d93f` to `codex/sjskills-startup-integration`. The [completed goal](goals/sjskills-startup-check.md) owns the closed delivery boundary.
- [Final promotion CI](https://github.com/sjunepark/agent-scripts/actions/runs/34937232876) passed on Windows x64 and both macOS targets. Local hook tests and bounded runtime review passed.
- The separately authorized promotion merged through [PR #25](https://github.com/sjunepark/agent-scripts/pull/25) as `15533e8` on 2026-09-15. The current Windows plugin installation and configured skill sync are complete. The [startup-check plan](plans/sjskills-startup-check.md) owns evidence and the remaining `/hooks` trust and fresh-session acceptance.
- Older installations may still inject the previous workflow. Existing plugin-data snapshots and update state remain intact.

## Previous CLI delivery

# sjskills modified-copy reconciliation

- Published in v1.2.0: project and global sync quarantine locally modified
  managed desired copies before verified replacement. Unknown ownership, source
  mismatch, untrusted provenance, unsafe paths, and changes after review remain
  conflicts. The [reconciliation contract](docs/skill-registry.md#ownership-and-reconciliation)
  owns the policy and restoration semantics.
- Go suite, isolated CLI flows for both scopes, rollback and subprocess crash
  recovery, registry checks, skill validation, and local skill discovery passed.
  Bounded independent code review found no actionable issues; affected docs are
  aligned. Hosted native reconciliation and installer tests passed on both
  macOS targets and Windows x64. Skill trials are recorded in
  `skills/sjskills/evals/evaluation-2026-09-10.json`.
- Committed and pushed as `4aad1dc`; the immutable release and local v1.2.0
  installation are verified in the [delivery record](plans/sjskills-release.md).
  Requested publication and CLI update are complete. Live skill reconciliation
  was not requested or performed.

# sjskills CLI version status delivery

- Feature implementation, tests, bounded independent review, and documentation
  are complete. The release bump was committed and pushed to main as
  `79baae2`.
- `sjskills-v1.1.0` is published as an immutable GitHub release from that
  commit.
  Hosted source checks and native tests/installer checks passed on macOS Intel,
  macOS Apple silicon, and Windows x64.
- Publication used exact CI artifacts with verified uploaded hashes through the
  authenticated administrator session. The workflow token's administration-read
  limitation and draft-lookup follow-up are recorded in the
  [release delivery record](plans/sjskills-release.md).
- That release did not change installed binaries or managed skill state.
  No v1.1.0 feature or publication work remains; the later CLI update is recorded
  above.

# Address issues skill

- Published in `fa8ed7c`: single-issue delivery and confirmed sequential
  Astra/medium task queues, registered in the dev profile in both registries.
- Skill validation, registry tests, the full Go suite, and bounded independent
  reviews passed. Synthetic behavior trials passed 6/6 and trigger cases 8/8;
  evidence and live-execution limits are in
  `skills/address-issues/evals/evaluation.md`.
- On 2026-09-10, synced this project's dev+go selection from the published source:
  both address-issues copies installed and verified; all 33 project placements
  and all 16 fixed-global placements are exact. No quarantines or blockers.
- That sync used one verified temporary binary built before the v1.1.0 release
  and did not replace the installed CLI. The later v1.2.0 installation above
  includes the updated registry.
- Publication and requested skill reconciliation are complete.
