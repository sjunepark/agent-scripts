---
name: ai-reviewer-setup
description: "Set up or update Greptile and CodeRabbit AI code review configuration in a project. Use when adding root greptile.json or .coderabbit.yaml files, tuning PR auto-review behavior, restricting Greptile to PRs targeting main, excluding generated files or bots, or documenting installation steps for these review bots."
---

# AI Reviewer Setup

Use this skill to add repo-local Greptile and CodeRabbit configuration without pretending repository files install or authorize the services. Keep the setup low-noise, reviewable, and explicit about any dashboard or GitHub App work the user still needs to do.

## Workflow

1. Inspect the project before editing.
   - Check `git status --short` and do not overwrite unrelated user changes.
   - Look for existing `greptile.json`, `.coderabbit.yaml`, `.coderabbit.yml`, `.github/`, agent docs, style guides, generated-output directories, bot usage, and current branch/default-branch evidence.
   - Merge with existing config files. Do not replace working settings wholesale.

2. Confirm service installation boundaries.
   - Greptile needs the GitHub/GitLab integration enabled and the repository selected for review.
   - CodeRabbit needs its GitHub/GitLab integration installed and authorized for the repository.
   - If local evidence cannot prove the app is installed, add a concise manual follow-up rather than blocking the config change.

3. Use official docs when exact behavior matters.
   - Greptile: `https://www.greptile.com/docs/code-review-bot/greptile-json`
   - Greptile trigger filters: `https://www.greptile.com/docs/code-review-bot/trigger-code-review`
   - CodeRabbit YAML config: `https://docs.coderabbit.ai/getting-started/yaml-configuration`
   - CodeRabbit auto-review controls: `https://docs.coderabbit.ai/configuration/auto-review`
   - CodeRabbit full schema: `https://docs.coderabbit.ai/reference/configuration`

4. Prefer low-usage review triggers unless the user asks for continuous review.
   - Avoid configuring both services to review every push on the same PR. It usually spends review quota quickly with low additional signal.
   - Greptile reviews only the initial PR by default. Keep `triggerOnUpdates` omitted or `false` unless the team explicitly wants a fresh Greptile review after each new commit.
   - CodeRabbit reviews subsequent pushes by default. Set `reviews.auto_review.auto_incremental_review: false` when the team wants CodeRabbit to review only when the PR is opened and rely on manual `@coderabbitai review` for later passes.
   - If the team still wants some automatic CodeRabbit follow-up, set `reviews.auto_review.auto_pause_after_reviewed_commits` to `1` or `2` instead of leaving the default higher threshold.
   - For manual-only CodeRabbit workflows, set `reviews.auto_review.enabled: false` and optionally add an opt-in label or description keyword.

5. Add or update Greptile.
   - Use `greptile.json` in the repository root.
   - Greptile reads settings from the PR source branch.
   - To satisfy "Greptile should run only for PRs in main", configure PR target-branch filtering:

```json
{
  "includeBranches": ["main"]
}
```

   - Keep `includeBranches: ["main"]` unless the user explicitly asks for additional target branches.
   - Do not rely on `excludeBranches` alone for the main-only requirement.
   - Prefer a compact, high-signal Greptile config such as:

```json
{
  "includeBranches": ["main"],
  "triggerOnUpdates": false,
  "triggerOnDrafts": false,
  "strictness": 2,
  "commentTypes": ["logic", "syntax"],
  "disabledLabels": ["do-not-review", "wip"],
  "excludeAuthors": ["dependabot[bot]", "renovate[bot]"],
  "instructions": "Prioritize correctness, security, and maintainability. Avoid style comments already covered by automated formatters."
}
```

6. Add Greptile options only when earned by the project.
   - Use `ignorePatterns` as a newline-separated `.gitignore`-style string for generated, vendored, build, coverage, snapshot, or lockfile paths that should not be reviewed.
   - Use `customContext.files` when the repository already has style guides, architecture docs, or agent instructions that should guide reviews.
   - Use `triggerOnUpdates: true` only when the team wants Greptile reviews after each new commit and accepts the added review volume.
   - Use `skipReview: "AUTOMATIC"` only when the team wants Greptile to run from manual `@greptileai` triggers instead of automatic PR-open reviews.
   - Use `statusCheck: true` only when the team wants a GitHub status check. Do not configure branch protection or merge-blocking checks unless the user asks.

7. Add or update CodeRabbit.
   - Use `.coderabbit.yaml` in the repository root unless the repo already has a deliberate `.coderabbit.yml`.
   - CodeRabbit reads the config from the feature branch under review.
   - CodeRabbit auto-reviews PRs targeting the default branch by default. `reviews.auto_review.base_branches` adds more target branches; it does not replace the default branch.
   - Use conservative defaults unless the user asks for stricter merge behavior:

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

8. Tailor CodeRabbit to the repository.
   - Add `reviews.path_filters` exclusions only for paths that actually exist or are clearly generated.
   - Add `reviews.path_instructions` for real subsystem-specific review rules, not generic language advice.
   - Keep `request_changes_workflow: false` unless the user wants CodeRabbit to request changes.
   - Keep `reviews.auto_review.auto_incremental_review: false` unless the user explicitly wants automatic re-reviews after each push.
   - Use `reviews.auto_review.auto_pause_after_reviewed_commits: 1` or `2` as a compromise only when automatic incremental reviews should stay enabled but stop quickly on busy branches.
   - Avoid pre-merge checks in `error` mode unless the user explicitly wants CodeRabbit to block merges.

9. Validate before finishing.
   - Validate JSON with `python3 -m json.tool greptile.json >/dev/null`.
   - Validate YAML with an available parser, for example `ruby -e 'require "yaml"; YAML.load_file(".coderabbit.yaml")'` or a repo-native YAML command.
   - Run any repository validation that covers config files.
   - If the services are installed and a test PR is practical, verify a PR targeting `main` triggers the intended initial reviews, a PR targeting another branch does not, and an additional push does not trigger automatic re-reviews when update reviews are disabled.

## Final Response

Report the files changed, the Greptile main-target-branch rule, the CodeRabbit incremental-review setting, validation results, and any remaining app-install/dashboard work. Mention skipped service-level verification if no installed app or test PR was available.
