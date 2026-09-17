---
name: create-pr
description: "Create or update GitHub PRs with gh, including stacked PR bases and CodeRabbit or Greptile review controls."
---

# Create PR

Create a reviewable PR from the intended diff and validation evidence. Requires
Git and authenticated GitHub CLI access.

## Workflow

1. Inspect the PR surface before opening anything.
   - Check current branch, upstream, uncommitted changes, existing PR status, base branch evidence, relevant diff, commits, and validation output.
   - A request to create or update a PR authorizes the scoped commits and push needed to publish its intended changes, unless the user limits the task to drafting or inspection. Reuse earlier authority and preserve unrelated changes. If the intended diff or publication target is materially unclear, inspect and validate the candidate diff and prepare the PR draft before asking about that boundary.
   - If a PR already exists for the branch, update it instead of creating a duplicate.
   - For a stacked PR, read
     [workflows/stacked-prs.md](workflows/stacked-prs.md) before creating or
     retargeting branches or PRs. Verify the integration branch, parent
     sequence, and landing order from repository evidence and existing instructions;
     ask only when a material choice remains unresolved.

2. Draft the PR from evidence.
   - Use the repo's PR template when present, but remove irrelevant prompts.
   - Keep the title specific and merge-history friendly.
   - Explain the concrete problem, resulting behavior, validation, and material
     risks. Use sections only when they help the reviewer.
   - Mention skipped validation plainly. Do not imply tests passed because the code looks plausible.
   - Use issue-closing keywords only when the user clearly wants the linked issue closed on merge.

3. Create or update with `gh`.
   - Resolve the AI review decision below and load any applicable bot-control
     guide before creating the PR; some controls must be present in the initial
     body or labels.
   - Prefer `gh pr create --base <base> --head <branch> --title <title> --body-file <file>` so multiline bodies and bot commands are preserved exactly.
   - Create a ready-for-review PR by default. Add `--draft` only when the user asks for a draft, the user calls the work WIP, repo instructions require drafts for this case, or the PR is explicitly meant for early visibility rather than review. Incomplete validation alone does not justify a draft.
   - Use other reviewer, label, assignee, and milestone flags only when requested or clearly supported by repo convention.
   - After creation or edit, verify with `gh pr view --json url,number,title,state,baseRefName,headRefName,isDraft,labels`.

## AI Review Decision

Before creating a PR, inspect `.coderabbit.yaml`, `.coderabbit.yml`, and
`greptile.json` if present. Apply the user's and repository's existing review
policy. In the absence of a more specific policy, request the initial CodeRabbit
review with `coderabbit-review` and let configuration or service defaults govern
remaining reviews. Suppress a review only when the selected policy calls for it;
draft status or incomplete validation alone does not override that policy.

Do not edit review-bot config just to open a PR unless the user asked for a config change or there is no PR-level way to achieve the requested behavior; when a config edit is needed, say so before making it.

When CodeRabbit may run, be suppressed, or be triggered manually—or existing
CodeRabbit configuration needs interpretation—read
[review-bots/coderabbit.md](review-bots/coderabbit.md) before finalizing the PR
body and labels.

When the user requests Greptile control, `greptile.json` exists, or repository
evidence shows that Greptile reviews PRs, read
[review-bots/greptile.md](review-bots/greptile.md) before allowing, suppressing,
or manually triggering it.

Resolve a material review-policy gap only when existing instructions do not
settle it, after preparing the PR and applicable controls. Honor requests for
suppression or manual-only review through the relevant guide; do not turn the
default policy into a mandatory choice prompt.

## Post-Create

After opening or updating the PR:

- Report the PR URL, base/head, draft status, and any reviewer/label assignments.
- For a stacked PR, report its position, dependency, integration branch, and next landing action.
- State exactly what happened with CodeRabbit and Greptile: defaulted, disabled, opt-in label/body keyword used, manual trigger posted, or config change needed but not made.
- If bot comments were requested, post them with `gh pr comment <number-or-url> --body '<command>'` only after the PR exists.
