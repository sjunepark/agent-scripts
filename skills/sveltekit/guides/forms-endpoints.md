# Forms and endpoints

- Prefer form actions for user-driven mutations that naturally come from a page form.
- Keep actions close to the route that owns the UI and validation unless multiple routes truly share the mutation.
- Return validation errors with `fail(...)` rather than throwing generic errors for expected bad input.
- Use generated route-specific action types from `$types`.
- Use `method="GET"` with URL parameters for idempotent filter and search forms.
- Use `use:enhance` for progressive enhancement, not to replace the native form flow with a custom client-side mutation system.
- Reach for `applyAction` or a custom enhance callback only when the default enhanced flow is insufficient.
- Use `+server.*` endpoints for API-style interactions, webhooks, non-form clients, or cases where the action model is awkward.
- If a form behaves differently with and without JavaScript, compare native action behavior with the custom `use:enhance` path.

## Official documentation

- [Form actions](https://svelte.dev/docs/kit/form-actions)
