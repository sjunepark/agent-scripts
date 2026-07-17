---
name: svelte
description: "Svelte 5 component work: build, review, refactor, and debug `.svelte` components. Use for runes, props, snippets, events, styling, context, reactive state, and reactivity gotchas, including SvelteKit apps when the issue is component-level rather than route/server-level."
---

# Svelte

Work from current Svelte 5 idioms. Skip basic-syntax explanations unless asked; give the shortest correct implementation.

## Scope

If the hard part is route boundaries, `load`, form actions, auth, cookies, SSR navigation behavior, or invalidation, switch to the `sveltekit` skill.

## Workflow

1. Establish the local style before changing code.
   - Inspect nearby components, shared UI primitives, and `package.json`.
   - Prefer the repo's current conventions in touched files unless the task is an explicit migration.
   - For new code, default to runes mode and current event syntax.

2. Choose the ownership boundary before writing code.
   - Decide whether truth should live in props, component state, context, or a shared reactive module.

3. Model data flow before polishing markup.
   - Name the single source of truth for each piece of state and derive the rest.

4. Check the runtime boundary.
   - Assume SvelteKit components render on the server first and avoid browser-only assumptions during render.

5. Implement the smallest coherent change, then validate it with the repository's existing checks and the relevant runtime path.
   - When behavior is version-sensitive, verify it against the installed version and current official documentation.

## Topic guides

- When choosing or debugging state, props, derived values, effects, context, or stores, read [guides/reactivity.md](guides/reactivity.md).
- When designing a component API, template structure, events, snippets, or styling boundary, read [guides/component-design.md](guides/component-design.md).
- When browser APIs, server rendering, or a hydration mismatch are involved, read [guides/ssr-hydration.md](guides/ssr-hydration.md).
- When touching pre-runes code or using version-sensitive features, read [guides/migration.md](guides/migration.md).
