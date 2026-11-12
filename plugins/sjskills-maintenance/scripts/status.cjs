"use strict";

const DAY = 24 * 60 * 60 * 1000;
// Agent Skills names: lowercase alphanumerics and inner hyphens, at most 64
// characters. Anything else is counted, never echoed, so CLI output cannot
// place arbitrary text in the user message or agent context.
const SKILL = /^[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?$/;
const CATEGORIES = ["missing", "update", "extra", "conflict"];
const NAMES = 3;
const object = (value) => value !== null && typeof value === "object" && !Array.isArray(value);
const stable = (value) => typeof value === "string" && /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(value);

function fresh(value, now) {
  const observed = Date.parse(value?.observedAt);
  return value?.freshness === "fresh" && typeof value.cached === "boolean" &&
    Number.isFinite(observed) && observed <= now && now - observed < DAY && !value.error;
}

// Names at most NAMES skills per category (sjskills status shows five) so the
// startup context stays short however large the drift is.
function describe(findings) {
  const parts = [];
  for (const category of CATEGORIES) {
    const skills = [...new Set(findings.filter((f) => f.category === category).map((f) => f.skill))];
    if (!skills.length) continue;
    const named = skills.filter((s) => SKILL.test(s)).sort();
    const shown = named.slice(0, NAMES), rest = skills.length - shown.length;
    if (!shown.length) parts.push(`${category} ${rest} skill${rest === 1 ? "" : "s"}`);
    else parts.push(`${category} ${shown.join(", ")}${rest ? ` and ${rest} more` : ""}`);
  }
  return parts.join("; ");
}

function classifyStatus(value, now = Date.now()) {
  const findings = [], incomplete = [];
  if (!object(value) || value.operation !== "status" || value.result !== "success" || value.error) {
    incomplete.push("status response");
  }
  const cli = value?.cliAdvisory;
  if (object(cli) && cli.comparison === "update") findings.push({ summary: "CLI update available" });
  const releaseEvidence = object(cli) && (cli.comparison === "no-release"
    ? [undefined, ""].includes(cli.availableVersion) && [undefined, ""].includes(cli.releaseURL)
    : ["equal", "ahead", "update"].includes(cli.comparison) && stable(cli.availableVersion));
  if (!releaseEvidence || !stable(cli.runningVersion) || !fresh(cli, now)) incomplete.push("CLI verification");

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
      object(f) && CATEGORIES.includes(f.category) &&
      typeof f.reason === "string" && f.reason.length > 0 && typeof f.skill === "string";
    const valid = Array.isArray(advisory?.findings) ? advisory.findings.filter(validFinding) : [];
    // The review command is fixed per scope rather than copied from CLI output.
    if (valid.length) findings.push({ summary: `${scope} skills differ (${describe(valid)})`,
      review: scope === "global" ? "sjskills plan --global" : "sjskills plan" });
    if (matches.length !== 1 || !Array.isArray(advisory.findings) || valid.length !== advisory.findings.length ||
        !fresh(advisory, now)) {
      incomplete.push(`${scope} verification`);
    }
  }
  return { findings, incomplete };
}

// The user message and the agent context carry the same fixed-template facts.
// The context only lets the agent mention them and offer a sync; it names no
// mutating command and grants no authority to act without the user's consent.
// Every finding is `{ summary, review? }`; a review names its fixed plan command.
function report(results, { context = false } = {}) {
  const findings = [], seen = new Set();
  for (const finding of results.flatMap((r) => r.findings)) {
    if (!seen.has(finding.summary)) { seen.add(finding.summary); findings.push(finding); }
  }
  const incomplete = [...new Set(results.flatMap((r) => r.incomplete))];
  if (!findings.length && !incomplete.length) return null;
  const parts = findings.map((f) => f.summary);
  if (incomplete.length) parts.push(`check incomplete (${incomplete.join(", ")})`);
  const reviews = [...new Set(findings.flatMap((f) => f.review ? [f.review] : []))];
  const manual = results.some((r) => r.plugin && (r.findings.length || r.incomplete.length))
    ? "Review sjskills status and the plugin configuration manually." : "Run sjskills status manually.";
  const next = [reviews.length ? `Review with ${reviews.join(" and ")}.` : "",
    incomplete.length || findings.some((f) => !f.review) ? manual : ""].filter(Boolean).join(" ");
  const output = { systemMessage: `sjskills: ${parts.join("; ")}. ${next}` };
  if (!context) return output;
  const lines = ["sjskills startup check (read-only) results for this session:"];
  for (const f of findings) lines.push(`- ${f.summary}${f.review ? `. Review: ${f.review}` : ""}.`);
  if (incomplete.length) {
    lines.push(`- Verification did not finish (${incomplete.join(", ")}), so skill state is unconfirmed. Suggest running sjskills status.`);
  }
  lines.push(`In your first reply, briefly mention these results and offer to ${findings.length ? "sync or update" : "run sjskills status"}. Change nothing unless the user agrees.`);
  output.hookSpecificOutput = { hookEventName: "SessionStart", additionalContext: lines.join("\n") };
  return output;
}

module.exports = { classifyStatus, report, describe, DAY };
