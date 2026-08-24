"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const {
  globalSkillEntries,
  validateSkillRegistry
} = require("./skill-registry");

function validRegistry() {
  return {
    version: 3,
    description: "Test registry.",
    global: {
      allowUnlistedSkills: false,
      profiles: {
        dev: { audiences: ["common", "dev"] },
        kicpa: { audiences: ["common", "kicpa"] }
      }
    },
    sources: {
      test: {
        kind: "external",
        location: "example/test-skills"
      }
    },
    skills: [
      {
        name: "common-alpha",
        source: "test",
        recommendation: {
          scope: "global",
          audience: "common",
          targets: [".agents"]
        },
        installation: { manager: "skills-cli", mode: "copy" }
      },
      {
        name: "dev-alpha",
        source: "test",
        recommendation: {
          scope: "global",
          audience: "dev",
          targets: [".agents"]
        },
        installation: { manager: "skills-cli", mode: "copy" }
      },
      {
        name: "kicpa-alpha",
        source: "test",
        recommendation: {
          scope: "global",
          audience: "kicpa",
          targets: [".agents"]
        },
        installation: { manager: "skills-cli", mode: "copy" }
      },
      {
        name: "project-alpha",
        source: "test",
        recommendation: {
          scope: "project",
          targets: [".agents"],
          when: "The target project opts in."
        },
        installation: { manager: "skills-cli", mode: "symlink" }
      }
    ]
  };
}

function validationErrors(mutate) {
  const registry = validRegistry();
  mutate(registry);
  return validateSkillRegistry(registry);
}

function assertErrorIncludes(errors, expected) {
  assert.ok(
    errors.some((error) => error.includes(expected)),
    `Expected an error containing ${JSON.stringify(expected)}; got:\n${errors.join("\n")}`
  );
}

test("version 3 registry resolves the dev and kicpa profile unions", () => {
  const registry = validRegistry();
  const devEntries = globalSkillEntries(registry, "dev");

  assert.deepEqual(validateSkillRegistry(registry), []);
  assert.deepEqual(
    devEntries.map((entry) => entry.name),
    ["common-alpha", "dev-alpha"]
  );
  assert.deepEqual(
    devEntries.map(({ name, targets }) => ({ name, targets })),
    [
      { name: "common-alpha", targets: [".agents"] },
      { name: "dev-alpha", targets: [".agents"] }
    ]
  );
  assert.deepEqual(
    globalSkillEntries(registry, "kicpa").map((entry) => entry.name),
    ["common-alpha", "kicpa-alpha"]
  );
});

test("live version 3 global profiles match the staged version 4 baseline unions", () => {
  const root = path.resolve(__dirname, "../..");
  const live = JSON.parse(fs.readFileSync(path.join(root, "skill-registry.json"), "utf8"));
  const staged = JSON.parse(fs.readFileSync(path.join(root, "internal/sjskills/data/registry-v4.json"), "utf8"));
  const stagedByName = new Map(staged.skills.map((skill) => [skill.name, skill]));

  function stagedEntry(name) {
    const skill = stagedByName.get(name);
    const targets = staged.targetExceptions?.[name] ?? staged.defaults.targets;
    return {
      name,
      source: staged.sources[skill.source].location,
      manager: skill.manager,
      mode: skill.mode,
      targets,
      fullDepth: skill.fullDepth === true
    };
  }

  function comparableLiveEntry(entry) {
    return {
      name: entry.name,
      source: entry.source,
      manager: entry.manager,
      mode: entry.mode,
      targets: entry.targets,
      fullDepth: entry.fullDepth
    };
  }

  for (const profile of ["dev", "kicpa"]) {
    const names = [...staged.global.baseline, ...staged.profiles[profile].skills]
      .sort();
    assert.deepEqual(
      globalSkillEntries(live, profile).map(comparableLiveEntry),
      names.map(stagedEntry),
      `${profile} profile drifted across the version 3/version 4 boundary`
    );
  }
});

test("profile resolution requires an explicit supported profile", () => {
  const registry = validRegistry();

  assert.throws(
    () => globalSkillEntries(registry),
    /profile is required; expected one of: dev, kicpa/
  );
  assert.throws(
    () => globalSkillEntries(registry, "other"),
    /unknown profile other; expected one of: dev, kicpa/
  );
});

test("older registry versions are rejected", () => {
  const errors = validationErrors((registry) => {
    registry.version = 2;
  });

  assertErrorIncludes(errors, "version must be 3");
});

test("strict desired state cannot allow unlisted global skills", () => {
  const errors = validationErrors((registry) => {
    registry.global.allowUnlistedSkills = true;
  });

  assertErrorIncludes(errors, "global.allowUnlistedSkills must be false for strict desired state");
});

test("both supported profiles are required and unknown profiles are rejected", () => {
  const missing = validationErrors((registry) => {
    delete registry.global.profiles.kicpa;
  });
  const unknown = validationErrors((registry) => {
    registry.global.profiles.other = { audiences: ["common"] };
  });

  assertErrorIncludes(missing, "global.profiles must define exactly: dev, kicpa");
  assertErrorIncludes(unknown, "global.profiles must define exactly: dev, kicpa");
});

