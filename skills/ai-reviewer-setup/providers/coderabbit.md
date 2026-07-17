# CodeRabbit Setup

## Installation and documentation

- Confirm that the GitHub or GitLab integration is installed and authorized
  for the repository.
- Use the official [YAML configuration
  guide](https://docs.coderabbit.ai/getting-started/yaml-configuration),
  [auto-review
  documentation](https://docs.coderabbit.ai/configuration/auto-review), and
  [configuration schema](https://docs.coderabbit.ai/reference/configuration)
  when exact behavior matters.

## Configuration location and review cadence

- Use `.coderabbit.yaml` in the repository root unless the repository already
  deliberately uses `.coderabbit.yml`. CodeRabbit reads the file from the
  feature branch under review.
- CodeRabbit reviews subsequent pushes by default. Set
  `reviews.auto_review.auto_incremental_review: false` to review only when the
  PR opens and rely on manual `@coderabbitai review` commands for later passes.
- If the team wants limited automatic follow-up, set
  `reviews.auto_review.auto_pause_after_reviewed_commits` to `1` or `2`.
- For manual-only review, set `reviews.auto_review.enabled: false` and
  optionally configure an opt-in label or description keyword.
- CodeRabbit reviews PRs targeting the default branch by default.
  `reviews.auto_review.base_branches` adds targets; it does not replace the
  default branch.

## Start conservatively

After removing path exclusions that do not match repository output, use a
compact baseline such as:

```yaml
language: "en-US"
reviews:
  profile: "chill"
  request_changes_workflow: false
  high_level_summary: true
  review_status: true
  review_details: false
  path_filters:
    - "!dist/**"
    - "!build/**"
    - "!coverage/**"
  auto_review:
    enabled: true
    auto_incremental_review: false
    drafts: false
    base_branches: []
    ignore_title_keywords:
      - "WIP"
      - "[skip review]"
    ignore_usernames:
      - "dependabot[bot]"
      - "renovate[bot]"
chat:
  auto_reply: true
```

## Tailor only to repository evidence

- Add `reviews.path_filters` exclusions only for paths that exist or are
  clearly generated.
- Add `reviews.path_instructions` only for concrete subsystem review rules.
- Keep `request_changes_workflow: false` unless the user wants CodeRabbit to
  request changes.
- Avoid pre-merge checks in `error` mode unless the user explicitly wants
  CodeRabbit to block merges.

## Validate

- Use a repository-native YAML command when available. Otherwise run
  `ruby -e 'require "yaml"; YAML.load_file(".coderabbit.yaml")'`, adjusting the
  filename when the repository deliberately uses `.coderabbit.yml`.
- When an installed app and test PR are available, verify initial, update,
  draft, author, and path-filter behavior affected by the configuration.
