---
name: review-campaign
description: "Run a stateful whole-repository review campaign with a persistent `reviews/` ledger. Explicit invocation only."
---

# Review Campaign

Run a stateful whole-repository review in small, resumable sessions. The skill is stateless and repo-agnostic; all campaign state lives in `reviews/` at the target repository root, on the campaign branch named in the repo profile. If it is not in the ledger, it did not happen—never rely on session memory.

## Vocabulary

- **Area**: a matrix row—a directory cluster or cross-cutting seam sized for one session.
- **Pass**: a matrix column—`structure` or one detail dimension.
- **Cell**: area × pass, the normal unit of work for one session.
- **Finding**: one numbered, actionable item in an area file.
- **Tier**: `auto` (fix without discussion) or `triage` (needs the user).

## Choose One Mode

Invoke as `$review-campaign <mode>`; no argument means `continue`.

Bootstrap guard: when `reviews/REVIEW.md` does not exist, treat an omitted mode or explicit `continue` as `plan`. Only `plan` bootstraps the ledger; `continue` must not create `reviews/REVIEW.md`.

Read and follow exactly the selected mode file:

- **plan** — bootstrap or replan: [modes/plan.md](modes/plan.md)
- **sync** — absorb base-branch drift without reviewing code: [modes/sync.md](modes/sync.md)
- **continue** — complete the next review cell: [modes/continue.md](modes/continue.md)
- **status** — report progress and drift without writing: [modes/status.md](modes/status.md)
- **fix** — apply only eligible `auto + open` findings: [modes/fix.md](modes/fix.md)
- **triage** — decide `triage + open` findings with the user: [modes/triage.md](modes/triage.md)

Before `plan`, `sync`, `continue`, `fix`, or `triage`, read [formats/ledger.md](formats/ledger.md) for the exact ledger schema, finding state transitions, and auto-tier rules. `status` may read the existing ledger directly; load the format only if the ledger is incomplete or ambiguous.

## Route the Selected Pass

In `continue` mode, read exactly the rubric for the selected pass. For a configured `small` area that batches passes, load rubrics one at a time as each pass runs; do not preload them all.

- `structure` → [rubrics/structure.md](rubrics/structure.md)
- `security` → [rubrics/security.md](rubrics/security.md)
- `tests` → [rubrics/tests.md](rubrics/tests.md)
- `errors` → [rubrics/errors.md](rubrics/errors.md)
- `logging` → [rubrics/logging.md](rubrics/logging.md)
- `effect` → [rubrics/effect.md](rubrics/effect.md)
- `naming` → [rubrics/naming.md](rubrics/naming.md)
- `diet` → [rubrics/diet.md](rubrics/diet.md)
- `docs` → [rubrics/docs.md](rubrics/docs.md)

Use this phase order:

1. `structure` for every area, by priority. Complete it before detail passes so later reviews see the post-refactor shape.
2. `security` for trust-boundary areas only.
3. Detail passes per area: `tests → errors → logging → effect → naming → diet → docs`.

## Campaign Invariants

- Follow the repo profile for branch/merge policy, accepted-finding export, verification, stack notes, agent instructions, and recorded architectural anchors.
- Apply user choices already supplied for the selected mode; ask only for
  unresolved campaign or triage decisions. Reverify recorded findings at the
  current code before an authorized fix.
- Verify every finding against code, history, or usage before recording it. A clean pass is valid; never manufacture findings.
- Keep one finding in the pass where its fix belongs. Cross-reference it elsewhere in at most one line.
- Surface blocker findings immediately. Security findings are never auto-tier.
- Keep campaign review-note commits limited to `reviews/`. Findings reach code only through `fix` or the repository's normal development flow after triage export; `sync` may integrate base drift but must not fix reviewed code.

## Long-Running Use

Goal prompt: `Run $review-campaign repeatedly until it reports the matrix complete or a blocker finding. Every few sessions run $review-campaign status; run $review-campaign sync when it reports base drift. Stop and surface blockers immediately.`

Review sessions may run unattended. Merge the campaign branch at milestones—a completed phase or triage batch—so exported work items and auto-fixes travel with the ledger.
