# CodeRabbit PR Controls

Use this guide after making the parent skill's AI review decision.

## Apply the selected control

- For the parent skill's normal ready-PR policy, add `coderabbit-review` with
  `--label coderabbit-review` during creation or
  `gh pr edit --add-label coderabbit-review` during an update.
- Disable CodeRabbit for one PR by placing `@coderabbitai ignore` in the PR
  description before creation.
- Trigger after creation with `@coderabbitai review` for an incremental review
  or `@coderabbitai full review` for a fresh full review.
- Pause or resume automatic reviews on an open PR with `@coderabbitai pause` or
  `@coderabbitai resume`.

## Respect repository configuration

- When auto-review is disabled, opt in with a configured
  `reviews.auto_review.labels` label or the exact
  `reviews.auto_review.description_keyword` in the PR body. Label triggers can
  run even when `reviews.auto_review.enabled` is `false`.
- Honor configured skip signals such as `ignore_title_keywords` or negative
  labels in `reviews.auto_review.labels`.
- Treat opt-in labels as execution controls, not ordinary metadata.

## Preserve the review budget

A trigger comment, opt-in label, or description keyword spends review quota.
Do not add or post a trigger beyond the parent skill's selected control unless
the user requested it or a trusted repository workflow clearly authorizes it
for this PR.
