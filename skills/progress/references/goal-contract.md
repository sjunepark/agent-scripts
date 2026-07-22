# Goal Contract

Use this reference only in `$progress` goal mode. It defines three independent
planes whose state may change at different times.

## Ownership

- **Project planning plane:** `ROADMAP.md`, `plans/`, and `tasks/` describe the
  project's truthful current state and preferred order. They may advance beyond
  an active goal. Recording or prioritizing work never authorizes execution.
- **Goal contract:** the immutable semantic outcome, included results,
  completion predicate, exclusions, derived-work rule, expansion authority,
  resume invariant, and delivery lifecycle define what one persistent run may
  execute. `$next-goal` emits this contract as prompt text without writing;
  `$progress` owns its durable initialization and recovery in the fresh goal
  session.
- **Goal progress:** completed included results, current checkpoint, next
  in-scope action, evidence, blockers, and decisions are mutable execution state
  measured against the contract. They never inherit membership from the project
  queue.

## Location and Lifecycle

Use the prompt-named path. Otherwise store ordinary goal state at
`goals/<stable-slug>.md` and worktree-scoped goal state at
`goals/<scope>/<stable-slug>.md`, using the same scope key as the roadmap,
plans, and tasks. Keep one active goal per planning scope unless the user
explicitly coordinates more and names the intended goal path.

Keep the file tracked with the goal's work unless applicable repository
instructions establish another durable tracked planning location. Retain a
completed file as concise history. Do not delete, ignore, or silently replace
it. A goal in another worktree's scope cannot grant membership in this one.

When the contract selects PR delivery, preflight direct metadata commit and push
permission on the candidate non-production integration branch. Use and push a
temporary integration branch when the established branch cannot support that
lifecycle. Commit and push the initialized file there before creating the first
work branch, and carry execution-status updates through the relevant slice PRs.
After the final included PR merges, commit and push only the terminal goal and
project-planning metadata directly on the integration branch. That metadata-only
commit is durable bookkeeping, not an unreviewed implementation tail; never
include code or new scope in it. If policy changes and blocks that terminal
commit or push, stop and ask before inventing another PR or leaving the goal
falsely active. For no-PR delivery, commit initialization and terminal status as
incremental units on the selected working branch.

## State Shape

```markdown
# Goal: <outcome-oriented title>

Status: active
Planning scope: <roadmap or scoped namespace>

## Original contract

<Paste the complete `Goal contract` block verbatim, beginning with its
`Goal contract` label and ending with its `Delivery` field.>

## Authorized amendments

_None._

## Execution status

### Completed included results
_None._

### Current in-scope result
<Current included result.>

### Next in-scope action
<Concrete action, never the next excluded project item.>

### Evidence and blockers
<Only durable decisions, validation, blockers, and source-path repairs.>
```

Copy the prompt's complete `Goal contract` block byte-for-byte under `Original
contract`; do not normalize its lists, headings, wrapping, `Goal state`, or
`Delivery`. Keep that block immutable. Source paths are anchors rather than
membership identities; record a path-only replacement under evidence when its
semantic content is unchanged. Append an authorized amendment with the user's
instruction and its effect instead of rewriting the original contract.
Effective scope and delivery are the original contract plus those append-only
amendments.

## Boundary Classification

Before a new plan, work branch, PR, review program, or independently reviewable
outcome becomes active, classify it against the effective contract:

```text
Candidate: <next action or outcome>
Classification: included | necessary | outside | ambiguous
Contract basis: <included result or completion condition>
Action: proceed | record for later | request user direction | complete goal
```

- **Included:** directly advances a named semantic result.
- **Necessary:** the smallest bounded work without which an included completion
  condition cannot become true.
- **Outside:** useful but unnecessary work; record it in project state without
  executing it.
- **Ambiguous:** materially changes the outcome, acceptance conditions,
  architecture, delivery program, or risk; stop for explicit user direction.

`Current`, queue order, a new plan, an unblocked state, or a review finding is
never contract evidence. When the completion predicate becomes true, complete
the goal before consulting the project queue. If exact state cannot be
recovered, fail closed rather than reconstructing authority from the roadmap or
conversation fragments.
