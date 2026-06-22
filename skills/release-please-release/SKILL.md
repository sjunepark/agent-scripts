---
name: release-please-release
description: "Prepare, audit, set up, and review Release Please-managed releases. Use when the user asks to add Release Please, or when a repo uses Release Please and the user asks to prepare a release, review or merge a Release Please PR, classify SemVer impact, check breaking changes, choose Conventional Commit messages, or document release criteria. Release work requires verified Release Please ownership; setup requires an explicit setup request."
---

# Release Please Release

Guide Release Please-owned versioning, changelog, tag, and release work. Treat Release Please as the release authority only after verifying local evidence.

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
- Use repository docs and local agent instructions first. Do not invent release commands.
- Validate with documented CI/test/build commands. For release-specific checks, use read-only commands or documented dry-runs unless the user has authorized write effects.
- Do not assume `docs:`, `refactor:`, `chore:`, or dependency commits produce releases; inspect Release Please config and docs.
- For implementation tasks, report the likely Release Please impact and propose the intended Conventional Commit message when relevant.

## Workflow

1. Read release context.
   - Read local agent instructions, release docs, Release Please config/manifest, package metadata, changelog, and release workflows.
   - Identify package/component paths, release types, tag format, changelog paths, publishing workflow, releasable commit types, and pre-1.0 policy.

2. Determine the comparison range.
   - Find the last Release Please-managed release tag or manifest version.
   - For manifest or monorepo setups, determine the range per configured component.
   - Inspect commits and diffs since the relevant release. If history is shallow or unavailable, say so.

3. Classify release impact.
   - Group commits as fixes, features, breaking changes, configured docs/chores, and non-release changes.
   - Classify per package/component instead of collapsing manifest releases into one bump.
   - Compare commit-message classification with actual diffs; hidden breaking changes still matter.

4. Prepare or review release inputs.
   - Recommend exact Conventional Commit or squash-merge messages when release input is wrong or missing.
   - For breaking changes, provide a concise migration note suitable for a `BREAKING CHANGE:` footer.
   - Mention `Release-As: x.y.z` only when the user explicitly needs a forced version.
   - Do not edit generated release PR files unless reviewing a generated PR or doing an explicit manual fallback.

5. Validate.
   - Run documented typecheck/test/build commands and safe release validation checks.
   - Confirm CI gates, publishing permissions, tag/version consistency, and required secrets when release automation is in scope.

6. Handle authorized automation.
   - If the user authorizes Release Please automation, wait for the relevant workflow and Release Please PR.
   - Review the PR before recommending merge: version bump, changelog, release notes, component scope, breaking-change notes, CI status, and downstream workflows triggered after merge.
   - Stop before merge unless the user authorizes merging that specific PR and confirms the exact release version after seeing the review summary.

7. Report.
   - Recommended SemVer bump and evidence.
   - Required commit-message or release-config changes.
   - Breaking-change migration notes, if any.
   - Validation results, blockers, and PR merge-readiness when applicable.

## SemVer Classification

Map releasable changes in Release Please terms after inspecting config:

- `patch`: bug fixes, plus configured patch-level docs/refactor/chore/dependency changes that do not change public behavior.
- `minor`: backward-compatible user-facing behavior, new commands/options/APIs, broader support, or additive output fields consumers can ignore safely.
- `major` / breaking: incompatible public contract changes. Require `!` in the Conventional Commit type/scope or a `BREAKING CHANGE:` footer.

For pre-1.0 projects, inspect options such as `bump-minor-pre-major` and `bump-patch-for-minor-pre-major`; if absent, state the configured or default behavior instead of guessing.

## Breaking Change Checklist

Audit the project-specific public surface. Common breaking changes include:

- CLI command, flag, argument, environment variable, config file, or default behavior removal/rename.
- Output format changes, especially JSON shape, field meaning, ordering guarantees, stdout/stderr behavior, or machine-readable envelopes.
- Exit-code or error-classification changes used by scripts.
- Public API signature, type, schema, route, protocol, event, database migration, or file format incompatibility.
- Runtime/support policy changes such as minimum Node/Python/OS/browser version or dropped binary/platform target.
- Package identity changes: package name, executable/bin name, import path, published files, artifact names, tag convention, or registry.
- Security/auth/permission changes that require users to reconfigure existing deployments.
- Data deletion, migration, or one-way state changes that affect existing users.

Use repository contract docs as the authoritative checklist when present.

## Setup Requests

When the user explicitly asks to add Release Please, inspect package/language, existing release docs, CI, publishing flow, secrets, branch protection, package manager, tag conventions, and current changelog/version files before proposing config.

Make setup changes only where the repository expects release automation to live. Update release docs and agent instructions when they define release ownership or manual fallback steps. Do not commit credentials or assume publishing permissions exist.

## Response Style

Be concise and decision-oriented. Lead with the release classification or merge-readiness answer, then evidence, blockers, and next actions. When uncertain, ask focused questions about product intent or compatibility guarantees rather than guessing.
