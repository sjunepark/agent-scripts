"use strict";

const DAY = 24 * 60 * 60 * 1000;
const object = (value) => value !== null && typeof value === "object" && !Array.isArray(value);
const stable = (value) => typeof value === "string" && /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(value);

function fresh(value, now) {
  const observed = Date.parse(value?.observedAt);
  return value?.freshness === "fresh" && typeof value.cached === "boolean" &&
    Number.isFinite(observed) && observed <= now && now - observed < DAY && !value.error;
}

function classifyStatus(value, now = Date.now()) {
  const findings = [], incomplete = [];
  if (!object(value) || value.operation !== "status" || value.result !== "success" || value.error) {
    incomplete.push("status response");
  }
  const cli = value?.cliAdvisory;
  if (object(cli) && cli.comparison === "update") findings.push("CLI update available");
  if (!object(cli) || !["equal", "ahead", "update"].includes(cli.comparison) ||
      !stable(cli.runningVersion) || !stable(cli.availableVersion) || !fresh(cli, now)) incomplete.push("CLI verification");

  const configuration = value?.status?.projectConfiguration;
  if (!["configured", "not-configured", "unavailable"].includes(configuration)) incomplete.push("project configuration");
  else if (configuration === "unavailable") incomplete.push("project configuration");
  const required = configuration === "not-configured" ? ["global"] : ["global", "project"];
  const advisories = Array.isArray(value?.advisories) ? value.advisories : [];
  if (advisories.some((a) => !object(a) || !required.includes(a.scope))) incomplete.push("scope response");
  for (const scope of required) {
    const matches = advisories.filter((a) => a?.scope === scope);
    const advisory = matches[0];
    const validFinding = (f) =>
      object(f) && ["update", "missing", "extra", "conflict"].includes(f.category) &&
      typeof f.reason === "string" && f.reason.length > 0 && typeof f.skill === "string";
    const validFindings = Array.isArray(advisory?.findings) && advisory.findings.every(validFinding);
    if (Array.isArray(advisory?.findings) && advisory.findings.some(validFinding)) findings.push(`${scope} skills differ`);
    if (matches.length !== 1 || !validFindings || !fresh(advisory, now)) incomplete.push(`${scope} verification`);
  }
  return { findings, incomplete };
}

function report(results) {
  const findings = [...new Set(results.flatMap((r) => r.findings))];
  const incomplete = [...new Set(results.flatMap((r) => r.incomplete))];
  if (!findings.length && !incomplete.length) return null;
  const parts = [...findings];
  if (incomplete.length) parts.push(`check incomplete (${incomplete.join(", ")})`);
  const next = results.some((r) => r.plugin && (r.findings.length || r.incomplete.length))
    ? "Review sjskills status and the plugin configuration manually." : "Run sjskills status manually.";
  return { systemMessage: `sjskills: ${parts.join("; ")}. ${next}` };
}

module.exports = { classifyStatus, report, DAY };
