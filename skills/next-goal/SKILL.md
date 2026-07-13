---
name: next-goal
description: "Choose a substantial, evidence-backed implementation goal from repository plans and current state, then return a concise prompt to paste after /goal in a fresh Codex session. Use only when the user explicitly invokes $next-goal; prefer a phase, milestone, or multiple connected plan slices over a small standalone slice."
---

# Next Goal

Generate a short fresh-session routing prompt. Keep this run read-only: inspect the repository
and return a prompt, but leave files, plans, goals, git state, and external systems unchanged.

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

## 3. Write the fresh-session prompt

Return, in this order:

1. **Recommended boundary** — a short scope statement and evidence-based rationale.
2. **Copy-paste `/goal` prompt** — one fenced block containing only the body to enter after
   `/goal`. Assume the new session has none of this conversation.
3. **Next excluded area** — one short statement when it clarifies the stopping boundary.

Use existing repository documents as the detailed specification instead of copying them into
the prompt. Usually write one to three short paragraphs containing only:

- the outcome and stopping boundary;
- the few authoritative plan documents to follow;
- verified state or a completion condition only when needed to disambiguate the plans;
- a direction to follow applicable `AGENTS.md`, preserve unrelated changes, keep plans current,
  and perform repository-required validation and review.

Leave implementation steps, file inventories, design guidance, invariants, test matrices,
commands, git choreography, and status recaps to the goal-running agent and cited repository
documents. Include an omitted detail only when the plans do not contain it and the goal would
otherwise be ambiguous or unsafe. Do not invent a token budget or turn uncertain choices into
instructions.

Before responding, verify that the prompt is actionable without this conversation and that
this run made no repository write, goal change, git mutation, or external publication.
