"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");

const {
  assertRemoteSource,
  executeApplyOperations,
  materializeExpectedContent,
  parseArgs,
  run,
  skillsCliAddArgs,
  skillsCliEnvironment
} = require("./audit-global-skills");
const { hashSkillDirectory } = require("./lib/global-skill-state");

function temporaryDirectory(t, prefix) {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), prefix));
  t.after(() => fs.rmSync(directory, { recursive: true, force: true }));
  return directory;
}

function writeSkill(root, name, content) {
  const directory = path.join(root, name);
  fs.mkdirSync(directory, { recursive: true });
  fs.writeFileSync(path.join(directory, "SKILL.md"), content);
  return directory;
}

function copySkill(source, homeDir, relativeRoot, name) {
  const target = path.join(homeDir, relativeRoot, name);
  fs.mkdirSync(path.dirname(target), { recursive: true });
  fs.cpSync(source, target, { recursive: true });
  return target;
}

function registryFixture(t) {
  const fixtureRoot = temporaryDirectory(t, "global-skill-audit-registry-");
  const registryPath = path.join(fixtureRoot, "skill-registry.json");
  const registry = {
    version: 2,
    description: "Audit fixture.",
    global: {
      allowUnlistedSkills: false,
      profiles: {
        dev: { audiences: ["common", "dev"] },
        kicpa: { audiences: ["common", "kicpa"] }
      }
    },
    sources: {
      public: { kind: "external", location: "https://github.com/example/public-skills" }
    },
    skills: [
      {
        name: "common-alpha",
        source: "public",
        recommendation: { scope: "global", audience: "common", agents: ["codex"] },
        installation: { manager: "skills-cli", mode: "copy" }
      },
      {
        name: "dev-alpha",
        source: "public",
        recommendation: {
          scope: "global",
          audience: "dev",
          agents: ["claude-code", "codex", "pi"]
        },
        installation: { manager: "skills-cli", mode: "copy" }
      },
      {
        name: "kicpa-alpha",
        source: "public",
        recommendation: { scope: "global", audience: "kicpa", agents: ["codex"] },
        installation: { manager: "skills-cli", mode: "copy" }
      }
    ]
  };
  fs.writeFileSync(registryPath, JSON.stringify(registry, null, 2));
  return { registryPath, repositorySkillNames: [], skillNames: registry.skills.map(({ name }) => name) };
}

function expectedFixture(t, names) {
  const sourceRoot = temporaryDirectory(t, "global-skill-audit-source-");
  const sources = {};
  const expectedContent = {};
  for (const name of names) {
    sources[name] = writeSkill(sourceRoot, name, `${name} expected\n`);
    expectedContent[name] = hashSkillDirectory(sources[name]);
  }
  return { sources, expectedContent };
}

function options(registryPath, homeDir, profile, overrides = {}) {
  return {
    registry: registryPath,
    profile,
    homeDir,
    apply: false,
    prune: false,
    restore: null,
    yes: false,
    help: false,
    ...overrides
  };
}

test("parses strict modes and keeps apply, prune, and restore separate", () => {
  assert.throws(() => parseArgs([]), /--profile is required/);
  assert.throws(
    () => parseArgs(["--profile", "dev", "--apply", "--prune"]),
    /separate operations/
  );
  assert.throws(
    () => parseArgs(["--profile", "dev", "--replace-unverified"]),
    /requires --yes/
  );
  assert.throws(() => parseArgs(["--profile", "dev", "--prune"]), /requires --yes/);
  assert.throws(() => parseArgs(["--restore", "/tmp/manifest.json"]), /requires --yes/);
  assert.equal(parseArgs(["--profile", "dev", "--apply"]).apply, true);
  assert.equal(
    parseArgs(["--profile", "dev", "--replace-unverified", "--yes"]).replaceUnverified,
    true
  );
  assert.equal(parseArgs(["--profile", "kicpa", "--prune", "--yes"]).prune, true);
});

