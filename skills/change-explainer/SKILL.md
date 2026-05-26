---
name: change-explainer
description: "Explain a code or document change set for reviewers and maintainers, focusing on design, decisions, contracts, tradeoffs, and maintenance implications. Use for understanding a git diff, unstaged changes, commit, commit range, PR patch, or two versions of a file or document."
---

# Change Explainer

Teach the change, not just the patch mechanics.

The goal is to help a reviewer or maintainer understand:
- what problem or purpose the change seems to address
- what changed at the design, responsibility, or contract level
- what the relevant part of the system now does after the change
- how the important edits fit together into one technical decision
- how control flow, data flow, ownership, or contracts changed
- what tradeoffs, risks, or maintenance consequences matter
- what mental model the reader should carry forward

This is primarily an explanatory skill, not an implementer tutorial. Do not explain syntax, line-by-line execution, or how to write the code unless the user explicitly asks. Do not turn it into a formal code review by default. Offer opinions only when the change appears to introduce a material issue or an important tradeoff the user should know about.

## First-Class Inputs

Handle these as normal entry points:
- `git diff`
- unstaged or staged local changes
- a commit
- a commit range
- a PR patch or review diff
- two versions of a file or document

If the user asks for an explanation but does not provide a diff, commit, range, patch, file pair, or other comparison target, default to the current repository's uncommitted changes. Treat staged changes, unstaged changes, and untracked files as part of that default target. Do not ask for clarification before checking local status. If there are no uncommitted changes, say that plainly and ask for the commit, range, patch, or file versions to explain.

If the user provides another concrete comparison format, adapt the same workflow.

## Core Behavior

1. Start from the reader's question.
- Identify what the user is trying to understand about the change, not just which files moved.
- Match the depth to the request.
- If the user asks broadly, narrow to the smallest useful mental model first.

2. Build a mental model before hunk details.
- Begin with the patch's purpose, scope, or visible behavior change.
- Then explain the main implementation flow.
- Then cover supporting edits such as helpers, tests, docs, or cleanup.
- Avoid starting with file-by-file or hunk-by-hunk narration unless the user explicitly wants that.

3. Teach the changed system, not just the delta.
- Reconstruct the smallest useful model of the code or document as it works after the patch.
- Use only the minimum necessary "before" state to make the change legible.
- For each conceptual change, prefer a `before -> after -> why it matters` arc over raw patch narration.
- Name the underlying abstraction, ownership shift, or contract change when one exists.

4. Anchor on the actual change.
- Read the diff first.
- Then read surrounding code, tests, and nearby docs only where needed to explain intent, ownership, or invariants.
- Do not speculate about the design before checking the code around the changed hunk.

5. Reorder the explanation for comprehension.
- Do not explain hunks in raw diff order unless the diff is already easy to follow.
- Group related edits into a small number of conceptual changes.
- Prefer an order like:
  - what this patch is trying to accomplish
  - user-visible behavior
  - interface or contract changes
  - main control-flow or state changes
  - data model or persistence changes
  - tests and docs

6. Teach relationships, not isolated hunks.
- Explain how edits connect across files.
- Show who calls what, where decisions moved, and how data or state now flows differently.
- If responsibilities were split, merged, or renamed, make that relationship explicit.
- Say what idea the cluster of edits expresses, not just what lines changed.

7. Use snippets only when they are crucial evidence.
- Snippets are rare by default.
- Include a snippet only when the exact code shape is important or crucial to understanding the change.
- Prefer prose, compact pseudocode, or a small diagram when the decision can be explained without quoting code.
- When a snippet is necessary, embed a short, focused excerpt directly in the response.
- Start each snippet with a comment that names the file path it came from.
- Show `before` and `after` snippets only when the contrast is important to the technical decision.
- Do not send the user to file paths or line numbers unless they explicitly ask for references.

8. Use ASCII diagrams when flow is easier to see than describe.
- Add a compact ASCII diagram when it materially improves understanding of:
  - control flow across changed components
  - data flow before vs after
  - ownership or boundary changes
  - state transitions affected by the patch
- Keep diagrams compact, readable, and ASCII-only.
- Do not add a diagram if prose and snippets already make the point clear.

9. Separate explanation from judgment.
- Stay mainly descriptive.
- Add judgment only for issues that look important enough to change how the user should read the patch:
  - correctness risk
  - behavior regression risk
  - broken contract or compatibility risk
  - meaningful security or data-handling concern
  - a design tradeoff that materially affects future work
- If there is no important concern, say so plainly and keep the focus on understanding.

## Workflow

1. Identify the comparison unit.
- Confirm whether you are explaining local changes, a specific commit, a commit range, a PR patch, or a file-to-file comparison.
- If the comparison is ambiguous or omitted, default to uncommitted local changes in the current repository.
- Start with `git status --short`, then inspect staged and unstaged diffs. Include untracked files when they are part of the visible local change set.
- If there are no uncommitted changes, stop and ask for the commit, range, patch, or file versions to explain.

2. Read the change before the surroundings.
- Start with the diff or patch.
- Pull in neighboring code only where the diff alone does not explain behavior, ownership, or invariants.

3. Build conceptual buckets.
- Merge scattered hunks that serve one idea.
- Split one large diff into separate conceptual changes when it mixes different concerns.
- Name each bucket by its purpose or technical decision, not by the file it touched.

