---
name: next-goal
description: "Choose a substantial, evidence-backed implementation goal from repository plans and current state, then, when /goal is warranted, return two concise prompts for a fresh Codex session: one requiring PR delivery and one prohibiting PR creation. Use only when the user explicitly invokes $next-goal; prefer a phase, milestone, or multiple connected plan slices over a small standalone slice."
---

# Next Goal

Select a substantial next goal and, only when `/goal` is warranted, generate two short fresh-session routing prompts. Keep this run read-only: inspect the repository and return the result, but leave files, plans, goals, git state, and external systems unchanged.

## 1. Establish Current State

1. Read applicable `AGENTS.md` files and resolve the authoritative active `PLAN`, `TODO`, `ROADMAP`, progress, or handoff documents. Honor user-named documents; otherwise follow repository conventions and links. When no plan exists, infer candidates from instructions, code, tests, and history.
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

First absorb naturally connected work that follows a small candidate. If no substantial unblocked boundary remains, say that `/goal` is not warranted yet instead of manufacturing a small goal. Name the recommended maximum boundary and the next meaningful excluded area.

## 3. Route the Result

- When `/goal` is **not warranted**, give the evidence-based reason and omit both prompt variants. Do not read the delivery-variant instructions.
- Only after the evidence establishes that `/goal` **is warranted**, read and follow [prompts/delivery-variants.md](prompts/delivery-variants.md). Keep its two variants identical in outcome, stopping boundary, authoritative documents, and completion conditions; vary only delivery mechanics.

Before responding, verify that this run made no repository write, goal change, git mutation, or external publication.
