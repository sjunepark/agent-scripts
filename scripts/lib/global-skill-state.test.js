"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");

const {
  DEFAULT_ROOT_POLICY,
  TREE_HASH_ALGORITHM,
  expectedGlobalPlacements,
  hashSkillDirectory,
  inspectSkillRoots,
  planGlobalSkillOperations,
  reconcileGlobalSkillState
} = require("./global-skill-state");

function temporaryHome(t) {
  const homeDir = fs.mkdtempSync(path.join(os.tmpdir(), "global-skill-state-"));
  t.after(() => fs.rmSync(homeDir, { recursive: true, force: true }));
  return homeDir;
}

function writeSkill(homeDir, relativeRoot, name, content) {
  const skillDir = path.join(homeDir, relativeRoot, name);
  fs.mkdirSync(skillDir, { recursive: true });
  fs.writeFileSync(path.join(skillDir, "SKILL.md"), content);
  return skillDir;
}

function copySkill(source, homeDir, relativeRoot, name) {
  const target = path.join(homeDir, relativeRoot, name);
  fs.mkdirSync(path.dirname(target), { recursive: true });
  fs.cpSync(source, target, { recursive: true });
  return target;
}

function writeLock(homeDir, records, rawContent) {
  const lockFile = path.join(homeDir, ".agents", ".skill-lock.json");
  fs.mkdirSync(path.dirname(lockFile), { recursive: true });
  fs.writeFileSync(
    lockFile,
    rawContent ?? JSON.stringify({ version: 1, skills: records }, null, 2)
  );
}

function desiredSkill(name, options = {}) {
  return {
    name,
    sourceId: options.sourceId ?? "test-source",
    source: options.source ?? "https://github.com/example/skills/tree/main/skills",
    manager: options.manager ?? "skills-cli",
    scope: "global",
    mode: "copy",
    agents: options.agents ?? ["codex"],
    fullDepth: false
  };
}

function expectedContentRecord(skillDir) {
  return hashSkillDirectory(skillDir);
}

function issue(report, type, skill, root) {
  return report.issues.find((candidate) =>
    candidate.type === type &&
    candidate.skill === skill &&
    (root === undefined || candidate.root === root)
  );
}

test("maps Codex and Pi compatibility to shared and Claude compatibility to Claude", () => {
  const placements = expectedGlobalPlacements([
    desiredSkill("claude-only", { agents: ["claude-code"] }),
    desiredSkill("codex-and-claude", { agents: ["codex", "claude-code", "pi"] }),
    desiredSkill("pi-only", { agents: ["pi"] })
  ]);

  assert.deepEqual(
    placements.map(({ skill, root, targetAgent }) => ({ skill, root, targetAgent })),
    [
      { skill: "claude-only", root: "claude", targetAgent: "claude-code" },
      { skill: "codex-and-claude", root: "claude", targetAgent: "claude-code" },
      { skill: "codex-and-claude", root: "shared", targetAgent: "codex" },
      { skill: "pi-only", root: "shared", targetAgent: "codex" }
    ]
  );
  assert.equal(placements.some(({ root, targetAgent }) => root === "pi" || targetAgent === "pi"), false);
});

test("requires an explicitly injected absolute fixture home", () => {
  assert.throws(
    () => inspectSkillRoots({ homeDir: "relative-fixture-home" }),
    /homeDir must be a non-empty absolute path/
  );
});

