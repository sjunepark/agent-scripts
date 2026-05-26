---
name: pi-extension-intake
description: "Evaluate, import, vendor, fork, and harden third-party pi extensions. Use when copying a pi extension into a repo, choosing vendoring vs internal fork vs patch-package vs direct dependency, auditing weakly maintained extension code, removing self-install or network-install behavior, or analyzing wrapper-package and npm package trust boundaries."
---

# Pi Extension Intake

Help the user choose the right ownership model for a third-party pi extension, then give a concrete intake plan that matches their control, maintenance, and security goals.

Default bias: if the extension is small, weakly maintained, likely to be customized, or raises trust concerns, prefer vendoring the source into the user's repo over patching a published package.

## Goals

Do three things well:
1. Identify the real dependency and trust boundary.
2. Recommend the right strategy: direct dependency, patch-package, internal fork, or vendored copy.
3. If the user wants to adopt it, produce a practical import and hardening plan.

## Core idea

A copied wrapper is not the same as owned behavior. If an extension shells out to another CLI, downloads binaries, auto-installs npm packages, or fetches code at runtime, that downstream tool is part of the real trust boundary and must be analyzed explicitly.

## Workflow

1. Clarify the user's goal.
- Determine whether the user cares most about security review, local control, low maintenance, easy upstream updates, or organization-wide reuse.
- Determine whether they want advice only or want the extension imported into their repo.

2. Inspect the extension before recommending a strategy.
- Read the package manifest and extension entry point.
- Check how the extension is loaded by pi.
- Identify runtime dependencies, shell-outs, process-spawn paths, package install or lifecycle scripts, downloads, auto-install prompts, browser/binary bootstrap steps, and remote network behavior.
- Distinguish the visible package from the true dependency chain.

3. Choose the ownership model.
- **Vendor into the repo** when the code is small, lightly maintained, customization is likely, or trust and auditability matter more than painless upgrades.
- **Internal fork** when multiple repos or teammates need the same extension and you still want a structured upstream sync path.
- **patch-package** only when you intentionally still want the upstream package, the diff is small, and the patch is likely temporary.
- **Direct dependency** only when the upstream is healthy, trusted, easy to update, and you do not need deep control.

4. Explain the tradeoff plainly.
- State what is gained and what is lost.
- Be explicit about whether copying the repo actually removes the risky dependency, or only copies a thin wrapper around it.

5. If the user wants to adopt it, give the smallest safe plan.
- Prefer local, repo-owned extension files under the project's `extensions/` directory or equivalent local source path.
- Preserve provenance and license information.
- Remove or gate self-install and network-install behavior.
- Pin, replace, or separately vendor any runtime dependency that remains in the trust boundary.
- Add a short maintenance note describing the upstream source and local changes.

## Decision guide

Use this default decision order:

### Recommend vendoring when most of these are true
- The extension is small or structurally simple.
- Upstream is niche, weakly maintained, or unlikely to review your fixes quickly.
- You already expect to edit the code.
- The extension fits your repo's local customization needs.
- You want code review, security review, and refactoring to happen in your own repo.

### Recommend an internal fork when most of these are true
- The extension will be shared across repos or teams.
- You want a stable canonical internal version.
- You may occasionally merge upstream fixes.
- You want clearer provenance than ad hoc copy-paste, but still want ownership.

### Recommend patch-package only when all of these are true
- You still want to install the upstream package from npm or another package source.
- Your change is small.
- The package is maintained enough that staying close to upstream is valuable.
- The patch is expected to be temporary or low-risk to rebase.

### Recommend against patch-package when any of these are true
- The user's main goal is to remove trust in the upstream package.
- The package is poorly maintained.
- The extension will need ongoing product-specific changes.
- The real runtime behavior lives in another CLI or binary that the patch does not eliminate.

## Pi-specific review checklist

When inspecting a pi extension, explicitly check for:
- `pi.exec(...)` calls and any other process-spawn paths such as `child_process`, `spawn`, `execFile`, `execa`, Bun process helpers, or wrapper scripts that shell out to external tools
- prompts, package lifecycle hooks, or setup scripts that install npm packages, global tools, or other runtime dependencies
- browser or binary download steps
- dynamic network fetches
- cleanup hooks that assume external tools exist
- broad passthrough commands that expose a large external CLI surface area
- peer dependencies that reveal pi/TUI/typebox coupling

If the extension wraps a separate CLI, say so clearly: the wrapper may be easy to vendor, but the wrapped CLI is still the real operational dependency.

## Import checklist for vendoring

When the recommendation is to vendor the extension, follow this checklist:

1. Copy only the source files you actually need.
- Prefer the extension entry point and any referenced local resources.
- Do not import unrelated docs, plans, or tooling unless they materially help maintainability.

2. Preserve provenance.
- Add a short header comment or nearby note with:
  - upstream repo URL
  - upstream commit or release
  - import date
  - original license
  - summary of local modifications

3. Remove hidden trust paths.
- Delete or disable auto-install flows.
- Replace silent downloads with explicit operator steps.
- Require known local binaries or a separately reviewed internal package.

4. Reassess the remaining dependency boundary.
- If the extension still depends on an external CLI or browser stack, either:
  - accept and pin that dependency explicitly,
  - fork or vendor it too, or
  - replace it with a trusted implementation.

5. Adapt to local conventions.
- Rename the extension if needed.
- Match repository structure, logging, error handling, and docs.
- Keep the implementation simple and obvious.

6. Add minimal verification.
- Confirm pi can load the local extension.
- Run a smoke test for the main user path.
- Document manual setup steps if the runtime dependency remains external.

## Response shape

When advising on a third-party pi extension, prefer this structure:

### Recommendation
- Name the preferred strategy.

### Why
- Summarize the deciding factors.
- Call out the real trust boundary.

### What you gain / what you lose
- Give the concrete tradeoff.

### Adoption plan
- List the next implementation steps.

### Risk notes
- Mention any remaining external binaries, downloads, or maintenance burden.

## Guardrails

- Do not imply that vendoring a thin wrapper eliminates the risk of the wrapped tool.
- Do not recommend patch-package as a security solution when the user wants ownership and auditability.
- Do not remove intentional functionality without saying what behavior is being dropped.
- Do not ignore licenses when copying code.
- Do not overcomplicate small extensions with heavy process unless the user asked for it.

## Example triggers

- "I want to copy this pi extension into our repo. Should I vendor it or use patch-package?"
- "This extension is tiny but it auto-installs another npm package. What's the real risk here?"
- "We keep manually copying third-party pi extensions. Is there a cleaner workflow?"
- "Please import this extension locally and strip out self-install behavior."
- "This repo is not well maintained. Should we fork it, patch it, or just own the code?"
