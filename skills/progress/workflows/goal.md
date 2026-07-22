# Goal Execution

Use this workflow only when `$progress` is explicitly invoked in goal mode or
with a `Goal contract`, or when the resolved planning scope contains an active
goal file. Apply the goal-contract reference loaded by `SKILL.md`; this workflow
overlays its closed authorization boundary on the mutable project queue for one
persistent `/goal` run.

## Initialize or Recover State

1. Resolve the planning scope before choosing a goal file. Use the concrete
   path supplied by the prompt. Otherwise use `goals/<stable-slug>.md` for
   ordinary work or `goals/<scope>/<stable-slug>.md` for an isolated worktree
   scope. Keep concurrently edited worktrees in separate goal directories just
   as their roadmaps, plans, and tasks are separate.
2. Initialize the file after any required delivery-base selection but before
   implementation, a work branch, or a PR. Copy the complete `Goal contract`
   block, from its label through `Delivery`, byte-for-byte under `Original
   contract`; do not summarize, normalize, broaden, or omit fields while
   persisting it. Keep the file in repository history with the work unless
   applicable repository instructions establish another durable tracked
   planning location. Do not delete it on completion.
3. On recovery, prefer the prompt-named path. If no path is available, continue
   only when exactly one active goal exists in the current planning scope. Stop
   for direction when no exact contract can be recovered, multiple active goals
   are plausible, or prompt and file disagree. Never substitute `ROADMAP.md`.
4. Keep the original contract block immutable and execution status mutable. If
   a source moves without a semantic scope change, record the replacement under
   evidence rather than rewriting the original contract.
5. At every resumed turn, automatic continuation, compaction recovery, or
   handoff, recover and verify the goal file before selecting candidate work.
   Treat the contract's repeated `$progress` invocation as continuing authority
   for this recovery; do not require the user to restate it on every turn.

## Check Every Transition

Before starting a new plan, work branch, PR, review program, or independently
reviewable outcome, state and apply the reference's boundary check. A generally
valuable improvement or broader assurance program is not `necessary` merely
because it increases confidence.

Queue position, `Current`, new plan creation, a review finding, and the fact that
work is unblocked are never contract evidence. Repository-required review and
validation of an included result may be necessary; a new review program that
can create its own remediation queue is outside unless explicitly included.

## Track Both Truth Planes

- Keep the project roadmap and item files accurate as work completes. They may
  advance to excluded work without authorizing it.
- Update goal status only with included or necessary work, durable evidence,
  blockers, decisions, and the next in-scope action. Do not copy the project's
  next item into goal status.
- Treat intermediate plans, commits, and PRs as checkpoints inside the contract,
  not automatic goal boundaries or automatic invitations to continue.
- Keep routine traces out of the file. Record a transition decision only when
  it prevents ambiguity or helps a later session resume safely.

## Amend or Complete the Goal

Append an amendment only after an explicit user instruction changes scope.
Record the instruction's concise substance, the contract fields it changes,
and why; retain the original contract instead of rewriting history. A bare
request to "continue" does not authorize an unclear expansion.

When every included result and the completion condition are satisfied:

1. Update project planning state truthfully, even when that makes excluded work
   `Current` or next.
2. Set goal `Status` to `complete`, set `Next in-scope action` to
   `None — goal complete`, and retain concise completion evidence.
3. Persist the terminal state using the lifecycle in the contract reference.
   For PR delivery, do this only after the final included PR merges and commit
   and push only goal and project-planning metadata directly on the preflighted
   integration branch; do not open another work branch or PR for this
   bookkeeping commit.
4. Mark the persistent platform goal complete when that mechanism is available.
5. Return to the user before reading, implementing, branching for, or opening a
   PR for the next project item.

If durable goal state is unavailable or cannot be trusted at any point, stop
and request direction. Failing closed is part of the goal contract.