test("does not enumerate roots or locks whose canonical paths escape the fixture home", (t) => {
  const homeDir = temporaryHome(t);
  const outside = fs.mkdtempSync(path.join(os.tmpdir(), "global-skill-state-outside-"));
  t.after(() => fs.rmSync(outside, { recursive: true, force: true }));
  writeSkill(outside, "skills", "escaped", "must not read\n");
  fs.writeFileSync(
    path.join(outside, ".skill-lock.json"),
    JSON.stringify({ version: 1, skills: { escaped: { source: "outside" } } })
  );
  fs.symlinkSync(outside, path.join(homeDir, ".agents"));

  const inventory = inspectSkillRoots({ homeDir });

  assert.equal(inventory.entries.some(({ name }) => name === "escaped"), false);
  assert.equal(inventory.locks.some(({ name }) => name === "escaped"), false);
  assert.equal(
    inventory.errors.some(({ reason }) => reason === "root-outside-home"),
    true
  );

  const lockHome = temporaryHome(t);
  const outsideLock = path.join(outside, "outside-lock.json");
  fs.writeFileSync(
    outsideLock,
    JSON.stringify({ version: 1, skills: { escaped: { source: "outside" } } })
  );
  fs.mkdirSync(path.join(lockHome, ".agents"), { recursive: true });
  fs.symlinkSync(outsideLock, path.join(lockHome, ".agents/.skill-lock.json"));

  const lockInventory = inspectSkillRoots({ homeDir: lockHome });
  assert.equal(lockInventory.locks.some(({ name }) => name === "escaped"), false);
  assert.equal(
    lockInventory.errors.some(({ reason }) => reason === "lock-file-outside-home"),
    true
  );
});

test("hashes skill trees independently of creation order and timestamps", (t) => {
  const homeDir = temporaryHome(t);
  const first = writeSkill(homeDir, ".fixture-source-a", "stable", "same\n");
  fs.mkdirSync(path.join(first, "references"));
  fs.writeFileSync(path.join(first, "references", "b.md"), "b\n");
  fs.writeFileSync(path.join(first, "references", "a.md"), "a\n");

  const second = writeSkill(homeDir, ".fixture-source-b", "stable", "same\n");
  fs.mkdirSync(path.join(second, "references"));
  fs.writeFileSync(path.join(second, "references", "a.md"), "a\n");
  fs.writeFileSync(path.join(second, "references", "b.md"), "b\n");
  fs.utimesSync(path.join(second, "SKILL.md"), new Date(0), new Date(0));

  assert.deepEqual(hashSkillDirectory(first), hashSkillDirectory(second));
});

test("inspects only explicit roots and verifies a canonical legacy symlink", (t) => {
  const homeDir = temporaryHome(t);
  const desired = desiredSkill("alpha", { agents: ["codex", "claude-code", "pi"] });
  const source = writeSkill(homeDir, ".fixture-source", "alpha", "alpha v1\n");
  const expectedContent = { alpha: expectedContentRecord(source) };
  const shared = copySkill(source, homeDir, ".agents/skills", "alpha");
  copySkill(source, homeDir, ".claude/skills", "alpha");
  const piRoot = path.join(homeDir, ".pi/agent/skills");
  fs.mkdirSync(piRoot, { recursive: true });
  fs.symlinkSync(path.relative(piRoot, shared), path.join(piRoot, "alpha"));
  writeSkill(homeDir, ".codex/skills/.system", "runtime-owned", "do not inspect\n");
  writeLock(homeDir, {
    alpha: {
      source: desired.source,
      ref: "main",
      skillPath: "skills/alpha",
      computedHash: expectedContent.alpha.hash,
      hashAlgorithm: TREE_HASH_ALGORITHM
    }
  });

  const rootPolicy = [
    ...DEFAULT_ROOT_POLICY,
    {
      id: "codex-system",
      relativePath: ".codex/skills/.system",
      kind: "protected"
    }
  ];
  const inventory = inspectSkillRoots({ homeDir, rootPolicy });
  const report = reconcileGlobalSkillState({
    desiredEntries: [desired],
    knownSkillNames: ["alpha"],
    expectedContent,
    inventory
  });

  assert.equal(inventory.locks.length, 1);
  assert.equal(inventory.locks[0].hashAlgorithm, TREE_HASH_ALGORITHM);
  assert.equal(inventory.locks[0].ref, "main");
  assert.equal(inventory.locks[0].skillPath, "skills/alpha");
  assert.equal(
    inventory.entries.some(({ path: entryPath }) => entryPath.endsWith("runtime-owned")),
    false
  );
  assert.ok(issue(report, "protected", null, "codex-system"));
  assert.ok(issue(report, "verified-legacy-duplicate", "alpha", "pi"));
  assert.equal(report.issues.some(({ type }) => type === "missing"), false);

  const quarantineRoot = path.join(homeDir, ".skill-quarantine", "run-1");
  const operations = planGlobalSkillOperations({ report, quarantineRoot });
  assert.deepEqual(operations.apply, []);
  assert.equal(operations.prune.length, 1);
  assert.deepEqual(operations.restore, [
    {
      type: "restore",
      skill: "alpha",
      from: path.join(quarantineRoot, "pi", "alpha"),
      to: path.join(homeDir, ".pi/agent/skills/alpha"),
      mustNotExist: true
    }
  ]);
});

