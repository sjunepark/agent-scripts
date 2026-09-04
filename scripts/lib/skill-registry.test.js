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

test("live registry is the validated version 4 contract embedded by sjskills", () => {
  const live = liveRegistry();
  const embedded = readJSON("internal/sjskills/data/registry-v4.json");
  const repositorySkillNames = fs
    .readdirSync(path.join(repositoryRoot, "skills"), { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name);

  assert.deepEqual(live, embedded);
  assert.deepEqual(validateSkillRegistry(live, { repositorySkillNames }), []);
});

test("version 4 resolves one fixed global baseline with target exceptions", () => {
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
      "skills-cli"
    ]
  );
  assert.deepEqual(entries.find((entry) => entry.name === "codex-cleanup").targets, [".agents"]);
  assert.deepEqual(entries.find((entry) => entry.name === "clarify").targets, [".agents", ".claude"]);
  assert.ok(entries.every((entry) => entry.manager === "skills-cli" && entry.mode === "copy"));
});

test("version 4 serializes a defined source for location-free baseline managers", () => {
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

test("version 4 rejects profile selection", () => {
  assert.throws(
    () => globalSkillEntries(liveRegistry(), "dev"),
    /one fixed global baseline; profile selection is not supported/
  );
});

test("version 4 validates desired-set integrity", () => {
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

test("version 4 validates targets and installable remote sources", () => {
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

test("version 4 rejects unclassified repository skills", () => {
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
