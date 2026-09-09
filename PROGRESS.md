# sjskills CLI version status delivery

- Feature implementation, tests, bounded independent review, and documentation
  are complete. Committed and pushed to main as `7e5e2ba`.
- `sjskills-v1.0.0` is published as an immutable GitHub release from that commit.
  Hosted source checks and native tests/installer checks passed on macOS Intel,
  macOS Apple silicon, and Windows x64.
- Publication used exact CI artifacts with verified uploaded hashes through the
  authenticated administrator session. The workflow token's administration-read
  limitation and draft-lookup follow-up are recorded in the
  [release delivery record](plans/sjskills-release.md).
- Installed binaries and managed skill state remain unchanged. No requested
  feature or publication work remains.

# Address issues skill

- Candidate: `skills/address-issues/`, with single-issue delivery and a confirmed,
  sequential Codex task queue using Astra/medium. Each issue ends at verified
  resolution after PR review and merge; blocked issues defer dependents while
  independent issues continue.
- Candidate authoring and bounded independent review are complete. Synthetic
  behavior trials passed 6/6 and fresh trigger classifications passed 8/8;
  evidence and limits are in `skills/address-issues/evals/evaluation.md`.
  App coordination is established by exposed tool schemas, not a live issue run.
- Registered in the dev profile in both canonical and CLI-embedded registries.
  Full skill validation, registry tests, and the full Go suite pass; bounded registry review found
  no remaining issues. Runtime instructions are unchanged from the evaluated
  candidate.
- Publication and configured project/global sync are authorized. Next: push the
  validated source, verify the remote skill tree, and reconcile both scopes
  using one temporary binary built from that published commit.
