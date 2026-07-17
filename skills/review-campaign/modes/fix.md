# Fix Mode

Fix mode handles only findings that are both `auto` and `open`. Everything else reaches code through triage, the repo profile's export target, and normal development flow.

1. Collect eligible findings for one area or a few configured `small` areas.
2. Re-verify each finding at `HEAD`, apply the minimal change, and follow the `AGENTS.md` chain for every touched path.
3. Verify with the repo profile's commands for the touched module: typecheck, lint, test, targeted file lint/format, and markdown checks when present.
4. Make one commit per coherent cluster, including the ledger status flips: `fix(<area>): <summary> (campaign <ids>)`.
5. If a fix grows beyond its finding or no longer satisfies every auto-tier rule, revert that attempt and retier the finding to `triage`; do not push it through.

Fixes ride the campaign branch. Run fix sessions near milestones, keep clusters small, and merge according to the profile promptly afterward. Report fixed ids, files, commits, validation, findings retiered or blocked, and remaining open auto count.
