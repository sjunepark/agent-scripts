# Status Mode

Keep this mode read-only.

1. Print the current matrix.
2. Count open findings by severity and tier.
3. Run a staleness sweep for every `✓ <sha>` cell with `git diff --stat <sha>..HEAD -- <paths>` and report changed cells without restamping them.
4. Report base drift with `git rev-list --count HEAD..<base>` using the base branch from the repo profile.
5. Suggest `triage` when open triage findings exceed about 15 or a column just completed; suggest a milestone merge when a phase completes; suggest `sync` when the base is ahead.

Do not edit ledger files, findings, cells, branches, or code. Lead with campaign completeness, stale-cell count, base drift, and the recommended next mode.
