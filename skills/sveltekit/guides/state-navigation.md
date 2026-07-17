# State, navigation, and invalidation

## State placement

- Put SSR-relevant or shareable state in the URL when it should survive reloads, deep links, and server rendering.
- Keep disposable UI state in the component when losing it on reload is acceptable.
- Use snapshots when ephemeral UI state should survive history navigation without becoming URL or server state.
- Prefer context over module-level singletons for state shared only within one rendered subtree.
- Avoid over-centralizing state; prefer the URL or page data when they explain route behavior more directly.

## Navigation and invalidation

- Use `goto` only for genuinely imperative navigation; prefer `href` and native forms when they already express the flow.
- After mutations, invalidate the narrowest dependency possible instead of reaching reflexively for `invalidateAll()`.
- Route components are reused across navigation, and rerunning a `load` updates props without recreating the component. When reset semantics matter, key the subtree with `{#key ...}`, derive from `page.url`, or use `afterNavigate`.

## Debugging

- If state unexpectedly persists across navigation, account for reused layouts and pages; key the subtree or derive from route state.
- If state unexpectedly resets, decide whether it belongs in component state, URL parameters, snapshots, or server data.

## Official documentation

- [State management](https://svelte.dev/docs/kit/state-management)
- [Invalidation](https://svelte.dev/tutorial/kit/invalidation)