test("classifies missing, outdated, modified, and unverifiable expected entries", (t) => {
  const homeDir = temporaryHome(t);
  const desiredEntries = ["missing", "modified", "outdated", "unverifiable"]
    .map((name) => desiredSkill(name));
  const expectedContent = {};
  const oldContent = {};

  for (const name of ["modified", "outdated", "unverifiable"]) {
    const expectedSource = writeSkill(homeDir, ".fixture-source", name, `${name} expected\n`);
    expectedContent[name] = expectedContentRecord(expectedSource);
    const oldSource = writeSkill(homeDir, ".fixture-old", name, `${name} old\n`);
    oldContent[name] = expectedContentRecord(oldSource);
  }

  copySkill(path.join(homeDir, ".fixture-old/modified"), homeDir, ".agents/skills", "modified");
  fs.writeFileSync(
    path.join(homeDir, ".agents/skills/modified/SKILL.md"),
    "modified locally\n"
  );
  copySkill(path.join(homeDir, ".fixture-old/outdated"), homeDir, ".agents/skills", "outdated");
  copySkill(
    path.join(homeDir, ".fixture-old/unverifiable"),
    homeDir,
    ".agents/skills",
    "unverifiable"
  );
  writeLock(homeDir, {
    modified: {
      source: desiredEntries[1].source,
      computedHash: oldContent.modified.hash,
      hashAlgorithm: TREE_HASH_ALGORITHM
    },
    outdated: {
      source: desiredEntries[2].source,
      computedHash: oldContent.outdated.hash,
      hashAlgorithm: TREE_HASH_ALGORITHM
    },
    unverifiable: {
      source: desiredEntries[3].source,
      computedHash: oldContent.unverifiable.hash
    }
  });

  const inventory = inspectSkillRoots({ homeDir });
  const report = reconcileGlobalSkillState({
    desiredEntries,
    knownSkillNames: desiredEntries.map(({ name }) => name),
    expectedContent,
    inventory
  });

  assert.ok(issue(report, "missing", "missing", "shared"));
  assert.ok(issue(report, "modified", "modified", "shared"));
  assert.ok(issue(report, "outdated", "outdated", "shared"));
  assert.match(issue(report, "unclassified", "unverifiable", "shared").reason, /hash/);
});

