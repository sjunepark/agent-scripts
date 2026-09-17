---
name: address-pr-feedback
description: "Address existing GitHub PR feedback from human or bot reviewers end to end, including stacked PR chains."
---

# Address PR Feedback

Handle existing PR feedback through assessment, fixes, validation, push, and
reviewer-facing replies. Requires Git and authenticated `gh`; the optional
collector also requires Python 3. PR feedback includes review bodies, issue
comments, bot summaries, outside-diff findings, inline threads, and follow-ups.

For dependent PRs, read [Stacked PRs](workflows/stacked-prs.md) before mapping or
changing the chain. Feedback handling alone does not authorize retargeting,
history rewrites, or merging; use the landing route only when authorized.

## Collect and assess

Identify the repository, PR, base/head, author, latest head SHA, and review state.
If no PR is named, infer the current branch PR with `gh pr view`.

Prefer `scripts/collect_pr_feedback.py` for complete feedback and reaction
collection:

```bash
python3 <skill-dir>/scripts/collect_pr_feedback.py <pr-url-or-number>
```

Read its Markdown report and use JSON for IDs, reply structure, or thread metadata.
If unavailable, gather equivalent surfaces with `gh`: PR body and reactions,
issue comments and reactions, reviews, inline comments and replies, commits,
files, and available thread-resolution state. Include reaction actor identities.
Inspect `Potential Outside-Diff Sources` and any actionable/nitpick or
"Prompt for all review comments" sections; keyword detection does not replace
reading the full feedback.

Apply the active-review gate below before assessment or edits. Then keep a
concise ledger of actionable items with source links/IDs, affected code, status,
handling, and reply target. Include findings without resolvable threads and
retain all source links when grouping duplicates. Verify proposals against
current code and PR intent; mark stale, already-fixed, incorrect, or harmful
suggestions with evidence instead of applying them blindly.

## Active-review gate

Manual Codex and CodeRabbit requests are one-time PR-level requests, not
per-push requirements. Do not retrigger for incremental follow-up pushes.
If neither service is active, use existing feedback without requiring an absent
review. Handle reviews that start automatically.

Classify state from fresh reactions, trigger/status comments, reviews, summaries,
and checks when needed, correlating the evidence with the current head:

- Only `chatgpt-codex-connector[bot]` or a verified replacement establishes
  Codex reaction state. Aggregate counts or other actors do not.
- Codex `eyes` on the PR body or an `@codex review` comment means accepted or
  in progress. Codex `+1` means completed with no findings, even without a
  review body; record target, actor, and timestamp. A Codex-authored review
  with findings also proves completion.
- A CodeRabbit processing/status message means in progress; its completed
  review or final summary proves completion.

Wait for every observed active review and refresh feedback before assessment,
ledger creation, checkout, edits, pushes, or replies. An active signal consumes
the request; do not trigger again. Diagnose an explicit failure or a review that
stalls beyond a reasonable wait window and report intake as blocked.

## Fix and publish

Check out the PR branch with `gh pr checkout <pr-url-or-number>` or the repository
workflow. Preserve unrelated work, fetch, and reconcile new remote commits
before editing; reassess affected feedback when the head changes.

Fix in cohesive groups with commits that map to the handled feedback. Use the
scope and delegated judgment to choose a narrow correction or a necessary
refactor; resolve consequential design decisions that remain unsettled.
Run targeted and required validation and one local implementation review over
the follow-up diff, fixing actionable findings.

Push after validation, or after a completed bounded group when a long-running
task needs remote visibility. Refresh feedback after each push; return to the
active-review gate if Codex or CodeRabbit starts another review.

## Reply and finish

Reply after the relevant commit is pushed. Use the review-comment reply endpoint
for inline findings, preserving the reply text in a file:

```bash
gh api -X POST repos/OWNER/REPO/pulls/PR_NUMBER/comments/COMMENT_ID/replies -F body=@/path/to/reply.md
```

For review-body, PR-level, bot-summary, and outside-diff findings without a thread,
post one concise PR comment grouping handled items by source, with pushed commits,
validation, and reasons for skipped or rejected suggestions. Write concrete
code-based replies as the PR operator; omit local skill and automation details.

Finish when every actionable item is fixed and replied to, explicitly skipped
with evidence, or identified as an unresolved user decision; all follow-up
commits are pushed; and validation/review evidence is recorded in replies.
No Codex or CodeRabbit review may remain active, and feedback must have been
refreshed after the latest review completed. For a stack, cover every selected
PR before declaring the selected scope complete.

Report pushed commits, validation, and any unresolved feedback or concrete
blocker. A blocked review or incomplete push is not completed feedback handling.