test("synthesizes credential-free remote commands with explicit supported targets", () => {
  const publicOperation = {
    source: "https://github.com/example/public-skills/tree/main/skills",
    skill: "alpha",
    targetAgent: "codex",
    fullDepth: false
  };
  const privateOperation = {
    ...publicOperation,
    source: "https://github.com/example/private-skills/tree/main/skills",
    targetAgent: "claude-code",
    fullDepth: true
  };

  assert.deepEqual(skillsCliAddArgs(publicOperation), [
    "skills",
    "add",
    publicOperation.source,
    "--skill",
    "alpha",
    "--copy",
    "--global",
    "--agent",
    "codex",
    "--yes"
  ]);
  assert.equal(skillsCliAddArgs(privateOperation).includes("--full-depth"), true);
  assert.equal(skillsCliAddArgs(privateOperation).includes("pi"), false);
  assert.throws(() => assertRemoteSource("./skills"), /must be remote/);
  assert.throws(
    () => assertRemoteSource("https://token@example.com/private-skills"),
    /credential-free/
  );
  assert.throws(
    () => assertRemoteSource("https://github.com/example/skills?token=secret"),
    /credential-free/
  );
  assert.throws(
    () => assertRemoteSource("https://github.com/example/skills#password=secret"),
    /credential-free/
  );
  assert.throws(() => assertRemoteSource("npm:example?token=secret"), /credential-free/);
});

test("points explicit Codex and Claude targets at the two managed roots", () => {
  const homeDir = "/tmp/example-home";
  const codex = skillsCliEnvironment(homeDir, "codex", {});
  const claude = skillsCliEnvironment(homeDir, "claude-code", {});

  assert.equal(codex.HOME, homeDir);
  assert.equal(codex.CODEX_HOME, path.join(homeDir, ".agents"));
  assert.equal(claude.CLAUDE_CONFIG_DIR, path.join(homeDir, ".claude"));
});

test("materializes remote expected content only in a temporary home", (t) => {
  const sourceRoot = temporaryDirectory(t, "global-skill-materialize-source-");
  const source = writeSkill(sourceRoot, "alpha", "alpha\n");
  let observedTemporaryHome;
  const expected = materializeExpectedContent({
    desiredEntries: [{
      name: "alpha",
      source: "example/skills",
      manager: "skills-cli",
      agents: ["pi"],
      fullDepth: false
    }],
    execSkillsCli({ args, homeDir, operation }) {
      observedTemporaryHome = homeDir;
      assert.equal(args.includes("pi"), false);
      copySkill(source, homeDir, ".agents/skills", operation.skill);
    }
  });

  assert.deepEqual(expected.alpha, hashSkillDirectory(source));
  assert.equal(fs.existsSync(observedTemporaryHome), false);
});

test("strict fixture audits pass for exact profiles and fail for cross-profile drift", (t) => {
  const homeDir = temporaryDirectory(t, "global-skill-audit-home-");
  const fixture = registryFixture(t);
  const expected = expectedFixture(t, fixture.skillNames);
  copySkill(expected.sources["common-alpha"], homeDir, ".agents/skills", "common-alpha");
  copySkill(expected.sources["dev-alpha"], homeDir, ".agents/skills", "dev-alpha");
  copySkill(expected.sources["dev-alpha"], homeDir, ".claude/skills", "dev-alpha");

  const devOutput = [];
  const devExit = run(options(fixture.registryPath, homeDir, "dev"), {
    expectedContent: expected.expectedContent,
    repositorySkillNames: fixture.repositorySkillNames,
    output: (line) => devOutput.push(line),
    now: () => new Date("2026-08-06T00:00:00.000Z")
  });
  assert.equal(devExit, 0);
  assert.equal(devOutput.includes("Strict result: pass"), true);

  const kicpaOutput = [];
  const kicpaExit = run(options(fixture.registryPath, homeDir, "kicpa"), {
    expectedContent: expected.expectedContent,
    repositorySkillNames: fixture.repositorySkillNames,
    output: (line) => kicpaOutput.push(line),
    now: () => new Date("2026-08-06T00:00:00.000Z")
  });
  assert.equal(kicpaExit, 1);
  assert.equal(kicpaOutput.some((line) => line.includes("missing: kicpa-alpha")), true);
  assert.equal(kicpaOutput.some((line) => line.includes("unexpected-managed: dev-alpha")), true);

  const kicpaHome = temporaryDirectory(t, "global-skill-audit-kicpa-home-");
  copySkill(expected.sources["common-alpha"], kicpaHome, ".agents/skills", "common-alpha");
  copySkill(expected.sources["kicpa-alpha"], kicpaHome, ".agents/skills", "kicpa-alpha");
  const exactKicpaOutput = [];
  const exactKicpaExit = run(options(fixture.registryPath, kicpaHome, "kicpa"), {
    expectedContent: expected.expectedContent,
    repositorySkillNames: fixture.repositorySkillNames,
    output: (line) => exactKicpaOutput.push(line),
    now: () => new Date("2026-08-06T00:00:00.000Z")
  });
  assert.equal(exactKicpaExit, 0);
  assert.equal(exactKicpaOutput.includes("Strict result: pass"), true);
});

