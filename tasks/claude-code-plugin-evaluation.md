# Evaluate Claude Code packaging for the Codex plugins

## Outcome

Decide whether any plugin under `plugins/` should also ship for Claude Code,
and in what form. Not started.

## Context

The plugins are Codex-only; nothing in the repository carries a
`.claude-plugin/` manifest. Claude Code plugins can now ship mods: function
hooks declared as modules in `hooks/hooks.json`, with handlers for events such
as `session.start`, `turn.start`, and `turn.complete` and per-session state, as
described in
[Getting started with Claude Code mods](https://claude.dev/blog/getting-started-with-claude-code-mods/).

## Candidates

- `codex-pushover-notify` is the closest fit: it times a turn across two
  command hooks and a state file, which maps onto turn events with in-memory
  state. It also ships an MCP server (`.mcp.json`) that any Claude packaging
  must account for.
- `sjskills-maintenance` already falls back to `CLAUDE_PLUGIN_ROOT` and
  `CLAUDE_PLUGIN_DATA` but has no Claude manifest. Mods run sandboxed without
  Node, so its process supervision would not carry over unchanged.
- `chezmoi-sync` is a session-start notice with a hardcoded checkout path; low
  value.

## Open questions

- Whether the existing command-style `hooks.json` can ship in a Claude plugin
  with only a manifest added, making a mod unnecessary.
- The mod API's process and notification surface, which the post does not
  document; read the Claude Code plugin documentation before designing.
- Whether a mod should also log skill invocations to give field evidence for
  implicit-skill selection.
