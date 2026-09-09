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
- Installed binaries and managed skill state remain unchanged; the published
  binary is available for future explicit installation. No requested feature or
  publication work remains.

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
- Sync used one verified temporary binary built before the v1.1.0 release. The
  installed CLI was not replaced and still embeds the pre-1.1.0 registry;
  future syncs need an explicit install of the updated release to retain the
  skill.
- Publication and requested skill reconciliation are complete.
