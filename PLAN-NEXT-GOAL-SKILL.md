# Next Goal Skill

Status: implementation complete 2026-07-22; live `/goal` compaction verification
pending.

Next action: run one real persistent `/goal` through automatic compaction and
verify that its retained prompt re-invokes `$progress` before continuation.
Publication, remote synchronization, or global installation remains separately
authorized work.

## Current state

- `skills/next-goal` remains a read-only, manually invoked selector. It chooses
  a substantial evidence-backed boundary and returns both PR and no-PR prompts
  without changing the repository or starting a goal.
- Each prompt now carries a closed execution contract: semantic outcome,
  durable state path, source anchors, individually named included results,
  observable completion, explicit exclusions, bounded derived-work authority,
  explicit-user-only expansion authority, and a repeated resume invariant.
- `skills/progress` owns durable initialization, recovery, amendments, mutable
  status, and completion. Goal authorization overlays its ordinary organize,
  orient, continue, and handoff workflows instead of replacing them.
- Project planning and goal execution are independent truths. A roadmap,
  `Current`, plan, branch, PR, or review finding may describe useful next work
  but cannot add it to an active goal.
- Exact goal state is recovered from `goals/<stable-slug>.md` or the current
  worktree scope's `goals/<scope>/<stable-slug>.md`. Missing or ambiguous state
  fails closed rather than reconstructing authority from conversation or the
  project queue.

## Ownership and lifecycle

There are three separate state planes:

1. Project planning (`ROADMAP.md`, `plans/`, and `tasks/`) records truthful
   project order and can advance beyond the active goal.
2. The original goal contract records immutable authorization. Source paths are
   anchors, not membership identities; path-only replacements are evidence,
   while semantic expansion requires an append-only authorized amendment.
3. Goal progress records completed results, the current checkpoint, next
   in-scope action, evidence, decisions, and blockers.

The goal-running session persists the tracked goal file before implementation
or the first work branch. PR delivery preflights a non-production integration
branch for direct metadata commit and push, then persists terminal status there
after the final included PR merges. Completed goal files remain concise history.
Parallel worktrees use isolated planning and goal namespaces, so another
worktree's queue or contract cannot grant scope in the current worktree.

## Transition invariant

Before starting a new plan, work branch, PR, review program, or independently
reviewable outcome, `$progress` classifies it as included, necessary, outside,
or ambiguous against the effective contract. It proceeds only with explicitly
included work or the smallest bounded prerequisite required by an included
completion condition. Outside work is recorded for later; ambiguous or
materially expansive work requires explicit user direction.

When the terminal condition holds, `$progress` updates project state truthfully,
marks the goal complete, and returns before consulting or starting the next
project item—even when that excluded item has become `Current`.

## Deliverables

- [x] Closed-contract selection and routing in `skills/next-goal/SKILL.md`.
- [x] Self-contained PR and no-PR prompt schemas in
  `skills/next-goal/prompts/delivery-variants.md`.
- [x] Goal ownership, lifecycle, state shape, and boundary rules in
  `skills/progress/references/goal-contract.md`.
- [x] Goal initialization, recovery, transition, tracking, amendment, and
  completion workflow in `skills/progress/workflows/goal.md`.
- [x] Goal-overlay hooks in the ordinary progress workflows.
- [x] Evaluation cases cover selection, scoped worktrees, queue drift,
  compaction recovery, review-created work, necessary prerequisites,
  ambiguity, amendments, completion, and concurrent worktrees; repository
  fixtures back the selection, initialization, queue-drift, ordinary-mode, and
  recovery regressions.
- [x] OpenAI metadata retains explicit invocation so the skill does not compete
  with ordinary work; the generated contract re-invokes it on every persistent
  goal continuation and recovery.

## Forward-test evidence

Fresh agents exercised two different planning layouts and goal names without
receiving the expected answer. Both generated prompts preserved the same
contract across delivery variants, named each included result, put exclusions
inside the copy-paste prompt, initialized durable state before work, and stopped
before the next excluded milestone.

Separate fresh agents resumed a fixture whose roadmap had already advanced to
an appealing integrated review. They recovered the narrower durable contract,
classified that review as outside, marked only the goal complete, kept the
roadmap truthful, and created no implementation branch or new plan.

Fixture-backed trials also cover planning-only work beside an active goal,
terminal status committed and pushed to a local integration remote, mid-goal
queue drift, and unrelated documents under `goals/`. A simulated automatic
continuation omitted both `$progress` and the original conversational contract;
it recovered durable state before opening plan contents, completed only the
goal, and pushed only terminal metadata. A real Codex `/goal` compaction remains
the final environment-level check because repository fixtures cannot trigger
platform compaction.

## Validation

The installed skill-creator validator, eval JSON parsing,
`scripts/validate-skills`, direct and full local catalog listings, and
`git diff --check` pass. The required `$code-review` pass used the system,
design, and diet lenses; its safe ownership, dispatch, selection-order,
persistence, and documentation findings were applied and revalidated. Do not
publish or install as part of this change unless separately requested.

## History

- 2026-07-13: created the read-only substantial-goal selector and delivery
  routing prompts.
- 2026-07-22: separated immutable goal authorization from mutable project
  queues and progress, added durable recovery and transition enforcement, and
  added regression fixtures for the scope-cascade failure.
