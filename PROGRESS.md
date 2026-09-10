# Recovered private KICPA registry draft

- Preserved seven unpublished registrations on `codex/integrate-private-kicpa-skills`
  and combined them with current main (50 registry entries, including Darty).
- Not ready to merge or publish: the draft adds private access requirements to
  `kicpa`, contrary to the selected separate private-profile design. Follow the
  [private-source plan](plans/sjskills-private-github-sources.md) before rollout.
- Targeted Go contract tests and 12 Node registry/wrapper tests pass. Bounded
  review and documentation alignment completed; private fetching remains unverified.
- Full original checkout recovery remains in Git stash commit
  `038118178cce4759e3acebf854f19b36723488fc`, including superseded hook drafts.

# sjskills maintenance plugin delivery

- User authorized automatic CLI updates and configured skill sync on every
  supported local host. Existing release support is macOS Intel/ARM and Windows
  x64; Linux is unsupported.
- User also authorized automatic adoption of this hook's published `main`
  updates. Migrated the standalone hook into `plugins/sjskills-maintenance/`;
  retired its installer. Maintenance still runs in the next agent turn.
- Added daily plugin-check state, a 15-minute failure cooldown, remote source/ref
  verification instructions, and target-only reinstall with installed-version
  verification. Canonical sjskills workflow references are bundled and checked.
- Plugin schema validation, isolated Codex CLI install/reinstall, and installed
  Windows hook invocation succeeded. All 21 targeted tests pass. Bounded
  independent review and scoped documentation alignment are complete.
- Native validation confirmed reinstall removes the old plugin cache. The hook
  now preserves content-addressed workflow/helper snapshots in plugin data;
  regression coverage verifies they survive removal and reject local edits.
  Explicit Windows command dispatch passes cmd.exe and PowerShell. Test fixture
  setup uses ordinary reads/writes after reproducing a Node 22.17.1 recursive-copy
  crash on Unicode Windows paths.
- Published implementation and catalog to `main` in `e0c60b4`. Registered the
  remote-backed personal marketplace on Windows and installed/enabled plugin
  version `0.1.0+codex.20260910121700`. Installed content matches publication;
  the installed Windows hook emits its maintenance context successfully.
- Hook trust still requires review through `/hooks`. Other hosts require their
  own installation. Full session-driven maintenance and native macOS execution
  remain unverified.

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
