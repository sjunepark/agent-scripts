# Codex Pushover Notify

Codex Pushover Notify sends a Pushover notification when a Codex turn stops.
The `UserPromptSubmit` hook records a start timestamp, and the `Stop` hook
sends the duration plus a preview of the latest assistant message.

It also exposes a Pushover MCP server so Codex can proactively notify you or
request a decision while it is working.

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
export CODEX_PUSHOVER_INCLUDE_CWD="1"
```

Runtime state is written to `PLUGIN_DATA/state.json` when installed as a
plugin. It can be overridden with `CODEX_PUSHOVER_STATE_FILE`.

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

Run the script directly for local checks:

```sh
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs status
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs test
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs off
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs on
```

Use `CODEX_PUSHOVER_DRY_RUN=1` to test without calling Pushover.

You can also dry-run the MCP server through an MCP client with
`CODEX_PUSHOVER_DRY_RUN=1`.
