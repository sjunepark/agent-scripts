#!/usr/bin/env node
"use strict";
const path = require("node:path");
const { nativeExecutable, runProcess, LIMIT } = require("./process.cjs");
const { classifyStatus, report } = require("./status.cjs");
const { observePlugin } = require("./observation.cjs");

async function check(input, options = {}) {
  if (input?.hook_event_name !== "SessionStart" || !["startup", "resume"].includes(input.source)) return null;
  const platform = options.platform || process.platform, arch = options.arch || process.arch;
  if (!((platform === "darwin" && ["x64", "arm64"].includes(arch)) || (platform === "win32" && arch === "x64"))) {
    return { systemMessage: "sjskills: check skipped (unsupported release target)." };
  }
  if (typeof input.cwd !== "string" || !path.isAbsolute(input.cwd)) return report([{ findings: [], incomplete: ["session directory"] }]);
  const env = options.env || process.env, signal = options.signal;
  const run = options.run || runProcess, resolve = options.resolve || nativeExecutable;
  const cli = async () => {
    try {
      const result = await run(resolve("sjskills", env), ["--json", "status"], { cwd: input.cwd, env, signal });
      const status = classifyStatus(JSON.parse(result.stdout), options.now);
      if (result.code !== 0) status.incomplete.push("status process");
      return status;
    } catch { return { findings: [], incomplete: ["native CLI status"] }; }
  };
  const plugin = async () => {
    try {
      return await (options.observe || observePlugin)({ root: path.resolve(__dirname, ".."),
        dataRoot: env.PLUGIN_DATA || env.CLAUDE_PLUGIN_DATA, env, signal, now: options.now });
    } catch { return { plugin: true, findings: [], incomplete: ["plugin verification"] }; }
  };
  return report(await Promise.all([cli(), plugin()]));
}

async function main() {
  const controller = new AbortController();
  const cancel = () => controller.abort();
  const timer = setTimeout(cancel, 35000);
  process.once("SIGINT", cancel); process.once("SIGTERM", cancel);
  let output;
  try {
    const chunks = []; let size = 0;
    const input = await new Promise((resolve, reject) => {
      const abort = () => { process.stdin.destroy(); reject(new Error("input cancelled")); };
      controller.signal.addEventListener("abort", abort, { once: true });
      process.stdin.on("data", (chunk) => {
        size += chunk.length;
        if (size > LIMIT) abort(); else chunks.push(chunk);
      });
      process.stdin.once("error", reject);
      process.stdin.once("end", () => {
        controller.signal.removeEventListener("abort", abort);
        resolve(Buffer.concat(chunks).toString("utf8"));
      });
    });
    const payload = JSON.parse(input);
    if (!payload || typeof payload !== "object" || Array.isArray(payload)) throw new Error("invalid input");
    output = await check(payload, { signal: controller.signal });
  } catch { output = report([{ findings: [], incomplete: ["hook input or deadline"] }]); }
  finally {
    clearTimeout(timer);
    process.removeListener("SIGINT", cancel); process.removeListener("SIGTERM", cancel);
  }
  if (output) process.stdout.write(JSON.stringify(output) + "\n");
}
if (require.main === module) main();
module.exports = { check, main };
