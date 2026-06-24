#!/usr/bin/env node
import {
  firstNonEmpty,
  isConfigured,
  isRecord,
  parsePositiveInt,
  readConfig,
  readState,
  sendPushover,
  stateFilePath,
  trimMessageBody,
  trimSummary,
} from "./pushover-client.mjs";

const SERVER_NAME = "codex-pushover-notify";
const SERVER_VERSION = "0.1.0";
const DEFAULT_PROTOCOL_VERSION = "2024-11-05";
const INSTRUCTIONS =
  "Use Pushover only when the user needs material attention. Use pushover_request_decision for non-obvious judgment, risky tradeoffs, or real blockers. Continue automatically for obvious safe work, routine progress, and minor findings.";

const tools = [
  {
    name: "pushover_notify",
    description: "Send a one-way Pushover notification to the user for material status or attention.",
    inputSchema: {
      type: "object",
      additionalProperties: false,
      properties: {
        title: {
          type: "string",
          description: "Notification title. Defaults to a Codex title.",
        },
        message: {
          type: "string",
          description: "Notification body to send.",
        },
        priority: {
          type: "integer",
          enum: [-2, -1, 0, 1],
          description: "Pushover priority for non-emergency notifications.",
          default: 0,
        },
        sound: {
          type: "string",
          description: "Optional Pushover sound override.",
        },
        device: {
          type: "string",
          description: "Optional Pushover device override.",
        },
        url: {
          type: "string",
          description: "Optional URL attached to the notification.",
        },
        url_title: {
          type: "string",
          description: "Optional title for the attached URL.",
        },
      },
      required: ["message"],
    },
  },
  {
    name: "pushover_request_decision",
    description: "Notify the user that Codex needs a decision, then wait for the user to answer in Codex.",
    inputSchema: {
      type: "object",
      additionalProperties: false,
      properties: {
        title: {
          type: "string",
          description: "Decision notification title. Defaults to 'Codex needs a decision'.",
        },
        question: {
          type: "string",
          description: "The concrete decision Codex needs from the user.",
        },
        context: {
          type: "string",
          description: "Brief context explaining why Codex cannot safely decide alone.",
        },
        options: {
          type: "array",
          items: { type: "string" },
          description: "Optional concise options for the user to choose from.",
        },
        urgency: {
          type: "string",
          enum: ["normal", "high", "emergency"],
          description: "Use emergency only for work that should repeatedly alert until acknowledged.",
          default: "high",
        },
        emergency_retry_seconds: {
          type: "integer",
          minimum: 30,
          description: "Emergency repeat interval. Pushover requires at least 30 seconds.",
          default: 60,
        },
        emergency_expire_seconds: {
          type: "integer",
          minimum: 30,
          maximum: 10800,
          description: "Emergency repeat window before giving up.",
          default: 3600,
        },
        url: {
          type: "string",
          description: "Optional URL attached to the notification.",
        },
        url_title: {
          type: "string",
          description: "Optional title for the attached URL.",
        },
      },
      required: ["question"],
    },
  },
  {
    name: "pushover_status",
    description: "Report whether Pushover is configured for this Codex session without exposing secrets.",
    inputSchema: {
      type: "object",
      additionalProperties: false,
      properties: {},
    },
  },
];

function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value));
}

function textContent(text) {
  return {
    content: [
      {
        type: "text",
        text,
      },
    ],
  };
}

function toolError(message) {
  return {
    isError: true,
    content: [
      {
        type: "text",
        text: message,
      },
    ],
  };
}

function normalizeArgs(value) {
  return isRecord(value) ? value : {};
}

async function handleNotify(rawArgs) {
  const args = normalizeArgs(rawArgs);
  const message = firstNonEmpty(args.message);
  if (!message) return toolError("message is required.");

  const config = readConfig();
  const title = firstNonEmpty(args.title) ?? "Codex notification";
  const priority = Number.isInteger(args.priority) && args.priority >= -2 && args.priority <= 1 ? args.priority : 0;
  await sendPushover(config, title, message, {
    priority,
    sound: firstNonEmpty(args.sound),
    device: firstNonEmpty(args.device),
    url: firstNonEmpty(args.url),
    urlTitle: firstNonEmpty(args.url_title),
  });

  return textContent(config.dryRun ? "Dry run completed; no Pushover notification was sent." : "Pushover notification sent.");
}

function formatDecisionMessage(args) {
  const lines = [trimMessageBody(args.question)];
  const context = firstNonEmpty(args.context);
  if (context) lines.push("", trimMessageBody(context, 700));
  if (Array.isArray(args.options) && args.options.length > 0) {
    const options = args.options
      .filter((option) => typeof option === "string" && option.trim())
      .slice(0, 6)
      .map((option, index) => `${index + 1}. ${trimSummary(option, 140)}`);
    if (options.length > 0) lines.push("", "Options:", ...options);
  }
  lines.push("", "Please answer in the Codex thread.");
  return trimMessageBody(lines.join("\n"));
}