test("apply installs only planned remote Codex and Claude targets and verifies content", (t) => {
  const homeDir = temporaryDirectory(t, "global-skill-apply-home-");
  const fixture = registryFixture(t);
  const expected = expectedFixture(t, fixture.skillNames);
  const calls = [];

  const exitCode = run(options(fixture.registryPath, homeDir, "dev", { apply: true }), {
    expectedContent: expected.expectedContent,
    repositorySkillNames: fixture.repositorySkillNames,
    execSkillsCli({ args, operation }) {
      calls.push({ args, operation });
      const relativeRoot = operation.root === "shared" ? ".agents/skills" : ".claude/skills";
      copySkill(expected.sources[operation.skill], homeDir, relativeRoot, operation.skill);
    },
    output: () => {},
    now: () => new Date("2026-08-06T00:00:00.000Z")
  });

  assert.equal(exitCode, 0);
  assert.deepEqual(
    calls.map(({ operation }) => `${operation.skill}:${operation.targetAgent}`),
    ["common-alpha:codex", "dev-alpha:claude-code", "dev-alpha:codex"]
  );
  assert.equal(calls.some(({ args }) => args.includes("pi")), false);
});

test("apply refuses to overwrite a target that changed after planning", (t) => {
  const homeDir = temporaryDirectory(t, "global-skill-apply-race-home-");
  const sourceRoot = temporaryDirectory(t, "global-skill-apply-race-source-");
  const source = writeSkill(sourceRoot, "alpha", "expected\n");
  const target = copySkill(source, homeDir, ".agents/skills", "alpha");
  let invoked = false;

  assert.throws(
    () => executeApplyOperations({
      operations: [{
        type: "install",
        skill: "alpha",
        source: "example/skills",
        root: "shared",
        targetAgent: "codex",
        fullDepth: false,
        targetPath: target,
        expectedCurrentKind: null,
        expectedCurrentHashAlgorithm: null,
        expectedCurrentHash: null
      }],
      expectedContent: { alpha: hashSkillDirectory(source) },
      homeDir,
      execSkillsCli() {
        invoked = true;
      }
    }),
    /target appeared after planning/
  );
  assert.equal(invoked, false);
});

test("apply refuses a missing managed root beneath a symlinked ancestor", (t) => {
  const homeDir = temporaryDirectory(t, "global-skill-apply-confined-home-");
  const outside = temporaryDirectory(t, "global-skill-apply-confined-outside-");
  const sourceRoot = temporaryDirectory(t, "global-skill-apply-confined-source-");
  const source = writeSkill(sourceRoot, "alpha", "expected\n");
  fs.symlinkSync(outside, path.join(homeDir, ".agents"));
  let invoked = false;

  assert.throws(
    () => executeApplyOperations({
      operations: [{
        type: "install",
        skill: "alpha",
        source: "example/skills",
        root: "shared",
        targetAgent: "codex",
        fullDepth: false,
        targetPath: path.join(homeDir, ".agents/skills/alpha"),
        expectedCurrentKind: null,
        expectedCurrentHashAlgorithm: null,
        expectedCurrentHash: null
      }],
      expectedContent: { alpha: hashSkillDirectory(source) },
      homeDir,
      execSkillsCli() {
        invoked = true;
      }
    }),
    /apply root resolves outside homeDir/
  );
  assert.equal(invoked, false);
  assert.equal(fs.existsSync(path.join(outside, "skills/alpha")), false);
});

