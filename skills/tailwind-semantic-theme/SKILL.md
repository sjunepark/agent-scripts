---
name: tailwind-semantic-theme
description: "Set up or refactor Tailwind v4 theming with shadcn-compatible semantic CSS tokens, without shadcn. Use only when explicitly named or when the exact methodology is requested: OKLCH `:root` and `.dark` variables, class-based dark mode, `@theme inline`, and shadcn-compatible token names."
---

# Tailwind Semantic Theme

Use shadcn's theming contract without adopting shadcn as a dependency.

Keep the token names and CSS structure. Adapt the file layout to the project in front of you.

## Scope

- Use this skill only when explicitly invoked by name or when the user explicitly asks for this methodology.
- Target Tailwind v4 projects.
- Prefer methodology first. If the user asks for implementation, turn the methodology into concrete edits in the existing project.
- Do not add shadcn just to get this pattern.

## Goals

- Define theme values once as semantic CSS variables.
- Keep light and dark mode on the same token names via `:root` and `.dark`.
- Expose those tokens to Tailwind through `@theme inline`.
- Preserve shadcn-compatible token names so future component adoption stays easy.
- Prefer OKLCH values.

## Workflow

1. Inspect the current styling boundary.
- Find the global stylesheet, Tailwind entrypoint, and dark-mode toggle.
- Check whether the project already uses CSS variables, `@theme`, or a different dark-mode selector.
- Reuse the existing global CSS file when possible. Do not create a competing theme layer.

2. Decide whether the user wants guidance or edits.
- For methodology-only requests, explain the structure, provide a scaffold, and give migration guidance.
- For implementation requests, update the existing global CSS and any dark-mode wiring needed for a `.dark` class contract.

3. Establish the token contract.
- Use the full token set:
  - `background`, `foreground`
  - `card`, `card-foreground`
  - `popover`, `popover-foreground`
  - `primary`, `primary-foreground`
  - `secondary`, `secondary-foreground`
  - `muted`, `muted-foreground`
  - `accent`, `accent-foreground`
  - `destructive`, `destructive-foreground`
  - `border`, `input`, `ring`
  - `chart-1` through `chart-5`
  - `sidebar`, `sidebar-foreground`
  - `sidebar-primary`, `sidebar-primary-foreground`
  - `sidebar-accent`, `sidebar-accent-foreground`
  - `sidebar-border`, `sidebar-ring`
  - `radius`
- Keep `name` / `name-foreground` pairs for surface/text compatibility.
- Keep raw color values in CSS variables, not in component class names.

4. Structure the CSS in this order.
- `@import "tailwindcss";`
- `@custom-variant dark (&:is(.dark *));`
- `@theme inline { ... }` mappings from semantic vars to Tailwind tokens
- `:root { ... }` light values
- `.dark { ... }` dark values
- `@layer base { ... }` only for a small reset such as body/background/text and shared border/ring defaults

5. Keep dark mode simple.
- Follow the shadcn default: class-based dark mode via `.dark`.
- Do not duplicate dark styling in component classes when token overrides are enough.
- Do not invent automatic dark-value generation unless the user asks for it. Author light and dark values explicitly.

6. Keep components semantic.
- Prefer `bg-background text-foreground`, `bg-card text-card-foreground`, `bg-primary text-primary-foreground`, and similar semantic utilities.
- If a raw palette class like `bg-zinc-900` or `text-slate-500` repeats across the app, promote that usage into a semantic token.
- Avoid naming tokens after palette families such as `slate`, `blue`, or `gray` when the real role is semantic.

7. Migrate incrementally.
- Start with the global token scaffold.
- Map the most reused surfaces and actions first.
- Replace repeated raw color utilities with semantic ones in touched files.
- Do not churn one-off decorative colors into the token system unless they are becoming reusable design primitives.

## Output

For methodology-only requests, provide:

- The target file or files to change.
- The recommended Tailwind v4 scaffold.
- The token contract and why it exists.
- The migration order for existing classes.
- Any dark-mode wiring needed for `.dark`.

For implementation requests, also:

- Edit the existing stylesheet instead of introducing parallel theme files.
- Preserve project-specific imports and layout.
- Summarize any tokens added beyond the standard contract.

## Reference Scaffold

When you need a concrete starting point, read [references/tailwind-v4-theme-scaffold.css](./references/tailwind-v4-theme-scaffold.css). Adapt it to the project's existing CSS file instead of copying it blindly.

## Guardrails

- Do not add shadcn as a dependency unless the user explicitly asks for shadcn.
- Do not use Tailwind config theme colors when `@theme inline` in global CSS is the simpler fit.
- Do not create a second theme system beside an existing one unless the user explicitly wants a migration.
- Do not hardcode light and dark palette classes into components when semantic tokens can express the same intent.
- Do not rename the standard tokens just to match a local preference if future shadcn-style compatibility matters.

## Example Triggers

- "`$tailwind-semantic-theme` help me set up shadcn-style theme variables in this Tailwind v4 app, but do not install shadcn."
- "`$tailwind-semantic-theme` refactor my raw Tailwind colors into semantic tokens with `.dark` mode."
- "`$tailwind-semantic-theme` show me the right globals.css structure for OKLCH tokens, `@theme inline`, and shadcn-compatible names."
