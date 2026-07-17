# Authentication and cookies

- Prefer hooks for broad auth gates that must run before protected loads.
- Prefer checks in `+page.server.*` for page-specific protection.
- Treat auth checks in `+layout.server.*` carefully: child pages benefit reliably only when they call `await parent()` before protected work.
- Keep secrets, trusted headers, and raw cookies on the server.
- Let the server decide auth state once and return the necessary result as page data instead of re-reading auth state in client code.
- After setting or deleting auth cookies in an action, update `event.locals` when later `load` logic depends on it. `handle` runs before the action and does not rerun before the post-action load.
- Set an explicit `path: '/'` when a cookie should be site-wide.

## Debugging

- If auth is inconsistent, verify that protected reads happen on the server and that cookie writes also update `event.locals` when the same request uses it.
- If cookies differ across routes, verify `path`, domain expectations, and whether the code runs in an action, endpoint, or server load.