test("apply refuses a managed root symlinked to another in-home directory", (t) => {
  const homeDir = temporaryDirectory(t, "global-skill-apply-in-home-");
  const sourceRoot = temporaryDirectory(t, "global-skill-apply-in-home-source-");
  const source = writeSkill(sourceRoot, "alpha", "expected\n");
  const alternateRoot = path.join(homeDir, "alternate-skills");
  fs.mkdirSync(path.join(homeDir, ".agents"), { recursive: true });
  fs.mkdirSync(alternateRoot, { recursive: true });
  fs.symlinkSync(alternateRoot, path.join(homeDir, ".agents/skills"));
  let invoked = false;

  assert.throws(
    () => executeApplyOperations({
      operations: [{
        type: "install",
        skill: "alpha",
        source: "example/skills",
        root: "shared",
        targetAgent: "codex",
        fullDepth: false,
        targetPath: path.join(homeDir, ".agents/skills/alpha"),
        modeledRootPath: path.join(homeDir, ".agents/skills"),
        expectedRootCanonicalPath: null,
        expectedCurrentKind: null,
        expectedCurrentHashAlgorithm: null,
        expectedCurrentHash: null
      }],
      expectedContent: { alpha: hashSkillDirectory(source) },
      homeDir,
      execSkillsCli() { invoked = true; }
    }),
    /apply root must not traverse a symlink/
  );
  assert.equal(invoked, false);
  assert.equal(fs.existsSync(path.join(alternateRoot, "alpha")), false);
});

test("apply adopts exact copies and later updates only recorded unmodified drift", (t) => {
  const homeDir = temporaryDirectory(t, "global-skill-adoption-home-");
  const fixture = registryFixture(t);
  const initial = expectedFixture(t, fixture.skillNames);
  copySkill(initial.sources["common-alpha"], homeDir, ".agents/skills", "common-alpha");
  copySkill(initial.sources["dev-alpha"], homeDir, ".agents/skills", "dev-alpha");
  copySkill(initial.sources["dev-alpha"], homeDir, ".claude/skills", "dev-alpha");
  let calls = 0;

  assert.equal(run(options(fixture.registryPath, homeDir, "dev", { apply: true }), {
    expectedContent: initial.expectedContent,
    repositorySkillNames: fixture.repositorySkillNames,
    execSkillsCli() { calls += 1; },
    output: () => {},
    now: () => new Date("2026-08-06T00:00:00.000Z")
  }), 0);
  assert.equal(calls, 0);
  const statePath = path.join(homeDir, ".agents/.global-skill-state.json");
  const state = JSON.parse(fs.readFileSync(statePath, "utf8"));
  assert.equal(state.records.length, 3);

  const updated = expectedFixture(t, fixture.skillNames);
  for (const name of fixture.skillNames) {
    fs.writeFileSync(path.join(updated.sources[name], "SKILL.md"), `${name} updated\n`);
    updated.expectedContent[name] = hashSkillDirectory(updated.sources[name]);
  }
  const operationTypes = [];
  assert.equal(run(options(fixture.registryPath, homeDir, "dev", { apply: true }), {
    expectedContent: updated.expectedContent,
    repositorySkillNames: fixture.repositorySkillNames,
    execSkillsCli({ operation }) {
      operationTypes.push(`${operation.type}:${operation.root}:${operation.skill}`);
      const target = path.join(
        homeDir,
        operation.root === "shared" ? ".agents/skills" : ".claude/skills",
        operation.skill
      );
      fs.rmSync(target, { recursive: true, force: true });
      copySkill(updated.sources[operation.skill], homeDir,
        operation.root === "shared" ? ".agents/skills" : ".claude/skills",
        operation.skill);
    },
    output: () => {},
    now: () => new Date("2026-08-07T00:00:00.000Z")
  }), 0);
  assert.deepEqual(operationTypes, [
    "update:shared:common-alpha",
    "update:claude:dev-alpha",
    "update:shared:dev-alpha"
  ]);
});

test("apply refuses to overwrite malformed reconciler provenance", (t) => {
  const homeDir = temporaryDirectory(t, "global-skill-invalid-state-home-");
  const fixture = registryFixture(t);
  const expected = expectedFixture(t, fixture.skillNames);
  copySkill(expected.sources["common-alpha"], homeDir, ".agents/skills", "common-alpha");
  copySkill(expected.sources["dev-alpha"], homeDir, ".agents/skills", "dev-alpha");
  copySkill(expected.sources["dev-alpha"], homeDir, ".claude/skills", "dev-alpha");
  const statePath = path.join(homeDir, ".agents/.global-skill-state.json");
  fs.writeFileSync(statePath, "{not-json");

  assert.throws(
    () => run(options(fixture.registryPath, homeDir, "dev", { apply: true }), {
      expectedContent: expected.expectedContent,
      repositorySkillNames: fixture.repositorySkillNames,
      execSkillsCli() {
        assert.fail("no install should run with malformed state");
      },
      output: () => {},
      now: () => new Date("2026-08-06T00:00:00.000Z")
    }),
    /refuses to overwrite invalid or untrusted reconciler state/
  );
  assert.equal(fs.readFileSync(statePath, "utf8"), "{not-json");
});

