"use strict";

const fs = require("node:fs");

const allowedTargets = new Set([".agents", ".claude"]);
const allowedManagers = new Set(["manual", "none", "skills-cli", "workflow"]);
const requiredProfiles = ["dev", "go", "kicpa", "rust"];

function isObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function arraysEqual(left, right) {
  return left.length === right.length &&
    left.every((value, index) => value === right[index]);
}

function rejectUnknownFields(value, allowed, label, errors) {
  if (!isObject(value)) return;
  for (const field of Object.keys(value)) {
    if (!allowed.has(field)) errors.push(`${label} has unsupported field ${field}`);
  }
}

function portableName(value) {
  return typeof value === "string" &&
    /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value);
}

// The Skills CLI clones git shorthand (owner/repo[/path]) or credential-free
// HTTPS remotes. Return null only for a source the adapter can use.
function skillsCliSourceProblem(location) {
  if (typeof location !== "string" || location.length === 0) return "empty";
  if (location.startsWith(".") || location.startsWith("/") || location.startsWith("~")) {
    return "local";
  }
  if (/^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+(?:\/[A-Za-z0-9_.\/-]+)?$/.test(location)) {
    const segments = location.split("/");
    if (!segments.some((segment) => segment === "." || segment === ".." || segment === "")) {
      return null;
    }
  }
  let parsed;
  try {
    parsed = new URL(location);
  } catch {
    return "unsupported";
  }
  if (parsed.protocol !== "https:" || parsed.username || parsed.password ||
      parsed.search || parsed.hash) {
    return "unsupported";
  }
  return null;
}

function validateSortedSelection(label, values, known, errors) {
  if (!Array.isArray(values) || values.length === 0) {
    errors.push(`${label} must be a non-empty array`);
    return;
  }
  if (!arraysEqual(values, [...values].sort())) errors.push(`${label} must be sorted`);
  const seen = new Set();
  for (const name of values) {
    if (!portableName(name)) errors.push(`${label} contains invalid skill ${name}`);
    if (seen.has(name)) errors.push(`${label} repeats a skill`);
    seen.add(name);
    if (!known.has(name)) errors.push(`${label} references unknown skill ${name}`);
  }
}

function validateTargets(label, targets, errors) {
  if (!Array.isArray(targets) || targets.length === 0 ||
      targets.some((target) => !allowedTargets.has(target)) ||
      new Set(targets).size !== targets.length ||
      !arraysEqual(targets, [...targets].sort())) {
    errors.push(`${label} must contain sorted unique supported targets`);
  }
}

