---
name: create-pr
description: "Create or update GitHub pull requests from a repository. Use when drafting PR title/body, opening PRs with gh, choosing draft versus ready state, selecting reviewers or labels, or deciding whether CodeRabbit or Greptile reviews should run, be skipped, or be manually triggered for a PR."
---

# Create PR

Create focused PRs that reviewers and review bots can act on without extra clarification. Prefer the repository's PR template and `gh` when available.

## Workflow

1. Inspect the PR surface before opening anything.
   - Check current branch, upstream, uncommitted changes, existing PR status, base branch evidence, relevant diff, commits, and validation output.
   - If uncommitted changes are part of the intended PR, stop and ask whether to commit them first. Do not create a PR that silently excludes visible work.
   - If a PR already exists for the branch, update it instead of creating a duplicate.

2. Draft the PR from evidence.
   - Use the repo's PR template when present, but remove irrelevant prompts.
   - Keep the title specific and merge-history friendly.
   - Include only sections with real content. Useful sections are `Summary`, `Validation`, `Review notes`, `Risk`, and `AI review`.
   - Mention skipped validation plainly. Do not imply tests passed because the code looks plausible.
   - Use issue-closing keywords only when the user clearly wants the linked issue closed on merge.

3. Create or update with `gh`.
   - Prefer `gh pr create --base <base> --head <branch> --title <title> --body-file <file>` so multiline bodies and bot commands are preserved exactly.
   - Create a ready-for-review PR by default. Add `--draft` only when the user asks for a draft, the user calls the work WIP, repo instructions require drafts for this case, or the PR is explicitly meant for early visibility rather than review.
   - When validation is incomplete, keep the PR ready by default and state the missing validation plainly in the body. Use a draft only when the incomplete validation makes the PR intentionally not ready for review.
   - Use `--reviewer`, `--label`, `--assignee`, and `--milestone` only when requested or clearly supported by repo convention.
   - After creation or edit, verify with `gh pr view --json url,number,title,state,baseRefName,headRefName,isDraft,labels`.

## AI Review Decision

Before creating a PR, inspect `.coderabbit.yaml`, `.coderabbit.yml`, and `greptile.json` if present. If the user asks to decide, prompt briefly with the available choices. If the user does not specify and the PR is ready for human review, let the repo config/defaults run; if the PR is draft, noisy, generated, dependency-only, or validation-incomplete, suppress automatic bot reviews when a low-risk PR-level control exists.

Do not edit review-bot config just to open a PR unless the user asked for a config change or there is no PR-level way to achieve the requested behavior.

### CodeRabbit

Use these PR-time controls:

- Disable CodeRabbit for one PR by putting `@coderabbitai ignore` in the PR description before creation.
- Trigger manually after creation with a PR comment: `@coderabbitai review` for incremental review or `@coderabbitai full review` for a fresh full review.
- Pause or resume automatic reviews on an open PR with `@coderabbitai pause` or `@coderabbitai resume`.
- Opt in when auto-review is disabled by config with a configured `reviews.auto_review.labels` label or exact `reviews.auto_review.description_keyword` in the PR body. CodeRabbit labels can trigger reviews even when `enabled: false`.
- Skip through existing config-aware signals such as configured `ignore_title_keywords` or negative labels in `reviews.auto_review.labels`.

Treat a CodeRabbit trigger comment as spending review quota. Do not post one unless the user asked for it or confirmed the review decision.

### Greptile

Use these PR-time controls:

- Trigger manually after creation with a PR comment containing `@greptileai`.
- Draft PRs are skipped by default, but `@greptileai` can manually trigger a draft review.
- Use configured `labels`, `disabledLabels`, `includeKeywords`, `ignoreKeywords`, `includeBranches`, `excludeBranches`, `includeAuthors`, or `excludeAuthors` when they already exist.
- To disable automatic Greptile review while preserving manual `@greptileai` triggers, set `"skipReview": "AUTOMATIC"` in `greptile.json` on the PR source branch, but only if a config diff is acceptable for this PR.
- `triggerOnUpdates: true` reviews each new commit; otherwise Greptile's default is an initial PR review.

Greptile reads `greptile.json` from the PR source branch, so branch-local config changes affect the PR being opened.

## Bot-Control Prompt

When the user wants a choice or review cost/noise is material, ask for a compact decision before creating the PR:

```text
AI review handling for this PR?
- default: let existing CodeRabbit/Greptile config decide
- disable: suppress automatic reviews where PR-level controls allow it
- manual: create the PR without automatic review, then trigger selected bots by comment when requested
- run: allow automatic review and optionally post manual trigger comments after creation
```

Translate the answer into concrete PR actions and report them in the final response. If a requested Greptile override requires changing `greptile.json`, say that explicitly before editing.

## Post-Create

After opening or updating the PR:

- Report the PR URL, base/head, draft status, and any reviewer/label assignments.
- State exactly what happened with CodeRabbit and Greptile: defaulted, disabled, opt-in label/body keyword used, manual trigger posted, or config change needed but not made.
- If bot comments were requested, post them with `gh pr comment <number-or-url> --body '<command>'` only after the PR exists.
