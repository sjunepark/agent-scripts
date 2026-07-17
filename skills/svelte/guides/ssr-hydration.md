# SSR and hydration

- Assume components inside SvelteKit render on the server before hydrating in the browser.
- Guard browser-only APIs such as `window`, `document`, storage, observers, and media APIs behind `onMount` or an explicit browser check.
- Keep render output deterministic across server and browser. Check browser-only branches, time, randomness, locale-sensitive formatting, and data that can change between SSR and hydration.
- Prefer stable server-rendered markup that upgrades in the browser over incompatible server and client branches.
- When hydration fails, compare the actual SSR markup with the first browser render before changing component state logic.
