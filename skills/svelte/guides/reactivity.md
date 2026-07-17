# Reactivity

## State and derivation

- Use runes-first patterns in new code: `$state`, `$derived`, and `$props`.
- Use `$state` only for values that need reactivity; keep plain variables plain.
- Prefer `$state.raw` for large objects or arrays that are replaced wholesale rather than mutated in place.
- Prefer `$derived` for computed values. Use `$derived.by(...)` when a derivation needs multiple statements.
- Remember that objects or arrays returned from `$derived` are not made deeply reactive.
- Derive, don't sync. If an effect writes `x` from `y`, make `x` derived or choose a different source of truth.

## Effects and external systems

- Treat `$effect` as an escape hatch for DOM APIs, timers, analytics, imperative libraries, or external subscriptions.
- Prefer event handlers, function bindings, `{@attach ...}`, or `createSubscriber(...)` before reaching for an effect.
- When diagnosing unexpected reruns, add `$inspect.trace(...)` to the relevant `$derived.by(...)` or `$effect` rather than leaving logging effects in place.

## Props and shared ownership

- Treat props as live values. Derive from them instead of snapshotting them into local mutable state unless the component intentionally forks from the parent value.
- Prefer `createContext` over ad hoc `setContext`/`getContext` pairs when context is the right boundary.
- Do not default to stores for local sharing. A reactive object or class with `$state` fields is often simpler in Svelte 5.
- Keep stores when the store contract matters, such as async streams, external subscriptions, or ecosystem integrations.

## Gotchas and debugging

- Component setup runs once. Make values that must follow props or state reactive with `$derived` instead of computing them once in plain script.
- Passing a primitive from `$state(...)` to a function passes its current value, not a live reference. Pass an updater or container object when mutation must flow back.
- Reactive churn often comes from depending on a whole object when only one stable field matters. Extract that field into its own `$derived` before building downstream state.
- Deeply reactive `$state` proxies large graphs. Prefer smaller state shapes or `$state.raw` when mutation granularity is unnecessary.
- Mutating a property on an object returned from `$derived` affects the original only when that object is already reactive upstream. `$derived` does not deepen reactivity.
- If UI is stale, verify the value is `$state` or `$derived`, not a one-time plain variable.
- If updates loop or become hard to follow, remove syncing effects and re-establish one source of truth.
- If child state drifts from parent props, stop snapshotting props into local mutable variables.
- If shared reactive state resets or recreates too often, check for downstream dependencies on an unstable object rather than a stable field.

## Official documentation

- [State](https://svelte.dev/docs/svelte/$state)
- [Derived state](https://svelte.dev/docs/svelte/$derived)
- [Effects](https://svelte.dev/docs/svelte/$effect)
- [Props](https://svelte.dev/docs/svelte/$props)
