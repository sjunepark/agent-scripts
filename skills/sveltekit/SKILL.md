---
name: sveltekit
description: "SvelteKit apps — build, review, refactor, and debug. Use for routes (`+page/+layout/+server`), `load`, form actions, hooks, cookies, auth, invalidation, SSR, hydration, navigation, and URL/server/client state boundaries, even for small bugfixes or features."
---

# SvelteKit

Optimize for route and data flows that stay correct under SSR, hydration, navigation, and progressive enhancement. Focus on boundaries and the shortest correct implementation.

## Scope

Use this skill when the hard part is app-level or route-level behavior. If the task is mostly about component internals, runes, snippets, local state, or styling, use the `svelte` skill.

## Workflow

1. Establish the local style before changing code.
   - Inspect `package.json`, `svelte.config.*`, the route tree, hooks, and nearby route files.
   - Prefer existing conventions in touched files unless the task is an explicit migration.
   - For new code, default to current SvelteKit APIs and typed route files.

2. Choose the server/client boundary before writing code.
   - Put server-only work in `+page.server.*`, `+layout.server.*`, `+server.*`, hooks, or `$lib/server/*`.
   - Put reusable view logic in components or `$lib`, not route files.
   - Keep request-specific mutable state out of shared modules so server state cannot leak across users.
   - Keep route flow obvious: params, URL, cookies, and locals -> load or action -> typed props -> component.

3. Solve data flow before UI details.
   - Decide whether truth belongs in the URL, server data, component state, snapshots, context, or a shared reactive object.
   - Derive state instead of synchronizing copies.
   - Decide whether data reruns automatically or only after targeted invalidation.

4. Make navigation and SSR behavior explicit.
   - Check whether code runs on the server, in the browser, or both.
   - Decide whether state survives reloads, deep links, history navigation, or only the current render.
   - Check whether route components are reused across navigation and whether anything must reset explicitly.

5. Validate the relevant server, browser, and progressively enhanced paths with the repository's existing checks.
   - When behavior is version-sensitive, verify it against the installed Svelte and SvelteKit versions and current official documentation.

## Topic guides

- When changing route files, `load`, parent data, dependency tracking, or serialization, read [guides/load-routing.md](guides/load-routing.md).
- When implementing a form, mutation, progressive enhancement, or API endpoint, read [guides/forms-endpoints.md](guides/forms-endpoints.md).
- When handling hooks, authentication, cookies, secrets, or request-local identity, read [guides/auth-cookies.md](guides/auth-cookies.md).
- When placing state or changing navigation, invalidation, history, or reset behavior, read [guides/state-navigation.md](guides/state-navigation.md).
- When browser-only code, server rendering, hydration, or pre-hydration interaction is involved, read [guides/ssr-hydration.md](guides/ssr-hydration.md).
- When adopting renamed APIs or another version-sensitive SvelteKit feature, read [guides/migration.md](guides/migration.md).