4. Organize the explanation as reviewer context.
- Start with the highest-level purpose of the patch.
- Then explain the changed system or document in the smallest useful scope.
- Then move through the main behavior, contract, ownership, or data-flow changes.
- Then cover only the supporting implementation details needed to understand the decision.
- Keep tests and docs near the concept they validate rather than treating them as a disconnected appendix.

5. Teach each conceptual bucket with a stable pattern.
- Prefer `before -> after -> why` when the contrast materially helps.
- If the old state is less important, explain the post-change design first and then name the delta.
- Explicitly say what new mental model the reader should adopt for this part of the patch.
- Avoid line-by-line behavior unless the user explicitly asks for it.

6. Teach with selective evidence.
- For each conceptual change, use prose first.
- Add a snippet only when the exact code is important or crucial to understanding the change.
- Prefer a compact ASCII diagram or pseudocode summary when it explains ownership, flow, or contracts more clearly than syntax.
- Connect any evidence back to the larger mental model of the patch.

7. Close the loop, not just the walkthrough.
- Summarize what the patch changed in the reader's mental model.
- Mention the most important thing to remember about the new behavior, boundary, or contract.
- Then mention what the change means for callers, operators, maintainers, or future edits.
- Label inferences as inferences when they are not directly stated in the code or tests.

8. Add warnings only when warranted.
- If a critical or important issue stands out, include it in a compact note.
- Do not pad the answer with minor style opinions or speculative nitpicks.

## Snippet Rules

- Snippets are rare by default.
- Include a snippet only when prose would make the change ambiguous or hide a crucial technical decision.
- Good reasons to include a snippet include an exact contract shape, condition, data shape, boundary, or API that the reviewer must see to understand the change.
- Do not include snippets just to prove that a file changed, restate syntax, or walk through routine implementation detail.
- Use fenced code blocks for necessary snippets.
- Put the source file path in the first line of the snippet as a comment.
- Match the comment style to the language when practical, for example `// src/auth/session.ts` or `# app/models/user.rb`.
- If the language is unclear, use a neutral comment style or label the snippet in surrounding prose.
- Keep each snippet narrowly scoped to the point being explained.
- Prefer the changed lines plus just enough surrounding context to make them legible.
- If a long function changed, excerpt only the relevant branch, condition, signature, or data shape instead of pasting the whole function.
- When comparing versions, make the contrast obvious:

```ts
// src/auth/session.ts
// before
if (!session) return null;

// after
if (!session || session.expiresAt < now) return null;
```

- If the exact old code is unavailable, show the current code and explain the inferred delta without pretending to quote the previous version.

## Diagram Rules

- Use diagrams only when they make the patch easier to understand.
- Use compact ASCII diagrams that render clearly in plain Markdown.
- Do not use Mermaid.
- Keep them small and purpose-built for one idea.
- Favor these diagram types:
  - before/after flow
  - module relationship map
  - state transition sketch
  - data transformation pipeline
- Put the diagram near the explanation it supports.
- Explain the diagram briefly instead of assuming it is self-explanatory.

Example:

```text
Before: Route -> Repo
After:  Route -> Service -> Repo
```

This works when the important change is a new ownership boundary rather than a line-by-line code delta.

## Output Shape

Use this shape unless the user asks for something else:

### Big Picture
- One short paragraph on what this patch is trying to do and where it matters.

### How the Change Works
- Explain each conceptual change in logical learning order.
- Add a snippet only when the exact code is crucial to understanding the change.
- Add a compact ASCII diagram if it makes the flow or relationships clearer.
- Focus on behavior, boundaries, contracts, ownership, and movement of control or data.

### Key Decisions
- Call out the few abstractions, invariants, contracts, or design decisions that make the patch make sense.

### Reviewer / Maintenance Focus
- List only the consequences that materially affect usage, behavior, compatibility, testing, maintenance, or future review.
- Mention important tradeoffs, risks, or follow-up questions from a maintainer's perspective.

### What to Remember
- State the new mental model in one short paragraph or a few tight bullets.
- Favor the one or two points that will help the reader understand future edits in this area.

### Important Concerns
- Include only if there is a critical or important issue.
- If there is no such issue, omit this section or say there are no material concerns.

## Communication Rules

- Optimize for reviewer understanding, not exhaustiveness.
- Prefer a coherent design explanation over a file-by-file or hunk-by-hunk tour.
- Do not default to flat bullet lists like "change 1, change 2, change 3" unless the user explicitly wants a recap.
- Do not let the explanation collapse into "file A changed X, file B changed Y" narration without naming the concept those edits serve.
- Do not default to review language like "finding" or "severity" unless the user explicitly wants review mode.
- Do not teach syntax, line-by-line behavior, or implementation-writing technique unless explicitly requested.
- Do not include code snippets unless they are important or crucial to understanding the change.
- Do not mention file names or line numbers as the primary navigation aid.
- Keep the explanation concrete enough that the user does not need to open an editor just to follow the answer.
- When inferring intent or tradeoffs, say that you are inferring from the change.
- If the user seems to want a concise overview or the change is tiny, stay brief and practical.
- If the user clearly wants deeper teaching, go further into mechanisms, design decisions, and tradeoffs.

## Example Triggers

- "Walk me through my unstaged changes and explain them in an order that makes sense."
- "Explain this diff like I'm joining the review late. Start with the big picture, then explain the design decisions."
- "I don't want a file-by-file recap. Help me understand this commit from a maintainer's perspective."
- "Compare these two versions and explain the technical decisions in the easiest order to understand."
- "Help me understand what changed, why, and what it implies. Only show snippets if they're crucial."
