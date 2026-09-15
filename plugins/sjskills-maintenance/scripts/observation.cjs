"use strict";
const fs = require("node:fs"), path = require("node:path"), os = require("node:os");
const { randomUUID } = require("node:crypto");
const { nativeExecutable, runProcess } = require("./process.cjs");
const { DAY } = require("./status.cjs");
const SOURCE = "https://github.com/sjunepark/agent-scripts.git";
const RETRY = 15 * 60 * 1000;
const unavailable = () => ({ plugin: true, findings: [], incomplete: ["plugin verification"] });
const validVersion = (v) => typeof v === "string" && /^\d+\.\d+\.\d+\+codex\.[A-Za-z0-9.-]{1,100}$/.test(v);

function safeDirectory(directory) {
  if (typeof directory !== "string" || !path.isAbsolute(directory)) throw new Error("unsafe directory");
  for (let current = directory; ; current = path.dirname(current)) {
    try {
      const stat = fs.lstatSync(current);
      if (!stat.isDirectory() || stat.isSymbolicLink()) throw new Error("unsafe directory");
    } catch (error) { if (error.code !== "ENOENT") throw error; }
    if (current === path.dirname(current)) break;
  }
}
function readBounded(file, limit) {
  const stat = fs.lstatSync(file);
  if (!stat.isFile() || stat.isSymbolicLink() || stat.nlink !== 1 || stat.size > limit) throw new Error("unsafe file");
  const fd = fs.openSync(file, fs.constants.O_RDONLY | (fs.constants.O_NOFOLLOW || 0));
  try {
    const opened = fs.fstatSync(fd);
    if (opened.dev !== stat.dev || opened.ino !== stat.ino) throw new Error("file changed");
    const bytes = Buffer.alloc(limit + 1);
    const size = fs.readSync(fd, bytes, 0, bytes.length, 0);
    if (size > limit) throw new Error("oversized file");
    return bytes.subarray(0, size).toString("utf8");
  } finally { fs.closeSync(fd); }
}
function mainRef(config) {
  // Codex validates the complete config during list. Recognize only its ordinary
  // serialized marketplace table; unfamiliar forms fail closed.
  if (config.includes('"""') || config.includes("'''")) return false;
  let active = false, count = 0;
  const fields = {};
  for (const line of config.split(/\r?\n/)) {
    if (/^\s*\[/.test(line)) {
      active = /^\s*\[marketplaces\.personal\]\s*(?:#.*)?$/.test(line);
      if (active) count++;
    } else if (active) {
      const match = /^\s*(source_type|source|ref)\s*=\s*(?:"([^"\\]*)"|'([^']*)')\s*(?:#.*)?$/.exec(line);
      if (match) {
        if (Object.hasOwn(fields, match[1])) return false;
        fields[match[1]] = match[2] ?? match[3];
      }
    }
  }
  return count === 1 && fields.source_type === "git" && fields.source === SOURCE && fields.ref === "main";
}
async function fetchJSON(url, signal) {
  const response = await fetch(url, { signal, redirect: "error", headers: { "User-Agent": "sjskills-startup-check", Accept: "application/json" } });
  if (!response.ok || !response.body) throw new Error("metadata unavailable");
  const reader = response.body.getReader();
  const chunks = []; let size = 0;
  try {
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      size += value.byteLength;
      if (size > 128 * 1024) throw new Error("metadata too large");
      chunks.push(Buffer.from(value));
    }
    return JSON.parse(Buffer.concat(chunks).toString("utf8"));
  } finally { await reader.cancel(); }
}
async function provenance({ root, env, signal, run = runProcess, resolve = nativeExecutable }) {
  const codexHome = env.CODEX_HOME || path.join(os.homedir(), ".codex");
  if (!path.isAbsolute(codexHome)) throw new Error("configuration unavailable");
  const listing = await run(resolve("codex", env), ["plugin", "list", "--marketplace", "personal", "--available", "--json"],
    { cwd: codexHome, env, signal });
  if (listing.code !== 0) throw new Error("configuration unavailable");
  const installed = JSON.parse(listing.stdout).installed;
  if (!Array.isArray(installed)) throw new Error("configuration unavailable");
  const matches = installed.filter((p) => p.pluginId === "sjskills-maintenance@personal");
  const entry = matches[0];
  const manifest = JSON.parse(readBounded(path.join(root, ".codex-plugin/plugin.json"), 16384));
  if (matches.length !== 1 || !entry.enabled || !entry.installed ||
      entry.marketplaceSource?.sourceType !== "git" || entry.marketplaceSource.source !== SOURCE ||
      manifest.name !== "sjskills-maintenance" || !validVersion(manifest.version) || entry.version !== manifest.version ||
      !mainRef(readBounded(path.join(codexHome, "config.toml"), 1024 * 1024))) throw new Error("configuration unavailable");
  return manifest.version;
}
async function observePlugin(options) {
  const { dataRoot, signal, now = Date.now(), get = fetchJSON } = options;
  let version;
  try { version = await (options.provenance || provenance)(options); }
  catch { return unavailable(); }
  const key = `${SOURCE}@main:${version}`;
  const result = (published) => ({ plugin: true, findings: published === version ? [] : ["plugin update available"], incomplete: [] });
  let file, lock, owned = false, writable = false;
  try {
    safeDirectory(dataRoot);
    file = path.join(dataRoot, "plugin-observation.json");
    let cached;
    try { cached = JSON.parse(readBounded(file, 4096)); }
    catch (error) { if (error.code !== "ENOENT" && !(error instanceof SyntaxError)) return unavailable(); }
    if (cached?.key === key && Number.isSafeInteger(cached.checkedAt) && now >= cached.checkedAt) {
      const age = now - cached.checkedAt;
      if (cached.outcome === "success" && validVersion(cached.version) && /^[a-f0-9]{40}$/.test(cached.commit) && age < DAY) return result(cached.version);
      if (cached.outcome === "failure" && age < RETRY) return unavailable();
    }
    lock = path.join(dataRoot, "plugin-observation.lock");
    try {
      fs.mkdirSync(dataRoot, { recursive: true, mode: 0o700 });
      safeDirectory(dataRoot);
      fs.mkdirSync(lock, { mode: 0o700 });
      owned = true; writable = true;
    } catch (error) {
      if (error.code === "EEXIST") return unavailable();
      if (!["EACCES", "EPERM", "EROFS"].includes(error.code)) return unavailable();
      // A disposable cache write failure must not discard fresh read results.
    }
    let observation, answer;
    try {
      const commit = await get("https://api.github.com/repos/sjunepark/agent-scripts/commits/main", signal);
      if (!/^[a-f0-9]{40}$/.test(commit?.sha)) throw new Error("invalid commit");
      const published = await get(`https://raw.githubusercontent.com/sjunepark/agent-scripts/${commit.sha}/plugins/sjskills-maintenance/.codex-plugin/plugin.json`, signal);
      if (published?.name !== "sjskills-maintenance" || !validVersion(published.version)) throw new Error("invalid manifest");
      observation = { key, checkedAt: now, outcome: "success", commit: commit.sha, version: published.version };
      answer = result(published.version);
    } catch {
      observation = { key, checkedAt: now, outcome: "failure" };
      answer = unavailable();
    }
    if (writable) {
      const temporary = path.join(dataRoot, `.observation-${randomUUID()}.tmp`);
      let created = false;
      try {
        safeDirectory(dataRoot);
        const fd = fs.openSync(temporary, "wx", 0o600); created = true;
        try { fs.writeFileSync(fd, JSON.stringify(observation) + "\n"); } finally { fs.closeSync(fd); }
        fs.renameSync(temporary, file);
      } catch { /* Cache persistence cannot invalidate a completed read. */ }
      finally { if (created && fs.existsSync(temporary)) fs.unlinkSync(temporary); }
    }
    return answer;
  } catch { return unavailable(); }
  finally { if (owned) { try { fs.rmdirSync(lock); } catch { /* Never remove unfamiliar contents. */ } } }
}
module.exports = { observePlugin, provenance, mainRef, fetchJSON, readBounded, safeDirectory, RETRY };