test("profile audiences must be nonempty, sorted, and deduplicated", () => {
  const empty = validationErrors((registry) => {
    registry.global.profiles.dev.audiences = [];
  });
  const unsorted = validationErrors((registry) => {
    registry.global.profiles.dev.audiences = ["dev", "common"];
  });
  const duplicate = validationErrors((registry) => {
    registry.global.profiles.dev.audiences = ["common", "dev", "dev"];
  });

  assertErrorIncludes(empty, "global profile dev audiences must be a non-empty array");
  assertErrorIncludes(unsorted, "global profile dev audiences must be sorted");
  assertErrorIncludes(duplicate, "global profile dev repeats audience dev");
});

test("profile compositions are fixed to common plus their machine audience", () => {
  const wrongDev = validationErrors((registry) => {
    registry.global.profiles.dev.audiences = ["dev"];
  });
  const unsupported = validationErrors((registry) => {
    registry.global.profiles.kicpa.audiences = ["common", "finance"];
  });

  assertErrorIncludes(wrongDev, "global profile dev audiences must be: common, dev");
  assertErrorIncludes(unsupported, "global profile kicpa has unsupported audience finance");
});

test("every global recommendation has exactly one supported audience", () => {
  const missing = validationErrors((registry) => {
    delete registry.skills[0].recommendation.audience;
  });
  const unsupported = validationErrors((registry) => {
    registry.skills[0].recommendation.audience = "finance";
  });
  const multiple = validationErrors((registry) => {
    registry.skills[0].recommendation.audience = ["common", "dev"];
  });

  assertErrorIncludes(missing, "common-alpha global recommendation must name one audience");
  assertErrorIncludes(unsupported, "common-alpha has unsupported audience finance");
  assertErrorIncludes(multiple, "common-alpha global recommendation must name one audience");
});

test("global and project recommendations use sorted installation targets", () => {
  const legacy = validationErrors((registry) => {
    delete registry.skills[0].recommendation.targets;
    registry.skills[0].recommendation.agents = ["codex"];
  });
  const unsupported = validationErrors((registry) => {
    registry.skills[0].recommendation.targets = ["codex"];
  });
  const duplicate = validationErrors((registry) => {
    registry.skills[0].recommendation.targets = [".agents", ".agents"];
  });
  const unsorted = validationErrors((registry) => {
    registry.skills[0].recommendation.targets = [".claude", ".agents"];
  });
  const missingProjectTargets = validationErrors((registry) => {
    delete registry.skills[3].recommendation.targets;
  });

  assertErrorIncludes(legacy, "common-alpha recommendation has unsupported field agents");
  assertErrorIncludes(legacy, "common-alpha global recommendation must name targets");
  assertErrorIncludes(unsupported, "common-alpha has unsupported target codex");
  assertErrorIncludes(duplicate, "common-alpha repeats target .agents");
  assertErrorIncludes(unsorted, "common-alpha recommendation targets must be sorted");
  assertErrorIncludes(
    missingProjectTargets,
    "project-alpha project recommendation must name targets"
  );
});

test("skills-cli installation requires a source the Skills CLI can clone", () => {
  const npmSpecifier = validationErrors((registry) => {
    registry.sources.test.location = "npm:test-skills@latest";
  });
  const localPath = validationErrors((registry) => {
    registry.sources.test.location = "./skills";
  });
  const credentialed = validationErrors((registry) => {
    registry.sources.test.location = "https://token@example.com/test-skills";
  });

  assertErrorIncludes(
    npmSpecifier,
    "common-alpha skills-cli installation requires a git shorthand or credential-free https source location, not npm:test-skills@latest"
  );
  // The check is total: project-scoped records never reach the reconciler's
  // runtime guard, so validation is their only protection.
  assertErrorIncludes(npmSpecifier, "project-alpha skills-cli installation requires");
  assertErrorIncludes(
    localPath,
    "common-alpha skills-cli installation requires a remote source, not the local path ./skills"
  );
  assertErrorIncludes(
    credentialed,
    "common-alpha skills-cli installation requires a git shorthand or credential-free https source location"
  );
});

test("non-skills-cli managers may provision from a source the Skills CLI cannot clone", () => {
  const registry = validRegistry();
  registry.sources = {
    installer: { kind: "external", location: "npm:installer@latest" },
    test: registry.sources.test
  };
  registry.skills.push({
    name: "workflow-alpha",
    source: "installer",
    recommendation: {
      scope: "project",
      targets: [".claude"],
      when: "The delegating workflow operates in a matching repository."
    },
    installation: { manager: "workflow", workflow: "delegate-ui-to-claude" }
  });

  assert.deepEqual(validateSkillRegistry(registry), []);
});

test("project and catalog recommendations cannot declare audiences", () => {
  const project = validationErrors((registry) => {
    registry.skills[3].recommendation.audience = "dev";
  });
  const catalog = validationErrors((registry) => {
    registry.skills[3].recommendation = {
      scope: "catalog",
      audience: "dev"
    };
    registry.skills[3].installation = { manager: "none" };
  });

  assertErrorIncludes(project, "project-alpha project recommendation must not define audience");
  assertErrorIncludes(catalog, "project-alpha catalog recommendation must not define audience");
});

test("catalog recommendations cannot declare installation targets", () => {
  const errors = validationErrors((registry) => {
    registry.skills[3].recommendation = {
      scope: "catalog",
      targets: [".agents"]
    };
    registry.skills[3].installation = { manager: "none" };
  });

  assertErrorIncludes(errors, "project-alpha catalog recommendation must not define targets");
});
