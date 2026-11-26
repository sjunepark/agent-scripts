"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const {
  globalSkillEntries,
  skillsCliSourceProblem,
  validateSkillRegistry
} = require("./skill-registry");

const repositoryRoot = path.resolve(__dirname, "../..");

function readJSON(relativePath) {
  return JSON.parse(fs.readFileSync(path.join(repositoryRoot, relativePath), "utf8"));
}

function liveRegistry() {
  return readJSON("skill-registry.json");
}

function validationErrors(mutate) {
  const registry = structuredClone(liveRegistry());
  mutate(registry);
  return validateSkillRegistry(registry);
}

function assertErrorIncludes(errors, expected) {
  assert.ok(
    errors.some((error) => error.includes(expected)),
    `Expected an error containing ${JSON.stringify(expected)}; got:\n${errors.join("\n")}`
  );
}

test("live registry is the validated version 5 contract published to sjskills", () => {
  const live = liveRegistry();
  const repositorySkillNames = fs
    .readdirSync(path.join(repositoryRoot, "skills"), { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name);

  assert.deepEqual(validateSkillRegistry(live, { repositorySkillNames }), []);
});

test("version 5 resolves one fixed global baseline with target exceptions", () => {
  const entries = globalSkillEntries(liveRegistry());

  assert.deepEqual(
    entries.map((entry) => entry.name),
    [
      "brainstorming",
      "clarify",
      "codex-cleanup",
      "distill-response",
      "interview",
      "next-goal",
      "pdf-to-markdown",
      "progress",
      "sjskills"
    ]
  );
  assert.deepEqual(entries.find((entry) => entry.name === "codex-cleanup").targets, [".agents"]);
  assert.deepEqual(entries.find((entry) => entry.name === "clarify").targets, [".agents", ".claude"]);
  assert.ok(entries.every((entry) => entry.manager === "skills-cli" && entry.mode === "copy"));
});

test("version 5 serializes a defined source for location-free baseline managers", () => {
  const registry = structuredClone(liveRegistry());
  const name = registry.global.baseline[0];
  registry.sources["location-free"] = { kind: "external" };
  const skill = registry.skills.find((entry) => entry.name === name);
  skill.source = "location-free";
  skill.manager = "manual";
  delete skill.mode;

  const entry = globalSkillEntries(registry).find((candidate) => candidate.name === name);
  assert.equal(entry.source, "");
  assert.equal(JSON.parse(JSON.stringify(entry)).source, "");
});

test("version 5 rejects profile selection", () => {
  assert.throws(
    () => globalSkillEntries(liveRegistry(), "dev"),
    /one fixed global baseline; profile selection is not supported/
  );
});

test("version 5 validates desired-set integrity", () => {
  assertErrorIncludes(
    validationErrors((registry) => registry.global.baseline.push("missing-skill")),
    "global.baseline references unknown skill missing-skill"
  );
  assertErrorIncludes(
    validationErrors((registry) => registry.profiles.dev.skills.push("brainstorming")),
    "brainstorming appears in multiple desired sets"
  );
  assertErrorIncludes(
    validationErrors((registry) => registry.global.baseline.push("brainstorming")),
    "global.baseline repeats a skill"
  );
});

test("version 5 validates targets and installable remote sources", () => {
  assertErrorIncludes(
    validationErrors((registry) => {
      registry.defaults.targets = "xx";
    }),
    "defaults.targets must be exactly: .agents, .claude"
  );
  assertErrorIncludes(
    validationErrors((registry) => {
      registry.targetExceptions.clarify = [".claude", ".agents"];
    }),
    "target exception clarify must contain sorted unique supported targets"
  );
  assertErrorIncludes(
    validationErrors((registry) => {
      registry.sources["agent-scripts"].location = "./skills";
    }),
    "skills-cli installation requires an installable remote source"
  );
});

test("version 5 rejects unclassified repository skills", () => {
  const registry = liveRegistry();
  const errors = validateSkillRegistry(registry, {
    repositorySkillNames: ["brainstorming", "not-classified"]
  });

  assertErrorIncludes(errors, "repository skill is not classified: not-classified");
  assertErrorIncludes(errors, "repository source references missing skills/clarify/SKILL.md");
});

test("skills-cli sources must be public git shorthands or credential-free HTTPS", () => {
  assert.equal(skillsCliSourceProblem("owner/repo"), null);
  assert.equal(skillsCliSourceProblem("owner/repo/skills/example"), null);
  assert.equal(skillsCliSourceProblem("https://github.com/owner/repo/tree/main/skills"), null);
  assert.equal(skillsCliSourceProblem("./skills"), "local");
  assert.equal(skillsCliSourceProblem("npm:package@latest"), "unsupported");
  assert.equal(skillsCliSourceProblem("https://token@example.com/repo"), "unsupported");
  assert.equal(skillsCliSourceProblem("https://example.com/repo?token=secret"), "unsupported");
});

test("access profiles remain explicit, composable, and GitHub-only", () => {
  const registry = liveRegistry();
  assert.equal(registry.profiles.kicpa.access, "public");
  assert.equal(registry.profiles["kicpa-private"].access, "github-authenticated");
  assert.equal(registry.profiles.kicpa.skills.length, 10);
  assert.equal(registry.profiles["kicpa-private"].skills.length, 10);
  for (const value of [null, "", "private", true, 1, [], {}]) {
    assertErrorIncludes(validationErrors((r) => r.profiles["kicpa-private"].access = value), "access must be");
  }
  assertErrorIncludes(validationErrors((r) => r.profiles.kicpa.access = "github-authenticated"), "must remain public");
  for (const source of ["https://example.com/o/r", "https://github.com:443/o/r", "https://github.com/o/r/%2e%2e", "https://github.com/o/r/../a", "https://github.com/o/r/a%2Fb", "owner/.git", "https://github.com/o/r/blob/main/skills", "https://github.com/o/r/tree", "https://github.com/o/r/issues"]) {
    assertErrorIncludes(validationErrors((r) => r.sources.kicpa.location = source), "requires a credential-free GitHub.com");
  }
  delete registry.profiles.kicpa.access;
  assert.deepEqual(validateSkillRegistry(registry), []);
  registry.profiles["other-private"] = registry.profiles["kicpa-private"];
  delete registry.profiles["kicpa-private"];
  registry.profiles = Object.fromEntries(Object.entries(registry.profiles).sort(([a],[b]) => a.localeCompare(b)));
  assert.deepEqual(validateSkillRegistry(registry), []);
});

test("version 5 validates registry-declared former source identities", () => {
  const setFormer = (former, location) => (registry) => {
    if (location !== undefined) registry.sources.kicpa.location = location;
    registry.sources.kicpa.formerIdentities = former;
  };
  assert.deepEqual(validationErrors(setFormer(["github:sjunepark/kicpa"])), []);
  for (const [former, expected, location] of [
    [[], "must be a non-empty array"],
    ["github:sjunepark/kicpa", "must be a non-empty array"],
    [["github:b/kicpa", "github:a/kicpa"], "must be sorted"],
    [["github:sjunepark/kicpa", "github:sjunepark/kicpa"], "must not repeat"],
    [["sjunepark/kicpa"], "must be a canonical identity"],
    [["github:SJunePark/kicpa"], "must be a canonical identity"],
    [["local:./kicpa"], "must be a canonical identity"],
    [["github:cpaikr/kicpa"], "must differ from the current identity"],
    [["github:cpaikr/kasb"], "is the current identity of source kasb"],
    [["github:sjunepark/kicpa"], "requires a remote location", ""]
  ]) {
    assertErrorIncludes(validationErrors(setFormer(former, location)), expected);
  }
  assertErrorIncludes(validationErrors((registry) => {
    registry.sources.kicpa.formerIdentities = ["github:sjunepark/kicpa"];
    registry.sources.kasb.formerIdentities = ["github:sjunepark/kicpa"];
  }), "is also claimed by source");
});

test("version 5 former identities compare identities exactly as sjskills derives them", () => {
  // Go keeps an uppercase .GIT suffix in shorthand, so these two sources have
  // different current identities and cannot share a former identity.
  assertErrorIncludes(validationErrors((registry) => {
    registry.sources.kicpa.location = "Acme/Tools.GIT";
    registry.sources.kicpa.formerIdentities = ["github:old/tools"];
    registry.sources.kasb.location = "acme/tools.git";
    registry.sources.kasb.formerIdentities = ["github:old/tools"];
  }), "is also claimed by source");
  // Explicit ports are kept but normalized to a plain decimal number.
  assertErrorIncludes(validationErrors((registry) => {
    registry.sources.kicpa.location = "https://example.com:0443/repo";
    registry.sources.kicpa.formerIdentities = ["https://example.com:0443/old"];
  }), "must be a canonical identity");
  // (kicpa-private's GitHub-only access rule still rejects this host.)
  assert.deepEqual(validationErrors((registry) => {
    registry.sources.kicpa.location = "https://example.com:0443/repo";
    registry.sources.kicpa.formerIdentities = ["https://example.com:443/old"];
  }).filter((error) => error.includes("formerIdentities")), []);
  // A malformed path escape yields validation results instead of throwing.
  assert.doesNotThrow(() => validationErrors((registry) => {
    registry.sources.kicpa.location = "https://example.com/%";
  }));
});
