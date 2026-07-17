# Next Goal Skill

Status: completed 2026-07-13.

Next action: none. Publication or global installation remains a separately authorized action.

## Current state

- `skills/next-goal` is a read-only, manually invoked skill that chooses work substantial
  enough to benefit from `/goal` and returns a concise routing prompt.
- It treats plan slices as building blocks rather than default stopping boundaries. It prefers
  an active phase, milestone, or several connected slices and says `/goal` is not warranted
  when only ordinary-turn work remains.
- Generated prompts point the next agent to authoritative plans and `AGENTS.md` instead of
  restating implementation steps, file inventories, invariants, test matrices, commands, or
  detailed git choreography. They still require coherent local commits as meaningful passing
  units finish, with messages that state what changed and why, including non-obvious decisions.
- A post-revision forward test against `../creo-valuation` selected the whole remaining M0
  milestone instead of only Slice 4B.3 and produced a two-paragraph prompt without writing to
  the target repository.
- Goal selection remains in `SKILL.md`; the PR/no-PR prompt contract now lives in
  `prompts/delivery-variants.md` and loads only after `/goal` is warranted. No scripts, assets,
  or eval file are needed, and this repository has no evaluation runner for `evals/evals.json`.
- Skill validation, catalog validation, direct and full local-source listing, and whitespace
  checks pass.

## Purpose

Create a manually invoked `$next-goal` skill that inspects a repository's plans and current
state, selects an appropriately substantial next implementation outcome, and emits a concise
prompt that routes a fresh `/goal` session through the repository's existing plans.

The skill is a read-only handoff generator. It does not execute the work, create a goal,
change a plan, or mutate the target repository.

## Existing seams

Use the repository's existing skills instead of duplicating their responsibilities:

- `progress-run` executes one substantial connected plan slice; `next-goal` instead chooses a
  larger persistent-execution boundary that may contain several such slices.
- `progress-handoff` supplies fresh-session handoff principles, but writes a durable plan;
  `next-goal` instead returns a prompt in the response.
- `progress-doc` supplies the requirement to verify plan claims against repository evidence.

Do not hardcode Creo Valuation, M0, manifests, SQLite, or any other one-repository concept.

## Deliverables

- [x] `skills/next-goal/SKILL.md`
- [x] `skills/next-goal/prompts/delivery-variants.md`
- [x] `skills/next-goal/agents/openai.yaml`
- [x] Evaluation decision: omit `evals/evals.json` because the repository has no workflow that
  runs it

Start the skill with the installed OpenAI `skill-creator` initializer. Keep scope selection's
judgment-heavy guidance in `SKILL.md`; disclose the delivery-variant branch rather than loading
it when no substantial goal is available. Do not add a README, scripts, assets, or evals without
a concrete need.

## Durable behavior

### 1. Establish the real current state

- Read every applicable `AGENTS.md`.
- Resolve the authoritative active `PLAN`, `TODO`, `ROADMAP`, progress, or handoff document.
- Inspect git status and recent history.
- Read only enough implementation and validation evidence to detect completed or stale plan
  items and identify real prerequisites.
- Treat unrelated working-tree changes as intentional and preserve them.

If no plan exists, infer the next outcome from repository instructions, current code, tests,
and recent history. Ask a question only when multiple choices would materially change the
outcome or no safe, unblocked implementation goal can be selected.

### 2. Select a substantial outcome

Choose the largest useful outcome a persistent agent can pursue autonomously. Prefer a whole
active phase or milestone, or several adjacent plan slices with settled requirements. Do not
stop at a plan heading merely because it names a reviewable slice.

A good goal:

- reaches a concrete, demonstrable project or user outcome;
- is materially larger than one ordinary interactive turn;
- benefits from persistent execution across multiple checkpoints;
- has a crisp stopping condition and an obvious next excluded milestone.

Split before:

- a genuine unresolved product or technical decision;
- required external authorization or coordination;
- a real blocker or materially unrelated next milestone.

Treat slices and checklist items as planning units rather than default goal boundaries. Fold in
naturally connected later work. If no substantial unblocked outcome remains, report that
`/goal` is not warranted instead of manufacturing a small prompt.

### 3. Generate the fresh-session prompt

Return:

1. A short recommended boundary and rationale.
2. One copy-paste prompt intended as the body entered after `/goal`.
3. A short statement of the next excluded area, when useful.

Keep the generated prompt to one to three short paragraphs in the normal case. Name the
outcome, stopping boundary, and few authoritative plan documents, then direct the next agent to
follow `AGENTS.md`, preserve unrelated changes, keep plans current, and use repository-required
validation and review. Require coherent local commits as meaningful passing units finish and
commit messages that state what changed and why, including non-obvious decisions. Add detail
only when it is absent from the plans and essential to avoid ambiguity or unsafe work.

