---
name: teach
description: Teach the user how code, a subsystem, or a relevant technical concept works in clear learning order for reviewers and maintainers, focusing on design, responsibilities, contracts, data flow, invariants, tradeoffs, and maintenance implications rather than syntax or line-by-line execution. Use whenever the user asks to understand a part of the current codebase, a feature flow, an architecture area, a module, an API boundary, a data flow, or a relevant library or framework concept needed to understand the code. Include code snippets only when they are important or crucial to understanding. If the request is specifically about what changed in a diff, commit, or patch, prefer `change-explainer`.
---

# Teach

Teach for maintainable understanding, not implementation mechanics.

The goal is to help a reviewer or maintainer understand:
- what the relevant code or concept is for
- how responsibilities, boundaries, and contracts fit together
- how control and data move through the system
- what assumptions, invariants, or edge cases matter
- what design decisions and tradeoffs explain the current shape
- what maintenance implications or confusing parts are worth noticing

This skill is primarily explanatory, not an implementer tutorial. Do not explain syntax, line-by-line execution, or how to write the code unless the user explicitly asks. Point out important design issues or confusing structure when that would materially improve understanding, but do not drift into generic review mode.

## Scope

This skill usually applies to the current codebase, but it may also cover:
- a library or framework used by that code
- a protocol, runtime mechanism, or pattern needed to understand the implementation
- a small amount of surrounding product or systems context if it clarifies the code

Prefer the smallest scope that lets the user understand the thing they asked about.

## First-Class Inputs

Handle these as normal teaching targets:
- a module or package
- a feature or user flow
- a subsystem or architecture area
- an API boundary or integration
- a runtime path such as request handling, background jobs, or event flow
- a data model or state flow
- a library or framework behavior that is directly relevant to the code

If the user is actually asking about a diff, commit, PR patch, or file-to-file comparison, use `change-explainer` instead.

## Core Behavior

1. Start from the learner's question.
- Identify what the user is trying to understand, not just which files are nearby.
- Match the explanation depth to the request.
- If the user asks a broad question, narrow it to the smallest coherent mental model first.

2. Build a mental model before details.
- Begin with purpose, role, and boundaries.
- Then explain the main flow, ownership, and contract relationships.
- Then drill into only the mechanisms, edge cases, and invariants needed to understand the design.
- Avoid starting with line-by-line code reading unless the user explicitly wants that.

3. Use snippets only when they are crucial evidence.
- Snippets are rare by default.
- Include a snippet only when the exact code shape is important or crucial to understanding the concept.
- Prefer prose, compact pseudocode, or a small diagram when the design can be explained without quoting code.
- When a snippet is necessary, embed a short, focused excerpt directly in the response.
- Start each snippet with a comment that names the file path it came from.
- Do not force the user back into the editor just to follow the explanation.

4. Use ASCII diagrams when structure is easier to see than describe.
- Add a small diagram when it materially improves understanding of:
  - control flow
  - data flow
  - component or module relationships
  - state transitions
  - layered architecture or request lifecycles
- Keep diagrams compact, readable, and ASCII-only.
- Do not add a diagram if prose already makes the point clear.

5. Teach relationships, not isolated facts.
- Explain who calls what, who owns what, and where decisions happen.
- Show how data changes shape as it crosses boundaries.
- Name the abstractions, responsibilities, and contracts that matter to the flow.

6. Surface important confusion points.
- Call out misleading names, blurred responsibilities, hidden invariants, or awkward control flow when they materially affect understanding.
- Treat these as teaching notes, not as a full review.
- If the structure is mostly sound, say so plainly.

7. Use external context only when it helps.
- If library or framework behavior is necessary to explain the code, include only the part that changes how the code should be read.
- Keep external explanation tied to the code in front of you.

## Workflow

1. Identify the teaching target.
- Confirm whether the user wants to understand a module, flow, subsystem, data path, architecture area, or relevant external concept.
- Infer the likely target from the request and nearby context when needed.

2. Read top-down.
- Start from entry points, exported symbols, route handlers, public interfaces, or the main flow the user is asking about.
- Then read supporting helpers, data structures, and tests only as needed.
- Pull in docs such as `AGENTS.md`, `ARCHITECTURE.md`, or nearby notes when they materially change the explanation.

3. Organize the lesson as reviewer context.
- Prefer an order like:
  - what this part exists to do
  - where it sits in the system
  - the main runtime or data flow
  - the key contracts, ownership boundaries, or abstractions
  - invariants, edge cases, tradeoffs, and maintenance implications
- Reorder the material so the explanation is easy to understand, not so it mirrors file order.

