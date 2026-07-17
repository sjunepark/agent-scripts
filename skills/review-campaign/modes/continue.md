# Continue Mode

Complete one review cell. A configured `small` area may batch all its applicable passes.

1. Start by reading `reviews/REVIEW.md` only, not the whole ledger.
2. Pick the next cell: stale (`~`) cells in the current phase first, then phase order and area priority. Within detail phases, finish one area's columns left to right before moving to the next area.
3. Read the selected area's file, creating it from the ledger template if missing. Read only the selected pass rubric; for a `small` batch, load each rubric only when its pass begins.
4. Review the configured paths against that rubric. Verify every claim by reading the code path, checking history, or grepping usage because findings are acted on later without re-verification.
   - For stale cells, scope to `git diff <sha>..HEAD -- <paths>` plus existing findings. Mark a finding `obsolete` when code is gone or rewritten past recognition; repoint `where:` references that merely moved; then review the delta. Never silently drop a finding.
5. Deduplicate before writing against existing findings and `decisions.md`. Never resurface a rejected finding unless the code changed since rejection.
6. Write findings, append a pass-log line, stamp the cell `✓ <HEAD sha7>`, and update the next-up pointer.
   - A clean pass is valid: record it in the pass log and stop. Do not manufacture findings; report only what changes what a maintainer does.
   - A finding belonging to another pass goes where its fix belongs; leave at most a one-line cross-reference.
   - Structure passes also update the Map, record keep-as-is verdicts, and propose row splits for areas too large for one detail session by adding rows to the `REVIEW.md` config and matrix.
7. Commit the notes: `review(<area>): <pass> pass notes`. Notes commits touch only `reviews/`; rubric changes belong in the skill source repository, not the target repository.
8. Report the finished cell, finding counts by severity, blockers explicitly, open auto-tier count, and next cell. If no matrix cells remain pending, report the campaign complete and suggest final triage and a milestone merge.
