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

export function firstNonEmpty(...values) {
  for (const value of values) {
    const trimmed = typeof value === "string" ? value.trim() : "";
    if (trimmed) return trimmed;
  }
  return undefined;
}

export function parsePositiveInt(value, fallback) {
  const parsed = Number.parseInt(value ?? "", 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

export function parseNonNegativeInt(value, fallback) {
  const parsed = Number.parseInt(value ?? "", 10);
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : fallback;
}

export function pluginDataDir() {
  return firstNonEmpty(process.env.PLUGIN_DATA, process.env.CODEX_PUSHOVER_DATA_DIR) ?? join(homedir(), ".codex", "codex-pushover-notify");
}

export function stateFilePath() {
  return firstNonEmpty(process.env.CODEX_PUSHOVER_STATE_FILE) ?? join(pluginDataDir(), "state.json");
}

export function readConfig() {
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

export function isRecord(value) {
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

export async function readState() {
  const state = await readJsonFile(stateFilePath(), {});
  return {
    enabled: state.enabled === false ? false : true,
    turns: isRecord(state.turns) ? state.turns : {},
    lastNotificationKey: typeof state.lastNotificationKey === "string" ? state.lastNotificationKey : "",
    lastNotificationAt: typeof state.lastNotificationAt === "number" ? state.lastNotificationAt : 0,
  };
}

export async function writeState(state) {
  await writeJsonFile(stateFilePath(), state);
}

export function turnKey(input) {
  const sessionId = firstNonEmpty(input.session_id) ?? "unknown-session";
  const turnId = firstNonEmpty(input.turn_id) ?? "unknown-turn";
  return `${sessionId}:${turnId}`;
}

export function trimSummary(text, maxLength = 160) {
  const collapsed = String(text ?? "").replace(/\s+/g, " ").trim();
  return collapsed.length > maxLength ? `${collapsed.slice(0, maxLength - 3)}...` : collapsed;
}

export function trimMessageBody(text, maxLength = PUSHOVER_MESSAGE_MAX_LENGTH) {
  const trimmed = String(text ?? "").trim();
  return trimmed.length > maxLength ? `${trimmed.slice(0, maxLength - 3)}...` : trimmed;
}

export function formatDuration(ms) {
  const totalSeconds = Math.max(1, Math.round(ms / 1000));
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes === 0) return `${totalSeconds}s`;
  if (seconds === 0) return `${minutes}m`;
  return `${minutes}m ${seconds}s`;
}

export function formatTitle(config, input) {
  const cwd = typeof input.cwd === "string" ? basename(input.cwd) : "";
  const title = config.includeCwd && cwd ? `${config.title} - ${cwd}` : config.title;
  return trimSummary(title, PUSHOVER_TITLE_MAX_LENGTH);
}

export function completionMessage(config, input, durationMs) {
  const base = durationMs >= 1000 ? `${config.message} (took ${formatDuration(durationMs)})` : config.message;
  const assistant = typeof input.last_assistant_message === "string" ? input.last_assistant_message : "";
  if (!assistant.trim()) return base;
  return trimMessageBody(`${base}\n\n${trimSummary(assistant, ASSISTANT_PREVIEW_MAX_LENGTH)}`);
}

export function pruneTurns(turns) {
  const entries = Object.entries(turns)
    .filter(([, value]) => isRecord(value) && typeof value.startedAt === "number")
    .sort((a, b) => b[1].startedAt - a[1].startedAt)
    .slice(0, MAX_STORED_TURNS);
  return Object.fromEntries(entries);
}

export function isConfigured(config) {
  return Boolean(config.token && config.user);
}

export async function sendPushover(config, title, message, options = {}) {
  if (config.dryRun) return { dryRun: true };
  if (!isConfigured(config)) {
    throw new Error("Pushover is not configured. Set PUSHOVER_APP_TOKEN and PUSHOVER_USER_KEY.");
  }

  const body = new URLSearchParams({
    token: config.token,
    user: config.user,
    title: trimSummary(title, PUSHOVER_TITLE_MAX_LENGTH),
    message: trimMessageBody(message),
  });
  const device = firstNonEmpty(options.device, config.device);
  const sound = firstNonEmpty(options.sound, config.sound);
  if (device) body.set("device", device);
  if (sound) body.set("sound", sound);
  if (Number.isInteger(options.priority)) body.set("priority", String(options.priority));
  if (Number.isInteger(options.retry)) body.set("retry", String(options.retry));
  if (Number.isInteger(options.expire)) body.set("expire", String(options.expire));
  if (options.url) body.set("url", String(options.url));
  if (options.urlTitle) body.set("url_title", String(options.urlTitle));

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

  return { dryRun: false };
}
