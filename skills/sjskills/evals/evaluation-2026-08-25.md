# sjskills replacement evaluation — 2026-08-25

## Scope and provenance

This evaluation covers the new explicit-only `sjskills` skill and replacement
of the tracked repository-local `.agents/skills/sync` predecessor. The source
material is authored in this repository at commit
`c0b23ba524e84e61d58a7fab1d375822df898b4b`; no third-party code, assets, or
templates were retained.

Reusable outcome: operate project and fixed-global `sjskills` reconciliation
without confusing local catalog state, published content, desired selections,
or managed placements.

Distinctive help: preserve the CLI's authority boundaries, plan/apply/restore
sequence, published-source precondition, quarantine handling, and evidence-bound
global rollout.

Expected reuse: project adoption, project reconciliation, quarantine recovery,
global inspection, and separately authorized global rollout on machines that
use `sjskills`.

At evaluation time, the replacement was catalog-only (`manager: none`) and
explicit-only. The current registry supersedes that installation classification:
`sjskills` is now part of the fixed global baseline and is managed as a copied
Skills CLI skill. Explicit-only activation remains unchanged. Direct Skills CLI
installs, plugin deployment, and repository publication mechanics remain outside
its scope.

## Migration decisions

| Predecessor capability | Decision | Destination or reason |
| --- | --- | --- |
| Local catalog, published source, project selection, and global baseline model | Adapt | `SKILL.md` opening model |
| Read-only inspection versus explicitly authorized reconciliation | Adapt | `SKILL.md` request classification and project workflow |
| Plan, apply, final plan, quarantine retention, and restore boundaries | Adapt | `SKILL.md` project workflow and restore sections |
| Fixed-global evidence and authorization boundary | Adapt and deepen | `references/global-rollout.md` |
| Published-ref verification before reconciliation | Adapt | `SKILL.md` publication precondition |
| Exact repository validation, Git staging, PR, and merge commands | Remove | Applicable repository instructions own publication mechanics |
| Plugin deployment and ad hoc `bunx skills` installs | Exclude | Separate plugin and Skills CLI workflows |
| Implicit discovery for broad sync requests | Replace | Explicit-only description and Codex adapter policy |
| Existing synthetic reconciliation cases | Adapt | `evals/evals.json` and `evals/trigger-cases.json` |

## Frozen candidate

| File | SHA-256 |
| --- | --- |
| `SKILL.md` | `893e7b19bc3140c007488779ec0caa2c5d23f6651218fe3743273a3152a278ff` |
| `references/global-rollout.md` | `f47c50640c0569ee1917ce0db6b89d8cabc88fb1bc6df1042b0b030e58d4592c` |
| `agents/openai.yaml` | `babf42abbc0001dd1b76a4a0b57e1efaf4d6122cb0bdf750af7c6703bf812aa4` |
| `evals/evals.json` | `8132aa8ea87726849d974e300390d8c5a966cf8e6d88eef933896c0beb25c1ee` |
| `evals/trigger-cases.json` | `2b264b5df0923b9c21d37d31f413b7912795b0c2dfeef04d1f38c0dd2bf202c9` |

## Static gate — final-tree revalidation, 2026-09-04

The following passed on the PR #17 tree based on `3cbd575`, after removal of
`.agents/skills/sync` and consolidation of the catalog:

- Skill Creator `quick_validate.py` — both `skills/sjskills` and
  `.agents/skills/sjskills` valid using temporary PyYAML dependencies.
- `scripts/validate-skills` — 31 catalog skills and the registry valid.
- `bunx skills add ./skills/sjskills --list` — one skill discovered.
- `bunx skills add ./skills --list` — 31 catalog skills discovered.
- `node --test scripts/lib/skill-registry.test.js scripts/audit-global-skills.test.js`
  — 12 of 12 tests passed.
- `go test ./...` — command and internal reconciler packages passed.
- `git diff --check` — passed.

This refresh covers static and repository gates; the behavior and trigger
results below remain the original candidate evaluation.

## Behavior gate

A fresh read-only evaluator compared the predecessor with the candidate against
cases 1–7 in `evals/evals.json`; case 8 was checked as the Skills CLI boundary.
All eight passed their observable assertions. The first review identified three
candidate gaps: the four-state model, pinned-source verification, and explicit
separation of drift from operational failure. After focused edits, fresh retests
of cases 2 and 7 passed and the evaluator found no material replacement blocker.

No live publication, project apply, restore, or real-home global mutation was
performed. Reconciler mechanics are covered separately by the repository's Go
and Node tests.

## Trigger gate

A separate fresh evaluator classified the three explicit positives and four
near misses in `evals/trigger-cases.json`. All seven matched the frozen labels.
The adapter explicitly sets `allow_implicit_invocation: false`; in-scope prompts
that do not invoke `$sjskills` remain negative cases.

One documented ambiguity remains portable across hosts: plain-language wording
such as “Use sjskills” may name the CLI rather than explicitly invoke the skill.
Codex enforces the intended boundary through adapter metadata, and the saved
cases use `$sjskills` for positive invocation.

## Authoring review

The authoring rubric passes with no blocker or major finding. The package has
one coherent purpose, directly routes its only conditional runtime reference,
keeps client policy in `agents/openai.yaml`, preserves mutation authority, and
states observable completion. Publication implementation details were removed
rather than duplicated from repository instructions.

Decision: adopt `skills/sjskills` and remove `.agents/skills/sync`. Installation,
publication, commit, push, and real-machine reconciliation remain outside this
change.
