---
name: address-pr-feedback
description: "Act on existing GitHub pull request feedback end to end. Use when Codex needs to read PR review comments, issue comments, review bodies, bot summaries, outside-diff review sections, fix code locally, run local review passes, make follow-up commits, push the PR branch, and reply to PR feedback with what was handled."
---

# Address PR Feedback

Handle an existing GitHub PR from review intake through local fixes, validation,
push, and reviewer-facing replies. Treat PR feedback as broader than inline
threads: review bodies, issue comments, bot summaries, "outside diff" sections,
and follow-up comments can all contain actionable items.

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

   - Read the generated Markdown report first, then inspect the JSON when body
     text, thread metadata, replies, or IDs are needed.
   - If the collector cannot run, manually gather the same surfaces with `gh`:
     PR body, issue comments, review bodies, review comments, commits, files,
     and review-thread resolution status when available.
   - Search the collected artifacts for `Outside diff`, `Outside Diff`,
     `Comments Outside Diff`, `outside the diff`, `Actionable comments`,
     `Nitpick comments`, `Prompt for all review comments`, and bot names.

3. Build a concise feedback ledger before editing.
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
2. Fetch and compare against the remote head before editing.
3. Fix feedback in small cohesive groups.
   - Prefer one commit per feedback cluster or subsystem.
   - Commit as progress is made; do not accumulate a single large follow-up
     commit when the work naturally splits.
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
   group when the task is long-running and remote visibility matters.

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

- all issue comments, review bodies, review comments, replies, and outside-diff
  sections were read;
- each actionable item is fixed, replied to, explicitly skipped, or left as a
  user decision;
- follow-up commits are local and pushed;
- local validation and local review results are captured in the final PR
  comment or thread replies;
- the final response to the user includes pushed commit hashes, validation, and
  any remaining PR feedback that needs human judgment.
