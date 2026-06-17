#!/usr/bin/env node
import { mkdir, readFile, rename, writeFile } from "node:fs/promises";
import { request } from "node:https";
import { homedir } from "node:os";
import { basename, dirname, join } from "node:path";

const PUSHOVER_API_HOST = "api.pushover.net";
const PUSHOVER_API_PATH = "/1/messages.json";
const DEFAULT_TITLE = "Codex finished";
const DEFAULT_MESSAGE = "Ready for your next prompt";
const DEFAULT_TIMEOUT_MS = 8000;
const DEFAULT_DEBOUNCE_MS = 3000;
const PUSHOVER_TITLE_MAX_LENGTH = 250;
const ASSISTANT_PREVIEW_MAX_LENGTH = 900;
const PUSHOVER_MESSAGE_MAX_LENGTH = 1024;
const MAX_STORED_TURNS = 50;

function firstNonEmpty(...values) {
  for (const value of values) {
    const trimmed = typeof value === "string" ? value.trim() : "";
    if (trimmed) return trimmed;
  }
  return undefined;
}

function parsePositiveInt(value, fallback) {
  const parsed = Number.parseInt(value ?? "", 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

function parseNonNegativeInt(value, fallback) {
  const parsed = Number.parseInt(value ?? "", 10);
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : fallback;
}

function pluginDataDir() {
  return firstNonEmpty(process.env.PLUGIN_DATA, process.env.CODEX_PUSHOVER_DATA_DIR) ?? join(homedir(), ".codex", "codex-pushover-notify");
}

function stateFilePath() {
  return firstNonEmpty(process.env.CODEX_PUSHOVER_STATE_FILE) ?? join(pluginDataDir(), "state.json");
}

function readConfig() {
  return {
    token: firstNonEmpty(process.env.PUSHOVER_APP_TOKEN, process.env.PUSHOVER_API_TOKEN),
    user: firstNonEmpty(process.env.PUSHOVER_USER_KEY),
    device: firstNonEmpty(process.env.PUSHOVER_DEVICE),
    sound: firstNonEmpty(process.env.PUSHOVER_SOUND),
    title: firstNonEmpty(process.env.CODEX_PUSHOVER_TITLE) ?? DEFAULT_TITLE,
    message: firstNonEmpty(process.env.CODEX_PUSHOVER_MESSAGE) ?? DEFAULT_MESSAGE,
    minDurationMs: parseNonNegativeInt(process.env.CODEX_PUSHOVER_MIN_MS, 0),
    debounceMs: parseNonNegativeInt(process.env.CODEX_PUSHOVER_DEBOUNCE_MS, DEFAULT_DEBOUNCE_MS),
    timeoutMs: parsePositiveInt(process.env.CODEX_PUSHOVER_TIMEOUT_MS, DEFAULT_TIMEOUT_MS),
    includeCwd: process.env.CODEX_PUSHOVER_INCLUDE_CWD !== "0",
    dryRun: process.env.CODEX_PUSHOVER_DRY_RUN === "1",
    verbose: process.env.CODEX_PUSHOVER_VERBOSE === "1",
  };
}

function isRecord(value) {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}

async function readJsonFile(path, fallback) {
  try {
    const parsed = JSON.parse(await readFile(path, "utf8"));
    return isRecord(parsed) ? parsed : fallback;
  } catch {
    return fallback;
  }
}

async function writeJsonFile(path, value) {
  await mkdir(dirname(path), { recursive: true });
  const tempPath = `${path}.${process.pid}.${Date.now()}.tmp`;
  await writeFile(tempPath, `${JSON.stringify(value, null, 2)}\n`, "utf8");
  await rename(tempPath, path);
}

async function readState() {
  const state = await readJsonFile(stateFilePath(), {});
  return {
    enabled: state.enabled === false ? false : true,
    turns: isRecord(state.turns) ? state.turns : {},
    lastNotificationKey: typeof state.lastNotificationKey === "string" ? state.lastNotificationKey : "",
    lastNotificationAt: typeof state.lastNotificationAt === "number" ? state.lastNotificationAt : 0,
  };
}

async function writeState(state) {
  await writeJsonFile(stateFilePath(), state);
}

function turnKey(input) {
  const sessionId = firstNonEmpty(input.session_id) ?? "unknown-session";
  const turnId = firstNonEmpty(input.turn_id) ?? "unknown-turn";
  return `${sessionId}:${turnId}`;
}

function trimSummary(text, maxLength = 160) {
  const collapsed = String(text ?? "").replace(/\s+/g, " ").trim();
  return collapsed.length > maxLength ? `${collapsed.slice(0, maxLength - 3)}...` : collapsed;
}

function trimMessageBody(text, maxLength) {
  const trimmed = String(text ?? "").trim();
  return trimmed.length > maxLength ? `${trimmed.slice(0, maxLength - 3)}...` : trimmed;
}

function formatDuration(ms) {
  const totalSeconds = Math.max(1, Math.round(ms / 1000));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes === 0) return `${totalSeconds}s`;
  if (seconds === 0) return `${minutes}m`;
  return `${minutes}m ${seconds}s`;
}

function formatTitle(config, input) {
  const cwd = typeof input.cwd === "string" ? basename(input.cwd) : "";
  const title = config.includeCwd && cwd ? `${config.title} - ${cwd}` : config.title;
  return trimSummary(title, PUSHOVER_TITLE_MAX_LENGTH);
}

function completionMessage(config, input, durationMs) {
  const base = durationMs >= 1000 ? `${config.message} (took ${formatDuration(durationMs)})` : config.message;
  const assistant = typeof input.last_assistant_message === "string" ? input.last_assistant_message : "";
  if (!assistant.trim()) return base;
  return trimMessageBody(`${base}\n\n${trimSummary(assistant, ASSISTANT_PREVIEW_MAX_LENGTH)}`, PUSHOVER_MESSAGE_MAX_LENGTH);
}

function pruneTurns(turns) {
  const entries = Object.entries(turns)
    .filter(([, value]) => isRecord(value) && typeof value.startedAt === "number")
    .sort((a, b) => b[1].startedAt - a[1].startedAt)
    .slice(0, MAX_STORED_TURNS);
  return Object.fromEntries(entries);
}

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
  if (!config.token || !config.user) return;

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

async function sendPushover(config, title, message) {
  if (config.dryRun) return;
  if (!config.token || !config.user) {
    throw new Error("Pushover is not configured. Set PUSHOVER_APP_TOKEN and PUSHOVER_USER_KEY.");
  }

  const body = new URLSearchParams({
    token: config.token,
    user: config.user,
    title,
    message,
  });
  if (config.device) body.set("device", config.device);
  if (config.sound) body.set("sound", config.sound);

  await new Promise((resolve, reject) => {
    const req = request(
      {
        host: PUSHOVER_API_HOST,
        path: PUSHOVER_API_PATH,
        method: "POST",
        headers: {
          "Content-Type": "application/x-www-form-urlencoded",
          "Content-Length": Buffer.byteLength(body.toString()),
        },
        timeout: config.timeoutMs,
      },
      (res) => {
        const chunks = [];
        res.on("data", (chunk) => chunks.push(chunk));
        res.on("end", () => {
          if (res.statusCode && res.statusCode >= 200 && res.statusCode < 300) {
            resolve();
            return;
          }
          const detail = Buffer.concat(chunks).toString("utf8").trim();
          reject(new Error(`Pushover request failed (${res.statusCode})${detail ? `: ${trimSummary(detail)}` : ""}`));
        });
      },
    );
    req.on("timeout", () => req.destroy(new Error("Pushover request timed out")));
    req.on("error", reject);
    req.end(body.toString());
  });
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
    console.log(`Configured: ${config.token && config.user ? "yes" : "no, set PUSHOVER_APP_TOKEN and PUSHOVER_USER_KEY"}`);
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