function validateSkillRegistry(registry, options = {}) {
  const errors = [];
  if (!isObject(registry)) return ["registry must be an object"];
  rejectUnknownFields(
    registry,
    new Set(["defaults", "description", "global", "profiles", "skills", "sources",
      "targetExceptions", "version"]),
    "registry",
    errors
  );
  if (registry.version !== 4) errors.push("version must be 4");
  if (typeof registry.description !== "string" || registry.description.trim() === "") {
    errors.push("description must be a non-empty string");
  }

  if (!isObject(registry.defaults)) {
    errors.push("defaults must be an object");
  } else {
    rejectUnknownFields(registry.defaults, new Set(["targets"]), "defaults", errors);
    if (!arraysEqual(registry.defaults.targets || [], [".agents", ".claude"])) {
      errors.push("defaults.targets must be exactly: .agents, .claude");
    }
  }
  if (!isObject(registry.global)) {
    errors.push("global must be an object");
  } else {
    rejectUnknownFields(registry.global, new Set(["baseline"]), "global", errors);
  }
  if (!isObject(registry.profiles)) errors.push("profiles must be an object");
  if (!isObject(registry.sources)) errors.push("sources must be an object");
  if (!isObject(registry.targetExceptions)) errors.push("targetExceptions must be an object");
  if (!Array.isArray(registry.skills)) errors.push("skills must be an array");
  if (errors.length > 0) return errors;

  const sourceIds = Object.keys(registry.sources);
  if (!arraysEqual(sourceIds, [...sourceIds].sort())) errors.push("sources must be sorted by id");
  const sourceUsage = new Map(sourceIds.map((id) => [id, 0]));
  for (const [id, source] of Object.entries(registry.sources)) {
    const label = `source ${id}`;
    if (!isObject(source)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    rejectUnknownFields(source, new Set(["catalogPath", "kind", "location"]), label, errors);
    if (!portableName(id)) errors.push(`${label} has an invalid id`);
    if (!new Set(["external", "repository"]).has(source.kind)) {
      errors.push(`${label} has unsupported kind ${source.kind}`);
    }
    if (source.location !== undefined &&
        (typeof source.location !== "string" || source.location.length === 0)) {
      errors.push(`${label} location must be a non-empty string when present`);
    }
    if (source.kind === "repository" &&
        (!source.location || source.catalogPath !== "skills" ||
         skillsCliSourceProblem(source.location))) {
      errors.push(`${label} repository source requires a public remote location and catalogPath skills`);
    }
    if (source.kind === "external" && source.catalogPath !== undefined) {
      errors.push(`${label} external source must not define catalogPath`);
    }
  }

  const known = new Map();
  const repositoryNames = new Set();
  const skillOrder = [];
  for (const [index, skill] of registry.skills.entries()) {
    const fallback = `skills[${index}]`;
    if (!isObject(skill)) {
      errors.push(`${fallback} must be an object`);
      continue;
    }
    const label = typeof skill.name === "string" ? skill.name : fallback;
    rejectUnknownFields(
      skill,
      new Set(["fullDepth", "manager", "mode", "name", "source", "workflow"]),
      label,
      errors
    );
    if (!portableName(skill.name)) errors.push(`${label} has an invalid name`);
    if (known.has(skill.name)) errors.push(`duplicate registry skill: ${skill.name}`);
    known.set(skill.name, skill);
    skillOrder.push(skill.name);

    const source = registry.sources[skill.source];
    if (!source) {
      errors.push(`${label} references unknown source ${skill.source}`);
    } else {
      sourceUsage.set(skill.source, sourceUsage.get(skill.source) + 1);
      if (source.kind === "repository") repositoryNames.add(skill.name);
    }
    if (!allowedManagers.has(skill.manager)) {
      errors.push(`${label} has unsupported installation manager ${skill.manager}`);
    }
    if (skill.manager === "skills-cli") {
      if (skill.mode !== "copy") errors.push(`${label} skills-cli installation must use copy mode`);
      if (!source?.location || skillsCliSourceProblem(source.location)) {
        errors.push(`${label} skills-cli installation requires an installable remote source`);
      }
      if (skill.fullDepth !== undefined && typeof skill.fullDepth !== "boolean") {
        errors.push(`${label} fullDepth must be a boolean when present`);
      }
      if (skill.workflow !== undefined) {
        errors.push(`${label} skills-cli installation must not define workflow`);
      }
    } else if (skill.manager === "workflow") {
      if (typeof skill.workflow !== "string" || skill.workflow.length === 0) {
        errors.push(`${label} workflow installation must name its workflow`);
      }
      if (skill.mode !== undefined || skill.fullDepth !== undefined) {
        errors.push(`${label} workflow installation must not define copy fields`);
      }
    } else if (skill.mode !== undefined || skill.fullDepth !== undefined ||
               skill.workflow !== undefined) {
      errors.push(`${label} ${skill.manager} installation must not define copy or workflow fields`);
    }
  }
  if (!arraysEqual(skillOrder, [...skillOrder].sort())) errors.push("skills must be sorted by name");
  for (const [id, usage] of sourceUsage) {
    if (usage === 0) errors.push(`unused source: ${id}`);
  }

  const profileNames = Object.keys(registry.profiles);
  if (!arraysEqual(profileNames, requiredProfiles)) {
    errors.push(`profiles must define exactly in sorted order: ${requiredProfiles.join(", ")}`);
  }
  validateSortedSelection("global.baseline", registry.global.baseline, known, errors);
  const membership = new Map();
  const addMembership = (name, origin) => {
    const origins = membership.get(name) || [];
    origins.push(origin);
    membership.set(name, origins);
  };
  for (const name of registry.global.baseline || []) addMembership(name, "global baseline");
  for (const profileName of requiredProfiles) {
    const profile = registry.profiles[profileName];
    const label = `profile ${profileName}`;
    if (!isObject(profile)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    rejectUnknownFields(profile, new Set(["skills"]), label, errors);
    validateSortedSelection(`${label} skills`, profile.skills, known, errors);
    for (const name of profile.skills || []) addMembership(name, label);
  }
  for (const [name, origins] of membership) {
    if (origins.length > 1) {
      errors.push(`${name} appears in multiple desired sets: ${origins.join(", ")}`);
    }
    if (known.get(name)?.manager === "none") errors.push(`${name} manager none cannot be selected`);
  }

  const exceptionNames = Object.keys(registry.targetExceptions);
  if (!arraysEqual(exceptionNames, [...exceptionNames].sort())) {
    errors.push("targetExceptions must be sorted by skill name");
  }
  for (const [name, targets] of Object.entries(registry.targetExceptions)) {
    if (!known.has(name)) errors.push(`target exception references unknown skill ${name}`);
    validateTargets(`target exception ${name}`, targets, errors);
  }

  if (options.repositorySkillNames) {
    const actual = new Set(options.repositorySkillNames);
    for (const name of actual) {
      if (!repositoryNames.has(name)) errors.push(`repository skill is not classified: ${name}`);
    }
    for (const name of repositoryNames) {
      if (!actual.has(name)) {
        errors.push(`repository source references missing skills/${name}/SKILL.md`);
      }
    }
  }
  return errors;
}

function readSkillRegistry(file, options = {}) {
  let registry;
  try {
    registry = JSON.parse(fs.readFileSync(file, "utf8"));
  } catch (error) {
    throw new Error(`Could not read skill registry ${file}: ${error.message}`);
  }
  const errors = validateSkillRegistry(registry, options);
  if (errors.length > 0) {
    throw new Error(`Invalid skill registry ${file}:\n- ${errors.join("\n- ")}`);
  }
  return registry;
}

function globalSkillEntries(registry, profile) {
  if (profile !== undefined) {
    throw new Error("version 4 has one fixed global baseline; profile selection is not supported");
  }
  const byName = new Map(registry.skills.map((skill) => [skill.name, skill]));
  return registry.global.baseline.map((name) => {
    const skill = byName.get(name);
    const source = registry.sources[skill.source];
    return {
      name,
      sourceId: skill.source,
      source: source.location,
      manager: skill.manager,
      scope: "global",
      mode: skill.mode,
      targets: registry.targetExceptions[name] || registry.defaults.targets,
      fullDepth: skill.fullDepth === true
    };
  });
}

module.exports = {
  globalSkillEntries,
  readSkillRegistry,
  skillsCliSourceProblem,
  validateSkillRegistry
};
