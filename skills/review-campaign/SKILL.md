---
name: review-campaign
description: "Long-running systematic codebase review with a persistent ledger in reviews/. Use to plan review areas, continue the next review pass, check campaign status, triage findings with the user, apply auto-tier fixes, or absorb base-branch drift into the campaign. Modes: plan, continue, status, triage, fix, sync (default continue)."
---

# Review Campaign

Stateful engine for reviewing a whole repo in small, resumable sessions. The skill is stateless and repo-agnostic; all campaign state lives in `reviews/` (the ledger) at the target repo's root, on the campaign branch named in the repo profile. If it is not in the ledger, it did not happen — never rely on session memory.

## Vocabulary

- **Area**: a matrix row — a directory cluster or a cross-cutting seam, sized for one session.
- **Pass**: a matrix column — `structure` or one detail dimension. Each pass has a rubric in `references/`.
- **Cell**: area × pass. The unit of work for one session.
- **Finding**: one numbered, actionable item in an area file.
- **Tier**: `auto` (fix without discussion) or `triage` (needs the user).

## Ledger contract

```text
reviews/
  REVIEW.md       dashboard: repo profile, phases, next-up pointer, area config table, matrix, sync log
  areas/<id>.md   per-area: Map / Findings / Pass log
  decisions.md    append-only triage log (accept/reject + reason)
```

Matrix cell states: `·` pending | `✓ <sha7>` done at that commit | `~ <sha7>` stale (area changed since) | `—` not applicable.

REVIEW.md opens with the **repo profile** — the per-repo facts every session reads instead of assuming: campaign branch and merge policy, export convention for accepted findings (TODO files, issue tracker, ...), verification commands (typecheck/lint/test, targeted file lint/format, markdown checks), stack notes per area, agent-instruction rules to honor, and anchors (existing review docs to fold in, recorded keep-as-is verdicts, growth domains).

Area file template:

```markdown
# <id> — review notes

paths: <from REVIEW.md config>

## Map
<structure pass writes this: responsibilities, dependency direction, public surface,
keep-as-is verdicts. Detail passes read it and do not re-litigate verdicts.>

## Findings

### <PREFIX>-001 · major · triage · open — <title>
- pass: structure @ <sha7>
- where: path:line[, path:line]
- problem: <what is wrong and why it matters>
- action: <concrete change>

## Pass log
- <date> · structure @ <sha7> · 2 findings (1 major, 1 nit)
- <date> · tests @ <sha7> · clean
```

Finding fields:

- **id**: `<PREFIX>-<NNN>` from the area's prefix in REVIEW.md, sequential, never reused.
- **severity**: `blocker` (correctness, security, data loss) | `major` (will cost real time or quality) | `minor` (worth fixing opportunistically) | `nit` (mention once, never argue).
- **tier**: `triage` by default; `auto` only per the eligibility rules below, with a one-line `why auto:` justification.
- **status**: `open → accepted | rejected | fixed | obsolete`. Auto tier may go `open → fixed` directly. Accepting exports the item per the profile's export convention with its id; the exported entry becomes the work item and the campaign is done with it.

## Modes

Invoked as `/review-campaign <mode>`; no argument means `continue`.

### plan — bootstrap or replan

1. Inventory: tracked-file counts per directory, languages and frameworks (e.g. Effect usage via import grep), test locations, colocated docs, existing review docs and TODO conventions.
2. Write or refresh the repo profile. At first bootstrap ask the user for the campaign branch + merge policy and the export convention — those are theirs to choose; derive verification commands from the repo's scripts and agent instructions.
3. Propose or update the area table: paths, id prefix, pass applicability, priority, `small` flag (small areas batch all their passes into one session).
4. Replan duty: for every `✓ <sha>` cell run `git diff --stat <sha>..HEAD -- <paths>`; flip changed cells to `~`, keeping the old sha.
5. Seed `reviews/` files that are missing. Confirm row changes with the user when present; otherwise only apply splits already recorded by structure passes.

### sync — absorb base-branch drift

Run when the base branch has moved while the campaign branch reviewed against an older base — at latest before each milestone merge. Sync owns only the merge and the ledger bookkeeping; it reviews nothing and writes no findings — the `~` cells it produces are reconciled cell-by-cell by `continue`.

1. Merge base → campaign with a true merge commit (`git merge <base>` on the campaign branch), never the reverse — the profile's merge policy governs only campaign → base milestone merges. Conflict rules: integrate the base faithfully without fixing reviewed code (findings reach code only through triage/fix); under `reviews/` the campaign side is authoritative; a resolution that undoes an applied auto fix flips that finding back to `open` and retiers it to `triage`.
2. Staleness sweep: run `plan`'s replan duty.
3. Module drift: paths the merge deleted get their rows removed from the config table and matrix (keep the area file for history); top-level modules the base introduced are listed in the sync log as proposed rows — adding rows is `plan`'s job and needs the user.
4. Record the sync in REVIEW.md's sync log: one line with date, merge commit sha7, cells flipped, rows removed or proposed.
5. Commit (`reviews/` only): `review(sync): merge <base> @ <sha7>, restale N cells`.

