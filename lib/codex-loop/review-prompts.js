const COMMON_REVIEW_REQUIREMENTS = [
  "Inspect the real repo state before deciding. For uncommitted changes, inspect git status, git diff, git diff --staged, and relevant untracked files.",
  "Use subagents when useful for review quality: delegate independent focus areas or broad reconnaissance to fresh-context reviewers, and ask them for concise findings with evidence, affected paths, and validation notes. Prefer the inherited model and reasoning effort unless a simple focused task clearly warrants lower effort.",
  "Distinguish issues introduced by the reviewed change from pre-existing debt.",
  "Prefer no finding over a weak finding.",
  "Do not create, read, or update a persistent review ledger.",
];

const BUCKET_DEFINITIONS = [
  "Bucket I means a concrete safe in-scope fix that can be applied without product, architecture, rollout, compatibility, churn, or risk judgment.",
  "Bucket II means a real improvement that needs user judgment before implementation.",
];

const FINDING_QUALITY_GATE =
  "Record only findings that pass this gate: discrete, actionable, material, concrete affected path, explicit trigger scenario, narrow evidence, and likely author would fix it.";

const PRIORITY_REQUIREMENT = "Use priorities P0, P1, P2, or P3.";

const DESIGN_SIGNAL_REQUIREMENT =
  "Use exactly one design signal: simple local mistake; weak validation or invariant gap; unclear ownership / boundary problem; weak type or schema model; state, lifecycle, concurrency, or ordering hazard; over-abstraction, under-abstraction, or duplicated logic; brittle integration or contract mismatch.";

function buildPlanReviewPrompt({ planPath, planRelPath, reviewPass }) {
  return `Run a strict post-implementation review on the current uncommitted changes.

Plan file: ${planPath}
Plan display path: ${planRelPath}
Review pass: ${reviewPass}

Requirements:
${renderRequirements([
  ...COMMON_REVIEW_REQUIREMENTS,
  "Review only the current uncommitted changes and nearby code needed to judge them.",
  ...BUCKET_DEFINITIONS,
  "Apply only Bucket I safe in-scope fixes that are root-cause-fixable now.",
  "Do not stage, commit, push, or implement Bucket II items.",
  PRIORITY_REQUIREMENT,
  "If Bucket I fixes were applied, set status to bucket1_applied.",
  "If no Bucket I remains and unresolved Bucket II exists, set status to bucket2_only unless it blocks remaining plan work.",
  "Mark blocks_remaining_plan true only when the decision should stop the wrapper before the next progress slice.",
  "If any Bucket II blocks remaining plan work, set status to blocked.",
  "If validation failed and should stop the wrapper, set status to validation_failed.",
  "If no meaningful issue remains, set status to clean.",
  `For Bucket II items, ${lowercaseFirst(FINDING_QUALITY_GATE)}`,
])}

Return only JSON matching the provided schema.`;
}

function buildReviewLoopReviewPrompt({ scope, iteration, reviewOnly }) {
  return `Run a strict post-implementation code review for this target:
${scope}

Iteration: ${iteration}
Mode: ${reviewOnly ? "review-only; do not edit files" : "review phase only; do not edit files"}

Requirements:
${renderRequirements([
  ...COMMON_REVIEW_REQUIREMENTS,
  "Do not stage, commit, push, or modify files in this phase.",
  ...BUCKET_DEFINITIONS,
  "Keep As-Is is for meaningful concerns rejected after inspection.",
  FINDING_QUALITY_GATE,
  PRIORITY_REQUIREMENT,
  DESIGN_SIGNAL_REQUIREMENT,
  "Set status to bucket1_candidates if any safe Bucket I candidate exists.",
  "Set status to bucket2_only if no Bucket I candidate exists and unresolved Bucket II exists.",
  "Set status to blocked when scope/context is insufficient or a blocking decision is needed.",
  "Set status to validation_failed when an existing validation failure prevents reliable review.",
  "Set status to clean when no meaningful issue remains.",
])}

Return only JSON matching the provided schema.`;
}

function buildReviewLoopFixPrompt({ scope, iteration, candidates }) {
  return `Apply only the accepted safe Bucket I fixes for this review target:
${scope}

Iteration: ${iteration}
Bucket I candidates:
${JSON.stringify(candidates, null, 2)}

Requirements:
${renderRequirements([
  "Re-verify every candidate against the real code before editing.",
  "Apply only concrete, safe, in-scope fixes that are root-cause-fixable now.",
  "Reject or skip anything requiring product, domain, architecture, rollout, compatibility, churn, or risk judgment.",
  "Keep edits tightly scoped to the reviewed change.",
  "Do not create, read, or update a persistent review ledger.",
  "Do not stage, commit, push, or implement Bucket II items.",
  "Run focused validation when practical.",
  "Set status to applied only when at least one fix was applied.",
  "Set status to nothing_applied when all candidates were rejected or already resolved.",
  "Set status to blocked when user input or external context is required.",
  "Set status to validation_failed when validation fails due to the applied fix or blocks safe continuation.",
  "Set status to failed for execution failure that should stop automation.",
])}

Return only JSON matching the provided schema.`;
}

function renderRequirements(requirements) {
  return requirements.map((requirement) => `- ${requirement}`).join("\n");
}

function lowercaseFirst(value) {
  return `${value.charAt(0).toLowerCase()}${value.slice(1)}`;
}

module.exports = {
  buildPlanReviewPrompt,
  buildReviewLoopFixPrompt,
  buildReviewLoopReviewPrompt,
};
