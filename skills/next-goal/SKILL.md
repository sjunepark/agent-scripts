---
name: next-goal
description: "Choose a substantial, evidence-backed implementation goal from repository plans and current state, then, when /goal is warranted, recommend PR delivery or no-PR aggregation and return both fresh-session prompt variants with a closed execution contract and $progress goal tracking. Use only when the user explicitly invokes $next-goal; prefer a phase, milestone, or multiple connected plan slices over a small standalone slice."
---

# Next Goal

Select a substantial next goal and, only when `/goal` is warranted, recommend PR delivery or later aggregation based on expected change size, then generate both fresh-session routing prompts with a closed execution contract. Keep this selection run read-only: inspect the repository and return the result, but leave files, plans, goals, git state, and external systems unchanged. Put `$progress` goal tracking in the generated prompts so the goal-running session, not this selection run, initializes durable goal state.

## 1. Establish Current State

1. Read applicable `AGENTS.md` files and resolve the authoritative active `PLAN`, `TODO`, `ROADMAP`, progress, or handoff documents. Honor user-named documents; otherwise follow repository conventions and links. When concurrent worktrees or scoped roadmaps exist, select the current worktree's planning namespace; read other scopes only for an explicitly requested aggregate goal. When no plan exists, infer candidates from instructions, code, tests, and history.
2. Inspect git status and recent history, then read only enough implementation and validation evidence to detect stale plan claims, completed work, real prerequisites, and blockers.
3. Ask only when plausible choices would materially change the outcome or no safe, unblocked goal can be supported by evidence.

This step is complete when one real next outcome and its constraints are verified against the repository rather than merely repeated from a plan.

## 2. Select the Goal Boundary

Choose the largest useful outcome that a persistent goal-running agent can pursue autonomously. Treat named slices and checklist items as planning units, not default stopping boundaries. Prefer completing an active phase or milestone, or several adjacent slices, when their requirements are settled in the same plans. A goal may span subsystems and multiple coherent commits when they lead to one meaningful project state.

Keep the selected boundary identical across delivery variants. In the PR variant, treat PRs as delivery checkpoints within the large goal rather than separate `/goal` boundaries.

The selected boundary must:

- be materially larger than work suited to one ordinary interactive turn;
- reach a concrete, demonstrable project or user outcome rather than only a prerequisite or internal seam;
- contain enough settled work to benefit from persistent execution across multiple checkpoints;
- stop at a genuine unresolved decision, external authorization, blocker, or materially unrelated next milestone—not merely at the next plan heading or reviewable slice.

First absorb naturally connected work that follows a small candidate. If no substantial unblocked boundary remains, say that `/goal` is not warranted yet instead of manufacturing a small goal.

Before routing, express the boundary as a closed goal contract:

- **Outcome** — the semantic project or user result.
- **Goal state** — one concrete durable path: `goals/<stable-slug>.md` for ordinary work or `goals/<scope>/<stable-slug>.md` for an isolated worktree planning scope.
- **Sources** — the few authoritative documents that supply implementation detail. Their contents, paths, and queue positions do not define goal membership.
- **Included results** — every authorized plan or checkpoint, named by stable semantic result rather than path alone.
- **Completion** — observable conditions that end the goal after all included results, validation, review, and selected delivery work finish.
- **Excluded work** — the first meaningful excluded area plus every explicit user exclusion needed to prevent a plausible continuation mistake.
- **Derived work and authority** — allow only the smallest bounded work necessary to make an included completion condition true. Record useful but unnecessary work for later; require explicit user direction for ambiguous or materially expansive work. Only an explicit user instruction may enlarge the contract.
- **Resume invariant** — at every resumed turn, automatic continuation,
  compaction recovery, or handoff, invoke `$progress` in goal mode and recover
  the named goal state before selecting or starting more work.

Treat goal membership as closed. Advancing a roadmap, changing `Current`, creating a plan, opening a branch or PR, or discovering review findings never adds work to the goal. Project planning state describes what the project should do next; the goal contract alone describes what this run is authorized to do.

## 3. Route the Result

- When `/goal` is **not warranted**, give the evidence-based reason and omit both prompt variants. Do not read the delivery-variant instructions.
- Only after the evidence establishes that `/goal` **is warranted**, read and follow [prompts/delivery-variants.md](prompts/delivery-variants.md). Put the same closed scope contract inside both prompt bodies; vary only delivery mechanics and the corresponding delivery field.

Before responding, verify that this run made no repository write, goal change, git mutation, or external publication.
