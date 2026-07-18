# Organize Work

Use this branch to change the work system without implementing product work.

## Establish the Index

1. Inventory existing root plans, roadmaps, task lists, and linked item files.
2. Adopt an existing root `ROADMAP.md` or `PLAN.md` only when it already routes
   current, scheduled, and unscheduled work through the core model. Otherwise
   propose an explicit mapping into one root index plus `plans/` and `tasks/`,
   then ask before moving files or replacing links. Preserve every active item
   during the migration.
3. When no convention exists, create root `ROADMAP.md`. Create `plans/` and
   `tasks/` when their first items are captured.
4. Use this minimal index shape:

```markdown
# Roadmap

## Current

_None._

## Plans

_None._

## Tasks

_None._
```

Use a numbered list of links for plans and a bullet list of links for tasks.
Use one link under `Current` during ordinary single-worktree work.

## Create Work Items

Create each item from this minimum shape:

```markdown
# <Outcome-oriented title>

## Outcome

<What becomes true and why it matters.>

## Current state

<Only durable evidence and constraints needed to resume.>

## Next action

<One concrete action or decision that advances the outcome.>
```

Add optional sections only when they change future execution. Keep a requested
piece of work in `tasks/` when it can be chosen at any time. Put it in `plans/`
when the user wants it placed in the project sequence. For example, a scheduled
authentication replacement is a plan; an anytime empty-state polish is a task.

## Apply Lifecycle Changes

- **Capture:** create the item file and add one index link. Infer plan versus
  task from scheduling intent; ask only when that intent is genuinely unclear.
- **Start:** move the selected link into `Current`. Return an interrupted plan
  to the front of the plan queue and an interrupted task to the task pool.
- **Reorder:** edit only the numbered plan links. Treat the order as intention,
  not as a dependency claim.
- **Schedule:** move a task file into `plans/`. If it is current, update only
  its `Current` link; otherwise insert its sole link at the user-selected plan
  position. Ask for the position when it was not implied.
- **Unschedule:** move a plan file into `tasks/`. If it is current, update only
  its `Current` link; otherwise return its sole link to the unordered pool.
- **Complete:** summarize the achieved outcome under `Current state`, set
  `Next action` to `None — complete`, and remove the active link. Leave
  `Current` empty rather than starting unrelated work automatically.

Preserve useful content when adopting or moving existing files. Finish only
when no active item was lost and the index satisfies the invariants in
`SKILL.md`.