test("first-run stale copies require an explicit recoverable replacement", (t) => {
  const homeDir = temporaryDirectory(t, "global-skill-first-run-stale-home-");
  const fixture = registryFixture(t);
  const stale = expectedFixture(t, fixture.skillNames);
  const current = expectedFixture(t, fixture.skillNames);
  for (const name of fixture.skillNames) {
    fs.writeFileSync(path.join(current.sources[name], "SKILL.md"), `${name} current\n`);
    current.expectedContent[name] = hashSkillDirectory(current.sources[name]);
  }
  copySkill(stale.sources["common-alpha"], homeDir, ".agents/skills", "common-alpha");
  copySkill(stale.sources["dev-alpha"], homeDir, ".agents/skills", "dev-alpha");
  copySkill(stale.sources["dev-alpha"], homeDir, ".claude/skills", "dev-alpha");
  const quarantineRoot = path.join(homeDir, ".skill-quarantine", "first-run-replace");
  const dryOutput = [];

  assert.equal(run(options(fixture.registryPath, homeDir, "dev"), {
    expectedContent: current.expectedContent,
    repositorySkillNames: fixture.repositorySkillNames,
    quarantineRoot,
    output: (line) => dryOutput.push(line),
    now: () => new Date("2026-08-06T00:00:00.000Z")
  }), 1);
  assert.equal(
    dryOutput.some((line) => line.includes("--replace-unverified --yes")),
    true
  );

  const calls = [];
  assert.equal(run(options(fixture.registryPath, homeDir, "dev", {
    replaceUnverified: true,
    yes: true
  }), {
    expectedContent: current.expectedContent,
    repositorySkillNames: fixture.repositorySkillNames,
    quarantineRoot,
    execSkillsCli({ operation }) {
      calls.push(`${operation.root}:${operation.skill}`);
      copySkill(
        current.sources[operation.skill],
        homeDir,
        operation.root === "shared" ? ".agents/skills" : ".claude/skills",
        operation.skill
      );
    },
    output: () => {},
    now: () => new Date("2026-08-06T00:00:00.000Z")
  }), 0);
  assert.deepEqual(calls, [
    "shared:common-alpha",
    "claude:dev-alpha",
    "shared:dev-alpha"
  ]);
  assert.equal(
    fs.readFileSync(path.join(quarantineRoot, "shared/common-alpha/SKILL.md"), "utf8"),
    "common-alpha expected\n"
  );
  assert.equal(fs.existsSync(path.join(quarantineRoot, "manifest.json")), true);
  assert.equal(fs.existsSync(path.join(homeDir, ".agents/.global-skill-state.json")), true);
});

