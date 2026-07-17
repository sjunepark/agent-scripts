# Migration and compatibility

- Do not churn existing pre-runes files into modern syntax unless migration is the task or the touched code becomes materially clearer.
- When touching legacy code, prefer local consistency plus targeted improvements over mixing paradigms inside one file.
- In new code, prefer runes over `$:` and `export let`, snippets over slots, current event attributes over `on:click`, direct component references over `<svelte:component>`, and reactive classes or objects over stores for local shared state.
- Use promise-in-template features such as await expressions only when the project is on a compatible Svelte version and the relevant experimental option is enabled.
- Verify version-sensitive behavior against the installed Svelte version and the [Svelte 5 migration guide](https://svelte.dev/docs/svelte/v5-migration-guide).
