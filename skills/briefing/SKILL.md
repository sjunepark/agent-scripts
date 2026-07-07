---
name: briefing
description: "Brief the user on the current task so they can decide next steps without reloading the codebase. Use only when they explicitly ask for a briefing or catch-up on current state, the likely implementation approach, the relevant code or architecture, or the facts needed for a decision."
---

# Briefing

Give the user a short, decision-ready view of the current task. Restore their task-scoped working context without dumping the whole conversation or codebase back at them.

## Sources

Read sources in this order and stop once every point you will brief can be tied to something you read; label anything else as inferred:

- the current conversation
- the smallest set of relevant files in the codebase
- relevant local docs such as `AGENTS.md`, `ARCHITECTURE.md`, or nearby module docs when they change how the task should be understood

Use file paths when they materially help the user reconnect to the code.

## Emphases

Infer the emphasis from the user's explicit request; default to `status`; if mixed, combine the needed emphases without turning the response into a menu.

- `status` — asked to catch up or for current state: center the current implementation state.
- `plan` — asked what you will likely do next: center the likely approach and touched areas.
- `teach` — asked to understand the relevant code: center roles, boundaries, and flow.
- `decision` — asked what matters for a decision: center constraints, tradeoffs, and high-impact unknowns.

## Non-Goals

Do not use this skill to produce:

- a full architecture summary for the repository
- a plan review, code review, or design review
- a project roadmap reconstruction unless the current task depends on it
- a verbose recap of the whole conversation
- a line-by-line walkthrough of implementation details unless the user explicitly asks

## Workflow

1. Identify the requested emphasis (see Emphases).

2. Recover the task boundary.
- State the current request in concrete terms.
- Distinguish the present task from earlier side discussions if needed.

3. Gather only the relevant context.
- Start with the current conversation.
- Read the smallest set of files and docs that explain the current state.
- Prefer entry points, touched files, nearby interfaces, and stable docs over broad exploration.

4. Synthesize for fast orientation.
- Collapse raw findings into the few facts the user needs now.
- Name the relevant files and modules directly.
- Explain why each one matters to the current task.

5. Trim aggressively.
- Remove repeated points.
- Remove code trivia and low-signal history.
- Remove speculative statements unless they are clearly labeled.
- Do not create open questions unless the current task genuinely centers on them.

## Output

Default to 8-14 bullets total, or the equivalent in short sections.

Use this shape unless the user asks otherwise:

### Current Task
- One to three bullets on what is being worked on.

### Relevant Code
- The key files or modules, with a short note on why each matters.

### What Matters Now
- The most important implementation facts, behavior, or constraints for this task.

### Decision Context
- Only the facts the user needs to make the next decision.
- Include a light next-step note only when it helps.
