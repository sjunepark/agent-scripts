"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");
const test = require("node:test");
const { main } = require("./audit-global-skills");

test("transition wrapper delegates its only supported operation to sjskills global plan", () => {
  let invocation;
  const status = main([], {
    spawnSync(command, args, options) {
      invocation = { command, args, options };
      return { status: 0 };
    }
  });

  assert.equal(status, 0);
  assert.equal(path.basename(invocation.command), "sjskills");
  assert.deepEqual(invocation.args, ["plan", "--global"]);
  assert.equal(invocation.options.stdio, "inherit");
});

test("transition wrapper propagates sjskills failure and startup errors", () => {
  assert.equal(main([], { spawnSync: () => ({ status: 3 }) }), 3);
  assert.equal(main([], {
    spawnSync: () => ({ error: new Error("missing") }),
    error: () => {}
  }), 2);
});

test("transition wrapper rejects every legacy argument", () => {
  for (const args of [
    ["--profile", "dev"],
    ["--apply"],
    ["--prune", "sha256:example"],
    ["--replace-unverified", "sha256:example"],
    ["--restore", "manifest.json"]
  ]) {
    assert.equal(main(args, { log: () => {}, error: () => {} }), 64, args.join(" "));
  }
});

test("transition wrapper retains help as a successful operation", () => {
  assert.equal(main(["--help"], { log: () => {} }), 0);
});
