# Hand Off Work

Use this branch to preserve continuation context without implementing code.

1. Resolve the target from a user-named item, `Current`, or the work discussed
   in the session. When no item exists, create one in the appropriate queue or
   pool with the `Outcome`, `Current state`, and `Next action` required by
   `SKILL.md`.
2. Inspect relevant diffs, git status, recent commits, and validation evidence
   only far enough to distinguish completed work from intention.
3. Rewrite `Current state` as a compact durable summary. Preserve decisions,
   constraints, unfinished changes, real blockers, and evidence that the next
   session cannot cheaply rediscover.
4. Set `Next action` to the first concrete action or unresolved decision. Keep
   checklist and validation sections only when they sharpen that action. With
   an active goal, name only an in-scope action; record excluded project work in
   its own item without turning it into the goal handoff.
5. Point `Current` at the unfinished item. If the outcome is complete, record
   completion and remove it from the active index.

Prefer current truth over chronology. Fold useful prior notes into the current
sections instead of appending a session transcript or routine progress log.
Keep code, git history, and external systems unchanged. Report the updated
paths and next action.
