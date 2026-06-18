const fs = require("fs");
const path = require("path");
const { spawn } = require("child_process");

function runCodexPhase({ repoRoot, runDir, prefix, options, schema, prompt, buildCodexArgs, commandName }) {
  const schemaFile = path.join(runDir, `${prefix}-schema.json`);
  const lastMessageFile = path.join(runDir, `${prefix}-last-message.json`);
  const jsonlFile = path.join(runDir, `${prefix}.jsonl`);
  const stderrFile = path.join(runDir, `${prefix}.stderr.log`);
  const transcriptFile = path.join(runDir, `${prefix}.transcript.log`);

  writeJson(schemaFile, schema);

  const args = buildCodexArgs({ options, schemaFile, lastMessageFile, prompt });
  console.error(`[${commandName}] starting ${prefix}`);

  return new Promise((resolve, reject) => {
    const stdout = fs.createWriteStream(jsonlFile);
    const stderr = fs.createWriteStream(stderrFile);
    const transcript = fs.createWriteStream(transcriptFile);
    const renderer = createPhaseRenderer({
      prefix,
      style: options.logStyle,
      transcript,
    });
    const child = spawn("codex", args, {
      cwd: repoRoot,
      env: {
        ...process.env,
        CODEX_PUSHOVER_DRY_RUN: "1",
      },
      stdio: ["ignore", "pipe", "pipe"],
    });

    child.stdout.on("data", (chunk) => {
      stdout.write(chunk);
      renderer.write(chunk);
    });

    child.stderr.on("data", (chunk) => {
      stderr.write(chunk);
    });

    child.on("error", reject);
    child.on("close", (code) => {
      renderer.end();
      Promise.all([finishStream(stdout), finishStream(stderr), finishStream(transcript)])
        .then(() => {
          if (code !== 0) {
            reject(new Error(`${prefix} failed with exit code ${code}; see ${stderrFile} and ${transcriptFile}`));
            return;
          }

          try {
            const raw = readFinalMessage(lastMessageFile, jsonlFile);
            const parsed = parseStructuredMessage(raw);
            writeJson(path.join(runDir, `${prefix}-parsed.json`), parsed);
            resolve(parsed);
          } catch (error) {
            reject(new Error(`${prefix} did not produce parseable structured output: ${error.message}`));
          }
        })
        .catch(reject);
    });
  });
}

