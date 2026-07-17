---
name: release-please-release
description: "Release Please release management. Use when the user asks to add Release Please, or when a repo using Release Please needs a release prepared, a Release Please PR reviewed or merged, SemVer or breaking-change impact classified, or Conventional Commit messages chosen."
---

# Release Please Release

Guide Release Please-owned versioning, changelog, tag, and release work.

## Hard Gates

### Verify Ownership

Before release prep or classification, verify that the repository uses Release Please. Accept evidence such as:

- `release-please-config.json`, `.release-please-manifest.json`, or equivalent manifest config.
- A Release Please workflow using `googleapis/release-please-action`.
- Project docs that state Release Please owns version bumps, changelog updates, tags, or releases.
- A direct dependency or documented CLI usage of `release-please`.

If none exists:

- Do not infer a custom release workflow.
- Ask whether the user wants to add Release Please or use another process.
- Add Release Please only when the user explicitly asks for setup.

### Protect Generated Artifacts

Release Please owns generated release artifacts. Do not manually edit versions, changelogs, manifests, release tags, or release notes unless the user explicitly asks for a documented manual fallback.

Prefer fixing release inputs: Conventional Commit messages, squash/merge messages, release config, or documented release criteria.

### Block Side Effects

Do not push, tag, publish, create GitHub releases, run release automation with write effects, merge Release Please PRs, or approve irreversible release steps unless the user explicitly authorizes the specific operation.

Before any authorized side-effecting release step, state the exact version number or per-component version numbers that will be released and get explicit confirmation.

## Core Rules

- Classify release impact from public/user-facing contracts, not just code volume.
- Treat possible breaking changes as decision points. Call them out plainly; ask when compatibility intent is ambiguous.
- Do not invent release commands; use only those documented in repository docs and local agent instructions.
- Do not assume `docs:`, `refactor:`, `chore:`, or dependency commits produce releases; inspect Release Please config and docs.
- For implementation tasks, report the likely Release Please impact and propose the intended Conventional Commit message when relevant.

## Choose the Workflow

- **Setup:** When the user explicitly asks to add Release Please, read and follow [workflows/setup.md](workflows/setup.md) instead of the release-preparation workflow below.
- **Release preparation or review:** When Release Please already owns releases, follow the workflow below.
- **Authorized automation:** When the user explicitly authorizes running automation, merging a Release Please PR, or another release side effect, first complete the applicable setup or release-preparation work, then read and follow [workflows/authorized-automation.md](workflows/authorized-automation.md). The Block Side Effects gate still applies to each specific operation.

## Release Preparation or Review Workflow

1. Read local agent instructions, release docs, Release Please config/manifest, package metadata, changelog, and release workflows. Identify component paths, release types, tag format, changelog paths, publishing workflow, releasable commit types, and pre-1.0 policy.
2. Find the last Release Please-managed release tag or manifest version. For manifest or monorepo setups, determine the comparison range per configured component. Inspect commits and diffs since each relevant release; if history is shallow or unavailable, say so.
3. Classify each component's fixes, features, breaking changes, configured docs/chores, and non-release changes. Before deciding a version, read and apply [checklists/release-impact.md](checklists/release-impact.md). Compare commit-message classification with actual diffs because hidden breaking changes still matter.
4. Prepare or review release inputs. Recommend exact Conventional Commit or squash-merge messages when input is wrong or missing. For breaking changes, provide a concise migration note suitable for a `BREAKING CHANGE:` footer. Mention `Release-As: x.y.z` only when the user explicitly needs a forced version. Do not edit generated release PR files unless reviewing a generated PR or performing an explicit manual fallback.
5. Validate with documented typecheck/test/build commands. For release-specific checks, use read-only commands or documented dry-runs unless write effects have been authorized. Confirm CI gates, publishing permissions, tag/version consistency, and required secrets when release automation is in scope.

## Report

Lead with the release classification or merge-readiness answer. Include:

- recommended SemVer bump and evidence;
- required commit-message or release-config changes;
- breaking-change migration notes, if any;
- validation results, blockers, and PR merge-readiness when applicable.
