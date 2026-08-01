"use strict";

const fs = require("fs");

const allowedAgents = new Set(["claude-code", "codex", "pi"]);
const allowedInstallationManagers = new Set(["manual", "none", "skills-cli", "workflow"]);
const allowedInstallationModes = new Set(["copy", "symlink"]);
const allowedRecommendationScopes = new Set(["catalog", "global", "project"]);
const allowedSourceKinds = new Set(["external", "repository"]);
const topLevelFields = new Set(["description", "global", "skills", "sources", "version"]);
const globalFields = new Set(["allowUnlistedSkills"]);
const sourceFields = new Set(["catalogPath", "kind", "location"]);
const skillFields = new Set(["installation", "name", "recommendation", "source"]);
const recommendationFields = new Set(["agents", "scope", "when"]);
const installationFields = new Set(["fullDepth", "manager", "mode", "workflow"]);

function isObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function rejectUnknownFields(value, allowed, label, errors) {
  if (!isObject(value)) return;
  for (const field of Object.keys(value)) {
    if (!allowed.has(field)) errors.push(`${label} has unsupported field ${field}`);
  }
}

function validateSkillRegistry(registry, options = {}) {
  const errors = [];
  const repositorySkillNames = options.repositorySkillNames
    ? new Set(options.repositorySkillNames)
    : null;

  if (!isObject(registry)) return ["registry must be an object"];
  rejectUnknownFields(registry, topLevelFields, "registry", errors);

  if (registry.version !== 1) errors.push("version must be 1");
  if (typeof registry.description !== "string" || registry.description.length === 0) {
    errors.push("description must be a non-empty string");
  }

  if (!isObject(registry.global)) {
    errors.push("global must be an object");
  } else {
    rejectUnknownFields(registry.global, globalFields, "global", errors);
    if (typeof registry.global.allowUnlistedSkills !== "boolean") {
      errors.push("global.allowUnlistedSkills must be a boolean");
    }
  }

  if (!isObject(registry.sources)) {
    errors.push("sources must be an object keyed by source id");
    return errors;
  }
  if (!Array.isArray(registry.skills)) {
    errors.push("skills must be an array");
    return errors;
  }

  const sourceIds = Object.keys(registry.sources);
  const sortedSourceIds = [...sourceIds].sort();
  if (sourceIds.some((sourceId, index) => sourceId !== sortedSourceIds[index])) {
    errors.push("sources must be sorted by id");
  }

  const sourceUsage = new Map(sourceIds.map((sourceId) => [sourceId, 0]));
  for (const [sourceId, source] of Object.entries(registry.sources)) {
    const label = `source ${sourceId}`;
    if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(sourceId)) {
      errors.push(`${label} id must contain lowercase letters, numbers, and single hyphens only`);
    }
    if (!isObject(source)) {
      errors.push(`${label} must be an object`);
      continue;
    }
    rejectUnknownFields(source, sourceFields, label, errors);
    if (!allowedSourceKinds.has(source.kind)) {
      errors.push(`${label} has unsupported kind ${source.kind}`);
    }
    if (source.location !== undefined &&
        (typeof source.location !== "string" || source.location.length === 0)) {
      errors.push(`${label} location must be a non-empty string when present`);
    }
    if (source.kind === "repository") {
      if (source.catalogPath !== "skills") {
        errors.push(`${label} must use catalogPath skills`);
      }
      if (!source.location) errors.push(`${label} must record its published location`);
    } else if (source.catalogPath !== undefined) {
      errors.push(`${label} external source must not define catalogPath`);
    }
  }

  const registeredNames = new Set();
  const repositoryNames = new Set();
  const registryOrder = [];

  for (const [index, skill] of registry.skills.entries()) {
    const fallbackLabel = `skills[${index}]`;
    if (!isObject(skill)) {
      errors.push(`${fallbackLabel} must be an object`);
      continue;
    }
    const label = typeof skill.name === "string" ? skill.name : fallbackLabel;
    rejectUnknownFields(skill, skillFields, label, errors);

    if (typeof skill.name !== "string" || !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(skill.name)) {
      errors.push(`${fallbackLabel}.name must be a portable skill name`);
      continue;
    }
    registryOrder.push(skill.name);
    if (registeredNames.has(skill.name)) errors.push(`duplicate registry skill: ${skill.name}`);
    registeredNames.add(skill.name);

    const source = registry.sources[skill.source];
    if (!source) {
      errors.push(`${label} references unknown source ${skill.source}`);
    } else {
      sourceUsage.set(skill.source, sourceUsage.get(skill.source) + 1);
      if (source.kind === "repository") repositoryNames.add(skill.name);
    }

    const recommendation = skill.recommendation;
    if (!isObject(recommendation)) {
      errors.push(`${label} must define recommendation`);
    } else {
      rejectUnknownFields(recommendation, recommendationFields, `${label} recommendation`, errors);
      if (!allowedRecommendationScopes.has(recommendation.scope)) {
        errors.push(`${label} has unsupported recommendation scope ${recommendation.scope}`);
      }
      if (recommendation.scope === "global" || recommendation.scope === "project") {
        if (!Array.isArray(recommendation.agents) || recommendation.agents.length === 0) {
          errors.push(`${label} ${recommendation.scope} recommendation must name agents`);
        }
      }
      if (recommendation.agents !== undefined) {
        if (!Array.isArray(recommendation.agents)) {
          errors.push(`${label} recommendation agents must be an array when present`);
        } else {
          const seenAgents = new Set();
          for (const agent of recommendation.agents) {
            if (!allowedAgents.has(agent)) errors.push(`${label} has unsupported agent ${agent}`);
            if (seenAgents.has(agent)) errors.push(`${label} repeats agent ${agent}`);
            seenAgents.add(agent);
          }
        }
      }
      if (recommendation.scope === "project") {
        if (typeof recommendation.when !== "string" || recommendation.when.length === 0) {
          errors.push(`${label} project recommendation must explain when it applies`);
        }
      } else if (recommendation.when !== undefined) {
        errors.push(`${label} ${recommendation.scope} recommendation must not define when`);
      }
    }

    const installation = skill.installation;
    if (!isObject(installation)) {
      errors.push(`${label} must define installation`);
      continue;
    }
    rejectUnknownFields(installation, installationFields, `${label} installation`, errors);
    if (!allowedInstallationManagers.has(installation.manager)) {
      errors.push(`${label} has unsupported installation manager ${installation.manager}`);
    }
    if (installation.mode !== undefined && !allowedInstallationModes.has(installation.mode)) {
      errors.push(`${label} has unsupported installation mode ${installation.mode}`);
    }
    if (installation.fullDepth !== undefined && typeof installation.fullDepth !== "boolean") {
      errors.push(`${label} installation fullDepth must be a boolean when present`);
    }
    if (installation.manager === "skills-cli") {
      if (!source?.location) errors.push(`${label} uses skills-cli but its source has no location`);
      if (!installation.mode) errors.push(`${label} skills-cli installation must define mode`);
      if (recommendation?.scope === "global" && installation.mode !== "copy") {
        errors.push(`${label} global skills-cli installation must use copy mode`);
      }
    } else {
      if (installation.mode !== undefined) {
        errors.push(`${label} ${installation.manager} installation must not define mode`);
      }
      if (installation.fullDepth !== undefined) {
        errors.push(`${label} ${installation.manager} installation must not define fullDepth`);
      }
    }
    if (installation.manager === "workflow") {
      if (typeof installation.workflow !== "string" || installation.workflow.length === 0) {
        errors.push(`${label} workflow installation must name its workflow`);
      }
      if (recommendation?.scope !== "project") {
        errors.push(`${label} workflow installation must be project-scoped`);
      }
    } else if (installation.workflow !== undefined) {
      errors.push(`${label} ${installation.manager} installation must not define workflow`);
    }
    if (recommendation?.scope === "catalog" && installation.manager !== "none") {
      errors.push(`${label} catalog-only recommendation must use installation manager none`);
    }
    if (installation.manager === "none" && recommendation?.scope !== "catalog") {
      errors.push(`${label} installation manager none requires catalog scope`);
    }
  }

  for (const [sourceId, usage] of sourceUsage) {
    if (usage === 0) errors.push(`unused source: ${sourceId}`);
  }
  if (repositorySkillNames) {
    for (const name of repositorySkillNames) {
      if (!repositoryNames.has(name)) errors.push(`repository skill is not classified: ${name}`);
    }
    for (const name of repositoryNames) {
      if (!repositorySkillNames.has(name)) {
        errors.push(`repository source references missing skills/${name}/SKILL.md`);
      }
    }
  }

  const sortedOrder = [...registryOrder].sort();
  if (registryOrder.some((name, index) => name !== sortedOrder[index])) {
    errors.push("skills must be sorted by name");
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

function globalSkillEntries(registry) {
  return registry.skills
    .filter((skill) => skill.recommendation.scope === "global")
    .map((skill) => {
      const source = registry.sources[skill.source];
      return {
        name: skill.name,
        sourceId: skill.source,
        source: source.location,
        manager: skill.installation.manager,
        scope: skill.recommendation.scope,
        mode: skill.installation.mode,
        agents: skill.recommendation.agents,
        fullDepth: skill.installation.fullDepth === true
      };
    });
}

module.exports = {
  globalSkillEntries,
  readSkillRegistry,
  validateSkillRegistry
};
