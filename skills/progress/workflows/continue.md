# Continue Work

Use this branch to implement one substantial coherent slice and preserve its
continuation state.

1. With an active goal overlay, recover its durable status first. If every
   included result and the terminal condition are already satisfied, complete
   the goal without selecting or reading the project queue's next item. When
   work remains, select the goal status's current or next in-scope result and
   classify it from the contract plus index metadata before opening the
   candidate file. Use project `Current` only to locate or reconcile that result;
   never let it override goal status. Read the candidate completely only after
   the boundary check allows it.
2. Without the goal overlay, select the target in this order: the user-named
   item, `Current`, then the first queued plan. Choose from the task pool only
   when the user names a task or explicitly asks to choose one. When the user
   names new work, capture it first; treat it as unscheduled unless the user
   places it in the plan sequence. If no target exists, report that work must
   first be captured or selected. Read the selected target completely.
3. Inspect the implementation evidence needed to test the allowed target's
   claims. Repair stale planning text when the correct state is clear; ask only
   when an external decision would materially change the work.
4. Move the allowed target into `Current` before changing implementation. Apply
   the queue invariants from `SKILL.md` if another item is interrupted.
5. Implement the broadest connected slice that remains bounded, reviewable,
   and validatable. Include adjacent checklist items only when they share the
   same subsystem, decision context, or validation boundary.
6. Update the item as facts change. Keep `Current state`, checklist state, and
   `Next action` accurate; record only durable decisions, blockers, and useful
   validation results.
7. Run relevant repository validation. Record skipped checks only when the
   reason matters to the next session.
8. Stop when the slice or item is complete, a real blocker or user decision is
   reached, or the next work is materially unrelated. On item completion,
   remove it from `Current` and leave the next queued plan unstarted. With an
   active goal, classify that next transition: continue only to another
   included result, or complete the goal before excluded project work.

Do not stage, commit, push, or publish unless the user separately asks. In the
response, report the implemented slice, planning updates, validation, stopping
reason, and next intended work.
