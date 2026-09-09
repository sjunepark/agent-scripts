# Prepare a Goal

Use for `prepare`, `spawn`, or a separately authorized equivalent planning and
commit request. Requires repository read/write and Git capabilities, plus the
`progress` skill for planning. If a prerequisite is unavailable, report the
affected step; do not claim a prepared handoff.

## Complete the Plan

1. Resolve the substantial outcome under the entry point before editing. Reuse
   the current discussion, selected project, authoritative planning scope, and
   prior decisions. If the scope is too small or externally blocked, finish
   under the entry point without inventing planning work.
2. Invoke `$progress` for planning only. Update the existing authoritative plan
   and index, or create them using its organize workflow when absent. Capture
   the outcome, current evidence, intended behavior and design, dependencies,
   implementation sequence, acceptance and validation, and next action to the
   detail needed by a fresh session. Let that skill enforce coherent end-state
   planning; keep implementation detail in the plan rather than the goal prompt.
3. Resolve implementer-owned choices from evidence. Honor existing delegation
   for consequential choices and record its scope. For remaining user-owned
   decisions, use the readiness gate's delegation-or-repair choice; `prepare`
   authorizes writing the plan but does not supply missing product decisions.
   Finish independent planning while an answer is pending. Never mark an
   unanswered decision settled or start implementation to fill a planning gap.
4. Run one bounded review of the preparation against the selected outcome and
   repository evidence, using `$code-review` when available and repository-
   required checks. Apply safe planning fixes, reconcile affected docs when
   required, and validate changed links and acceptance conditions. Reuse a
   sufficient existing plan without cosmetic rewrites or another review when
   current evidence already establishes that it passed unchanged.

## Commit the Preparation

1. Inspect status, staged and unstaged diffs before writing and again before
   committing. Identify the exact planning paths/hunks this operation owns and
   separately authorized existing work. Preserve unrelated staged entries,
   unstaged edits, and untracked files. Do not use blanket staging or include
   all pre-existing changes merely because the user requested preparation.
2. Commit only reviewed preparation and separately authorized work with a
   descriptive message explaining the outcome and material decisions. Use an
   isolated index or equally scoped method when the real index contains
   unrelated changes; preserve its entries and partial staging. If an edited
   file mixes unrelated user work that cannot be safely separated, finish other
   preparation and resolve that boundary before committing the mixed content.
   Honor no-commit instructions; `spawn` may continue only if its required
   handoff is already present in the selected committed state.
3. Skip an empty commit when the necessary preparation is already committed.
   Record the prepared commit, branch/ref, planning scope, and source paths.
   Verify the commit's actual path/content scope and the preservation of
   unrelated work. Confirm every handoff source and required implementation
   prerequisite is readable from that commit. Uncommitted required work needs
   separate commit authority; do not stash, discard, or silently omit it.
4. Revalidate the selected outcome and readiness from the prepared state. If
   preparation changed its evidence, preserve the selected boundary unless it
   became invalid. Review and commit any necessary correction before handing
   off. Report preparation evidence in commentary, then route the final result
   through the entry point.

Preparation does not authorize push, installation, publication, PR creation, or
goal execution in the current session. The selected goal's later delivery
lifecycle applies in the executing session. If a check or commit fails, retain
completed preparation and report the specific failure without claiming success.
