---
name: next-goal
description: "Choose a substantial, evidence-backed implementation goal and emit one compact fresh-session routing envelope with a closed scope contract, $progress recovery, and an evidence-based PR or no-PR delivery choice. Use only when the user explicitly invokes $next-goal; return alternate delivery prompts only when requested, and prefer a settled phase or milestone over a small standalone slice."
---

# Next Goal

Select a substantial next goal and, only when `/goal` is warranted, recommend PR delivery or later aggregation based on expected change size, then generate one compact fresh-session routing envelope with a closed execution contract. Generate a non-recommended delivery prompt only when the user explicitly requests that variant, alone or alongside the recommendation.

Keep the goal-selection phase read-only. A combined request may separately authorize prerequisite mutation, such as committing completed planning work. Complete that distinct phase first under its applicable workflow, report it separately, and then select from the resulting repository state without further mutation. Put `$progress` goal tracking in the generated prompt so the goal-running session, not the selection phase, initializes durable goal state.

## 1. Establish Current State

1. Read applicable `AGENTS.md` files and resolve the authoritative active `PLAN`, `TODO`, `ROADMAP`, progress, or handoff documents. Honor user-named documents; otherwise follow repository conventions and links. When concurrent worktrees or scoped roadmaps exist, select the current worktree's planning namespace; read other scopes only for an explicitly requested aggregate goal. When no plan exists, infer candidates from instructions, code, tests, and history.
2. Inspect git status and recent history, then read only enough implementation and validation evidence to detect stale plan claims, completed work, real prerequisites, and blockers.
3. Ask only when plausible choices would materially change the outcome or no safe, unblocked goal can be supported by evidence.

This step is complete when one real next outcome and its constraints are verified against the repository rather than merely repeated from a plan.

## 2. Select the Goal Boundary

Choose the largest useful outcome that a persistent goal-running agent can pursue autonomously. Treat named slices and checklist items as planning units, not default stopping boundaries. Prefer completing an active phase or milestone, or several adjacent slices, when their requirements are settled in the same plans. A goal may span subsystems and multiple coherent commits when they lead to one meaningful project state.

When both delivery variants are requested, keep their selected boundary identical. In the PR variant, treat PRs as delivery checkpoints within the large goal rather than separate `/goal` boundaries.

The selected boundary must:

- be materially larger than work suited to one ordinary interactive turn;
- reach a concrete, demonstrable project or user outcome rather than only a prerequisite or internal seam;
- contain enough settled work to benefit from persistent execution across multiple checkpoints;
- stop at a genuine unresolved decision, external authorization, blocker, or materially unrelated next milestone—not merely at the next plan heading or reviewable slice.

First absorb naturally connected work that follows a small candidate. If no substantial unblocked boundary remains, say that `/goal` is not warranted yet instead of manufacturing a small goal.

Before routing, reduce the boundary to a closed routing envelope:

- **Outcome** — the semantic project or user result in one sentence.
- **Goal state** — one concrete durable path: `goals/<stable-slug>.md` for ordinary work or `goals/<scope>/<stable-slug>.md` for an isolated worktree planning scope.
- **Included results and sources** — every authorized result as a short, stable semantic label paired with the few authoritative documents that supply its implementation and acceptance detail. Labels define membership; paths and queue positions do not.
- **Completion** — one compact predicate requiring each named result to achieve its cited outcome and any applicable completion criteria, plus only cross-cutting validation, review, freshness, and delivery conditions not already carried by those sources.
- **Excluded work** — exactly the immediate next out-of-scope milestone plus exclusions stated directly by the user. Authority supplies the complete boundary for every later, unrelated, or merely plan-documented item.
- **Authority** — allow only the smallest bounded work necessary for an included result; record unnecessary work for later and require explicit user direction for ambiguous, expansive, or scope-enlarging work.
- **Resume invariant** — at every resumed turn, automatic continuation,
  compaction recovery, or handoff, invoke `$progress` in goal mode and recover
  the named goal state before selecting or starting more work.
- **Delivery** — the selected lifecycle and its skill routing, kept inside the contract so recovery preserves it.

Treat goal membership as closed. Advancing a roadmap, changing `Current`, creating a plan, opening a branch or PR, or discovering review findings never adds work to the goal. Project planning state describes what the project should do next; the goal contract alone describes what this run is authorized to do.

## 3. Route the Result

- When `/goal` is **not warranted**, give the evidence-based reason and omit the prompt. Do not read the delivery-variant instructions.
- Only after the evidence establishes that `/goal` **is warranted**, read and follow [prompts/delivery-variants.md](prompts/delivery-variants.md).
- By default, return only the recommended prompt and one sentence offering the alternate delivery variant.
- Honor an explicit request for one named delivery variant even when it differs from the recommendation; keep the recommendation truthful and label the prompt as the requested variant.
- When the user explicitly requests both variants, put the same closed scope contract inside both prompt bodies and vary only the delivery mechanics and `Delivery` field.

Before responding, verify that the selection phase made no repository write, goal change, git mutation, or external publication. When the request included prerequisite mutation, verify that it finished before selection began and report it separately.
