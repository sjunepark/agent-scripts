# Continue Work

Use this branch to implement one substantial coherent slice and preserve its
continuation state.

1. Select the target in this order: the user-named item, `Current`, then the
   first queued plan. Choose from the task pool only when the user names a task
   or explicitly asks to choose one. When the user names new work, capture it
   first; treat it as unscheduled unless the user places it in the plan
   sequence. If no target exists, report that work must first be captured or
   selected.
2. Read the target completely and inspect the implementation evidence needed
   to test its claims. Repair stale planning text when the correct state is
   clear; ask only when an external decision would materially change the work.
3. Move the target into `Current` before changing implementation. Apply the
   queue invariants from `SKILL.md` if another item is interrupted.
4. Implement the broadest connected slice that remains bounded, reviewable,
   and validatable. Include adjacent checklist items only when they share the
   same subsystem, decision context, or validation boundary.
5. Update the item as facts change. Keep `Current state`, checklist state, and
   `Next action` accurate; record only durable decisions, blockers, and useful
   validation results.
6. Run relevant repository validation. Record skipped checks only when the
   reason matters to the next session.
7. Stop when the slice or item is complete, a real blocker or user decision is
   reached, or the next work is materially unrelated. On item completion,
   remove it from `Current` and leave the next queued plan unstarted.

Do not stage, commit, push, or publish unless the user separately asks. In the
response, report the implemented slice, planning updates, validation, stopping
reason, and next intended work.
