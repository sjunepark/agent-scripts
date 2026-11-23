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

function githubSourceProblem(source) {
  if (skillsCliSourceProblem(source)) return true;
  let path = source;
  if (source.startsWith("https://")) {
    // Match the raw authority/path too: URL() silently normalizes ports,
    // escaped traversal and backslashes that the credential boundary rejects.
    if (!source.startsWith("https://github.com/") || source.includes("%") || source.includes("\\")) return true;
    path = source.slice("https://github.com/".length);
  }
  path = path.replace(/\/$/, "");
  if (!/^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+(?:\/[A-Za-z0-9_.-]+)*$/.test(path)) return true;
  const parts = path.split("/");
  if (source.startsWith("https://") && parts.length > 2 && (parts.length < 4 || parts[2] !== "tree")) return true;
  return parts.some((part) => part === "." || part === "..") ||
    ["", ".", ".."].includes(parts[1].replace(/\.git$/, ""));
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

// Mirror of sjskills' canonical provenance identity for remote locations:
// GitHub sources reduce to github:<owner>/<repo>, other HTTPS sources keep
// their lowercase host and path.
function remoteSourceIdentity(location) {
  if (typeof location !== "string" || skillsCliSourceProblem(location)) return null;
  if (!location.includes("://")) {
    // Shorthand strips a case-sensitive .git before lowercasing, as sjskills does.
    const [owner, repository, ...rest] = location.split("/");
    return `github:${[owner, repository.replace(/\.git$/, ""), ...rest].join("/").toLowerCase()}`;
  }
  const parsed = new URL(location);
  let parts;
  try {
    parts = decodeURIComponent(parsed.pathname).split("/").filter(Boolean);
  } catch {
    return null;
  }
  const host = parsed.hostname.toLowerCase();
  if (host === "github.com" && parsed.port === "") {
    if (parts.length < 2) return null;
    return `github:${parts[0].toLowerCase()}/${parts[1].toLowerCase().replace(/\.git$/, "")}`;
  }
  // URL() drops a default :443; sjskills keeps any explicit port, written as
  // a plain decimal number.
  const authority = location.slice("https://".length).split(/[/?#]/)[0];
  const explicitPort = /:(\d+)$/.exec(authority);
  const hostPort = explicitPort ? `${host}:${Number(explicitPort[1])}` : host;
  const path = parts.join("/").replace(/\.git$/, "");
  return `https://${hostPort}${path ? `/${path}` : ""}`;
}

const formerIdentityPattern = /^github:[a-z0-9_.-]+\/[a-z0-9_.-]+(?:\/[a-z0-9_.-]+)*$/;

function canonicalFormerIdentity(identity) {
  if (typeof identity !== "string") return false;
  if (identity.startsWith("github:")) {
    return formerIdentityPattern.test(identity) &&
      !identity.split("/").some((part) => part === "." || part === "..");
  }
  return identity.startsWith("https://") && remoteSourceIdentity(identity) === identity;
}

// Former identities let installs recorded under a renamed or transferred
// repository migrate to the source's current identity. Keep them explicit and
// unambiguous: sorted canonical GitHub identities that no source currently
// uses, claimed only by sources sharing one current identity.
function formerIdentityProblems(sources) {
  const errors = [];
  const current = new Map();
  for (const [id, source] of Object.entries(sources)) {
    const identity = remoteSourceIdentity(source?.location);
    if (identity) current.set(identity, id);
  }
  const claimed = new Map();
  for (const [id, source] of Object.entries(sources)) {
    if (!isObject(source) || source.formerIdentities === undefined) continue;
    const label = `source ${id} formerIdentities`;
    const identities = source.formerIdentities;
    if (!Array.isArray(identities) || identities.length === 0) {
      errors.push(`${label} must be a non-empty array when present`);
      continue;
    }
    const identity = remoteSourceIdentity(source.location);
    if (!identity) {
      errors.push(`${label} requires a remote location`);
      continue;
    }
    if (!arraysEqual(identities, [...identities].sort())) errors.push(`${label} must be sorted`);
    if (new Set(identities).size !== identities.length) errors.push(`${label} must not repeat an identity`);
    for (const former of identities) {
      if (!canonicalFormerIdentity(former)) {
        errors.push(`${label} entry ${former} must be a canonical identity such as github:owner/repo`);
      } else if (former === identity) {
        errors.push(`${label} entry ${former} must differ from the current identity`);
      } else if (current.has(former)) {
        errors.push(`${label} entry ${former} is the current identity of source ${current.get(former)}`);
      } else if (claimed.has(former) && claimed.get(former).identity !== identity) {
        errors.push(`${label} entry ${former} is also claimed by source ${claimed.get(former).id}`);
      } else if (!claimed.has(former)) {
        claimed.set(former, { id, identity });
      }
    }
  }
  return errors;
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
  if (registry.version !== 5) errors.push("version must be 5");
  if (typeof registry.description !== "string" || registry.description.trim() === "") {
    errors.push("description must be a non-empty string");
  }

  if (!isObject(registry.defaults)) {
    errors.push("defaults must be an object");
  } else {
    rejectUnknownFields(registry.defaults, new Set(["targets"]), "defaults", errors);
    if (!Array.isArray(registry.defaults.targets) ||
        !arraysEqual(registry.defaults.targets, [".agents", ".claude"])) {
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
    rejectUnknownFields(source, new Set(["catalogPath", "formerIdentities", "kind", "location"]), label, errors);
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
  errors.push(...formerIdentityProblems(registry.sources));

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
  if (!arraysEqual(profileNames, [...profileNames].sort())) errors.push("profiles must be sorted by name");
  for (const name of requiredProfiles) {
    if (!Object.hasOwn(registry.profiles, name)) errors.push(`required profile ${name} is missing`);
  }
  validateSortedSelection("global.baseline", registry.global.baseline, known, errors);
  const membership = new Map();
  const addMembership = (name, origin) => {
    const origins = membership.get(name) || [];
    origins.push(origin);
    membership.set(name, origins);
  };
  for (const name of registry.global.baseline || []) addMembership(name, "global baseline");
  for (const profileName of profileNames) {
    const profile = registry.profiles[profileName];
    const label = `profile ${profileName}`;
    if (!isObject(profile)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    rejectUnknownFields(profile, new Set(["access", "skills"]), label, errors);
    if (!portableName(profileName)) errors.push(`${label} has an invalid name`);
    const access = profile.access === undefined ? "public" : profile.access;
    if (!["public", "github-authenticated"].includes(access)) errors.push(`${label} access must be public or github-authenticated`);
    if (requiredProfiles.includes(profileName) && access !== "public") errors.push(`${label} must remain public`);
    validateSortedSelection(`${label} skills`, profile.skills, known, errors);
    for (const name of Array.isArray(profile.skills) ? profile.skills : []) {
      addMembership(name, label);
      const skill = known.get(name);
      if (access === "github-authenticated" && skill?.manager === "skills-cli" &&
          githubSourceProblem(registry.sources[skill.source]?.location)) {
        errors.push(`${label} ${name} requires a credential-free GitHub.com repository`);
      }
    }
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
    throw new Error("version 5 has one fixed global baseline; profile selection is not supported");
  }
  const byName = new Map(registry.skills.map((skill) => [skill.name, skill]));
  return registry.global.baseline.map((name) => {
    const skill = byName.get(name);
    const source = registry.sources[skill.source];
    return {
      name,
      sourceId: skill.source,
      source: source.location || "",
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
