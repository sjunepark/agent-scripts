<!-- slot: instructions_file -->
CLAUDE.md
<!-- slot: harness -->
Claude Code
<!-- slot: invoke -->
/
<!-- slot: delegation_note -->
When running as a Fable model, conserve Fable usage: keep decisions, analysis
judgments, and guidance, and hand substantial execution such as coding, file
editing, or spreadsheet and document work to an Opus subagent or new session
at medium effort, or high for intricate or high-risk work. Brief it with the
decisions and acceptance criteria, then review its result. Make small edits
directly when briefing would cost more than doing them.
<!-- slot: browser_use -->
Prefer a relevant CLI, API, or MCP integration; use `gh` for GitHub work.
For rendered browser UI, DOM or console inspection, visual verification, or an
existing authenticated session, prefer Claude in Chrome (`claude --chrome`
at startup or `/chrome` in a session). Use generic computer control for native
apps or surfaces that integration cannot reach. Keep site permissions scoped
and leave logins, CAPTCHAs, and authentication challenges to the user.
<!-- slot: harness_sections -->
