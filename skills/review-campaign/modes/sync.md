# Sync Mode

Use this mode when the base branch has moved while the campaign branch reviewed against an older base, at latest before each milestone merge. Sync owns only the merge and ledger bookkeeping. It reviews nothing and writes no findings; stale cells are reconciled later by `continue`.

1. Merge base → campaign with a true merge commit (`git merge <base>` on the campaign branch), never the reverse. The profile's merge policy governs only campaign → base milestone merges.
   - Integrate the base faithfully without fixing reviewed code; findings reach code only through triage or fix.
   - Under `reviews/`, the campaign side is authoritative.
   - If conflict resolution undoes an applied auto fix, flip that finding back to `open` and retier it to `triage`.
2. Sweep staleness: for every `✓ <sha>` cell run `git diff --stat <sha>..HEAD -- <paths>`, flip changed cells to `~`, and keep the old sha.
3. Handle module drift: remove config-table and matrix rows for paths deleted by the merge, but keep their area files for history. List newly introduced top-level modules in the sync log as proposed rows; only `plan`, with the user, may add them.
4. Append one sync-log line to `REVIEW.md` with the date, merge commit sha7, cells flipped, and rows removed or proposed.
5. Commit only `reviews/`: `review(sync): merge <base> @ <sha7>, restale N cells`.

Report the merge commit, conflicts and resolutions, cells made stale, removed/proposed rows, reverted auto fixes, validation of the merge shape, and the next required mode.
