# Migration and compatibility

- Use current names and helpers on compatible versions, such as `$app/state` instead of older store-based access and `PageProps` or `LayoutProps` where available.
- Keep version-sensitive features behind compatibility checks; some depend on particular Svelte or SvelteKit versions or configuration.
- Verify the installed versions and local migration state before mixing old and current APIs in a touched route.
