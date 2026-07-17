# Greptile PR Controls

Use this guide after making the parent skill's AI review decision.

## Apply the selected control

- Trigger a review after creation with a PR comment containing `@greptileai`.
- Draft PRs are skipped by default, but `@greptileai` can manually trigger a
  draft review.
- Use configured `labels`, `disabledLabels`, `includeKeywords`,
  `ignoreKeywords`, `includeBranches`, `excludeBranches`, `includeAuthors`, or
  `excludeAuthors` when they already exist.
- To disable automatic review while preserving manual `@greptileai` triggers,
  set `"skipReview": "AUTOMATIC"` in `greptile.json` on the PR source branch.

## Account for branch-local configuration

Greptile reads `greptile.json` from the PR source branch, so a branch-local
configuration change affects the PR being opened. With
`triggerOnUpdates: true`, Greptile reviews each new commit; otherwise it reviews
the initial PR by default.