test("separates wrong-root, unexpected managed, legacy-only, and unknown entries", (t) => {
  const homeDir = temporaryHome(t);
  const canonical = desiredSkill("canonical");
  const legacyOnly = desiredSkill("legacy-only");
  const sourceCanonical = writeSkill(homeDir, ".fixture-source", "canonical", "canonical\n");
  const sourceLegacy = writeSkill(homeDir, ".fixture-source", "legacy-only", "legacy\n");
  const expectedContent = {
    canonical: expectedContentRecord(sourceCanonical),
    "legacy-only": expectedContentRecord(sourceLegacy)
  };

  copySkill(sourceCanonical, homeDir, ".agents/skills", "canonical");
  copySkill(sourceCanonical, homeDir, ".claude/skills", "canonical");
  copySkill(sourceLegacy, homeDir, ".pi/agent/skills", "legacy-only");
  writeSkill(homeDir, ".agents/skills", "dev-only-extra", "extra\n");
  writeSkill(homeDir, ".agents/skills", "mystery", "unknown\n");

  const inventory = inspectSkillRoots({ homeDir });
  const report = reconcileGlobalSkillState({
    desiredEntries: [canonical, legacyOnly],
    knownSkillNames: ["canonical", "dev-only-extra", "legacy-only"],
    expectedContent,
    inventory
  });

  assert.ok(issue(report, "misplaced", "canonical", "claude"));
  assert.ok(issue(report, "missing", "legacy-only", "shared"));
  assert.ok(issue(report, "misplaced", "legacy-only", "pi"));
  assert.ok(issue(report, "unexpected-managed", "dev-only-extra", "shared"));
  assert.ok(issue(report, "unclassified", "mystery", "shared"));
  assert.equal(issue(report, "verified-legacy-duplicate", "legacy-only", "pi"), undefined);
});

test("does not follow a skill symlink outside the explicit roots", (t) => {
  const homeDir = temporaryHome(t);
  const outside = writeSkill(homeDir, "outside-root", "linked", "outside\n");
  fs.mkdirSync(path.join(homeDir, ".agents/skills"), { recursive: true });
  fs.symlinkSync(outside, path.join(homeDir, ".agents/skills/linked"));

  const inventory = inspectSkillRoots({ homeDir });
  const entry = inventory.entries.find(({ name }) => name === "linked");
  assert.equal(entry.kind, "symlink");
  assert.equal(entry.hash, null);
  assert.equal(entry.hashStatus, "target-outside-explicit-roots");

  const report = reconcileGlobalSkillState({
    desiredEntries: [desiredSkill("linked")],
    knownSkillNames: ["linked"],
    expectedContent: {
      linked: expectedContentRecord(outside)
    },
    inventory
  });
  assert.ok(issue(report, "misplaced", "linked", "shared"));
});

test("plans installs and updates only for skills-cli managed unambiguous drift", (t) => {
  const homeDir = temporaryHome(t);
  const desiredEntries = [
    desiredSkill("cli", { agents: ["codex", "claude-code", "pi"] }),
    desiredSkill("manual", { manager: "manual" }),
    desiredSkill("workflow", { manager: "workflow", agents: ["pi"] })
  ];
  const inventory = inspectSkillRoots({ homeDir });
  const report = reconcileGlobalSkillState({
    desiredEntries,
    knownSkillNames: desiredEntries.map(({ name }) => name),
    expectedContent: {},
    inventory
  });
  const operations = planGlobalSkillOperations({
    report,
    quarantineRoot: path.join(homeDir, ".skill-quarantine", "run-2")
  });

  assert.deepEqual(
    operations.apply.map(({ skill, targetAgent }) => ({ skill, targetAgent })),
    [
      { skill: "cli", targetAgent: "claude-code" },
      { skill: "cli", targetAgent: "codex" }
    ]
  );
  assert.equal(operations.apply.some(({ targetAgent }) => targetAgent === "pi"), false);
  assert.deepEqual(
    operations.manual.map(({ skill, manager }) => ({ skill, manager })),
    [
      { skill: "manual", manager: "manual" },
      { skill: "workflow", manager: "workflow" }
    ]
  );
  assert.throws(
    () => planGlobalSkillOperations({
      report,
      quarantineRoot: path.join(homeDir, ".agents/skills/quarantine")
    }),
    /quarantineRoot must not overlap skill root/
  );
  assert.throws(
    () => planGlobalSkillOperations({ report, quarantineRoot: homeDir }),
    /quarantineRoot must be a child of the inspected homeDir/
  );
});

