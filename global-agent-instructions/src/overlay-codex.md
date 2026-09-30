<!-- slot: instructions_file -->
AGENTS.md
<!-- slot: harness -->
Codex
<!-- slot: invoke -->
$
<!-- slot: delegation_note -->
Use ordinary subagents for non-implementation work.
<!-- slot: browser_use -->
Prefer a relevant CLI, API, or connector; use `gh` for GitHub work.
Use a browser when the task needs rendered UI, an authenticated session, or
visual verification. A URL alone does not require a browser.
<!-- slot: harness_sections -->
## KICPA files

On macOS, look for KICPA files in `/Volumes/Audit` and `/Volumes/Learning`.
Reconnect a missing volume with
`open -g 'smb://macshare@100.101.192.39/Users/user/Documents/Audit'` or
`open -g 'smb://macshare@100.101.192.39/Users/user/Documents/Learning'`.
Use saved Keychain credentials; never embed passwords in commands.