### continue — one session, one cell

1. Read `reviews/REVIEW.md` **only** (not the whole ledger).
2. Pick the next cell: stale (`~`) cells in the current phase first, then by phase order and area priority. Within detail phases, finish one area's columns (left to right) before moving to the next area.
3. Read that area's file (create from template if missing) and the one rubric for the pass. Load nothing else.
4. Review per the rubric. Verify every claim before writing it (read the code path, run the git history check, grep the usage) — findings are acted on later without re-verification. For stale cells, scope to `git diff <sha>..HEAD -- <paths>` plus existing findings: mark `obsolete` where the code is gone or rewritten past recognition, repoint `where:` refs that merely moved, then review the delta. Never silently drop a finding.
5. Write findings (dedup first — see session discipline), append a pass-log line, stamp the cell `✓ <HEAD sha7>`, update the next-up pointer.
6. Commit the notes: `review(<area>): <pass> pass notes`. Notes commits touch only `reviews/`; rubric sharpenings belong in the skill's source repo, not the target repo.
7. Report: cell finished, finding counts by severity, blockers called out explicitly, open auto-tier count, next cell. If the matrix has no pending cells, report the campaign complete and suggest a final triage and milestone merge.

### status — read-only

Print the matrix, open-finding counts by severity and tier, a staleness sweep (diff each `✓` cell, report without restamping), and base drift (`git rev-list --count HEAD..<base>`). Suggest `triage` when open triage findings exceed ~15 or a column just completed; suggest a milestone merge when a phase completes; suggest `sync` when the base is ahead.

### fix — auto tier only

Everything else reaches code through triage → the profile's export target → normal dev flow. Fix mode handles only `auto + open` findings:

1. Collect them for one area (or a few `small` areas).
2. For each: re-verify it still holds at HEAD; apply the minimal change; follow the AGENTS.md chain for every touched path.
3. Verify with the repo profile's verification commands for the touched module — typecheck, lint, test, plus targeted file lint/format and markdown checks when the repo has them.
4. One commit per cluster, including the ledger status flips: `fix(<area>): <summary> (campaign <ids>)`.
5. A fix that grows beyond its finding is retiered to `triage` and reverted, not pushed through.

Fixes ride the campaign branch, so run fix sessions near milestones and keep clusters small — merge per the profile's branch model promptly after.

### triage — requires the user

Walk `triage + open` findings, blockers first, batched by area. Outcomes per finding: **accept** (export per the profile with id, mark `accepted`), **reject** (one-line reason to `decisions.md`, mark `rejected`), or **retier to auto** (user says it's obvious). Close by appending all decisions to `decisions.md`, and when a phase column is complete, propose the milestone merge per the profile's branch model.

## Phase order

1. `structure` — every area, by priority. Completes before any detail pass starts, so detail reviews see the post-refactor shape.
2. `security` — trust-boundary areas only.
3. Detail passes per area: `tests → errors → logging → effect → naming → diet → docs`.

## Session discipline

- Load only REVIEW.md, the current area file, and one rubric. Never the whole ledger or other areas' notes.
- One cell per session; `small` areas may batch all their passes.
- **A clean pass is a valid result.** Record it in the pass log and stop. Do not manufacture findings; report only what would change what a maintainer does.
- Dedup before writing: check the area's existing findings and `decisions.md`. Never resurface a rejected finding unless the code changed since the rejection.
- A finding that belongs to another pass goes where the fix is (one finding, one home) — leave a one-line cross-reference at most.
- Structure passes additionally: update the Map, record keep-as-is verdicts, and propose row splits for areas too big for one session (add rows to REVIEW.md config + matrix).

## Auto-tier eligibility

Default is `triage`. Mark `auto` only when **all** hold:

1. Small blast radius: one file or one tight cluster; no exported API, contract, or schema change; no behavior change beyond restoring obviously intended behavior.
2. Mechanically verifiable: existing checks or tests prove the fix, or the defect is self-evident (dead import, broken doc link, log field typo).
3. No design opinion involved: if you can imagine a reasonable maintainer arguing, it is triage.
4. Not a removal of intentional-looking code. Only trivially dead artifacts qualify (unused import, unreferenced local, byte-identical duplicate test).
5. Not from the security pass — security findings are never auto.

## Long-running use

Goal prompt: `Run /review-campaign repeatedly until it reports the matrix complete or a blocker finding. Every few sessions run /review-campaign status; run sync when it reports base drift. Stop and surface blockers immediately.`

Cadence: review sessions run unattended; `triage` runs only with the user; merge the campaign branch at milestones (phase complete or after a triage batch), which also carries exported work items and auto fixes. Run `status` every few sessions; run `sync` when it reports base drift, and always before a milestone merge.
