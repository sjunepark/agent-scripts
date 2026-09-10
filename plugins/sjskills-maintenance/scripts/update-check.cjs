#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");
const DAY = 24 * 60 * 60 * 1000;
const RETRY = 15 * 60 * 1000;

function checkDirectory(dataRoot) {
  if (typeof dataRoot !== "string" || !path.isAbsolute(dataRoot)) throw new Error("missing absolute plugin data directory");
  for (let directory = dataRoot; ; directory = path.dirname(directory)) {
    try {
      const stat = fs.lstatSync(directory);
      if (!stat.isDirectory() || stat.isSymbolicLink()) throw new Error("unsafe plugin data directory");
    } catch (error) { if (error.code !== "ENOENT") throw error; }
    if (directory === path.dirname(directory)) break;
  }
}

function statePath(dataRoot) {
  checkDirectory(dataRoot);
  const file = path.join(dataRoot, "update-check.json");
  try {
    const stat = fs.lstatSync(file);
    if (!stat.isFile() || stat.isSymbolicLink() || stat.size > 4096) throw new Error("unsafe plugin update state");
  } catch (error) { if (error.code !== "ENOENT") throw error; }
  return file;
}

function updateStatus(dataRoot, now = Date.now()) {
  const file = statePath(dataRoot);
  let state;
  try { state = JSON.parse(fs.readFileSync(file, "utf8")); } catch (error) {
    if (error.code === "ENOENT" || error instanceof SyntaxError) return { due: true };
    throw error;
  }
  if (!state || !Number.isSafeInteger(state.checkedAt) || !["success", "failure"].includes(state.outcome)) return { due: true };
  const age = now - state.checkedAt;
  const interval = state.outcome === "success" ? DAY : RETRY;
  return { due: age < 0 || age >= interval, lastOutcome: state.outcome, checkedAt: state.checkedAt };
}

function recordCheck(dataRoot, outcome, now = Date.now()) {
  if (!["success", "failure"].includes(outcome)) throw new Error("expected success or failure");
  const file = statePath(dataRoot);
  fs.mkdirSync(dataRoot, { recursive: true, mode: 0o700 });
  const temporary = `${file}.${process.pid}.tmp`;
  let created = false;
  try {
    const fd = fs.openSync(temporary, "wx", 0o600);
    created = true;
    try { fs.writeFileSync(fd, JSON.stringify({ checkedAt: now, outcome }) + "\n"); } finally { fs.closeSync(fd); }
    fs.renameSync(temporary, file);
  } finally {
    if (created) {
      try { fs.unlinkSync(temporary); } catch (error) { if (error.code !== "ENOENT") throw error; }
    }
  }
}

if (require.main === module) {
  try {
    const [action, dataRoot, ...extra] = process.argv.slice(2);
    if (extra.length || !["status", "success", "failure"].includes(action)) throw new Error("usage: update-check.cjs status|success|failure PLUGIN_DATA");
    if (action === "status") console.log(JSON.stringify(updateStatus(dataRoot)));
    else recordCheck(dataRoot, action);
  } catch (error) {
    console.error(`sjskills plugin update check: ${error.message}`);
    process.exitCode = 1;
  }
}

module.exports = { updateStatus, recordCheck, checkDirectory };