function writeJson(file, value) {
  fs.writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function finishStream(stream) {
  return new Promise((resolve, reject) => {
    stream.on("error", reject);
    stream.end(resolve);
  });
}

function createPhaseRenderer({ prefix, style, transcript }) {
  let buffer = "";

  function emit(message) {
    const line = `[${prefix}] ${message}\n`;
    transcript.write(line);
    if (style === "live") {
      process.stderr.write(line);
    }
  }

  emit("phase started");

  return {
    write(chunk) {
      const text = chunk.toString("utf8");
      if (style === "jsonl") {
        process.stderr.write(text);
      }

      buffer += text;
      const lines = buffer.split(/\r?\n/);
      buffer = lines.pop() || "";

      for (const line of lines) {
        renderJsonLine(line, emit);
      }
    },
    end() {
      if (buffer.trim()) {
        renderJsonLine(buffer, emit);
      }
      emit("phase finished");
    },
  };
}

function renderJsonLine(line, emit) {
  if (!line.trim()) return;

  let event;
  try {
    event = JSON.parse(line);
  } catch {
    emit(`unparsed event: ${truncate(line.trim(), 500)}`);
    return;
  }

  if (event.type === "thread.started") {
    emit(`thread started: ${event.thread_id || "unknown"}`);
    return;
  }

  if (event.type === "turn.started") {
    emit("turn started");
    return;
  }

  if (event.type === "turn.completed") {
    const usage = event.usage ? ` ${JSON.stringify(event.usage)}` : "";
    emit(`turn completed${usage}`);
    return;
  }

  if (event.type === "turn.failed" || event.type === "error") {
    emit(`${event.type}: ${event.message || event.error || "unknown error"}`);
    return;
  }

  const item = event.item || {};
  if (!item.type) {
    emit(`${event.type || "event"}: ${truncate(JSON.stringify(event), 1000)}`);
    return;
  }

  if (event.type === "item.started" && item.type === "command_execution") {
    emit(`running: ${compactCommand(item.command || "")}`);
    return;
  }

  if (event.type === "item.completed" && item.type === "command_execution") {
    const exitText = item.exit_code === null || item.exit_code === undefined ? "" : ` exit ${item.exit_code}`;
    emit(`command ${item.status || "completed"}${exitText}: ${compactCommand(item.command || "")}`);
    if (item.aggregated_output && item.aggregated_output.trim()) {
      emitBlock("output", item.aggregated_output, emit, 1200);
    }
    return;
  }

  if (item.type === "agent_message") {
    emitBlock("agent", formatAgentMessage(item.text || ""), emit, 4000);
    return;
  }

  emit(`${item.type}: ${truncate(JSON.stringify(item), 1000)}`);
}

function compactCommand(command) {
  return String(command || "")
    .replace(/\s+/g, " ")
    .trim();
}

function formatAgentMessage(text) {
  const trimmed = String(text || "").trim();
  if (!trimmed.startsWith("{")) return trimmed;

  try {
    const value = JSON.parse(trimmed);
    if (!value || typeof value !== "object" || Array.isArray(value)) {
      return trimmed;
    }

    const lines = [];
    if (value.status || value.summary) {
      lines.push([value.status, value.summary].filter(Boolean).join(" - "));
    }
    if (typeof value.bucket1_applied_count === "number") {
      lines.push(`Bucket I applied: ${value.bucket1_applied_count}`);
    }
    if (Array.isArray(value.bucket1_candidates)) {
      lines.push(`Bucket I candidates: ${value.bucket1_candidates.length}`);
    }
    if (Array.isArray(value.bucket2_items)) {
      lines.push(`Bucket II: ${value.bucket2_items.length}`);
    }
    if (typeof value.applied_count === "number") {
      lines.push(`Applied: ${value.applied_count}`);
    }
    if (value.validation) {
      lines.push(`Validation: ${value.validation}`);
    }
    if (value.next_action) {
      lines.push(`Next: ${value.next_action}`);
    }
    return lines.length ? lines.join("\n") : trimmed;
  } catch {
    return trimmed;
  }
}

function emitBlock(label, text, emit, maxLength) {
  const rendered = truncate(String(text || "").trim(), maxLength);
  if (!rendered) return;
  const lines = rendered.split(/\r?\n/);
  emit(`${label}: ${lines[0]}`);
  for (const line of lines.slice(1)) {
    emit(`  ${line}`);
  }
}

function truncate(text, maxLength) {
  if (text.length <= maxLength) return text;
  return `${text.slice(0, maxLength).trimEnd()}... [truncated ${text.length - maxLength} chars]`;
}

function readFinalMessage(lastMessageFile, jsonlFile) {
  if (fs.existsSync(lastMessageFile)) {
    const raw = fs.readFileSync(lastMessageFile, "utf8").trim();
    if (raw) return raw;
  }

  if (!fs.existsSync(jsonlFile)) {
    throw new Error("missing final message and JSONL log");
  }

  const lines = fs.readFileSync(jsonlFile, "utf8").split(/\r?\n/);
  let lastAgentMessage = "";
  for (const line of lines) {
    if (!line.trim()) continue;
    try {
      const event = JSON.parse(line);
      const item = event.item;
      if (item && item.type === "agent_message" && typeof item.text === "string") {
        lastAgentMessage = item.text;
      }
    } catch {
      // Ignore malformed log lines; the final message file is authoritative.
    }
  }

  if (!lastAgentMessage) {
    throw new Error("no agent message found");
  }
  return lastAgentMessage.trim();
}

function parseStructuredMessage(raw) {
  try {
    return JSON.parse(raw);
  } catch {
    const fenced = raw.match(/```(?:json)?\s*([\s\S]*?)```/i);
    if (fenced) {
      return JSON.parse(fenced[1].trim());
    }
    throw new Error("final message was not valid JSON");
  }
}

module.exports = {
  runCodexPhase,
};