async function handleRequestDecision(rawArgs) {
  const args = normalizeArgs(rawArgs);
  const question = firstNonEmpty(args.question);
  if (!question) return toolError("question is required.");

  const urgency = ["normal", "high", "emergency"].includes(args.urgency) ? args.urgency : "high";
  const priority = urgency === "emergency" ? 2 : urgency === "high" ? 1 : 0;
  const options = {
    priority,
    url: firstNonEmpty(args.url),
    urlTitle: firstNonEmpty(args.url_title),
  };
  if (priority === 2) {
    options.retry = clamp(parsePositiveInt(args.emergency_retry_seconds, 60), 30, 86400);
    options.expire = clamp(parsePositiveInt(args.emergency_expire_seconds, 3600), 30, 10800);
  }

  const config = readConfig();
  const title = firstNonEmpty(args.title) ?? "Codex needs a decision";
  await sendPushover(config, title, formatDecisionMessage({ ...args, question }), options);
  const sent = config.dryRun ? "Dry run completed; no decision notification was sent." : "Decision notification sent.";
  return textContent(`${sent} Wait for the user to answer in the Codex thread before making this decision.`);
}

async function handleStatus() {
  const config = readConfig();
  const state = await readState();
  return {
    content: [
      {
        type: "text",
        text: [
          `Configured: ${isConfigured(config) ? "yes" : "no"}`,
          `Dry run: ${config.dryRun ? "yes" : "no"}`,
          `Hook notifications: ${state.enabled ? "on" : "off"}`,
          `Device override: ${config.device ? "set" : "not set"}`,
          `Sound override: ${config.sound ? "set" : "not set"}`,
          `State file: ${stateFilePath()}`,
        ].join("\n"),
      },
    ],
  };
}

async function callTool(params) {
  const name = params?.name;
  const args = params?.arguments;
  if (name === "pushover_notify") return handleNotify(args);
  if (name === "pushover_request_decision") return handleRequestDecision(args);
  if (name === "pushover_status") return handleStatus();
  return toolError(`Unknown tool: ${name}`);
}

async function handleRequest(message) {
  switch (message.method) {
    case "initialize":
      return {
        protocolVersion: message.params?.protocolVersion ?? DEFAULT_PROTOCOL_VERSION,
        capabilities: { tools: {} },
        serverInfo: {
          name: SERVER_NAME,
          version: SERVER_VERSION,
        },
        instructions: INSTRUCTIONS,
      };
    case "tools/list":
      return { tools };
    case "tools/call":
      return callTool(message.params);
    case "ping":
      return {};
    default:
      throw Object.assign(new Error(`Method not found: ${message.method}`), {
        code: -32601,
      });
  }
}

function writeMessage(message) {
  process.stdout.write(`${JSON.stringify(message)}\n`);
}

function respond(id, result) {
  writeMessage({ jsonrpc: "2.0", id, result });
}

function respondError(id, error) {
  writeMessage({
    jsonrpc: "2.0",
    id,
    error: {
      code: error?.code ?? -32603,
      message: error instanceof Error ? error.message : String(error),
    },
  });
}

let inputBuffer = Buffer.alloc(0);

function processMessageBody(body) {
  const message = JSON.parse(body);
  if (!isRecord(message) || message.jsonrpc !== "2.0") return;
  if (message.id === undefined) return;

  Promise.resolve(handleRequest(message))
    .then((result) => respond(message.id, result))
    .catch((error) => respondError(message.id, error));
}

function processContentLengthBuffer() {
  while (true) {
    const headerEnd = inputBuffer.indexOf(Buffer.from("\r\n\r\n"));
    if (headerEnd === -1) return;

    const header = inputBuffer.subarray(0, headerEnd).toString("ascii");
    const match = /(?:^|\r\n)Content-Length:\s*(\d+)/i.exec(header);
    if (!match) {
      inputBuffer = Buffer.alloc(0);
      throw new Error("Missing Content-Length header.");
    }

    const length = Number.parseInt(match[1], 10);
    const bodyStart = headerEnd + 4;
    const bodyEnd = bodyStart + length;
    if (inputBuffer.length < bodyEnd) return;

    const body = inputBuffer.subarray(bodyStart, bodyEnd).toString("utf8");
    inputBuffer = inputBuffer.subarray(bodyEnd);
    processMessageBody(body);
  }
}

function processLineBuffer() {
  while (true) {
    const newline = inputBuffer.indexOf(0x0a);
    if (newline === -1) return;
    const line = inputBuffer.subarray(0, newline).toString("utf8").trim();
    inputBuffer = inputBuffer.subarray(newline + 1);
    if (line) processMessageBody(line);
  }
}

process.stdin.on("data", (chunk) => {
  inputBuffer = Buffer.concat([inputBuffer, chunk]);
  try {
    if (inputBuffer.includes(Buffer.from("Content-Length:"))) {
      processContentLengthBuffer();
    } else {
      processLineBuffer();
    }
  } catch (error) {
    respondError(null, error);
  }
});

process.stdin.resume();
