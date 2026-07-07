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

3. Model data flow before polishing markup. Name the single source of truth for each piece of state; compute the rest with `$derived`.

4. Assume SvelteKit components render on the server first; avoid browser-only assumptions during render (see Common Gotchas).

## Core Svelte Patterns

- Use runes-first patterns in new code: `$state`, `$derived`, `$props`, snippets, and current event attributes such as `onclick={...}`.
- Use `$state` only for values that truly need reactivity. Keep plain variables plain.
- Prefer `$state.raw` for large objects or arrays that are usually replaced wholesale rather than mutated in place.
- Prefer `$derived` for computed values. Use `$derived.by(...)` when the derivation needs multiple statements.
- Remember that objects or arrays returned from `$derived` are not made deeply reactive.
- Treat `$effect` as an escape hatch for DOM APIs, timers, analytics, imperative libraries, or external subscriptions.
- Derive, don't sync: do not use `$effect` to keep one piece of state in sync with another — if you write `x` from `y` in an effect, you want a derived value or a different source of truth.
- Prefer event handlers, function bindings, `{@attach ...}`, or `createSubscriber(...)` before reaching for an effect.
- Treat props as live values: derive from them instead of snapshotting into local mutable state, unless the component intentionally forks from the parent value.
- Prefer snippets and explicit children rendering over legacy slot patterns in new code.
- Top-level snippets can also be referenced from `<script>` when that is clearer than creating a tiny helper component.
- Use `<svelte:window>` and `<svelte:document>` for global listeners instead of wiring them up manually in `onMount` or `$effect`.
- Prefer keyed `{#each}` blocks for stable identity. Do not use array index as the key.
- Avoid destructuring each-block items when the template mutates them.
- Prefer CSS custom properties and `style:--token={value}` for JS-to-CSS handoff and parent-controlled styling.
- Reach for `:global(...)` only when styling third-party output or a boundary you do not control.
- Prefer `createContext` over ad hoc `setContext`/`getContext` pairs when context is the right tool.
- Do not default to stores for local sharing. In Svelte 5, a reactive object or class with `$state` fields is often simpler.
- Keep stores for cases where the store contract matters: async streams, external subscriptions, or ecosystem integrations.

## Component Design Best Practices

- Keep presentational components narrow: props in, callbacks out, minimal ownership of app state.
- Prefer explicit props and callback props over "smart" components with many optional behaviors.
- In new Svelte 5-style code, prefer callback props over `createEventDispatcher`.
- Prefer a few well-named props over catch-all config objects unless the object is a real domain shape.
- When a component exposes many booleans or variant combinations, split it or redesign the API around clearer modes.
- Keep render helpers close when they are tiny; extract components only when they create a real boundary.
- Put async loading, auth checks, and persistence in higher-level modules rather than leaf UI components.
- Watch for components that both orchestrate data flow and render complex UI; split responsibilities before the file turns into a control tower.

## Common Gotchas

- Component setup runs once. If a value should update when props or state change, make it reactive with `$derived` instead of computing it once in plain script.
- Passing a primitive from `$state(...)` to a function passes the current value, not a live reference. Pass an updater or a container object when mutation must flow back.
- Reactive churn often comes from depending on a whole object when only one stable field matters. Extract the stable field into its own `$derived` before building downstream derived state.
- Deeply reactive `$state` is convenient, but proxying large graphs has a cost. Prefer smaller state shapes or `$state.raw` when mutation granularity is not needed.
- Mutating a property on an object returned from `$derived` affects the original object only if that original object is already reactive upstream. Do not assume `$derived` deepens reactivity.
- Browser-only APIs such as `window`, `document`, storage, observers, and media APIs must be guarded behind `onMount` or an explicit browser check when SSR is in play.

## Debugging Checklist

- If UI is stale, verify the value is `$state` or `$derived` rather than a one-time plain variable.
- If updates are looping or hard to follow, remove syncing effects and re-establish one source of truth (derive, don't sync).
- If something reruns more often than expected, add `$inspect.trace(...)` to the relevant `$derived.by(...)` or `$effect` instead of leaving logging effects in place.
- If child state drifts from parent props, stop snapshotting props into local mutable variables.
- If shared reactive state resets or recreates too often, check whether downstream logic depends on an unstable object instead of a stable field.
- If hydration fails, compare SSR markup with browser-only branches, time/randomness, and locale-sensitive formatting.

## Migration and Compatibility

- For existing pre-runes code, do not churn files into modern syntax unless the task is migration or the touched code becomes materially clearer.
- When touching legacy code, prefer local consistency plus small targeted improvements over mixing paradigms inside one file.
- Prefer modern replacements in new code: runes over `$:` and `export let`, snippets over slots, current event attributes over `on:click`, direct component references over `<svelte:component>`, and reactive classes or objects over stores when sharing local reactive state.
- Use promise-in-template features such as await expressions only when the project is on a compatible Svelte version and has the relevant experimental option enabled.

## Official References

Consult the current Svelte docs when behavior matters or seems version-sensitive:
- `https://svelte.dev/docs/svelte/best-practices`
- `https://svelte.dev/docs/svelte/$state`
- `https://svelte.dev/docs/svelte/$derived`
- `https://svelte.dev/docs/svelte/$effect`
- `https://svelte.dev/docs/svelte/$props`
