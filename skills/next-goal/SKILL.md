---
name: next-goal
description: "Choose a substantial, evidence-backed implementation goal from repository plans and current state, then, when /goal is warranted, return two concise prompts for a fresh Codex session: one requiring PR delivery and one prohibiting PR creation. Use only when the user explicitly invokes $next-goal; prefer a phase, milestone, or multiple connected plan slices over a small standalone slice."
---

# Next Goal

Select a substantial next goal and, when `/goal` is warranted, generate two short fresh-session
routing prompts. Keep this run read-only: inspect the repository and return the result, but leave
files, plans, goals, git state, and external systems unchanged.

## 1. Establish current state

1. Read applicable `AGENTS.md` files and resolve the authoritative active `PLAN`, `TODO`,
   `ROADMAP`, progress, or handoff documents. Honor user-named documents; otherwise follow
   repository conventions and links. When no plan exists, infer candidates from instructions,
   code, tests, and history.
2. Inspect git status and recent history, then read only enough implementation and validation
   evidence to detect stale plan claims, completed work, real prerequisites, and blockers.
3. Ask only when plausible choices would materially change the outcome or no safe, unblocked
   goal can be supported by evidence.

This step is complete when one real next outcome and its constraints are verified against the
repository rather than merely repeated from a plan.

## 2. Select the goal boundary

Choose the largest useful outcome that a persistent goal-running agent can pursue autonomously.
Treat named slices and checklist items as planning units, not default stopping boundaries.
Prefer completing an active phase or milestone, or several adjacent slices, when their
requirements are settled in the same plans. A goal may span subsystems and multiple coherent
commits when they lead to one meaningful project state.

Keep the selected boundary identical across delivery variants. In the PR variant, treat PRs as
delivery checkpoints within the large goal rather than separate `/goal` boundaries.

The selected boundary must:

- be materially larger than work suited to one ordinary interactive turn;
- reach a concrete, demonstrable project or user outcome rather than only a prerequisite or
  internal seam;
- contain enough settled work to benefit from persistent execution across multiple checkpoints;
- stop at a genuine unresolved decision, external authorization, blocker, or materially
  unrelated next milestone—not merely at the next plan heading or reviewable slice.

First absorb naturally connected work that follows a small candidate. If no substantial
unblocked boundary remains, say that `/goal` is not warranted yet instead of manufacturing a
small goal. Name the recommended maximum boundary and the next meaningful excluded area.

## 3. Write the fresh-session prompts

When `/goal` is warranted, return in this order:

1. **Recommended boundary** — a short scope statement and evidence-based rationale.
2. **Copy-paste `/goal` prompts** — exactly two labeled fenced blocks, each containing only the
   body to enter after `/goal`:
   - **PR delivery** — require sequential PR creation, review, and merge.
   - **No PR** — require completion without PR creation.
   Keep the outcome, stopping boundary, authoritative documents, and completion conditions
   identical between them; vary only delivery mechanics. Assume each new session has none of
   this conversation.
3. **Next excluded area** — one short statement when it clarifies the stopping boundary.

When `/goal` is not warranted, give the evidence-based reason and omit both prompt variants.

Make each prompt self-contained as a routing instruction, not as a detailed specification. Point
to existing repository documents for detailed guidance; do not summarize or duplicate their
contents. Usually write one to three short paragraphs per variant containing only:

- the outcome and stopping boundary;
- the few authoritative plan documents to follow;
- verified state or a completion condition only when needed to disambiguate the plans;
- a direction to follow applicable `AGENTS.md`, preserve unrelated changes, keep plans current,
  and perform repository-required validation and review;
- a requirement to commit incrementally as meaningful, passing units finish; do not defer all
  commits until goal completion or collapse the run into one final commit. Make each message state
  what changed and why, including non-obvious decisions needed later.

For **PR delivery**, also require:

- the fewest substantial, cohesive PRs that keep each review manageable. Combine small or tightly
  coupled slices instead of making one PR per plan item, but do not let the whole large goal
  accumulate into one oversized PR;
- a non-production integration branch: use the repository's established branch such as `dev`, or
  create a temporary integration branch when none exists. For each PR, create a fresh working
  branch from the current integration branch and target the PR back to it. Do not mutate `main` or
  open or merge PRs targeting it. Production promotion is outside the goal; leave the integration
  branch intact for that later workflow;
- a per-PR completion loop: open the PR, invoke `$address-pr-feedback`, wait for active Codex and
  CodeRabbit reviews, address all feedback and follow-ups, resolve every review thread, and confirm
  required validation and checks. Merge only then, update the integration branch, and create the
  next working branch from that merged state. Keep commits incremental within each PR. Do not
  create stacked or dependent PRs, start downstream work before the current PR merges, or leave an
  unreviewed final tail.

For **No PR**, require completing the selected boundary through implementation, validation, and
repository-required review. Explicitly state that the goal must not open or create a PR or invoke
PR-only feedback workflows; leave PR delivery for a later task.

Leave implementation steps, file inventories, design guidance, invariants, test matrices,
commands, git details beyond the selected variant's required delivery lifecycle, and status
recaps to the goal-running agent and cited repository documents. Include an omitted detail only
when the plans do not contain it and the goal would otherwise be ambiguous or unsafe. Do not
invent a token budget or turn uncertain choices into instructions.

Before responding, verify that every emitted prompt is actionable without this conversation and
that this run made no repository write, goal change, git mutation, or external publication.
