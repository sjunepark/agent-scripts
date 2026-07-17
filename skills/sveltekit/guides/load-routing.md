# Load and routing

## Route and load rules

- Use `+page.server.*` or `+layout.server.*` when code needs secrets, direct database access, cookies, or trusted request context.
- Use universal `load` only when the logic is safe on both server and client and rerunning it in the browser is acceptable.
- Keep `load` side-effect free. Fetch, read, and derive there; mutate in actions, endpoints, or other server handlers.
- Use the `fetch` provided by SvelteKit inside `load` so cookies, auth, SSR inlining, and internal request shortcuts work correctly.
- Return typed, already-shaped data from `load` instead of making every consumer reshape it.
- Minimize serialized data. Anything returned from server `load` joins the SSR payload, is visible to the client, and adds HTML and hydration cost.
- Use generated `$types` helpers such as `PageProps`, `LayoutProps`, and route-specific load types.
- Use `page.data` when parent layouts need child data, but do not create hidden coupling casually.
- Use `parent()` deliberately. Start independent async work first to avoid route-level waterfalls from sequential `await`s or `parent()` ordering.
- Keep layout loads lean; do not fetch data most child routes never use.
- Use `depends()` and targeted invalidation when automatic reruns are insufficient.
- Keep request-specific data out of shared modules; mutable module state can leak across users on the server.

## Debugging

- If data is stale, identify which `load` owns it, which dependencies it tracks, whether `depends()` or invalidation is missing, and whether the code is universal or server-only.
- If hydration is slow, inspect the SSR payload for data returned from `load` that the page does not use.

## Official documentation

- [Loading data](https://svelte.dev/docs/kit/load)