test("partial first-run replacement failures remain fully recoverable", (t) => {
  const homeDir = temporaryDirectory(t, "global-skill-partial-replace-home-");
  const fixture = registryFixture(t);
  const stale = expectedFixture(t, fixture.skillNames);
  const current = expectedFixture(t, fixture.skillNames);
  for (const name of fixture.skillNames) {
    fs.writeFileSync(path.join(current.sources[name], "SKILL.md"), `${name} current\n`);
    current.expectedContent[name] = hashSkillDirectory(current.sources[name]);
  }
  copySkill(stale.sources["common-alpha"], homeDir, ".agents/skills", "common-alpha");
  copySkill(stale.sources["dev-alpha"], homeDir, ".agents/skills", "dev-alpha");
  copySkill(stale.sources["dev-alpha"], homeDir, ".claude/skills", "dev-alpha");
  const quarantineRoot = path.join(homeDir, ".skill-quarantine", "partial-replace");
  const replaceOutput = [];
  let installAttempt = 0;

  assert.throws(
    () => run(options(fixture.registryPath, homeDir, "dev", {
      replaceUnverified: true,
      yes: true
    }), {
      expectedContent: current.expectedContent,
      repositorySkillNames: fixture.repositorySkillNames,
      quarantineRoot,
      execSkillsCli({ operation }) {
        installAttempt += 1;
        if (installAttempt === 2) throw new Error("simulated install failure");
        copySkill(
          current.sources[operation.skill],
          homeDir,
          operation.root === "shared" ? ".agents/skills" : ".claude/skills",
          operation.skill
        );
      },
      output: (line) => replaceOutput.push(line),
      now: () => new Date("2026-08-06T00:00:00.000Z")
    }),
    /simulated install failure/
  );

  const manifestPath = path.join(quarantineRoot, "manifest.json");
  assert.equal(
    replaceOutput.some((line) => line.includes(`Restore manifest: ${manifestPath}`)),
    true
  );
  for (const [root, skill] of [
    ["shared", "common-alpha"],
    ["claude", "dev-alpha"],
    ["shared", "dev-alpha"]
  ]) {
    assert.equal(fs.existsSync(path.join(quarantineRoot, root, skill, "SKILL.md")), true);
  }

  const activeReplacement = path.join(homeDir, ".agents/skills/common-alpha");
  const failedReplacementRoot = path.join(homeDir, ".failed-skill-replacements");
  fs.mkdirSync(failedReplacementRoot, { recursive: true });
  fs.renameSync(activeReplacement, path.join(failedReplacementRoot, "common-alpha"));

  assert.equal(run(options(fixture.registryPath, homeDir, null, {
    restore: manifestPath,
    yes: true
  }), { output: () => {} }), 0);
  assert.equal(
    fs.readFileSync(path.join(homeDir, ".agents/skills/common-alpha/SKILL.md"), "utf8"),
    "common-alpha expected\n"
  );
  assert.equal(
    fs.readFileSync(path.join(homeDir, ".agents/skills/dev-alpha/SKILL.md"), "utf8"),
    "dev-alpha expected\n"
  );
  assert.equal(
    fs.readFileSync(path.join(homeDir, ".claude/skills/dev-alpha/SKILL.md"), "utf8"),
    "dev-alpha expected\n"
  );
});

test("prune quarantines only a verified duplicate and restore refuses overwrite", (t) => {
  const homeDir = temporaryDirectory(t, "global-skill-prune-home-");
  const fixture = registryFixture(t);
  const expected = expectedFixture(t, fixture.skillNames);
  copySkill(expected.sources["common-alpha"], homeDir, ".agents/skills", "common-alpha");
  copySkill(expected.sources["dev-alpha"], homeDir, ".agents/skills", "dev-alpha");
  copySkill(expected.sources["dev-alpha"], homeDir, ".claude/skills", "dev-alpha");
  const legacy = copySkill(expected.sources["dev-alpha"], homeDir, ".pi/agent/skills", "dev-alpha");
  const quarantineRoot = path.join(homeDir, ".skill-quarantine", "fixed-run");
  const pruneOutput = [];

  const pruneExit = run(options(fixture.registryPath, homeDir, "dev", {
    prune: true,
    yes: true
  }), {
    expectedContent: expected.expectedContent,
    repositorySkillNames: fixture.repositorySkillNames,
    quarantineRoot,
    output: (line) => pruneOutput.push(line),
    now: () => new Date("2026-08-06T00:00:00.000Z")
  });
  assert.equal(pruneExit, 0);
  assert.equal(fs.existsSync(legacy), false);
  const manifestPath = path.join(quarantineRoot, "manifest.json");
  assert.equal(fs.existsSync(manifestPath), true);
  assert.equal(
    pruneOutput.some((line) => line.includes(`--home '${homeDir}' --yes`)),
    true
  );
  const pendingManifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
  pendingManifest.entries[0].status = "pending";
  fs.writeFileSync(manifestPath, JSON.stringify(pendingManifest, null, 2));

  const restoreExit = run(options(fixture.registryPath, homeDir, null, {
    restore: manifestPath,
    yes: true
  }), { output: () => {} });
  assert.equal(restoreExit, 0);
  assert.equal(fs.existsSync(legacy), true);

  const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
  manifest.entries[0].status = "quarantined";
  fs.mkdirSync(path.dirname(manifest.entries[0].quarantinedPath), { recursive: true });
  fs.renameSync(legacy, manifest.entries[0].quarantinedPath);
  fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2));
  writeSkill(path.dirname(legacy), path.basename(legacy), "replacement\n");
  assert.throws(
    () => run(options(fixture.registryPath, homeDir, null, {
      restore: manifestPath,
      yes: true
    }), { output: () => {} }),
    /restore destination already exists/
  );
});
