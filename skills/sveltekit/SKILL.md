---
name: sveltekit
description: "SvelteKit apps — build, review, refactor, and debug. Use for routes (`+page/+layout/+server`), `load`, form actions, hooks, cookies, auth, invalidation, SSR, hydration, navigation, and URL/server/client state boundaries, even for small bugfixes or features."
---

# SvelteKit

Optimize for route and data flows that stay correct under SSR, hydration, navigation, and progressive enhancement.

Focus on boundaries, best practices, gotchas, and the shortest correct implementation.

## Scope

Use this skill when the hard part is app-level or route-level behavior.

- If the task is mostly about component internals, runes, snippets, local state, or styling, use the `svelte` skill.

## Workflow

1. Establish the local style before changing code.
- Inspect `package.json`, `svelte.config.*`, the route tree, hooks, and nearby route files.
- Prefer the repo's existing conventions in touched files unless the task is an explicit migration.
- For new code, default to current SvelteKit APIs and typed route files.

2. Choose the server/client boundary before writing code.
- Put server-only work in `+page.server.*`, `+layout.server.*`, `+server.*`, hooks, or `$lib/server/*`.
- Put reusable view logic in components or `$lib`, not route files.
- Keep route flow obvious: params/url/cookies/locals -> load/action -> typed props -> component.

3. Solve data flow before UI details.
- Decide where truth should live: URL, server data, component state, snapshots, context, or a shared reactive object.
- Prefer deriving state over synchronizing copies.
- Be explicit about whether data should rerun automatically or only after targeted invalidation.

4. Be explicit about navigation and SSR behavior.
- Check whether code runs on the server, in the browser, or both.
- Check whether state should survive reloads, deep links, back/forward, or only the current render.
- Check whether the route component is reused across navigation and whether anything should reset explicitly.

## Route and Load Rules

- Use `+page.server.*` or `+layout.server.*` when code needs secrets, direct database access, cookies, or trusted request context.
- Use universal `load` only when the logic is safe on both server and client and rerunning in the browser is acceptable.
- Keep `load` side-effect free. Fetch, read, and derive there; mutate in actions, endpoints, or other server handlers.
- Use the `fetch` provided by SvelteKit inside `load` so cookies, auth, SSR inlining, and internal request shortcuts work correctly.
- Prefer returning typed, already-shaped data from `load` instead of pushing reshaping work into every consumer.
- Minimize serialized data. Anything returned from server `load` becomes part of the SSR payload — visible to the client and paid for in HTML size and hydration time. Shape returns to what the route renders.
- Use generated `$types` helpers like `PageProps`, `LayoutProps`, and route-specific load/action types.
- Use `page.data` when parent layouts need child data, but do not create hidden coupling casually.
- Use `parent()` deliberately and start independent async work first; route-level waterfalls come from avoidable sequential `await`s and `parent()` ordering.
- Keep layout `load`s lean; do not fetch data that most child routes never use.
- Use `depends()` and targeted invalidation when automatic reruns are insufficient.
- Do not store request-specific data in shared modules. Mutable module state can leak across users on the server.

## Actions, Forms, and Endpoints

- Prefer form actions for user-driven mutations that naturally come from a page form.
- Keep actions close to the route that owns the UI and validation unless multiple routes truly share the mutation.
- Return validation errors with `fail(...)` rather than throwing generic errors for expected bad input.
- Use `method="GET"` plus URL params for idempotent filter and search forms.
- Use `use:enhance` for progressive enhancement, not to replace the native form flow with a custom client-side mutation system.
- Reach for `applyAction` or custom enhance callbacks only when the default enhanced flow is insufficient.
- Use `+server.*` endpoints for API-style interactions, webhooks, non-form clients, or cases where the action model is awkward.
- After setting or deleting auth cookies in an action, update `event.locals` as well when later `load` logic depends on it. `handle` runs before the action and does not rerun before the post-action `load`.
- When a cookie should be site-wide, set an explicit `path: '/'`.

## Auth and Request Boundaries

- Prefer hooks for broad auth gates that should run before protected loads.
- Prefer route-specific checks in `+page.server.*` for page-specific protection.
- Be careful with auth checks in `+layout.server.*`: child pages only reliably benefit when they call `await parent()` before protected work.
- Keep secrets, trusted headers, and raw cookies on the server side.
- Do not read auth state in client code when the server can decide once and return the result as page data.

## State Placement

- Put SSR-relevant or shareable state in the URL when it should survive reloads, deep links, and server rendering.
- Keep disposable UI state in the component if losing it on reload is fine.
- Use snapshots when ephemeral UI state should survive history navigation without becoming URL or server state.
- If state is shared only within one rendered subtree, prefer context over module-level singletons.
- Avoid over-centralizing state; prefer URL or page data when they explain route behavior more directly.

## SSR and Hydration Gotchas

- Guard browser-only APIs such as `window`, `document`, storage, observers, and media APIs behind `onMount` or a browser check.
- Hydration mismatches usually come from non-deterministic rendering: browser-only branches during SSR, time or randomness during render, locale differences, or data that changes between SSR and hydration.
- Prefer stable SSR markup that upgrades in the browser over branching into incompatible markup.
- Critical interactions should still work before hydration when possible. Prefer real links and forms for important user actions, or render clear disabled/loading affordances for JS-only controls.
- Universal `load` runs on the server for SSR and then again in the browser for hydration. Do not put server-only assumptions there.

## Navigation and Invalidation

- Use `goto` only when navigation is genuinely imperative; prefer `href` and native forms that already express the flow.
- After mutations, invalidate the narrowest dependency you can — not `invalidateAll()` as a reflex when one dependency or URL key would do.
- Route components are reused across navigation, and rerunning a `load` updates props without recreating the component. When reset semantics matter, key the subtree with `{#key ...}`, derive from `page.url`, or use `afterNavigate`.

## Debugging Checklist

- If data is stale, check which `load` owns it, what dependencies it tracks, whether `depends()` or invalidation is missing, and whether the code is universal or server-only.
- If state unexpectedly persists across navigation, remember that layouts and pages may be reused. Key the subtree or derive from route state.
- If state unexpectedly resets, verify whether it belongs in component state, URL params, snapshots, or server data.
- If a form behaves differently with and without JavaScript, compare native action behavior against custom `use:enhance` logic.
- If auth behaves inconsistently, verify that protected reads happen on the server and that `event.locals` is updated after cookie writes.
- If cookies appear inconsistent across routes, verify `path`, domain expectations, and whether the code is running in an action, endpoint, or server load.
- If hydration feels slow or flaky, inspect SSR payload size and look for data returned from `load` that the page does not actually use.

## Migration and Compatibility

- Use current names and helpers on compatible versions, such as `$app/state` over older store-based access and `PageProps`/`LayoutProps` where available.
- Keep version-sensitive features behind compatibility checks. Some features depend on specific Svelte or SvelteKit versions or config options.

## Official References

Consult the current SvelteKit docs when behavior matters or seems version-sensitive:
- `https://svelte.dev/docs/kit/load`
- `https://svelte.dev/docs/kit/form-actions`
- `https://svelte.dev/docs/kit/state-management`
- `https://svelte.dev/tutorial/kit/invalidation`
