---
name: briefing
description: "Brief the user on the current task or session so they can decide next steps without reloading the codebase. Use only when they explicitly ask for a briefing, catch-up, current state, relevant architecture, implementation context, or task-scoped context recovery."
---

# Briefing

Give the user a short, decision-ready view of the current task. Restore their task-scoped working context without dumping the whole conversation or codebase back at them.

The user is the decision-maker. Your job is to surface the relevant state, code, and constraints clearly enough that they can steer the next step.

## Core Job

Restore the user's working context for the current task so they can make the next decision.

This skill is not for broad repository explanation. It is for reconnecting the user to the few details that matter now.

## Core Behavior

1. Work from the current request first.
- Treat the current task or question as the scope boundary.
- Read the current conversation before reading code.
- Pull in only the files and docs needed to explain the current state.

2. Stay task-scoped.
- Focus on the current request or session, not the full project history.
- Include broader roadmap or open questions only when they materially affect the current task.
- Avoid generic repository summaries.

3. Keep the briefing brief.
- Default to 8-14 bullets total, or the equivalent in short sections.
- Prefer high-signal facts over exhaustive detail.
- Do not walk line by line through code unless the user explicitly asks.

4. Be mostly descriptive.
- Explain the current state, relevant code, and constraints.
- Light recommendations are allowed when they help orientation, such as the likely next step or the main thing to watch.
- Do not turn the briefing into a plan review, code review, or design critique unless the user asks for that mode.

## Sources

Read sources in this order and stop as soon as the current task can be explained confidently:

- the current conversation
- the smallest set of relevant files in the codebase
- relevant local docs such as `AGENTS.md`, `ARCHITECTURE.md`, or nearby module docs when they change how the task should be understood

Use file paths when they materially help the user reconnect to the code.

## Briefing Styles

Default to a `status` briefing.

Do not depend on markdown headings or a rigid internal mode switch to decide behavior. Instead, infer the user's emphasis from their explicit request:

- If they ask to catch up, brief current state.
- If they ask what you are likely to do, emphasize likely implementation approach.
- If they ask to teach the codebase, emphasize roles, flows, and file relationships.
- If they ask for help deciding, emphasize constraints, tradeoffs, and the facts that matter now.

Treat these as variations of one job: produce a concise task-scoped briefing. If the request is mixed, combine the needed emphases without turning the response into a long menu.

## Non-Goals

Do not use this skill to produce:

- a full architecture summary for the repository
- a code review or design review
- a project roadmap reconstruction unless the current task depends on it
- a verbose recap of the whole conversation
- a line-by-line walkthrough of implementation details unless the user explicitly asks

## Workflow

1. Identify the user's requested emphasis.
- If the user explicitly asks for current state, use a `status` emphasis.
- If they ask what you will likely do next, use a `plan` emphasis.
- If they ask to understand the relevant code, use a `teach` emphasis.
- If they ask what matters for a decision, use a `decision` emphasis.
- If the request is ambiguous, default to `status`.

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

## Output

Default to short sections with flat bullets.

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

Adjust emphasis based on the user's request:

- `status`: center the current implementation state.
- `plan`: center the likely approach and touched areas.
- `teach`: center roles, boundaries, and flow.
- `decision`: center constraints, tradeoffs, and high-impact unknowns.

## Communication Rules

- Prefer the smallest useful briefing over a comprehensive one.
- Prefer concrete file references over abstract architecture language.
- Prefer "what matters now" over "everything that exists."
- Keep explanations compact but complete enough for decision-making.
- Do not repeat the conversation back to the user.
- Do not pad the output with generic advice.
- Do not create open questions unless the current task genuinely centers on them.
- If evidence is partial, say what is known and what is inferred.

## Example Triggers

- "Use $briefing and catch me up on what matters for this task."
- "Before you implement this, give me a briefing on the relevant code."
- "Brief me on the current state of this feature."
- "Teach me the part of the codebase related to this request."
- "Give me a decision briefing for this change."
