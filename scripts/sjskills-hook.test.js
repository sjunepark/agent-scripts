"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const { test } = require("node:test");
const { sessionContext, snapshotResources } = require("../plugins/sjskills-maintenance/scripts/session-start.cjs");
const { updateStatus, recordCheck } = require("../plugins/sjskills-maintenance/scripts/update-check.cjs");

const repo = path.resolve(__dirname, "..");
const plugin = path.join(repo, "plugins/sjskills-maintenance");
const runtime = path.join(plugin, "scripts/session-start.cjs");
const event = { hook_event_name: "SessionStart", source: "startup", cwd: repo };

function copyFixture(from, to) {
  // Node 22.17.1's native recursive copy crashes on these Unicode Windows
  // paths. Keep the fixture copy in ordinary reads/writes so quoting is tested.
  fs.mkdirSync(to, { recursive: true });
  for (const entry of fs.readdirSync(from, { withFileTypes: true })) {
    const source = path.join(from, entry.name), destination = path.join(to, entry.name);
    if (entry.isDirectory()) copyFixture(source, destination);
    else fs.writeFileSync(destination, fs.readFileSync(source));
  }
}

function fixture(t) {
  const temporaryRoot = fs.realpathSync(os.tmpdir());
  const root = fs.mkdtempSync(path.join(temporaryRoot, "sjskills-plugin-"));
  t.after(() => {
    const relative = path.relative(temporaryRoot, fs.realpathSync(root));
    assert.ok(relative && !relative.startsWith("..") && !path.isAbsolute(relative));
    fs.rmSync(root, { recursive: true, force: true });
  });
  const installed = path.join(root, "user's 한글 $workspace & tools", "plugin");
  const data = path.join(root, "plugin-data");
  copyFixture(plugin, installed);
  return { root, installed, data };
}

test("plugin bundles the exact sync workflow outside skill discovery", () => {
  const manifest = JSON.parse(fs.readFileSync(path.join(plugin, ".codex-plugin/plugin.json")));
  assert.equal(manifest.name, "sjskills-maintenance");
  assert.equal(manifest.skills, undefined);
  assert.equal(fs.existsSync(path.join(plugin, "skills")), false);
  for (const relative of ["SKILL.md", "references/global-rollout.md"]) {
    assert.deepEqual(fs.readFileSync(path.join(plugin, "references/sjskills", relative)),
      fs.readFileSync(path.join(repo, "skills/sjskills", relative)));
  }
  const result = spawnSync(process.execPath, [path.join(repo, "scripts/sync-sjskills-plugin"), "--check"], { encoding: "utf8", windowsHide: true });
  assert.equal(result.status, 0, result.stderr);
});

test("real plugin command handles shell-sensitive installation paths", (t) => {
  const f = fixture(t);
  const hooks = JSON.parse(fs.readFileSync(path.join(f.installed, "hooks/hooks.json")));
  const group = hooks.hooks.SessionStart[0];
  assert.equal(group.matcher, "^(startup|resume)$");
  const command = process.platform === "win32" ? group.hooks[0].commandWindows : group.hooks[0].command;
  const shells = process.platform === "win32"
    ? [["powershell.exe", ["-NoProfile", "-NonInteractive", "-Command", command]],
      [process.env.COMSPEC || "cmd.exe", ["/d", "/s", "/c", command]]]
    : [["/bin/sh", ["-c", command]]];
  for (const [shell, args] of shells) {
    const result = spawnSync(shell, args, {
      input: JSON.stringify({ ...event, cwd: f.root }), encoding: "utf8", cwd: f.root, windowsHide: true,
      env: { ...process.env, PLUGIN_ROOT: f.installed, PLUGIN_DATA: f.data },
    });
    assert.equal(result.status, 0, `${shell}: ${result.stderr}`);
    const output = JSON.parse(result.stdout);
    if (process.platform === "linux") assert.match(output.systemMessage, /not a supported/);
    else {
      assert.equal(output.hookSpecificOutput.hookEventName, "SessionStart");
      const context = output.hookSpecificOutput.additionalContext;
      const resources = JSON.parse(context.slice(context.lastIndexOf("\n") + 1));
      assert.equal(resources.sessionDirectory, f.root);
      assert.equal(resources.pluginData, f.data);
      assert.equal(resources.pluginUpdate.due, true);
      assert.equal(fs.existsSync(resources.syncSkill), true);
      assert.equal(fs.existsSync(resources.updateHelper), true);
    }
  }
  assert.equal(fs.existsSync(path.join(f.data, "update-check.json")), false);
});

