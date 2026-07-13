# Next Goal Skill

Status: completed 2026-07-13.

Next action: none. Publication or global installation remains a separately authorized action.

## Current state

- `skills/next-goal` is implemented as a concise, read-only procedural skill with generated
  OpenAI metadata and manual-only invocation policy.
- No scripts, references, assets, or eval file were added; the judgment-heavy workflow does
  not need them, and this repository has no evaluation runner for `evals/evals.json`.
- Three disposable-repository forward tests covered a clear active plan, stale plan claims,
  and an ambiguous dirty worktree. Each produced a self-contained goal-sized prompt, retained
  repository-specific constraints, and made no writes.
- Skill validation, catalog validation, direct and full local-source listing, and whitespace
  checks pass.

## Purpose

Create a manually invoked `$next-goal` skill that inspects a repository's plans and current
state, selects an appropriately substantial next implementation outcome, and emits a
self-contained prompt to paste after `/goal` in a fresh Codex session.

The skill is a read-only handoff generator. It does not execute the work, create a goal,
change a plan, or mutate the target repository.

## Existing seams

Use the repository's existing skills instead of duplicating their responsibilities:

- `progress-run` supplies the rule for selecting the broadest connected slice that remains
  bounded, reviewable, and validated.
- `progress-handoff` supplies fresh-session handoff principles, but writes a durable plan;
  `next-goal` instead returns a prompt in the response.
- `progress-doc` supplies the requirement to verify plan claims against repository evidence.

Do not hardcode Creo Valuation, M0, manifests, SQLite, or any other one-repository concept.

## Deliverables

- [x] `skills/next-goal/SKILL.md`
- [x] `skills/next-goal/agents/openai.yaml`
- [x] Evaluation decision: omit `evals/evals.json` because the repository has no workflow that
  runs it

Start the skill with the installed OpenAI `skill-creator` initializer. Do not add a README,
scripts, assets, or references unless a concrete need emerges. Scope selection is primarily
judgment-heavy guidance; prefer a concise `SKILL.md` over machinery.

## Required workflow

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

### 2. Select one goal-sized outcome

Choose the broadest connected implementation scope that remains coherent, reviewable, and
validatable. Prefer work that shares a phase, subsystem, ownership seam, files, decision
context, or validation suite.

A good goal:

- ends in one concrete, demonstrable artifact or behavior;
- can progress through multiple coherent passing commits;
- has settled enough requirements to proceed autonomously;
- includes small prerequisites needed to make the outcome real;
- has a crisp stopping condition and an obvious next adjacent goal.

Split before:

- an unrelated phase or independently useful outcome;
- multiple independent high-risk seams or new dependencies;
- a data migration, security boundary, platform commitment, or external coordination step;
- a product or compatibility decision not implied by repository evidence;
- a whole roadmap or milestone whose pieces cannot be reviewed coherently.

Do not select a single tiny checkbox when adjacent work is naturally connected. Do not select
an entire roadmap merely because `/goal` can run for a long time. Explain the recommended
maximum boundary and name meaningful work deliberately left for the next goal.

### 3. Generate the fresh-session prompt

Return:

1. A short recommended boundary and rationale.
2. One copy-paste prompt intended as the body entered after `/goal`.
3. A short statement of the next excluded area, when useful.

The generated prompt must be self-contained; assume the new session has none of the current
conversation. Include:

- the concrete objective and completion artifact;
- authoritative repository documents and relevant paths to read first;
- verified current state without relying on brittle commit hashes unless they are necessary;
- explicit in-scope work and non-goals;
- settled invariants, dependency rules, and decision constraints supported by repository
  evidence;
- acceptance criteria and repository-native validation commands;
- the duty to keep the active plan or progress document current;
- a stopping condition and the next deferred boundary;
- instructions to inspect and preserve pre-existing changes.

Require the goal-running session to:

- capture its own starting git ref before editing so final review and whitespace checks cover
  progressively committed work as well as the remaining worktree diff;
- make local commits as coherent, passing slices complete;
- stage explicit paths and never sweep unrelated changes into a commit;
- use conventional, descriptive commit subjects consistent with the repository;
- run applicable focused validation and the repository's final validation;
- follow repository-specific review instructions discovered in `AGENTS.md`, including skills
  such as `$code-review` when required;
- on successful completion, leave all completed goal-owned work committed while preserving
  any pre-existing dirty state; if blocked, report partial or uncommitted work rather than
  committing a failing slice;
- avoid pushing, opening a PR, merging, or performing other external publication unless the
  user separately authorizes it.

Do not invent a token budget. Include one only when the user explicitly requests a budget.
Do not require an initially clean worktree or erase existing changes.

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
- Suggested short description: `Build a fresh-session implementation goal`.
- The default prompt should explicitly invoke `$next-goal` and ask for a copy-paste `/goal`
  prompt.
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

- the chosen scope is larger than a trivial checkbox but smaller than an incoherent roadmap;
- the prompt is actionable without access to the originating conversation;
- inclusions, exclusions, commits, validation, plan updates, and stop conditions are explicit;
- repository-specific instructions survive into the prompt without generic hardcoding;
- the skill performed no writes or goal creation.

Revise and repeat any case that exposes missing context, over-broad scope, plan-trusting, or
unearned prompt boilerplate.

## Acceptance criteria

- The skill reliably chooses one evidence-backed, goal-sized implementation outcome.
- Its output can be pasted into a new `/goal` session without manual reconstruction.
- It preserves dirty-tree and publication boundaries and requests progressive local commits.
- It remains read-only and does not overlap the execution responsibilities of existing
  progress skills.
- Its `SKILL.md` and metadata are concise, valid, and free of repository-specific assumptions.
- All forward tests and repository validation pass.

## Validation

Run:

```sh
python3 ~/.codex/skills/.system/skill-creator/scripts/quick_validate.py skills/next-goal
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
- Validation passed via isolated `PyYAML` for `quick_validate.py`, `scripts/validate-skills`,
  direct and full `bunx skills add ... --list` checks, and git whitespace checks.
- The bounded implementation review with the diet lens found no correctness or complexity
  issues; one safe wording fix clarified why the optional eval file was omitted.
- Next slice: none; the planned skill slice is complete.
