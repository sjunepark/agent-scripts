# Greptile Setup

## Installation and documentation

- Confirm that the GitHub or GitLab integration is enabled and the repository
  is selected for review.
- Use the official [`greptile.json`
  documentation](https://www.greptile.com/docs/code-review-bot/greptile-json)
  and [trigger-filter
  documentation](https://www.greptile.com/docs/code-review-bot/trigger-code-review)
  when exact behavior matters.

## Configuration location and review cadence

- Use `greptile.json` in the repository root. Greptile reads it from the PR
  source branch.
- Greptile reviews the initial PR by default. Keep `triggerOnUpdates` omitted
  or `false` unless the team explicitly wants a fresh review after each commit.
- Use `skipReview: "AUTOMATIC"` when the team wants only manual
  `@greptileai` triggers.

## Target the intended branches

- To review only PRs targeting the repository's default branch, set
  `includeBranches` to the default branch discovered during inspection.
- Keep `includeBranches` limited to that branch unless the user asks for more
  targets. `excludeBranches` alone does not establish a default-branch-only
  policy.
- Replace `<default-branch>` below with the discovered branch before writing
  the file.

```json
{
  "includeBranches": ["<default-branch>"],
  "triggerOnUpdates": false,
  "triggerOnDrafts": false,
  "strictness": 2,
  "commentTypes": ["logic", "syntax"],
  "disabledLabels": ["do-not-review", "wip"],
  "excludeAuthors": ["dependabot[bot]", "renovate[bot]"],
  "instructions": "Prioritize correctness, security, and maintainability. Avoid style comments already covered by automated formatters."
}
```

## Add options only when earned

- Use `ignorePatterns` as a newline-separated `.gitignore`-style string for
  generated, vendored, build, coverage, snapshot, or lockfile paths.
- Use `customContext.files` when existing style guides, architecture docs, or
  agent instructions should guide reviews.
- Use `statusCheck: true` only when the team wants a GitHub status check. Leave
  branch protection and merge-blocking checks unchanged unless the user asks.

## Validate

- Run `python3 -m json.tool greptile.json >/dev/null`.
- When an installed app and test PR are available, verify that a PR targeting
  the configured branch triggers the intended initial review, a PR targeting
  another branch does not, and an additional push does not trigger a new
  automatic review when update reviews are disabled.