test("session resources survive plugin-cache removal and reject modified snapshots", (t) => {
  const f = fixture(t);
  const snapshot = snapshotResources(f.installed, f.data);
  assert.equal(snapshotResources(f.installed, f.data), snapshot);
  const skill = fs.readFileSync(path.join(snapshot, "references/sjskills/SKILL.md"));
  const relative = path.relative(f.root, f.installed);
  assert.ok(relative && !relative.startsWith("..") && !path.isAbsolute(relative));
  fs.rmSync(f.installed, { recursive: true, force: true });
  const result = spawnSync(process.execPath, [path.join(snapshot, "scripts/update-check.cjs"), "success", f.data], { encoding: "utf8", windowsHide: true });
  assert.equal(result.status, 0, result.stderr);
  assert.equal(updateStatus(f.data).due, false);
  assert.deepEqual(fs.readFileSync(path.join(snapshot, "references/sjskills/SKILL.md")), skill);
  copyFixture(plugin, f.installed);
  fs.appendFileSync(path.join(snapshot, "references/sjskills/SKILL.md"), "modified");
  assert.throws(() => snapshotResources(f.installed, f.data), /modified workflow snapshot/);
});

test("supported platforms agree with releases; compact and subagents do nothing", (t) => {
  const f = fixture(t);
  const targets = JSON.parse(fs.readFileSync(path.join(repo, "packaging/targets.json")))
    .map(({ os, arch }) => `${os === "windows" ? "win32" : os}/${arch === "amd64" ? "x64" : arch}`);
  for (const platform of ["win32", "darwin", "linux"]) {
    for (const arch of ["x64", "arm64", "ia32"]) {
      const output = sessionContext(event, platform, arch, f.installed, f.data);
      assert.equal(Boolean(output.hookSpecificOutput), targets.includes(`${platform}/${arch}`));
    }
  }
  assert.ok(sessionContext({ ...event, source: "resume" }, "darwin", "arm64", f.installed, f.data).hookSpecificOutput);
  for (const source of ["compact", "clear", "startup-extra"]) assert.equal(sessionContext({ ...event, source }), null);
  assert.equal(sessionContext({ ...event, hook_event_name: "SubagentStart" }), null);
});

test("daily success and failed-attempt cooldown are based on completed checks", (t) => {
  const { data } = fixture(t);
  const now = Date.now();
  assert.equal(updateStatus(data, now).due, true);
  recordCheck(data, "success", now);
  assert.equal(updateStatus(data, now + 24 * 60 * 60 * 1000 - 1).due, false);
  assert.equal(updateStatus(data, now + 24 * 60 * 60 * 1000).due, true);
  recordCheck(data, "failure", now);
  assert.equal(updateStatus(data, now + 15 * 60 * 1000 - 1).due, false);
  assert.equal(updateStatus(data, now + 15 * 60 * 1000).due, true);
  assert.equal(updateStatus(data, now - 1).due, true);
});

test("corrupt state is due; unsafe state is reported without suppressing skill maintenance", (t) => {
  const f = fixture(t);
  fs.mkdirSync(f.data);
  const state = path.join(f.data, "update-check.json");
  for (const text of ["broken", "null", "[]", '{"outcome":"success"}', '{"checkedAt":1,"outcome":"unknown"}']) {
    fs.writeFileSync(state, text);
    assert.equal(updateStatus(f.data).due, true);
  }
  fs.unlinkSync(state);
  fs.mkdirSync(state);
  assert.throws(() => updateStatus(f.data), /unsafe plugin update state/);
  assert.throws(() => recordCheck(f.data, "success"), /unsafe plugin update state/);
  const output = sessionContext(event, "win32", "x64", f.installed, f.data);
  assert.ok(output.hookSpecificOutput);
  assert.match(output.hookSpecificOutput.additionalContext, /unsafe plugin update state/);
});

test("linked plugin data directories never receive state writes", (t) => {
  const f = fixture(t);
  const outside = path.join(f.root, "outside");
  fs.mkdirSync(outside);
  fs.symlinkSync(outside, f.data, process.platform === "win32" ? "junction" : "dir");
  assert.throws(() => recordCheck(f.data, "success"), /unsafe plugin data directory/);
  assert.deepEqual(fs.readdirSync(outside), []);
});

test("state helper CLI records outcomes and rejects invalid invocations", (t) => {
  const f = fixture(t);
  const helper = path.join(f.installed, "scripts/update-check.cjs");
  const run = (...args) => spawnSync(process.execPath, [helper, ...args], { encoding: "utf8", windowsHide: true });
  assert.equal(run("status", f.data).status, 0);
  assert.equal(fs.existsSync(f.data), false);
  assert.equal(run("success", f.data).status, 0);
  assert.equal(JSON.parse(run("status", f.data).stdout).due, false);
  assert.equal(run("failure", f.data).status, 0);
  assert.equal(JSON.parse(run("status", f.data).stdout).lastOutcome, "failure");
  assert.equal(run("invalid", f.data).status, 1);
  assert.equal(run("success", "relative").status, 1);
});

test("malformed runtime input emits a nonblocking diagnostic", () => {
  for (const input of ["broken", "null", "[]", JSON.stringify({ ...event, cwd: "relative" }), "x".repeat(1024 * 1024 + 1)]) {
    const result = spawnSync(process.execPath, [runtime], { input, encoding: "utf8", windowsHide: true });
    assert.equal(result.status, 0, result.stderr);
    const output = JSON.parse(result.stdout);
    assert.match(output.systemMessage, /maintenance unavailable/);
    assert.equal(output.continue, undefined);
  }
});
