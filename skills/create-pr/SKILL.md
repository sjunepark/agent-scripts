---
name: create-pr
description: "Create or update GitHub pull requests with gh. Use when drafting a PR title/body, choosing draft versus ready state, or deciding whether CodeRabbit or Greptile reviews run, are skipped, or are manually triggered."
---

# Create PR

Create focused PRs that reviewers and review bots can act on without extra clarification.

## Workflow

1. Inspect the PR surface before opening anything.
   - Check current branch, upstream, uncommitted changes, existing PR status, base branch evidence, relevant diff, commits, and validation output.
   - If uncommitted changes are part of the intended PR, stop and ask whether to commit them first.
   - If a PR already exists for the branch, update it instead of creating a duplicate.

2. Draft the PR from evidence.
   - Use the repo's PR template when present, but remove irrelevant prompts.
   - Keep the title specific and merge-history friendly.
   - Include only sections with real content. Useful sections are `Summary`, `Validation`, `Review notes`, `Risk`, and `AI review`.
   - Mention skipped validation plainly. Do not imply tests passed because the code looks plausible.
   - Use issue-closing keywords only when the user clearly wants the linked issue closed on merge.

3. Create or update with `gh`.
   - Prefer `gh pr create --base <base> --head <branch> --title <title> --body-file <file>` so multiline bodies and bot commands are preserved exactly.
   - Create a ready-for-review PR by default. Add `--draft` only when the user asks for a draft, the user calls the work WIP, repo instructions require drafts for this case, or the PR is explicitly meant for early visibility rather than review. Incomplete validation alone does not justify a draft.
   - Apply the `coderabbit-review` label by default for ready-for-review PRs: use `--label coderabbit-review` when creating, or `gh pr edit --add-label coderabbit-review` when updating. Omit it only when the user asks to skip CodeRabbit, the PR is a draft/WIP, the PR is noisy/generated/dependency-only, validation is materially incomplete, or repo instructions say not to run CodeRabbit for this PR.
   - Use other reviewer, label, assignee, and milestone flags only when requested or clearly supported by repo convention.
   - Treat review-bot opt-in labels as bot execution controls; `coderabbit-review` is the default approved CodeRabbit control for normal ready PRs.
   - After creation or edit, verify with `gh pr view --json url,number,title,state,baseRefName,headRefName,isDraft,labels`.

## AI Review Decision

Before creating a PR, inspect `.coderabbit.yaml`, `.coderabbit.yml`, and `greptile.json` if present. If the user does not specify and the PR is ready for human review, apply `coderabbit-review` and let the remaining repo config/defaults run; if the PR is draft, noisy, generated, dependency-only, or validation-incomplete, suppress automatic bot reviews when a low-risk PR-level control exists.

Do not edit review-bot config just to open a PR unless the user asked for a config change or there is no PR-level way to achieve the requested behavior; when a config edit is needed, say so before making it.

### CodeRabbit

Use these PR-time controls:

- Disable CodeRabbit for one PR by putting `@coderabbitai ignore` in the PR description before creation.
- Trigger manually after creation with a PR comment: `@coderabbitai review` for incremental review or `@coderabbitai full review` for a fresh full review.
- Pause or resume automatic reviews on an open PR with `@coderabbitai pause` or `@coderabbitai resume`.
- Opt in when auto-review is disabled by config with a configured `reviews.auto_review.labels` label or exact `reviews.auto_review.description_keyword` in the PR body. Use `coderabbit-review` as the default opt-in label for normal ready PRs. CodeRabbit labels can trigger reviews even when `enabled: false`.
- Skip through existing config-aware signals such as configured `ignore_title_keywords` or negative labels in `reviews.auto_review.labels`.

Treat a CodeRabbit trigger comment, opt-in label, or description keyword as spending review quota. For normal ready PRs, `coderabbit-review` is approved by default and should be applied. Do not add or post other CodeRabbit triggers unless the user asked for them, confirmed the review decision, or the repository has a clear trusted workflow/policy that auto-applies them for this PR.

### Greptile

Use these PR-time controls:

- Trigger manually after creation with a PR comment containing `@greptileai`.
- Draft PRs are skipped by default, but `@greptileai` can manually trigger a draft review.
- Use configured `labels`, `disabledLabels`, `includeKeywords`, `ignoreKeywords`, `includeBranches`, `excludeBranches`, `includeAuthors`, or `excludeAuthors` when they already exist.
- To disable automatic Greptile review while preserving manual `@greptileai` triggers, set `"skipReview": "AUTOMATIC"` in `greptile.json` on the PR source branch.
- `triggerOnUpdates: true` reviews each new commit; otherwise Greptile's default is an initial PR review.

Greptile reads `greptile.json` from the PR source branch, so branch-local config changes affect the PR being opened.

### Bot-Control Prompt

When the user wants a choice or review cost/noise is material, ask for a compact decision before creating the PR:

```text
AI review handling for this PR?
- default: apply `coderabbit-review` and let remaining CodeRabbit/Greptile config decide
- disable: suppress automatic reviews where PR-level controls allow it
- manual: create the PR without automatic review, then trigger selected bots by comment when requested
- run: allow automatic review and optionally post manual trigger comments after creation
```

Translate the answer into concrete PR actions.

## Post-Create

After opening or updating the PR:

- Report the PR URL, base/head, draft status, and any reviewer/label assignments.
- State exactly what happened with CodeRabbit and Greptile: defaulted, disabled, opt-in label/body keyword used, manual trigger posted, or config change needed but not made.
- If bot comments were requested, post them with `gh pr comment <number-or-url> --body '<command>'` only after the PR exists.
