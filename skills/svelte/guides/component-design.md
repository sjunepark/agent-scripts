# Component design

## APIs and ownership

- Keep presentational components narrow: props in, callbacks out, and minimal ownership of app state.
- Prefer explicit props and callback props over smart components with many optional behaviors.
- In new Svelte 5-style code, prefer callback props over `createEventDispatcher`.
- Prefer a few well-named props over catch-all config objects unless the object is a real domain shape.
- When a component exposes many booleans or variant combinations, split it or redesign its API around clearer modes.
- Keep tiny render helpers close; extract components only when they create a real boundary.
- Put async loading, auth checks, and persistence in higher-level modules rather than leaf UI components.
- Split components that both orchestrate data flow and render complex UI before they become control towers.

## Templates and events

- Prefer snippets and explicit children rendering over legacy slot patterns in new code.
- Reference top-level snippets from `<script>` when that is clearer than creating a tiny helper component.
- Use current event attributes such as `onclick={...}` in new code.
- Use `<svelte:window>` and `<svelte:document>` for global listeners instead of wiring them manually in `onMount` or `$effect`.
- Prefer keyed `{#each}` blocks for stable identity; do not use an array index as the key.
- Avoid destructuring each-block items when the template mutates them.

## Styling boundaries

- Prefer CSS custom properties and `style:--token={value}` for JavaScript-to-CSS handoff and parent-controlled styling.
- Reach for `:global(...)` only when styling third-party output or another boundary you do not control.

## Official documentation

- [Best practices](https://svelte.dev/docs/svelte/best-practices)
