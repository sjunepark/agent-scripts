#!/usr/bin/env node
import { execFileSync } from "node:child_process";
import { createReadStream } from "node:fs";
import { mkdir, readFile, rename, writeFile } from "node:fs/promises";
import { homedir } from "node:os";
import { dirname, join } from "node:path";
import { createInterface } from "node:readline";

const THRESHOLDS = [50, 60, 70, 80];
const MAX_STATE_SESSIONS = 100;
const DEFAULT_STATE_DIR = join(homedir(), ".codex", "context-handoff");

function isRecord(value) {
  return Boolean(value && typeof value === "object" && !Array.isArray(value));
}

function firstNonEmpty(...values) {
  for (const value of values) {
    const trimmed = typeof value === "string" ? value.trim() : "";
    if (trimmed) return trimmed;
  }
  return undefined;
}

function pluginDataDir() {
  return firstNonEmpty(process.env.PLUGIN_DATA, process.env.CODEX_CONTEXT_HANDOFF_DATA_DIR) ?? DEFAULT_STATE_DIR;
}

function stateFilePath() {
  return firstNonEmpty(process.env.CODEX_CONTEXT_HANDOFF_STATE_FILE) ?? join(pluginDataDir(), "state.json");
}

async function readStdin() {
  const chunks = [];
  for await (const chunk of process.stdin) chunks.push(chunk);
  return Buffer.concat(chunks).toString("utf8");
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
    sessions: isRecord(state.sessions) ? state.sessions : {},
  };
}

async function writeState(state) {
  const entries = Object.entries(state.sessions)
    .filter(([, value]) => isRecord(value))
    .sort((a, b) => Number(b[1].updatedAt ?? 0) - Number(a[1].updatedAt ?? 0))
    .slice(0, MAX_STATE_SESSIONS);
  await writeJsonFile(stateFilePath(), { sessions: Object.fromEntries(entries) });
}

function sessionKey(input) {
  return firstNonEmpty(input.session_id, process.env.CODEX_THREAD_ID) ?? "unknown-session";
}

function currentSessionState(state, key) {
  const session = isRecord(state.sessions[key]) ? state.sessions[key] : {};
  return {
    policyEmitted: session.policyEmitted === true,
    ignoredUsageKey: typeof session.ignoredUsageKey === "string" ? session.ignoredUsageKey : undefined,
    emittedBuckets: Array.isArray(session.emittedBuckets)
      ? session.emittedBuckets.filter((value) => Number.isInteger(value))
      : [],
    stopContinuationTurns: isRecord(session.stopContinuationTurns) ? session.stopContinuationTurns : {},
    updatedAt: Date.now(),
  };
}

function hookOutput(fields = {}) {
  process.stdout.write(`${JSON.stringify({ continue: true, ...fields })}\n`);
}

function hookContextOutput(eventName, additionalContext) {
  hookOutput({
    hookSpecificOutput: {
      hookEventName: eventName,
      additionalContext,
    },
  });
}

function stopContinuationOutput(reason) {
  process.stdout.write(`${JSON.stringify({ decision: "block", reason })}\n`);
}

function asNumber(value) {
  return typeof value === "number" && Number.isFinite(value) ? value : undefined;
}

function tokenValue(breakdown, snakeName, camelName) {
  if (!isRecord(breakdown)) return undefined;
  return asNumber(breakdown[snakeName]) ?? asNumber(breakdown[camelName]);
}

function normalizeTokenInfo(info) {
  if (!isRecord(info)) return undefined;
  const window = asNumber(info.model_context_window) ?? asNumber(info.modelContextWindow);
  const last = isRecord(info.last_token_usage)
    ? info.last_token_usage
    : isRecord(info.lastTokenUsage)
      ? info.lastTokenUsage
      : info.last;
  if (!window || !isRecord(last)) return undefined;
  const totalTokens = tokenValue(last, "total_tokens", "totalTokens");
  const inputTokens = tokenValue(last, "input_tokens", "inputTokens");
  if (!totalTokens && !inputTokens) return undefined;
  const usedTokens = totalTokens ?? inputTokens;
  return {
    modelContextWindow: window,
    inputTokens,
    totalTokens,
    usedTokens,
    percent: (usedTokens / window) * 100,
  };
}

async function readLatestTokenUsage(transcriptPath) {
  if (!transcriptPath) return undefined;
  let latest;
  let lineNumber = 0;
  try {
    const rl = createInterface({
      input: createReadStream(transcriptPath, { encoding: "utf8" }),
      crlfDelay: Infinity,
    });

    for await (const line of rl) {
      lineNumber += 1;
      if (!line.includes("token_count") && !line.includes("tokenUsage") && !line.includes("model_context_window")) {
        continue;
      }
      try {
        const event = JSON.parse(line);
        const payload = event.payload;
        if (!isRecord(payload)) continue;
        const usage =
          payload.type === "token_count"
            ? normalizeTokenInfo(payload.info)
            : payload.tokenUsage
              ? normalizeTokenInfo(payload.tokenUsage)
              : undefined;
        if (usage) latest = { ...usage, lineNumber };
      } catch {
        // Ignore unstable transcript records that are not JSONL.
      }
    }
  } catch {
    return latest;
  }
  return latest;
}