test("preserves manual ownership for ambiguous drift", (t) => {
  const homeDir = temporaryHome(t);
  const desired = desiredSkill("manual-modified", { manager: "manual" });
  const source = writeSkill(homeDir, ".fixture-source", desired.name, "expected\n");
  const installed = copySkill(source, homeDir, ".agents/skills", desired.name);
  const locked = expectedContentRecord(installed);
  fs.writeFileSync(path.join(installed, "SKILL.md"), "locally modified\n");
  writeLock(homeDir, {
    [desired.name]: {
      source: desired.source,
      computedHash: locked.hash,
      hashAlgorithm: TREE_HASH_ALGORITHM
    }
  });

  const report = reconcileGlobalSkillState({
    desiredEntries: [desired],
    knownSkillNames: [desired.name],
    expectedContent: { [desired.name]: expectedContentRecord(source) },
    inventory: inspectSkillRoots({ homeDir })
  });
  const operations = planGlobalSkillOperations({
    report,
    quarantineRoot: path.join(homeDir, ".skill-quarantine", "run-3")
  });

  assert.deepEqual(operations.blocked, []);
  assert.deepEqual(
    operations.manual.map(({ skill, manager, issue: issueType }) => ({
      skill,
      manager,
      issue: issueType
    })),
    [{ skill: desired.name, manager: "manual", issue: "modified" }]
  );
});

test("reports contradictory source provenance without treating content as managed", (t) => {
  const homeDir = temporaryHome(t);
  const desired = desiredSkill("source-mismatch");
  const source = writeSkill(homeDir, ".fixture-source", "source-mismatch", "same\n");
  copySkill(source, homeDir, ".agents/skills", "source-mismatch");
  const expectedContent = { "source-mismatch": expectedContentRecord(source) };
  writeLock(homeDir, {
    "source-mismatch": {
      source: "https://github.com/other/private-skills",
      computedHash: expectedContent["source-mismatch"].hash,
      hashAlgorithm: TREE_HASH_ALGORITHM
    }
  });

  const inventory = inspectSkillRoots({ homeDir });
  const report = reconcileGlobalSkillState({
    desiredEntries: [desired],
    knownSkillNames: [desired.name],
    expectedContent,
    inventory
  });

  assert.equal(
    issue(report, "unclassified", "source-mismatch", "shared").reason,
    "source-provenance-mismatch"
  );
});

test("does not use a shared-root lock as Claude provenance", (t) => {
  const homeDir = temporaryHome(t);
  const desired = desiredSkill("root-local-lock", { agents: ["claude-code"] });
  const expected = writeSkill(homeDir, ".fixture-source", desired.name, "expected\n");
  const old = writeSkill(homeDir, ".fixture-old", desired.name, "old\n");
  copySkill(old, homeDir, ".claude/skills", desired.name);
  const oldContent = expectedContentRecord(old);
  writeLock(homeDir, {
    [desired.name]: {
      source: desired.source,
      computedHash: oldContent.hash,
      hashAlgorithm: TREE_HASH_ALGORITHM
    }
  });

  const report = reconcileGlobalSkillState({
    desiredEntries: [desired],
    knownSkillNames: [desired.name],
    expectedContent: { [desired.name]: expectedContentRecord(expected) },
    inventory: inspectSkillRoots({ homeDir })
  });

  assert.equal(
    issue(report, "unclassified", desired.name, "claude").reason,
    "lock-hash-unverifiable"
  );
  assert.equal(issue(report, "outdated", desired.name, "claude"), undefined);
});

test("turns malformed lock provenance into an unclassified issue", (t) => {
  const homeDir = temporaryHome(t);
  writeLock(homeDir, {}, "{not-json");

  const inventory = inspectSkillRoots({ homeDir });
  const report = reconcileGlobalSkillState({
    desiredEntries: [],
    knownSkillNames: [],
    expectedContent: {},
    inventory
  });

  const lockIssue = report.issues.find(({ reason }) => reason === "lock-file-invalid");
  assert.equal(lockIssue.type, "unclassified");
  assert.equal(lockIssue.skill, null);
  assert.match(lockIssue.path, /\.agents\/\.skill-lock\.json$/);
});