Do not copy plan content into the prompt. Leave implementation steps, detailed design rules,
file lists, acceptance matrices, commands, detailed git workflow, and progress recaps to the
goal-running agent and repository documents. Do not invent a token budget.

### 4. Remain read-only

While producing the prompt, do not:

- implement the selected work;
- edit the repository or its planning documents;
- create or update a `/goal` in the current session;
- stage, commit, push, open a PR, or merge;
- turn uncertain design choices into instructions without evidence.

## Skill metadata

- Name and folder: `next-goal`.
- Suggested display name: `Next Goal`.
- Short description: `Choose a substantial implementation goal`.
- The default prompt explicitly invokes `$next-goal` and asks for a concise, substantial
  `/goal` prompt.
- Set `policy.allow_implicit_invocation: false` so this user-invoked workflow does not compete
  with `progress-run`, `progress-handoff`, or ordinary planning requests.
- Keep all trigger guidance in the `SKILL.md` frontmatter description and keep the body
  procedural and concise.

Read the skill-creator `openai_yaml` reference before generating metadata. Generate
`agents/openai.yaml` with the skill-creator tooling rather than hand-writing stale metadata.

## Forward tests

Forward-test the completed skill in fresh subagent contexts without leaking the expected
answer. Cover at least:

1. One repository with a clear active plan and several adjacent checklist items.
2. One repository where plan claims conflict with current code or recent commits.
3. One dirty repository or repository with multiple plausible next areas.

For each result, verify that:

- the chosen scope is substantial enough to warrant persistent `/goal` execution;
- adjacent plan slices are included unless a genuine boundary prevents it;
- the prompt is actionable through its repository references without copying their details;
- the normal prompt is one to three short paragraphs;
- repository-specific instructions survive into the prompt without generic hardcoding;
- the skill performed no writes or goal creation.

Revise and repeat any case that exposes missing context, over-broad scope, plan-trusting, or
unearned prompt boilerplate.

## Acceptance criteria

- The skill reliably chooses one evidence-backed, substantial implementation outcome or says
  `/goal` is not warranted when no such outcome exists.
- Its output can be pasted into a new `/goal` session without manual reconstruction.
- It delegates execution detail to the active plans, `AGENTS.md`, and the goal-running agent.
- It requires progressive, coherent local commits with messages that state what changed, why,
  and any non-obvious decisions needed to understand each completed unit later.
- It remains read-only and does not overlap the execution responsibilities of existing
  progress skills.
- Its `SKILL.md`, disclosed prompt contract, and metadata are concise, valid, and free of
  repository-specific assumptions.
- All forward tests and repository validation pass.

## Validation

Run:

```sh
uv run --with pyyaml python ~/.codex/skills/.system/skill-creator/scripts/quick_validate.py skills/next-goal
scripts/validate-skills
bunx skills add ./skills/next-goal --list
bunx skills add ./skills --list
git diff --check
git diff --check <starting-ref>..HEAD
```

Then run `$code-review` in implementation-review mode with the diet lens over
`<starting-ref>..HEAD`, apply safe findings, and re-run affected validation. Do not publish or
globally install the skill as part of this ticket unless separately requested.

## Progress log

### 2026-07-13

- Initialized `next-goal` with the installed OpenAI skill-creator, then replaced the template
  with the evidence, goal-sizing, prompt-contract, and read-only steps in `SKILL.md`.
- Forward-tested the skill in three fresh subagent contexts. The clear-plan case selected an
  end-to-end CSV error-reporting phase; the stale-plan case recognized completed config work;
  the dirty/ambiguous case selected the active API plan while preserving the experimental
  overlap as a decision gate. Fixture status checks confirmed no writes.
- Revised the boundary and prompt contract after user feedback. A fresh read-only run against
  Creo Valuation selected all remaining M0 work—Slices 4B.3–4B.6, Slice 5, and the refixing
  representation check—and routed the goal through existing plans in two short paragraphs.
- Restored one concise change-management requirement: commit meaningful passing units as work
  progresses and state what changed, why, and any non-obvious decisions in each message.
- Validation passed via isolated `PyYAML` for `quick_validate.py`, `scripts/validate-skills`,
  direct and full `bunx skills add ... --list` checks, and git whitespace checks.
- The bounded implementation review with the diet lens found no remaining correctness or
  complexity issues after synchronizing the forward-test and validation evidence.
- Next slice: none; the planned skill slice is complete.
