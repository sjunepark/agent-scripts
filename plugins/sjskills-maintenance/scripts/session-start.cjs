#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");
const { createHash } = require("node:crypto");
const { updateStatus, checkDirectory } = require("./update-check.cjs");

function snapshotResources(root, dataRoot) {
  const paths = ["scripts/update-check.cjs", "references/sjskills/SKILL.md", "references/sjskills/references/global-rollout.md"];
  const files = paths.map((relative) => ({ relative, bytes: fs.readFileSync(path.join(root, relative)) }));
  const hash = createHash("sha256");
  for (const { relative, bytes } of files) hash.update(relative).update("\0").update(bytes).update("\0");
  const snapshot = path.join(dataRoot, "workflows", hash.digest("hex"));
  // Codex deletes old plugin caches on reinstall. Content-addressed references
  // in PLUGIN_DATA remain valid for active and suspended sessions.
  for (const { relative, bytes } of files) {
    const destination = path.join(snapshot, relative);
    checkDirectory(path.dirname(destination));
    try {
      const stat = fs.lstatSync(destination);
      if (!stat.isFile() || stat.isSymbolicLink() || !fs.readFileSync(destination).equals(bytes)) throw new Error("unsafe or modified workflow snapshot");
      continue;
    } catch (error) { if (error.code !== "ENOENT") throw error; }
    fs.mkdirSync(path.dirname(destination), { recursive: true, mode: 0o700 });
    const temporary = `${destination}.${process.pid}.tmp`;
    let created = false;
    try {
      const fd = fs.openSync(temporary, "wx", 0o600);
      created = true;
      try { fs.writeFileSync(fd, bytes); } finally { fs.closeSync(fd); }
      fs.renameSync(temporary, destination);
    } finally {
      if (created) {
        try { fs.unlinkSync(temporary); } catch (error) { if (error.code !== "ENOENT") throw error; }
      }
    }
  }
  return snapshot;
}

function sessionContext(input, platform = process.platform, arch = process.arch,
  root = path.resolve(__dirname, ".."), dataRoot = process.env.PLUGIN_DATA || process.env.CLAUDE_PLUGIN_DATA) {
  if (input.hook_event_name !== "SessionStart" || !["startup", "resume"].includes(input.source)) return null;
  const supported = (platform === "darwin" && ["x64", "arm64"].includes(arch)) ||
    (platform === "win32" && arch === "x64");
  if (!supported) {
    return { systemMessage: `sjskills startup maintenance skipped: ${platform}/${arch} is not a supported sjskills release target.` };
  }
  const instructions = fs.readFileSync(path.join(root, "references", "maintenance.md"), "utf8");
  let pluginUpdate;
  try { pluginUpdate = updateStatus(dataRoot); } catch (error) {
    pluginUpdate = { due: false, error: error.message };
  }
  let resourcesRoot = root;
  if (!pluginUpdate.error) {
    try { resourcesRoot = snapshotResources(root, dataRoot); } catch (error) {
      pluginUpdate = { due: false, error: `workflow snapshot unavailable: ${error.message}` };
    }
  }
  // Paths are data, not shell fragments. The agent uses its ordinary file tools.
  const resources = JSON.stringify({
    sessionDirectory: input.cwd,
    syncSkill: path.join(resourcesRoot, "references", "sjskills", "SKILL.md"),
    pluginRoot: root,
    pluginData: dataRoot,
    updateHelper: path.join(resourcesRoot, "scripts", "update-check.cjs"),
    pluginUpdate,
  });
  return {
    hookSpecificOutput: {
      hookEventName: "SessionStart",
      additionalContext: `${instructions}\n\nLocal resource paths (JSON data):\n${resources}`,
    },
  };
}

async function run() {
  let input = "";
  for await (const chunk of process.stdin) {
    input += chunk;
    if (Buffer.byteLength(input) > 1024 * 1024) throw new Error("hook input exceeds 1 MiB");
  }
  const payload = JSON.parse(input);
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) throw new Error("invalid hook input");
  if (typeof payload.cwd !== "string" || !path.isAbsolute(payload.cwd)) throw new Error("missing absolute session directory");
  const result = sessionContext(payload);
  if (result) process.stdout.write(`${JSON.stringify(result)}\n`);
}

async function main() {
  try { await run(); } catch (error) {
    // Maintenance must not prevent the user's session from starting.
    process.stdout.write(`${JSON.stringify({ systemMessage: `sjskills startup maintenance unavailable: ${error.message}` })}\n`);
  }
}

if (require.main === module) main();
module.exports = { sessionContext, snapshotResources, main };
