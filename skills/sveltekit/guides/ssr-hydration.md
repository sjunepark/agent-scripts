# SSR and hydration

- Guard browser-only APIs such as `window`, `document`, storage, observers, and media APIs behind `onMount` or a browser check.
- Keep server and browser rendering deterministic. Hydration mismatches usually come from browser-only render branches, time, randomness, locale differences, or data that changes between SSR and hydration.
- Prefer stable SSR markup that upgrades in the browser over incompatible server and client markup.
- Keep critical interactions usable before hydration when practical. Prefer real links and forms, or render clear disabled or loading affordances for JavaScript-only controls.
- Remember that universal `load` runs on the server for SSR and again in the browser for hydration; keep server-only assumptions out of it.
