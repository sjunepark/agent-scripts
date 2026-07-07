---
name: ui-lab
description: "Build or extend a lightweight in-app UI Lab: a dev-only route rendering real production UI states from deterministic fixtures shared with UI tests. Use when the user asks for a UI lab, scenario gallery, or Storybook-like in-app preview; wants to see a dialog, panel, or hard-to-reach UI state without clicking through the app; wants fixtures shared between a visual review page and UI tests; or is adding scenarios to an existing lab."
---

# UI Lab

A UI Lab is a small dev-only page inside a web app that renders real production UI in deterministic scenarios. It is for states that are hard or slow to reach by clicking through the live app: expired auth, empty/error states, long-content overflow, loading/success/failure panels, large tables, dialogs, popovers, and similar product surfaces.

Keep it lightweight. Start with the smallest useful path and add polish only when the scenario set grows.

Adapt naming and APIs to the target framework, build tool, router, and test runner.

Not for: a real Storybook/Ladle install the user explicitly wants; a full design-system catalog; backend-only work.

## Basic shape

Use these five pieces unless the existing app already has a clearer equivalent.

1. **Scenario contract** — a minimal record describing one renderable UI state.
   - Include: `id`, `title`, optional `description`, render target/component/view, and a fresh fixture factory such as `createProps`, `createState`, or `setup`.
   - Use ids like `<group>/<state>` so they work in URLs and tests.
   - The fixture factory must create fresh data for every render so mutable state does not leak between scenarios or tests.
   - Do not add controls, viewport metadata, knobs, or custom schema until there is a real need. If a scenario needs layout, size, an anchor, or a controller, use a small wrapper/stage component instead.

2. **Manual registry** — an explicit ordered list of groups and scenarios.
   - Prefer hand-written imports over auto-discovery while the lab is small.
   - Validate duplicate group ids and scenario ids at load time.
   - Keep ordering useful for humans: by feature area, then by surface/state.

3. **Shared scenario host** — the same mount wrapper used by the lab page and tests.
   - Install the real app providers the UI expects: theme, router context, i18n, stores, query/cache client, modal/toast roots, feature flags, or dependency injection.
   - Stub external services at this boundary. Scenarios should not call live network, analytics, auth, storage, IPC, or other production side effects.
   - Keep the host small; it exists to make production UI mount correctly, not to simulate the whole app.

4. **Dev-only lab route** — a route/page such as `/__ui-lab`, `/dev/ui-lab`, or `#/ui-lab`.
   - Gate it with the app's dev/build flag so lab code is excluded from production builds or release artifacts.
   - Deep-link to selected scenarios (`/__ui-lab/<id>` or `#/ui-lab/<id>`) so reviewers and tests can open a state directly.
   - Keep the shell simple: scenario list, preview pane, current title/description, and a remount/reset button. Add filter, keyboard shortcuts, theme toggles, screenshots, or visual-regression hooks only when needed.

5. **Smoke test** — one UI/component/browser test that imports the registry and shared host, then mounts every scenario.
   - Assert every scenario renders without crashing and without live services.
   - Capture obvious failures: thrown errors, unhandled promises, console errors if the test stack supports it, and scenario setup failures.
   - Add at least one example assertion against real product DOM to show tests can reuse a scenario without the lab shell.

## Guardrails

- **No production dependency on lab code.** Production UI may be imported by the lab; production code must not import the lab.
- **Do not ship the lab.** Use a static dev-only build guard where the stack supports it, then verify with the app's normal production build/checks.
- **Deterministic scenarios only.** No live network, randomness, dates based on now, or live credentials. Avoid timers that affect rendered state; use fixed timestamps, stable ids, inert callbacks, and immediate success/failure values.
- **Prefer fixture state over scripted setup.** Seed props/state directly when possible. Use DOM interaction replay only for states that truly require interaction behavior such as focus, keyboard, pointer, drag/drop, or portal positioning.
- **Do not add test-only props to production components.** Add props only when they are legitimate component boundaries. Otherwise keep lab-specific setup in wrappers.
- **Delete stale scenarios.** If the product state no longer exists, remove the scenario instead of preserving dead fixtures.

## Selecting scenarios

Favor product states that are valuable to review but expensive to reach live:

- auth/session/device states: pending, denied, expired, disconnected
- async states: loading, success, empty, error, retrying, partial data
- validation states: conflict, duplicate name, invalid input, permission denied
- content stress: long labels, long descriptions, overflow, empty lists, no search results
- scale states: large tables/lists, pagination, dense rows, many chips/tags
- transient surfaces: dialogs, popovers, toasts, banners, command palettes

Skip states that are genuinely live-only, and leave a short comment explaining why if the omission is non-obvious.

## Build order

1. Confirm the app's framework, router, build/dev flag, provider setup, and UI test runner.
2. Find existing test fixtures, domain builders, mock data, and provider/test utilities; reuse them in scenario fixtures instead of inventing parallel mock data.
3. Add the scenario contract, shared host, manual registry, and duplicate-id validation.
4. Wire the dev-only lab route without changing the normal app boot path.
5. Add 1–2 useful scenario groups end to end.
6. Add the smoke test that mounts all registered scenarios through the host.
7. Add more scenarios only after the first path is validated.

## Validate

Run the repository's normal checks over touched files: type-check, lint, format, production build, and the relevant UI/component/browser tests.
