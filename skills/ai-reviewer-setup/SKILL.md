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

4. Add or update Greptile.
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
  "triggerOnDrafts": false,
  "strictness": 2,
  "commentTypes": ["logic", "syntax"],
  "disabledLabels": ["do-not-review", "wip"],
  "excludeAuthors": ["dependabot[bot]", "renovate[bot]"],
  "instructions": "Prioritize correctness, security, and maintainability. Avoid style comments already covered by automated formatters."
}
```

5. Add Greptile options only when earned by the project.
   - Use `ignorePatterns` as a newline-separated `.gitignore`-style string for generated, vendored, build, coverage, snapshot, or lockfile paths that should not be reviewed.
   - Use `customContext.files` when the repository already has style guides, architecture docs, or agent instructions that should guide reviews.
   - Use `triggerOnUpdates: true` only when the team wants reviews after each new commit.
   - Use `statusCheck: true` only when the team wants a GitHub status check. Do not configure branch protection or merge-blocking checks unless the user asks.

6. Add or update CodeRabbit.
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

7. Tailor CodeRabbit to the repository.
   - Add `reviews.path_filters` exclusions only for paths that actually exist or are clearly generated.
   - Add `reviews.path_instructions` for real subsystem-specific review rules, not generic language advice.
   - Keep `request_changes_workflow: false` unless the user wants CodeRabbit to request changes.
   - Avoid pre-merge checks in `error` mode unless the user explicitly wants CodeRabbit to block merges.

8. Validate before finishing.
   - Validate JSON with `python3 -m json.tool greptile.json >/dev/null`.
   - Validate YAML with an available parser, for example `ruby -e 'require "yaml"; YAML.load_file(".coderabbit.yaml")'` or a repo-native YAML command.
   - Run any repository validation that covers config files.
   - If the services are installed and a test PR is practical, verify a PR targeting `main` triggers Greptile and a PR targeting another branch does not.

## Final Response

Report the files changed, the Greptile main-target-branch rule, validation results, and any remaining app-install/dashboard work. Mention skipped service-level verification if no installed app or test PR was available.
