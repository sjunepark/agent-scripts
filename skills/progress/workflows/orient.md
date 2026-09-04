# Orient to Work

Use this read-only branch to regain context, audit stale planning claims, or
brief the project's current state and best-supported next action. A planning
index improves routing but is not required for a useful local status briefing.

## Establish the snapshot

1. Capture the current branch, upstream and ahead/behind state, staged,
   unstaged, and untracked work, and a small recent history window before
   relying on planning documents. Summarize work by purpose only when the diff
   or history supports that inference.
2. Resolve the selected scope's index and linked current, queued, and pooled
   items. If no index exists, report that fact and continue from repository
   evidence; recommend initialization only when durable planning would
   materially help. During parallel work, keep other worktree namespaces out of
   scope unless the user requests an aggregate report.
3. With an active goal, recover and test its terminal state before opening
   linked project-item contents. Report goal current and next separately from
   project current and next, and do not open an excluded project item's contents
   when the goal is already terminal.

## Triangulate material claims

For every claim that could change the headline status or next action, compare
the plan or conversational claim with at least one independent source: the
working-tree diff, commit history, implementation and wiring, relevant tests,
generated artifacts, recorded validation, or CI evidence. Inspect only far
enough to establish delivery state.

Classify each headline claim as:

- **Verified:** independent evidence supports it.
- **Drift:** current repository evidence materially contradicts it.
- **Unverified:** affordable evidence is insufficient; absence of evidence alone
  is not contradiction.

For an explicit queue audit, also classify every in-scope item as `current`,
`queued`, `available`, `blocked`, `complete`, or `stale`. Keep desired direction
separate from observed implementation reality and do not edit a stale claim in
this read-only branch.

Run a targeted check only when it is an established, reasonably fast,
side-effect-free diagnostic whose scope is understood. Prefer existing CI
evidence when local validation would be slow or environment-dependent, and name
important checks that remain unrun.

## Add relevant remote state

When the repository has a GitHub remote and access is available, inspect open
pull requests and connect the current branch or current work to its review and
check state. Include an issue only when plans, commits, pull requests, an active
milestone, or the user's scope establishes relevance. Do not inventory the
general backlog. When remote access is unavailable, complete the local briefing
and state the blind spot.

## Recommend one action

Choose the smallest concrete next action by weighing, in order: direct user
direction; coherent in-progress local or pull-request work; blockers, failing
checks, or review gates; verified current planning intent; and the
best-supported queued work. Dirty or remote work is evidence of activity, not
automatic proof of priority or ownership. When materially different actions
remain equally plausible, state the decision needed instead of inventing
certainty.

Return only useful parts of this order:

- bottom-line current state and one recommended next action;
- current work, branch or worktree condition, and health signals;
- recent meaningful local or remote movement;
- verified, drifting, and unverified planning claims;
- relevant pull requests and connected issues;
- queued plans and unordered tasks when an index exists; and
- blockers, blind spots, missing files, and the smallest useful documentation
  repair.

Complete only when headline status and the recommendation are evidence-backed.
Keep the repository, planning state, Git state, and external systems unchanged,
and confirm that the run introduced no persistent changes.
