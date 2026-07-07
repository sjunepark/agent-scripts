---
name: progress-doc
description: "Read-only progress document review. Use when checking a PLAN, TODO, ROADMAP, or other progress document against repository evidence to explain what is done, left, or stale."
---

# Progress Doc

Review progress documents against the actual repository state. Treat the document as the starting claim, not the source of truth. Stay read-only: do not edit files, stage, commit, or push unless the user explicitly asks after the report.

## Workflow

1. Identify the target progress document.
   - If multiple progress docs are plausible and the target is unclear, ask which one to review.

2. Read the target document completely.
   - Inspect only the implementation evidence needed to verify status: referenced files, modules, tests, routes, commands, docs, or relevant git state.
   - Use git status or diffs only when the working tree may affect the answer.

3. Classify every item in the document.
   - `done`: clearly implemented, completed, or superseded by completed work.
   - `partially done`: meaningful progress exists, but acceptance criteria or follow-through remain.
   - `left`: not implemented, blocked, or still explicitly planned.
   - `stale or unclear`: the document conflicts with code evidence or lacks enough detail to verify.

4. Keep doc claims separate from verified evidence.
   - Name the concrete evidence for important classifications.

## Output

Use this shape and omit empty sections except `Recommended next step`:

```markdown
## Progress status

One short summary naming the reviewed progress doc.

## Done

- ...

## Partially done

- ...

## Left

- ...

## Stale or unclear

- ...

## Recommended next step

The most useful next action, including whether the progress doc should be updated.
```
