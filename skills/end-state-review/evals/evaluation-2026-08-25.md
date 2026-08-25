# Scope and retention evaluation — 2026-08-25

## Decision

Retain `end-state-review` as a distinct, explicit-invocation skill and move it
from the fixed global baseline to the `dev` project profile. Do not merge it
into `code-review` and do not remove it on the current evidence.

This is a static and test-definition result, not a completed behavioral trial.
The evaluation cases are frozen in `evals.json` and `trigger-cases.json`, but
fresh isolated worker runs were unavailable under the execution boundary for
this review. No behavior pass rate or trigger accuracy rate is claimed.

## Skill contract

- **Reusable outcome:** reconstruct a plan around its intended final state
  after requirements, prototypes, migrations, or rollout history have added
  residue.
- **Distinctive help:** classify residue separately from live compatibility,
  persistence, migration, security, and rollout obligations while enforcing a
  planning-only mutation boundary.
- **Expected reuse:** pre-implementation plan reviews after requirement churn,
  prototype graduation, migration design, and staged-rollout cleanup.

## Boundary against code-review

The two skills share system-design and simplification concerns, but their
execution contracts differ materially:

| Concern | `end-state-review` | `code-review` |
| --- | --- | --- |
| Primary target | Plan, roadmap item, specification, or migration proposal | Completed code change or implementation governed by planning intent |
| Default mutation | Review-only | Apply obvious safe implementation fixes unless the edit policy forbids it |
| Permitted approved edits | Planning artifacts only | In-scope implementation and tests for Bucket I fixes |
| Required stopping point | Coherent proposal, plan amendments, validation map, and unresolved decisions | Findings, safe fixes, decision items, and executed or blocked validation |
| Governing-plan behavior | Review and optionally revise the plan itself | Resolve and inspect the implementation governed by the plan; do not reduce the target to the document alone |

Merging these workflows would require a planning-only branch that reverses
`code-review`'s implementation target and safe-fix defaults. That would make
the common review path broader and less predictable without removing the need
for two distinct authority policies.

## Frozen behavior assertions

The five cases in `evals.json` cover removable historical residue, live
migration obligations, an unresolved external contract, a completed-code near
miss, and an explicitly invoked request that exceeds the skill's authority.

Critical assertions:

- state the intended end-state contract without silently changing product
  behavior;
- distinguish current obligations from historical residue with evidence;
- retain temporary obligations with explicit exit conditions;
- identify unresolved load-bearing decisions instead of guessing;
- map retained obligations and deleted assumptions to validation; and
- never modify implementation in this workflow.

The behavior gate passes only if fresh candidate runs satisfy every critical
assertion and materially outperform or complement `code-review` on the three
planning cases without being selected for the completed-code case.

## Frozen trigger assertions

`trigger-cases.json` contains three explicit positives and six near misses.
Because the skill is manual-only:

- all explicit positive forms must activate;
- in-scope prompts that do not name the skill must not activate;
- completed implementation reviews must remain with `code-review`;
- naming `end-state-review` must not grant implementation authority; and
- status reporting and next-task selection must remain outside the boundary.

The portable description states `Explicit invocation only`, and Codex metadata
sets `policy.allow_implicit_invocation` to `false`. Static policy alignment
therefore passes; runtime selection consistency remains to be measured in fresh
contexts.

## Authoring and portability review

- **Pass — scope:** one coherent planning-reconstruction job with an explicit
  implementation exclusion.
- **Pass — trigger contract:** capability, use conditions, and manual-only
  activation are stated, with separate positive and near-miss cases.
- **Pass — authority boundary:** review, planning-edit approval, and code
  exclusion are explicit and observable.
- **Pass — portability:** runtime guidance uses portable frontmatter, no
  client-specific commands or absolute paths, and isolates Codex metadata.
- **Pass — package size:** the self-contained entry point has no unused runtime
  resources or indirect routing.
- **Incomplete — behavior evidence:** representative cases and assertions now
  exist, but fresh paired trials have not been executed.
- **Incomplete — empirical trigger evidence:** static policy is consistent, but
  activation rates and variance have not been measured.

## Validation

- `scripts/validate-skills`: passed for 36 skills and the registry.
- `node --test scripts/lib/skill-registry.test.js scripts/audit-global-skills.test.js`:
  passed, 12 tests.
- `go test ./...`: passed for `cmd/sjskills` and `internal/sjskills`.
- `bunx skills add ./skills/end-state-review --list`: discovered exactly
  `end-state-review`.
- Root and embedded version 4 registries are byte-identical.
- Both evaluation JSON files parse successfully.

## Remaining evidence

Run the frozen cases in fresh isolated contexts before claiming measured
behavioral value. Compare the three planning cases against `code-review` and
use the completed implementation case as a holdout. If repeated trials show no
material planning-quality or authority-boundary advantage, prefer the simpler
catalog and remove `end-state-review`; until then, the demonstrated contract
difference supports retention in the narrower `dev` profile.
