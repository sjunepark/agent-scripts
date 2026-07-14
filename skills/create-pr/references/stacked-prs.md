# Stacked PRs

Use a dedicated integration branch as the stack's safety boundary. The preferred name is `staging/<stack-name>`; follow an equivalent repository convention when one exists. A per-stack branch keeps unrelated work isolated better than a shared `staging` branch.

If the repository uses `dev` as an alternative integration base to `main`, fetch
both remote branches and verify that `origin/dev` and `origin/main` resolve to
the same commit before building the stack. Do not assume `dev` is current: if
the tips differ, stop and report that `dev` must be synchronized or that the
stack needs an explicitly chosen base. Do not update, merge, or rewrite `dev`
without separate authorization.

## Build the stack

1. Create `staging/<stack-name>` from the repository's normal protected base, usually the default branch, before creating any slice branches.
2. Branch the first slice from the integration branch and each later slice from its immediate predecessor.
3. Open the first PR against the integration branch. Open each later PR against its predecessor so every review initially shows one slice.
4. Record the stack root, dependency, successor, and landing order in each PR body. Update these links after GitHub assigns PR numbers.

Example:

```text
main
└── staging/search-index
    └── stack/search-index/01-schema      PR #101 → staging/search-index
        └── stack/search-index/02-writer  PR #102 → stack/search-index/01-schema
            └── stack/search-index/03-api PR #103 → stack/search-index/02-writer
```

Use explicit bases with `gh pr create --base <parent> --head <slice>`. Verify every base/head pair after creation.

## Shape good slices

- Give each PR one coherent purpose and enough validation to review on its own.
- Put foundations before consumers: types or schema, implementation, then integration or UI.
- Separate mechanical refactors from behavior changes when that makes later diffs easier to trust.
- Keep every intermediate commit and branch buildable when practical. State unavoidable temporary limitations in the PR body.
- Use titles such as `[1/3] Add search index schema` when stack position would otherwise be unclear.
- Keep issue-closing keywords for the eventual promotion PR when the issue should close only after the stack reaches the default branch.

Suggested PR-body block:

```markdown
## Stack

- Integration branch: `staging/search-index`
- Depends on: #101
- Followed by: #103
- Landing order: #101 → #102 → #103
- Promotion to default branch: separate, explicitly authorized PR
```

## Land bottom-up

1. Select the lowest ready PR whose base is the integration branch.
2. Before merging it, retarget its immediate child to the integration branch with `gh pr edit <child> --base staging/<stack-name>`. The child's diff will temporarily include both slices.
3. Merge the selected PR, then verify that the child's diff contracts to its intended slice. Recheck its commits, conflicts, and required checks; refresh the branch when repository policy requires it.
4. Repeat until every slice has landed in the integration branch.

Treat an instruction to merge a stack or its PRs as authorization for these in-stack merges only. Keep the integration branch as the destination throughout agent-managed landing.

## Promote separately

Promotion is a distinct PR from `staging/<stack-name>` to the default branch. Create or merge it only when the user explicitly requests promotion to that named base; do not infer that authorization from instructions to create, update, or merge the stack.

Before promotion, validate the integrated result, summarize the complete user-facing change, and confirm that the promotion diff contains only the intended stack. Retain or remove the integration branch according to repository policy after promotion is complete.
