# Stacked Pull Requests

Use this branch when multiple open PRs form a linear base/head chain. Process
the chain serially from the PR closest to the original base through the most
downstream PR. Keep going until every PR in the selected stack is merged or one
PR reaches a concrete blocker.

## Invariants

- Work on one PR at a time. Merge it before preparing its downstream PR.
- Put the stack behind a staging branch before merging the first PR when its
  base is the repository's default or protected integration branch.
- When `dev` is used as an alternative integration base to `main`, fetch both
  remote branches and verify that `origin/dev` and `origin/main` resolve to the
  same commit. If the tips differ, stop and report that `dev` must be
  synchronized or that the stack needs an explicitly chosen base. Do not
  update, merge, or rewrite `dev` without separate authorization.
- Leave the completed staging branch for the user. Never merge it into the
  original base or delete it as part of this workflow.
- Avoid explicitly deleting stack head branches. Record immutable head SHAs and
  downstream fork points so repository-level branch auto-deletion cannot break
  the next integration.
- Spend at most one manual CodeRabbit trigger and one manual Codex trigger per
  PR. An existing completed or in-progress review counts. Never retrigger after
  an incremental push.
- Treat automatically posted follow-up feedback as feedback to address, even
  though the agent did not trigger another review.
- Stop at the current PR when a required review cannot run, a required check
  fails, the PR is not mergeable, permissions are insufficient, or the stack
  topology is ambiguous. Leave downstream PRs untouched and report the exact
  blocker.

## Map the Stack

1. List open PRs and capture at least `number`, `url`, `baseRefName`,
   `headRefName`, `headRefOid`, `isDraft`, and merge/check state. Use an
   explicitly supplied PR list when the user provided one; otherwise traverse
   open PRs whose base branch is another PR's head branch.
2. Build the base-to-head graph and identify the selected linear chain. The
   most upstream PR is the node whose base is not another selected PR's head.
3. Confirm that every selected PR has exactly one predecessor and one successor
   at most. Ask the user to choose a path when the graph forks or when unrelated
   open PRs target a stack branch.
4. Fetch the selected heads and record each successor's dependency fork point
   as the merge base between its head and its predecessor's current head. Do
   this before any stack head is rebased or otherwise rewritten.
5. Create a stack ledger containing the original base and SHA, staging branch,
   ordered PRs, original base/head pairs, current head SHA, dependency fork
   points, review evidence, trigger comment IDs, validation/check state, and
   eventual merge commit.

The map is complete when every selected PR appears exactly once in upstream-to-
downstream order and no dependency edge is inferred only from PR titles or
branch-name resemblance.

## Establish the Staging Base

When the upstream PR targets the original default or protected integration
branch:

1. Reuse a user-specified staging branch. Otherwise choose a collision-resistant
   name such as `staging/pr-stack-<upstream-pr-number>`.
2. Create the staging ref at the upstream PR's original base SHA. If that name
   already exists, reuse it only when its purpose and current SHA are compatible;
   otherwise choose a new name. For a new ref, use:

   ```bash
   gh api --method POST repos/{owner}/{repo}/git/refs \
     -f ref="refs/heads/$staging_branch" \
     -f sha="$original_base_sha"
   ```

3. Retarget only the upstream PR to staging with
   `gh pr edit <pr> --base <staging-branch>`.
4. Verify that staging still points at the recorded original base SHA and that
   the upstream PR's effective patch is unchanged.

If the upstream PR already targets a deliberate staging branch, record and use
that branch after verifying its relationship to the intended original base.
The staging gate is complete only when the first PR can no longer merge directly
into the original base and its patch remains the intended one.

## Process Each PR

Repeat this section for every ledger entry. Do not start its successor early.

### 1. Prepare the branch

For the first PR, use the staging retarget above. For each later PR, only after
its predecessor is merged:

1. Before rewriting the current PR head, refresh its successor when one exists
   and update that successor's fork point against the current pre-integration
   head.
2. Record the PR's old base, then retarget it to the staging branch.
3. Fetch the staging branch and PR head. Use the dependency fork point recorded
   before the predecessor merged; do not assume its branch ref still exists.
4. Integrate the updated staging tip into the PR head using the repository's
   established branch policy. Prefer rebasing the downstream-only commits onto
   staging from the recorded fork point when history rewrites are allowed;
   otherwise merge staging into the head. Use `--force-with-lease` for a
   rewritten head.
5. Resolve conflicts in favor of the combined upstream fixes plus the PR's own
   intent, then run validation appropriate to the integration.
6. Push and verify that the PR diff against staging contains its downstream
   change only, while its working tree contains every merged upstream fix.

Preparation is complete when the PR targets staging, includes the current
staging tip, and no already-merged upstream patch appears as an accidental new
change in its diff.

### 2. Ensure one review from each service

Inspect reviews, issue comments, and prior trigger comments before posting
anything. Record the current head SHA with the evidence.

- CodeRabbit is satisfied by a completed bot-authored review or final summary.
  A processing/status comment is only in-progress evidence. If neither review
  nor prior trigger exists, post exactly one `@coderabbitai full review` PR
  comment.
- Codex is satisfied by its standard GitHub review. If neither review nor prior
  trigger exists, post exactly one `@codex review` PR comment. The exact trigger
  comes from the OpenAI Codex GitHub review documentation.
- A previous trigger without a completed review means wait for that request or
  diagnose its failure; it does not authorize a second trigger.
- A completed no-findings review satisfies the requirement. Do not require a
  comment when the reviewer explicitly found nothing.

Wait until both reviews complete or explicitly fail. When a service is not
installed, does not react, reports an authorization/configuration error, or
never completes within the task's reasonable wait window, block the current PR
instead of merging without that review.

### 3. Address the complete feedback surface

Run the parent skill's intake, local fix loop, and reply workflow for the current
PR after both required reviews are available. Include all human feedback and
all bot feedback, not only CodeRabbit and Codex findings.

After each push, refresh comments, reviews, threads, and checks. Add any
automatically created findings to the ledger and address them, but preserve the
one-trigger budget. This step is complete only when every actionable item on the
current PR is fixed and replied to, explicitly rejected with evidence, or
escalated for user judgment.

### 4. Merge the current PR

Refresh the PR immediately before merging. Require all of the following:

- the PR still targets the staging branch and contains the expected patch;
- completed CodeRabbit and Codex review evidence is recorded;
- the feedback ledger has no unhandled actionable item;
- required checks and local validation pass;
- the PR is ready, mergeable, and free of an unmet required approval; and
- every follow-up commit and reviewer reply is pushed.

Use the repository's required merge method. Keep the head branch while a
downstream PR still depends on it when repository settings allow. Before
merging a PR with a successor, verify the successor's recorded fork point is
still an ancestor of its head. Keep that pre-rewrite boundary; recomputing it
against a rewritten predecessor can replay already-merged commits. Verify the
current PR reached `MERGED`, record its merge commit and the new staging tip,
then begin preparing the next ledger entry.

## Completion

Finish only when every selected PR is merged into staging. Report:

- the original base and the staging branch left for the user;
- each PR in order, with review evidence or the single trigger comment for each
  service, pushed fix commits, validation, and merge commit;
- the final staging tip and its diff/check state relative to the original base;
- any automatically generated feedback handled after incremental pushes; and
- the explicit handoff that integrating staging into the original base remains
  a user-controlled action.
