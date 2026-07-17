# Snippets

Use a snippet only when exact code is load-bearing for the concept and prose would be ambiguous or hide a crucial decision.

## Selection rules

- Embed the excerpt directly in the response so the learner does not need to return to the editor.
- Prefer exact contract shapes, conditions, data shapes, boundaries, state transitions, or APIs.
- Omit snippets that merely prove a file exists, restate syntax, or walk through routine implementation detail.
- Prefer prose or pseudocode when the relationship matters more than exact syntax.

## Presentation rules

- Use a fenced code block and keep the excerpt narrowly scoped.
- Put the source path on the first line as a language-appropriate comment, such as `// src/dashboard/load.ts` or `# backend/jobs/sync.py`.
- Prefer signatures, branching points, conditions, state transitions, and interface boundaries over long contiguous excerpts.
- Follow the snippet with the design fact it demonstrates.

Example:

```ts
// src/dashboard/loadUserDashboard.ts
export async function loadUserDashboard(userId: string) {
  const account = await accountRepo.getByUserId(userId);
  return buildDashboardView(account);
}
```

This shows an orchestration boundary: the repository fetches data, the view builder shapes it, and the function coordinates them without owning either responsibility.
