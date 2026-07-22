# Goal: Complete the Launch Path

Status: active
Planning scope: ROADMAP.md

## Original contract

Goal contract

- Outcome: Complete the launch path through validated renderer rollout.
- Goal state: `goals/launch-path.md`
- Sources (implementation detail, not scope authority): `ROADMAP.md` and
  `plans/renderer-rollout.md`.
- Included results:
  - Storage baseline.
  - Renderer rollout.
- Complete when: Both results are implemented and validated.
- Explicitly excluded: The assurance review and its remediation.
- Derived work: Absorb only bounded work necessary for an included completion
  condition; record useful outside work and ask before ambiguous expansion.
- Expansion authority: explicit user instruction only
- Resume invariant: invoke $progress in goal mode and recover Goal state before
  selecting work after every resume, automatic continuation, compaction, or
  handoff
- Delivery: No PR

## Authorized amendments

_None._

## Execution status

### Completed included results

- Storage baseline.

### Current in-scope result

Renderer rollout.

### Next in-scope action

Continue the renderer rollout from its recorded plan state.

### Evidence and blockers

Storage validation passed. Project planning has independently advanced its
`Current` pointer to the excluded assurance review.