4. Teach with selective evidence.
- For each important point, use prose first.
- Add a snippet only when the exact code is important or crucial to understanding the concept.
- Prefer a compact ASCII diagram or pseudocode summary when it explains ownership, flow, or contracts more clearly than syntax.
- Connect any evidence back to the larger mental model.

5. Close the loop.
- Summarize the model the user should now have.
- Mention the most important thing to remember about the design.
- If there is a major confusing or fragile area, call it out directly.

## Snippet Rules

- Snippets are rare by default.
- Include a snippet only when prose would make the concept ambiguous or hide a crucial technical decision.
- Good reasons to include a snippet include an exact contract shape, condition, data shape, boundary, state transition, or API that the reader must see to understand the design.
- Do not include snippets just to prove that a file exists, restate syntax, or walk through routine implementation detail.
- Use fenced code blocks for necessary snippets.
- Put the source file path in the first line of the snippet as a comment.
- Match the comment style to the language when practical, for example `// src/dashboard/load.ts` or `# backend/jobs/sync.py`.
- If the language is unclear, use a neutral comment style or label the snippet in surrounding prose.
- Keep snippets narrowly scoped to the mechanism or contract being taught.
- Prefer signatures, conditions, branching points, state transitions, and interface boundaries over long contiguous code.
- When a relationship matters more than the exact syntax, summarize the flow in prose or pseudocode instead of quoting code.
- If helpful and important, pair a snippet with a brief explanation like:

```ts
// src/dashboard/loadUserDashboard.ts
export async function loadUserDashboard(userId: string) {
  const account = await accountRepo.getByUserId(userId);
  return buildDashboardView(account);
}
```

This shows the boundary clearly: the function does orchestration, not heavy business logic. The repository fetches data, and the view builder shapes it for the caller.

## Diagram Rules

- Use diagrams only when they make the lesson easier to understand.
- Use compact ASCII diagrams that render clearly in plain Markdown.
- Do not use Mermaid.
- Keep them small and purpose-built for one idea.
- Favor these diagram types:
  - request or event flow
  - module relationship map
  - state transition sketch
  - data transformation pipeline
- Put the diagram near the explanation it supports.
- Explain the diagram briefly instead of assuming it is self-explanatory.

Example:

```text
Route -> Service -> Repo -> DB
```

This works when the main teaching problem is ownership or call flow rather than syntax.

## Output Shape

Use this shape unless the user asks for something else:

### Big Picture
- One short paragraph on what this part of the system is for and where it fits.

### How It Works
- Explain the main flow in logical learning order.
- Add a snippet only when the exact code is crucial to understanding the concept.
- Add a compact ASCII diagram if it makes the flow or relationships clearer.
- Focus on roles, boundaries, contracts, ownership, and movement of control or data.

### Key Decisions
- Call out the few abstractions, invariants, contracts, or design decisions that make the system make sense.

### Reviewer / Maintenance Focus
- List only the consequences that materially affect usage, behavior, compatibility, testing, maintenance, or future review.
- Mention important tradeoffs, risks, confusing boundaries, or follow-up questions from a maintainer's perspective.

### What to Remember
- State the mental model in one short paragraph or a few tight bullets.
- Favor the one or two points that will help the reader understand future work in this area.

### Important Confusion Points
- Include only when something materially affects understanding.
- Mention important design issues, awkward boundaries, or misleading structure briefly and concretely.

## Communication Rules

- Optimize for reviewer understanding, not exhaustiveness.
- Prefer a coherent design explanation over a file-by-file tour.
- Do not default to path or line references as the primary navigation aid.
- Do not teach syntax, line-by-line behavior, or implementation-writing technique unless explicitly requested.
- Do not include code snippets unless they are important or crucial to understanding the concept.
- Do not dump every related file into the answer.
- Label inferences as inferences when the code does not prove intent directly.
- If the user seems to want a concise overview or the target is tiny, stay brief and practical.
- If the user clearly wants deeper teaching, go further into mechanisms, design decisions, and tradeoffs.

## Non-Goals

Do not use this skill as the default choice for:
- diff, commit, or patch explanation
- formal code review
- task-status catch-up for the current session
- broad external-library research disconnected from the current code

Use `change-explainer` for change-focused teaching, `briefing` for task-state recovery, and review-oriented skills when the user is asking for critique rather than understanding.

## Example Triggers

- "Teach me how this feature works from a maintainer's perspective."
- "Explain this subsystem so I can understand the design and responsibilities."
- "I want to learn the request flow from the route down to the database. Only show snippets if they're crucial."
- "Teach me this flow with a diagram if that would make it easier to follow."
- "Teach me this module in a logical order, not file order."
- "Help me understand how this library is being used here and why the code is structured this way."
