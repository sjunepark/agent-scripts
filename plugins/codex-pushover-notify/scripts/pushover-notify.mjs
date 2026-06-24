#!/usr/bin/env node
import {
  completionMessage,
  formatTitle,
  isConfigured,
  isRecord,
  pruneTurns,
  readConfig,
  readState,
  sendPushover,
  stateFilePath,
  turnKey,
  writeState,
} from "./pushover-client.mjs";

async function handleUserPromptSubmit(input) {
  const state = await readState();
  state.turns[turnKey(input)] = {
    startedAt: Date.now(),
    cwd: typeof input.cwd === "string" ? input.cwd : undefined,
  };
  state.turns = pruneTurns(state.turns);
  await writeState(state);
}

async function handleStop(input, config) {
  const state = await readState();
  if (!state.enabled) return;
  if (!isConfigured(config)) return;

  const key = turnKey(input);
  const turn = isRecord(state.turns[key]) ? state.turns[key] : undefined;
  const durationMs = typeof turn?.startedAt === "number" ? Math.max(0, Date.now() - turn.startedAt) : 0;
  if (durationMs < config.minDurationMs) return;

  const title = formatTitle(config, input);
  const message = completionMessage(config, input, durationMs);
  const notificationKey = `${key}\n${title}\n${message}`;
  const now = Date.now();
  if (notificationKey === state.lastNotificationKey && now - state.lastNotificationAt < config.debounceMs) return;

  await sendPushover(config, title, message);
  state.lastNotificationKey = notificationKey;
  state.lastNotificationAt = now;
  delete state.turns[key];
  state.turns = pruneTurns(state.turns);
  await writeState(state);
}

function hookOutput(fields = {}) {
  process.stdout.write(`${JSON.stringify({ continue: true, ...fields })}\n`);
}

async function readStdin() {
  const chunks = [];
  for await (const chunk of process.stdin) chunks.push(chunk);
  return Buffer.concat(chunks).toString("utf8");
}

async function handleHook() {
  const config = readConfig();
  const raw = await readStdin();
  if (!raw.trim()) {
    hookOutput();
    return;
  }

  const input = JSON.parse(raw);
  if (!isRecord(input)) {
    hookOutput();
    return;
  }

  try {
    if (input.hook_event_name === "UserPromptSubmit") {
      await handleUserPromptSubmit(input);
    } else if (input.hook_event_name === "Stop") {
      await handleStop(input, config);
    }
    hookOutput();
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    hookOutput(config.verbose ? { systemMessage: `codex-pushover-notify: ${message}` } : {});
  }
}

async function handleCli(args) {
  const action = args[0] ?? "status";
  const state = await readState();
  const config = readConfig();

  if (action === "on" || action === "off") {
    state.enabled = action === "on";
    await writeState(state);
    console.log(`Pushover notifications ${state.enabled ? "enabled" : "disabled"}.`);
    return;
  }

  if (action === "status") {
    console.log(`Pushover notifications: ${state.enabled ? "on" : "off"}`);
    console.log(`Configured: ${isConfigured(config) ? "yes" : "no, set PUSHOVER_APP_TOKEN and PUSHOVER_USER_KEY"}`);
    console.log(`State file: ${stateFilePath()}`);
    return;
  }

  if (action === "test") {
    await sendPushover(config, formatTitle(config, { cwd: process.cwd() }), "Test notification from Codex Pushover Notify");
    console.log(config.dryRun ? "Dry run completed." : "Sent Pushover test notification.");
    return;
  }

  console.error("Usage: pushover-notify.mjs [on|off|status|test]");
  process.exitCode = 2;
}

if (process.argv.length > 2) {
  handleCli(process.argv.slice(2)).catch((error) => {
    console.error(error instanceof Error ? error.message : String(error));
    process.exitCode = 1;
  });
} else {
  handleHook().catch((error) => {
    const message = error instanceof Error ? error.message : String(error);
    hookOutput(readConfig().verbose ? { systemMessage: `codex-pushover-notify: ${message}` } : {});
  });
}
