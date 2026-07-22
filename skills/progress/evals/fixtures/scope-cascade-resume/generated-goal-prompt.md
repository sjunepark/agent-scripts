# Generated `/goal` Prompt

Complete this closed goal without expanding it.

Goal contract

- Outcome: Complete the canonical publication foundation through validated
  viewer scaling.
- Goal state: `goals/publication-foundation.md`
- Sources (implementation detail, not scope authority): `ROADMAP.md` and
  `plans/publication-foundation.md`.
- Included results:
  - Canonical artifacts.
  - Evaluation data plane.
  - Local preview.
  - Publication viewer scaling.
- Complete when: All four results are implemented, validated, reviewed, and
  merged to the integration branch with the project roadmap updated accurately.
- Explicitly excluded: The integrated whole-codebase review and official
  performance/Windows evidence.
- Derived work: Absorb only bounded work necessary for an included completion
  condition; record useful outside work and ask before ambiguous expansion.
- Expansion authority: explicit user instruction only
- Resume invariant: invoke $progress in goal mode and recover Goal state before
  selecting work after every resume, automatic continuation, compaction, or
  handoff
- Delivery: PR delivery

Preflight a non-production integration branch for direct goal-metadata commits
and pushes, using a temporary integration branch if needed. Invoke `$progress`
in goal mode there before implementation to initialize, commit, and push the
named goal-state file from this exact contract, and do so again at every resume
before selecting work. Keep the project roadmap truthful, but never use its
queue, `Current`, a new plan, review findings, or delivery state to expand the
goal. Check every new plan, branch, PR, review program, and independently
reviewable outcome against the contract. After the final included PR merges,
use `$progress` to complete the goal, then commit and push only terminal goal and
project-planning metadata on the integration branch. Return before inspecting
or starting the integrated review. If the exact contract cannot be recovered or
the terminal metadata cannot be persisted, fail closed and request direction.
