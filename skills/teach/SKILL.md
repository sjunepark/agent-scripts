---
name: teach
description: "Teach an existing codebase, subsystem, architecture, or feature flow to reviewers and maintainers. Use for a requested design-level mental model of responsibilities, contracts, data flow, invariants, and tradeoffs. Do not use for quick factual questions, implementation requests, library examples, diffs, commits, or patches."
---

# Teach

Teach for maintainable understanding, not implementation mechanics. Build the smallest coherent mental model that explains purpose, boundaries, ownership, contracts, control and data flow, invariants, edge cases, and consequential tradeoffs.

This skill is explanatory. Explain syntax, line-by-line execution, or how to write the code only when the user asks. Include external library, protocol, runtime, or product context only when it changes how the code should be understood.

## Workflow

1. Identify the teaching target.
   - Infer whether the learner needs a module, feature flow, subsystem, architecture area, API boundary, runtime path, data model, state flow, or directly relevant external concept.
   - For a broad request, choose the smallest coherent model first.

2. Read top-down.
   - Start from entry points, exported symbols, route handlers, public interfaces, or the requested flow.
   - Read supporting helpers, data structures, tests, and nearby `AGENTS.md`, `ARCHITECTURE.md`, or notes only as needed.
   - Continue until you can trace the main flow end to end and name every contract or boundary it crosses without guessing.

3. Build the lesson in learning order.
   - Start with purpose, role, and boundaries; then explain flow, ownership, and contracts.
   - Explain who calls and owns what, where decisions happen, and how data changes shape across boundaries.
   - Drill into mechanisms, invariants, edge cases, and tradeoffs only as needed to make the design intelligible.
   - Reorder the explanation for understanding rather than mirroring file order.

4. Use selective evidence.
   - Explain each important point in prose first and connect evidence back to the larger model.
   - If exact code is load-bearing for a contract, condition, data shape, boundary, state transition, or API, read [guides/snippets.md](guides/snippets.md) before including a snippet.
   - If a relationship is materially clearer as a flow, map, transition, or pipeline, read [guides/diagrams.md](guides/diagrams.md) before drawing a diagram.

5. Surface important confusion points.
   - Call out misleading names, blurred responsibilities, hidden invariants, or awkward control flow only when they affect understanding.
   - Frame them as teaching notes rather than a full review; say plainly when the structure is sound.

6. Close the loop.
   - Summarize the model the learner should retain and the one or two design facts that matter most for future work.

## Output shape

Use this shape unless the user asks for something else.

### Big Picture

Give one short paragraph on what this part of the system is for and where it fits.

### How It Works

Explain the main flow in logical learning order. Focus on roles, boundaries, contracts, ownership, and movement of control or data.

### Key Decisions

Call out the few abstractions, invariants, contracts, or design decisions that make the system make sense.

### Reviewer / Maintenance Focus

List only consequences that materially affect usage, behavior, compatibility, testing, maintenance, or future review. Include important tradeoffs, risks, confusing boundaries, or maintainer questions.

### What to Remember

State the mental model in one short paragraph or a few tight bullets. Favor the one or two points that will help the reader understand future work in this area.

### Important Confusion Points

Include this section only when something materially affects understanding. Describe design issues, awkward boundaries, or misleading structure briefly and concretely.

## Communication rules

- Optimize for reviewer understanding, not exhaustiveness.
- Match depth to the request: stay brief for tiny targets or overview asks; go deeper when the learner asks about mechanisms, decisions, or tradeoffs.
- Do not use path or line references as the primary navigation aid.
- Label inferences when the code does not prove intent directly.

## Routing

- Use `change-explainer` for diff-, commit-, patch-, or file-comparison teaching.
- Use a review-oriented skill when the user wants critique rather than understanding.
