# TODO

- [skills/review-campaign] Add a base-branch sync flow (merge + restale) to the skill.
  - Need observed in the creo campaign (2026-06-10): `main` progressed substantially while review sessions ran on the campaign branch; the skill has no instructions for absorbing that drift. `plan` step 4 (replan duty) covers restamping, but nothing tells a session to merge first, which direction to merge, or how findings interact with moved/deleted code mid-campaign.
  - Proposed: a `sync` mode (or a "base-branch sync" section under `plan`) that codifies:
    1. Merge base → campaign (`git merge main` on the campaign branch), never the reverse; resolve conflicts without fixing reviewed code — the merge carries no opinions, findings still reach code only via triage/fix.
    2. After the merge, run the staleness sweep: for every `✓ <sha>` matrix cell, `git diff --stat <sha>..HEAD -- <paths>`; unchanged cells stay `✓`, changed cells flip to `~ <old-sha>`.
    3. Findings pass: for each open finding in stale areas, re-verify `where:` at HEAD — mark `obsolete` when the code is gone/rewritten, repoint line refs when it merely moved, leave open otherwise. Never silently drop a finding.
    4. Area-table pass: new top-level modules brought by the base get new rows; deleted paths get their rows retired (note in pass log, keep area file for history).
    5. Record the sync in REVIEW.md (one line: date, merge commit, cells flipped) so later sessions know why cells went stale.
  - Also: `continue` already handles `~` cells (diff-scoped review); the new flow should explicitly hand off to it rather than re-reviewing whole areas.
