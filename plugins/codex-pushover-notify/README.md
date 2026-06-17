# Codex Pushover Notify

Codex Pushover Notify sends a Pushover notification when a Codex turn stops.
The `UserPromptSubmit` hook records a start timestamp, and the `Stop` hook
sends the duration plus a preview of the latest assistant message.

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

## Manual Commands

Run the script directly for local checks:

```sh
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs status
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs test
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs off
node plugins/codex-pushover-notify/scripts/pushover-notify.mjs on
```

Use `CODEX_PUSHOVER_DRY_RUN=1` to test without calling Pushover.
