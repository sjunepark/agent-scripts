---
name: ui-lab
description: "Build an in-app UI lab — a dev-only catalog that mounts real production components in hard-to-reach states with deterministic fixtures, reusable from browser tests. Use when asked to set up a Storybook-style scenario gallery, isolate/preview UI states (dialogs, panels, empty/error/overflow states), or share fixtures between a visual lab and automated tests, in any app regardless of framework."
---

# UI Lab

A UI Lab is a dev-only scenario renderer that lives **inside the app's own renderer/build**, not a separate Storybook install. It mounts the real production components with static fixture props so you can open any UI state directly — and reuse the exact same fixtures from browser tests.

The point is not a primitive catalog (buttons, inputs). The point is **whole UI states that are hard or slow to reach by driving the live app**: a device-pairing-expired dialog, a failed run transcript, an empty list, a 60-row paginated table, long-content overflow, a toolchain-installing panel.

Build this only when asked. Adapt every concrete choice below to the target app's framework, router, build tool, and test runner — the architecture is portable; the APIs are not.

## When to use

- "Set up a UI lab / scenario gallery / in-app Storybook."
- "Let me preview this dialog/panel/state in isolation without clicking through the app."
- "Share these test fixtures with a visual review surface."
- Reviewing or extending an existing lab (add a scenario, add a group).

Not for: a real Storybook/Ladle install the user explicitly wants; a primitive component catalog; backend-only work.

## Core architecture

Seven pieces. Keep them in one dev-only subtree (e.g. `<renderer>/dev/ui-lab/`) so the guardrails stay auditable by directory.

1. **Scenario contract** — a minimal, type-erased record: `{ id, title, description, component, createProps }`.
   - `id`: globally unique, convention `<group>/<state>`; used in the deep-link URL.
   - `createProps`: a **factory**, not a value — every mount gets fresh props so stateful fixtures never leak across mounts/tests.
   - A `defineScenario` helper type-checks `createProps`'s return against the real component's props, then erases to a heterogeneous type so one registry can hold all scenarios.
   - Keep it minimal. No width/viewport/controls metadata — surfaces that need framing get a **wrapper component** instead (see below).

2. **Registry** — a manually curated array of groups (`{ id, title, scenarios[] }`). Explicit imports, ordered by feature area. Validate unique group + scenario ids at module load (throw on duplicate). Avoid bundler auto-discovery magic until the count outgrows manual curation.

3. **Scenario host** — mounts one scenario wrapped in the **minimum app-level providers** production components assume (theme, shortcuts, i18n, store context). This host is **shared with tests** — that sharing is what guarantees a scenario mounts identically in the lab and in a test.

4. **Lab shell** — dev tooling UI: grouped sidebar with filter, a preview pane, theme switch, a remount control, and URL-synced selection. The shell is developer-facing; keep its own copy in the team's working language even if scenarios render localized product copy.

5. **Boot + entry gate** — gate the lab behind a **static dev-only check** so packaged builds compile the whole branch (and the dynamically imported lab chunk) away. A hash route (`#/ui-lab`, `#/ui-lab/<id>`) is ideal: dynamically import the lab boot when the build is dev AND the hash matches; otherwise run the normal app boot. Hash (not query param) gives stable per-scenario URLs for later screenshot automation.

6. **Fixtures** — colocated per scenario group. **Copy** useful builders from the test suite into lab-local fixture modules (with provenance comments) rather than importing test code: the dependency direction must stay **tests → lab**, never lab → tests. Import the app's own domain constructors (branded ids, builders, defaults) directly.

7. **Smoke test** — one browser test that imports the registry + host (not the shell) and asserts **every** registered scenario mounts cleanly with no live services. This is the contract that keeps fixtures reusable. Add fixture-level assertion examples to show tests reading production DOM through a scenario.

## Hard rules (the guardrails)

- **Production must never import lab code**, and **lab must never ship in packaged builds.** Enforce by directory (dev-only subtree) + static build guard.
- **Determinism, always.** No live IPC, network, timers, or randomness. Callbacks are inert stubs. Timestamps are fixed. For async/result-shaped props use the framework's immediate-success/failure primitive. States that intrinsically start a live timer (e.g. an elapsed-time counter) are represented timer-free or skipped — and the skip is documented in the scenario file.
- **Wrapper components model real mounting contexts.** When a surface needs a footprint the host doesn't provide (fixed-height container, anchored popover stage, a controller + toaster, a labelled grid), add a small group-local wrapper component rather than bloating the scenario contract.
- **Internal-state surfaces are driven by replayed DOM interactions**, using the same interactions the browser tests use, with bounded polling (e.g. `requestAnimationFrame`) — never timers/network/randomness. Keep this the exception, not the default.
- **Stale scenarios die with the product state they represent.** Delete them; don't keep dead fixtures.

## Selecting scenarios

Favor states that are **hard to reach live**: auth pairing/denied/expired, install/failed toolchain, failed and looped/retried transcripts, conflict and rename validation, read-only and empty states, long-content overflow, large-N pagination, no-results search. Cover the status/variant matrix for each surface in one scenario when it reads clearly. Skip what's genuinely live-only and record why in the scenario file's comments.

## Lab shell UI guidance

The shell is internal tooling — make it fast to navigate, visually quiet, and out of the way of the previewed component.

- **Layout:** fixed sidebar (navigation) + flexible preview pane. Sidebar ~`18rem`. Preview pane scrolls independently; give it generous padding so component edges/shadows aren't clipped.
- **Sidebar:** collapsible groups with per-group counts and a total. A live substring **filter** over scenario + group titles/ids, focusable from anywhere via `/`, clearable via Escape, with match highlighting. While filtering, force matching groups open. Show a "no matches" empty state.
- **Selection:** drive selection from the URL (hash) so every state is a shareable deep link; sync back on `hashchange`. Mark the active item (accent dot + weight, `aria-current`).
- **Header:** scenario title + description + a click-to-copy scenario id (for deep links and test lookups).
- **Controls (top-right, minimal):** theme mode toggle (light/dark/system) and a **Remount** button — dialog/dismissable scenarios need a clean re-mount, done by bumping an epoch key on the mount wrapper. Add nothing the visual review doesn't need.
- **Empty/unknown states:** "pick a scenario" prompt; for an unknown id from a stale URL, say so and point back to the sidebar.
- **Visual restraint:** lean on typography, spacing, and a single accent; avoid nesting the preview in heavy bordered cards — the component under review is the subject, the chrome is not. Use `aria-label`/`title` on icon-only controls; the filter is a labelled input; theme buttons expose `aria-pressed`.

## Build order

1. Confirm framework, build tool, router/entry mechanism, and browser test runner.
2. Find the test suite's existing fixtures and the app's domain constructors — these seed scenarios.
3. Write the contract + `defineScenario` helper, the host (with real providers), and the registry with id validation.
4. Wire the dev-only entry gate and boot; verify the normal app boot is untouched.
5. Add 1–2 scenario groups end to end (fixtures → wrapper if needed → scenarios → registry), then the lab shell.
6. Add the smoke test asserting all scenarios mount; run it.
7. Fill in remaining groups. Run the app's `check`/`lint`/`format`/test commands. Manually review representative scenarios in light and dark.

## Validate

Run the repo's own quality gates (type-check, lint, format, browser tests) over touched files. The smoke test passing = the lab/test reuse contract holds. Record what you validated.
