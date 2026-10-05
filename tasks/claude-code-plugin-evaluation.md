# Evaluate Claude Code packaging for the Codex plugins

## Outcome

Decide whether any plugin under `plugins/` should also ship for Claude Code,
and in what form. `sjskills-maintenance` ships; the others are undecided.

## Decisions

- Command-style `hooks/hooks.json` ships unchanged with only a
  `.claude-plugin/plugin.json` added; no mod is needed. Claude tolerates the
  Codex-only `statusMessage` and `commandWindows` fields and runs `command`
  through Git Bash on Windows (Claude Code 2.1.286, probed 2026-10-06).
- A manifest `hooks` path supplements `hooks/hooks.json` rather than replacing
  it, so a per-harness hooks file would run both.
- `sjskills-maintenance` runs only CLI status under Claude; see
  [docs/sjskills-startup-hook.md](../docs/sjskills-startup-hook.md#claude-code).

## Remaining candidates

- `codex-pushover-notify`: times a turn across two command hooks and a state
  file, and ships an MCP server (`.mcp.json`) that Claude packaging must
  account for. Not requested yet.
- `chezmoi-sync`: a session-start notice with a hardcoded checkout path; low
  value.

## Open questions

- Whether a mod should also log skill invocations to give field evidence for
  implicit-skill selection.
