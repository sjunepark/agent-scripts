# Codex Pushover Notify

Codex Pushover Notify sends a Pushover notification when a Codex turn stops.
The `UserPromptSubmit` hook records a start timestamp, and the `Stop` hook
sends the duration plus a preview of the latest assistant message.

It also exposes a Pushover MCP server so Codex can proactively notify you or
request a decision while it is working.

## Install

This plugin is published through the repository's `personal` marketplace. From
a machine where that remote-backed marketplace is registered, install it with:

```sh
codex plugin add codex-pushover-notify@personal
```

Repository maintainers should follow the plugin publication and reinstall
workflow in [Settings Sync](../../docs/settings-sync.md#codex-plugins) after
changing the plugin.

## Configuration

Set these environment variables where Codex can read them:

```sh
export PUSHOVER_APP_TOKEN="..."
export PUSHOVER_USER_KEY="..."
```

Optional settings:

```sh
export PUSHOVER_DEVICE="iphone"
export PUSHOVER_SOUND="pushover"
export CODEX_PUSHOVER_TITLE="Codex finished"
export CODEX_PUSHOVER_MESSAGE="Ready for your next prompt"
export CODEX_PUSHOVER_MIN_MS="0"
export CODEX_PUSHOVER_DEBOUNCE_MS="3000"
export CODEX_PUSHOVER_TIMEOUT_MS="8000"
export CODEX_PUSHOVER_INCLUDE_CWD="0"
```

The working-directory name is included in completion titles by default; set
`CODEX_PUSHOVER_INCLUDE_CWD=0` to omit it. The other values above show their
defaults except for the optional device and sound overrides.

Runtime state is written to `$PLUGIN_DATA/state.json` when Codex provides the
plugin data directory. Outside an installed plugin, it defaults to
`~/.codex/codex-pushover-notify/state.json`. Set `CODEX_PUSHOVER_DATA_DIR` to
change the fallback directory or `CODEX_PUSHOVER_STATE_FILE` to set the exact
path.

If the Pushover credentials are absent, completion hooks continue without
sending a notification. Set `CODEX_PUSHOVER_VERBOSE=1` to surface hook errors
as Codex system messages.

## MCP Tools

The bundled MCP server exposes:

- `pushover_notify`: send a one-way notification for material status or attention.
- `pushover_request_decision`: call the user for non-obvious judgment, risky tradeoffs, or real blockers.
- `pushover_status`: check configuration without exposing secrets.

`pushover_request_decision` is intentionally one-way. Codex should send the
notification, then wait for the user to answer in the Codex thread.

Use emergency urgency only when the phone should repeatedly alert until the
notification is acknowledged.

## Manual Commands

From the repository root, run the script directly for local checks or to toggle
completion-hook notifications:

```sh
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs status
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs test
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs off
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs on
```

Use `CODEX_PUSHOVER_DRY_RUN=1` to test without calling Pushover.

You can also dry-run the MCP server through an MCP client with
`CODEX_PUSHOVER_DRY_RUN=1`.