function usageKey(usage) {
  return [
    usage.lineNumber ?? "?",
    usage.modelContextWindow,
    usage.inputTokens ?? "",
    usage.totalTokens ?? "",
    usage.usedTokens,
  ].join(":");
}

function sqlString(value) {
  return `'${String(value).replaceAll("'", "''")}'`;
}

function readThreadMetadataFromState(sessionId) {
  const codexHome = firstNonEmpty(process.env.CODEX_HOME) ?? join(homedir(), ".codex");
  const dbPath = firstNonEmpty(process.env.CODEX_CONTEXT_HANDOFF_DB) ?? join(codexHome, "state_5.sqlite");
  try {
    const sql = [
      "select model, reasoning_effort, git_branch",
      "from threads",
      `where id = ${sqlString(sessionId)}`,
      "limit 1",
    ].join(" ");
    const raw = execFileSync("sqlite3", ["-json", dbPath, sql], {
      encoding: "utf8",
      timeout: 1000,
      stdio: ["ignore", "pipe", "ignore"],
    });
    const rows = JSON.parse(raw || "[]");
    return isRecord(rows[0]) ? rows[0] : {};
  } catch {
    return {};
  }
}

function gitBranch(cwd) {
  if (!cwd) return undefined;
  try {
    const branch = execFileSync("git", ["-C", cwd, "rev-parse", "--abbrev-ref", "HEAD"], {
      encoding: "utf8",
      timeout: 1000,
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    return branch || undefined;
  } catch {
    return undefined;
  }
}

function metadata(input, sessionId) {
  const row = readThreadMetadataFromState(sessionId);
  const cwd = firstNonEmpty(input.cwd);
  return {
    cwd: cwd ?? "unknown",
    branch: firstNonEmpty(row.git_branch, gitBranch(cwd)) ?? "unknown",
    model: firstNonEmpty(input.model, row.model) ?? "unknown",
    reasoningEffort: firstNonEmpty(input.reasoning_effort, input.reasoningEffort, row.reasoning_effort) ?? "unknown",
  };
}

function formatPercent(value) {
  return `${Math.round(value * 10) / 10}%`;
}

function formatNumber(value) {
  return typeof value === "number" ? value.toLocaleString("en-US") : "unknown";
}

function usageLine(usage) {
  return `Context usage: ${formatPercent(usage.percent)} (${formatNumber(usage.usedTokens)} / ${formatNumber(usage.modelContextWindow)} tokens from the latest token-count event).`;
}

function bucketFor(percent) {
  let bucket;
  for (const threshold of THRESHOLDS) {
    if (percent >= threshold) bucket = threshold;
  }
  return bucket;
}

function severityMessage(bucket, includePolicy) {
  if (bucket >= 80) {
    return "Do not continue substantial work in this thread. Create or update the handoff and start a fresh thread if possible. If thread creation is unavailable, stop with the handoff path and the exact fresh-thread prompt.";
  }
  if (bucket >= 70) {
    return "Strong warning: prepare migration to a fresh thread now. Finish only the current atomic step, create or update the handoff, and draft or spawn the fresh-thread continuation using the same cwd, branch, model, and reasoning effort.";
  }
  if (bucket >= 60) {
    return includePolicy
      ? "Handoff preparation is strongly recommended before expanding scope. Keep work focused and use the standard context handoff policy below."
      : "Handoff preparation is strongly recommended before expanding scope. Keep work focused and follow the standard context handoff policy already provided.";
  }
  return includePolicy
    ? "Context pressure is becoming relevant. Conserve context and consider preparing a handoff if this task may continue for several more turns. Use the standard context handoff policy below."
    : "Context pressure is becoming relevant. Conserve context and consider preparing a handoff if this task may continue for several more turns. Follow the standard context handoff policy already provided.";
}

function fullPolicy(meta) {
  return [
    "Standard context handoff policy is now active.",
    "",
    "Create or update a concise repo-local handoff before this thread becomes too full. Inspect repository conventions first: plans/, docs/plans/, .pi/plans/, existing PLAN*.md, or similar handoff files. If no convention exists, use PLAN-<SPECIFIC-SLUG>.md at the repository root.",
    "",
    "The handoff must be implementation-ready, not a transcript. Include:",
    "- Purpose: what is being done and why.",
    "- Current state: confirmed decisions, constraints, changed files, completed work, blockers, assumptions, and open questions.",
    "- Next implementation slice: small reviewable checklist.",
    "- Files to inspect first: max 3-5 paths or areas, each with a reason.",
    "- Validation: relevant commands, results, blockers, or manual QA notes.",
    "- Fresh-thread prompt: exact prompt to continue in a new thread.",
    "",
    "Keep it compact. Prefer latest actionable state over historical logs. Reference existing plans, issues, ADRs, commits, diffs, and docs by path or URL instead of duplicating them. Do not paste full command output. Redact secrets and private data.",
    "",
    "The fresh-thread prompt must tell the next thread to use the same working directory, branch, model, and reasoning effort:",
    `- cwd: ${meta.cwd}`,
    `- branch: ${meta.branch}`,
    `- model: ${meta.model}`,
    `- reasoning_effort: ${meta.reasoningEffort}`,
    "",
    "The next thread should read the handoff first, check git status, avoid broad re-scanning, preserve confirmed decisions unless files contradict them, and continue from the listed next implementation slice.",
  ].join("\n");
}

function contextMessage({ usage, bucket, includePolicy, meta }) {
  const parts = [usageLine(usage), "", severityMessage(bucket, includePolicy)];
  if (includePolicy) {
    parts.push("", fullPolicy(meta));
  }
  return parts.join("\n");
}

function markBucketsThrough(session, bucket) {
  const emitted = new Set(session.emittedBuckets);
  for (const threshold of THRESHOLDS) {
    if (threshold <= bucket) emitted.add(threshold);
  }
  session.emittedBuckets = [...emitted].sort((a, b) => a - b);
}

async function resetSessionState(input) {
  const latestUsage = await readLatestTokenUsage(input.transcript_path);
  const state = await readState();
  const key = sessionKey(input);
  const session = currentSessionState(state, key);
  session.policyEmitted = false;
  session.ignoredUsageKey = latestUsage ? usageKey(latestUsage) : undefined;
  session.emittedBuckets = [];
  session.stopContinuationTurns = {};
  session.updatedAt = Date.now();
  state.sessions[key] = session;
  await writeState(state);
}

async function maybeInjectThreshold(input) {
  const usage = await readLatestTokenUsage(input.transcript_path);
  if (!usage) {
    hookOutput();
    return;
  }

  const bucket = bucketFor(usage.percent);
  if (!bucket) {
    hookOutput();
    return;
  }

  const state = await readState();
  const key = sessionKey(input);
  const session = currentSessionState(state, key);
  if (session.ignoredUsageKey === usageKey(usage)) {
    hookOutput();
    return;
  }
  if (session.emittedBuckets.includes(bucket)) {
    hookOutput();
    return;
  }

  const meta = metadata(input, key);
  const includePolicy = !session.policyEmitted;
  const message = contextMessage({ usage, bucket, includePolicy, meta });
  session.policyEmitted = true;
  session.ignoredUsageKey = undefined;
  markBucketsThrough(session, bucket);
  session.updatedAt = Date.now();
  state.sessions[key] = session;
  await writeState(state);
  hookContextOutput(input.hook_event_name, message);
}

async function maybeContinueForStop(input) {
  const usage = await readLatestTokenUsage(input.transcript_path);
  const bucket = usage ? bucketFor(usage.percent) : undefined;
  if (!usage || !bucket || bucket < 70 || input.stop_hook_active === true) {
    hookOutput();
    return;
  }

  const state = await readState();
  const key = sessionKey(input);
  const session = currentSessionState(state, key);
  if (session.ignoredUsageKey === usageKey(usage)) {
    hookOutput();
    return;
  }
  const turnKey = firstNonEmpty(input.turn_id) ?? "unknown-turn";
  if (session.stopContinuationTurns[turnKey] === true) {
    hookOutput();
    return;
  }

  const meta = metadata(input, key);
  const includePolicy = !session.policyEmitted;
  const message = contextMessage({ usage, bucket, includePolicy, meta });
  session.policyEmitted = true;
  session.ignoredUsageKey = undefined;
  markBucketsThrough(session, bucket);
  session.stopContinuationTurns[turnKey] = true;
  session.updatedAt = Date.now();
  state.sessions[key] = session;
  await writeState(state);
  stopContinuationOutput(message);
}

async function handleHook() {
  if (process.env.CODEX_CONTEXT_HANDOFF_DISABLED === "1") {
    hookOutput();
    return;
  }

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

  if (input.hook_event_name === "PostCompact") {
    await resetSessionState(input);
    hookOutput();
    return;
  }

  if (input.hook_event_name === "SessionStart") {
    const source = firstNonEmpty(input.source);
    if (source === "clear" || source === "compact") {
      await resetSessionState(input);
      hookOutput();
      return;
    }
    await maybeInjectThreshold(input);
    return;
  }

  if (input.hook_event_name === "UserPromptSubmit" || input.hook_event_name === "PostToolUse") {
    await maybeInjectThreshold(input);
    return;
  }

  if (input.hook_event_name === "Stop") {
    await maybeContinueForStop(input);
    return;
  }

  hookOutput();
}

async function handleCli(args) {
  const action = args[0] ?? "status";
  const state = await readState();
  if (action === "status") {
    console.log(`Context handoff state: ${stateFilePath()}`);
    console.log(`Tracked sessions: ${Object.keys(state.sessions).length}`);
    return;
  }
  if (action === "reset") {
    await writeJsonFile(stateFilePath(), { sessions: {} });
    console.log("Context handoff state reset.");
    return;
  }
  console.error("Usage: context-handoff.mjs [status|reset]");
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
    hookOutput(process.env.CODEX_CONTEXT_HANDOFF_VERBOSE === "1" ? { systemMessage: `context-handoff: ${message}` } : {});
  });
}
