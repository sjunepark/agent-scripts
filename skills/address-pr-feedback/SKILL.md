---
name: address-pr-feedback
description: "Address existing GitHub PR feedback end to end for one pull request or a stacked PR chain. Use when asked to act on reviewer or bot feedback, fix it locally, push follow-up commits, reply to each comment or thread, or process dependent PRs upstream-to-downstream through review and merge."
---

# Address PR Feedback

Handle an existing GitHub PR from review intake through local fixes, validation,
push, and reviewer-facing replies. Treat PR feedback as broader than inline
threads: review bodies, issue comments, bot summaries, "outside diff" sections,
and follow-up comments can all contain actionable items.

If the task covers multiple dependent or stacked PRs, read
[`references/stacked-prs.md`](references/stacked-prs.md) before changing any PR
base or branch. Use that orchestration workflow to select and prepare one PR at
a time, then apply the intake, fix, reply, and per-PR completion checks below
to the current PR. Defer the user-facing completion response until the entire
selected stack satisfies the reference's completion gate.

## Intake

1. Start from a clean understanding of the PR.
   - Identify the PR URL/number, repository, base branch, head branch, author,
     review state, and latest head commit.
   - If no PR is specified, infer the current branch PR with `gh pr view`.
   - Do not overwrite unrelated local changes. If the working tree is dirty,
     separate user changes from PR-fix changes before editing.

2. Collect the complete PR feedback surface.
   - Prefer the bundled collector:

     ```bash
     python3 <skill-dir>/scripts/collect_pr_feedback.py <pr-url-or-number>
     ```

   - Read the generated Markdown report first, then inspect the JSON when
     thread metadata, reply structure, or comment IDs are needed.
   - If the collector cannot run, manually gather the same surfaces with `gh`:
     PR body, PR-body and issue-comment reactions with actor identities, issue
     comments, review bodies, review comments, commits, files, and review-thread
     resolution status when available.
   - Check the report's `Potential Outside-Diff Sources` section, then search
     the collected artifacts case-insensitively for `outside diff`,
     `outside the diff`, `Actionable comments`, `Nitpick comments`,
     `Prompt for all review comments`, and bot names.

3. Gate action on active Codex and CodeRabbit reviews.
   - Determine each service's state for the latest head commit from the freshly
     collected reactions, trigger comments, status comments, reviews, and final
     summaries. Inspect PR checks or status contexts when those artifacts do
     not make the current state clear.
   - Count only reactions authored by the Codex connector account
     (`chatgpt-codex-connector[bot]`) or a verified replacement identity.
   - Treat 👀 (`eyes`) on the PR body or an `@codex review` comment as accepted
     or in-progress review evidence.
   - Treat 👍 (`+1`) as a completed Codex review with no findings, even when
     Codex posted no review body or inline comment. Record the reaction target,
     actor, and timestamp as completion evidence and do not retrigger.
   - Treat a Codex-authored review with findings as completed review evidence.
   - Treat a CodeRabbit processing or status message that says a review is
     underway as in-progress evidence; treat its completed review or final
     summary as completed evidence.
   - Reactions from other actors and aggregate reaction counts without actor
     identities do not establish Codex state.
   - While either service has an active review, wait and rerun the collector at
     a reasonable interval. Begin assessment, ledger creation, checkout, edits,
     pushes, and replies only after every observed active review completes and
     the feedback surface has been refreshed. An active signal consumes the
     current request; wait instead of triggering the service again.
   - Keep intake blocked when an active review explicitly fails or stalls beyond
     a reasonable task wait window; diagnose and report the review failure. If
     neither service has an active review, proceed with the feedback already
     available; this gate does not require triggering an absent review.

4. Critically assess each finding before planning fixes.
   - Treat suggested patches from CodeRabbit, other bots, or reviewers as
     proposals, not instructions; do not apply them blindly.
   - Decide whether the proper action is no-op with evidence, a narrow fix,
     added validation, or a larger refactor that addresses an underlying code
     smell or design flaw.
   - Challenge review claims against current code, requirements, and PR intent.
     Mark incorrect, stale, duplicate, or harmful suggestions explicitly in the
     ledger.

5. Build a concise feedback ledger before editing.
   - Track every actionable item with source, URL or comment ID, path/line when
     available, current status, planned handling, and eventual reply target.
   - Include actionable findings embedded in review bodies or bot summary
     comments, even when there is no separate GitHub comment to resolve.
   - Mark duplicate findings together, but keep all source links so replies can
     acknowledge every place the issue was raised.
   - Verify each item against current code before changing anything. Skip
     stale or already-fixed items with a brief reason.

## Local Fix Loop

1. Check out the PR branch with `gh pr checkout <pr-url-or-number>` or the
   repository's existing branch workflow.
2. Fetch and confirm the local branch matches the remote head before editing;
   if the remote has new commits, reconcile and re-check the feedback ledger
   first.
3. Fix feedback in small cohesive groups.
   - Commit as progress is made, one commit per feedback cluster or subsystem.
   - Keep each commit message specific enough to map back to handled feedback.
4. After each meaningful fix group, run targeted validation for the touched
   code. Keep expanding validation only when the blast radius warrants it.
5. Run a local implementation review pass over the follow-up diff before
   pushing.
   - Check correctness, regressions, validation gaps, edge cases, tests,
     contracts, and avoidable complexity introduced by the fixes.
   - Apply obvious safe fixes locally and commit them in the relevant group or
     in a small follow-up cleanup commit.
   - Leave broader design decisions for the user unless the PR feedback
     clearly requires them.
6. Push the branch after local validation passes or after a completed bounded
   group when the task is long-running and remote visibility matters. Refresh
   the feedback surface after each push and re-enter the intake review gate if
   Codex or CodeRabbit starts another review.

## Reply Workflow

1. Reply only after the relevant commit is pushed.
2. For inline review comments, reply to the thread with the GitHub review
   comment reply endpoint, for example:

   ```bash
   gh api -X POST repos/OWNER/REPO/pulls/PR_NUMBER/comments/COMMENT_ID/replies \
     -f body="$(cat /tmp/reply.md)"
   ```

3. For PR-level comments, review-body findings, bot summary findings, and
   outside-diff items without a resolvable thread, add one concise PR comment
   that lists:
   - commit hash or hashes pushed;
   - handled items grouped by source;
   - validation run and any blocked validation;
   - items intentionally skipped as stale, duplicate, or not applicable.
4. Do not mention skill names, local automation internals, or implementation
   process details in PR comments. Write as the PR author/operator explaining
   what changed and how it was validated.
5. Prefer concrete replies over vague closure language:
   - `Addressed in abc1234: dry-run now performs the same conflict check as write mode, with focused CLI coverage.`
   - `Verified current code already has a separate health-check timeout in def5678, so this thread is stale after the latest push.`
6. If a finding is rejected, explain the current code evidence and tradeoff
   briefly. Do not argue with bot style comments; keep the reply factual.

## Completion Check

Before finishing, confirm:

- no Codex or CodeRabbit review remains active, and the feedback surface was
  refreshed after the latest review completed;
- all PR-body and issue-comment reactions, issue comments, review bodies, review
  comments, replies, and outside-diff sections were read;
- each actionable item is fixed, replied to, explicitly skipped, or left as a
  user decision;
- every follow-up commit is pushed;
- local validation and local review results are captured in the final PR
  comment or thread replies;
- the final response to the user includes pushed commit hashes, validation, and
  any remaining PR feedback that needs human judgment.
