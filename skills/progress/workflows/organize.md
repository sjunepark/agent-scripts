# Organize Work

Use this branch to change the work system without implementing product work.

## Establish the Index

1. Inventory roadmaps, plan and task directories, task lists, and linked item
   files. Resolve the ordinary or worktree-specific planning scope before
   choosing anything to edit.
2. Adopt an existing index only when it already routes current, scheduled, and
   unscheduled work through the core model. Otherwise propose an explicit
   mapping into the selected scope's index and item directories. Ask before
   moving files or replacing links unless the user already authorized that
   migration. Preserve every active item during the migration.
3. When no convention exists, create the selected scope's roadmap. Create its
   plan and task directories when their first items are captured. For ordinary
   work these are `ROADMAP.md`, `plans/`, and `tasks/`; parallel work uses the
   isolated paths selected during resolution.
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
Use at most one link under `Current` in each index.

## Create Work Items

Create each item in the selected scope's plan or task directory from this
minimum shape:

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
piece of work in the selected task directory when it can be chosen at any time.
Put it in the selected plan directory when the user wants it placed in the
project sequence. For example, a scheduled authentication replacement is a
plan; an anytime empty-state polish is a task.

## Apply Lifecycle Changes

- **Capture:** create the item file and add one index link. Infer plan versus
  task from scheduling intent; ask only when that intent is genuinely unclear.
  With an active goal, capturing the item does not add it to goal scope.
- **Start:** with an active goal, run its boundary check before moving the
  selected link into `Current`. Return an interrupted plan to the front of the
  plan queue and an interrupted task to the task pool.
- **Reorder:** edit only the numbered plan links. Treat the order as intention,
  not as a dependency claim.
- **Schedule:** move a task file into the same scope's plan directory. If it is
  current, update only its `Current` link; otherwise insert its sole link at
  the user-selected plan position. Ask for the position when it was not
  implied.
- **Unschedule:** move a plan file into the same scope's task directory. If it
  is current, update only its `Current` link; otherwise return its sole link to
  the unordered pool.
- **Complete:** summarize the achieved outcome under `Current state`, set
  `Next action` to `None — complete`, and remove the active link. Leave
  `Current` empty rather than starting unrelated work automatically. With an
  active goal, complete the goal when its terminal condition now holds; do not
  promote the next project item into goal status.

Preserve useful content when adopting or moving existing files. Finish only
when no active item was lost and the index satisfies the invariants in
`SKILL.md`.
