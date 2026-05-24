---
name: release-please-release
description: Prepare, audit, set up, and guide releases for repositories that use or are explicitly adopting Release Please. Use when the user asks to release, prepare a release PR, set up or add Release Please, classify SemVer impact, decide whether a change is breaking, write Release Please-compatible Conventional Commit guidance, review a Release Please PR, or document release criteria. Release work requires existing Release Please configuration; setup work requires an explicit setup request.
---

# Release Please Release

Guide release preparation for projects whose versioning and changelog are owned by Release Please. Treat Release Please as a required dependency: if the repository does not have Release Please configured or the user is not explicitly asking to add it, stop and ask whether to set up Release Please before doing release work.

## Required Dependency Gate

Before classifying or preparing a release, verify that the project uses Release Please.

Accept evidence such as:

- `release-please-config.json`, `.release-please-manifest.json`, or equivalent manifest config.
- A Release Please workflow using `googleapis/release-please-action`.
- Project docs that state Release Please owns version bumps, changelog updates, tags, or releases.
- A direct dependency or documented CLI usage of `release-please`.

If none exists:

- Do not infer a custom release workflow.
- Ask whether the user wants to add Release Please or use another process.
- If they want to add Release Please, inspect package/language, existing release docs, CI, publishing flow, and tag conventions before proposing config.

## Core Rules

- Release Please is the version authority. Do not manually edit versions, changelogs, manifests, release tags, or release notes unless the user explicitly asks for a manual fallback.
- Conventional Commits are the release input. Prefer fixing commit messages or recommending squash/merge messages over hand-editing generated release artifacts.
- Classify release impact from public/user-facing contracts, not just code volume.
- Treat possible breaking changes as decision points. Call them out plainly and ask when product intent is ambiguous.
- Use repository docs and local agent instructions first. Do not invent release commands.
- Validate with the repo's documented CI/test/build commands. For release-specific commands, use read-only checks or documented dry-runs unless the user explicitly authorizes pushing, tagging, publishing, creating releases, merging release PRs, or other irreversible operations.
- Do not merge Release Please PRs by default. Merging may create tags, GitHub releases, publish packages, or trigger deployment automation, so merge only when the user explicitly authorizes merging the specific PR after seeing the review summary.
- Always get explicit user confirmation of the exact version number(s) that will be released before any side-effecting release step, including merging a Release Please PR, tagging, publishing, creating a GitHub release, or running release automation. For manifest or monorepo releases, confirm each component/package version separately.

## Release Impact Classification

First inspect Release Please config, changelog sections, and documented releasable commit behavior. Do not promise that `docs:`, `refactor:`, `chore:`, or dependency commits will produce a release or release notes unless the repository's configuration does so.

Then map releasable changes to SemVer in Release Please terms:

- `patch`: bug fixes, plus configured patch-level docs/refactor/chore/dependency changes that do not alter public behavior.
- `minor`: new backward-compatible user-facing behavior, new commands/options/APIs, broader support, additive output fields that consumers can safely ignore.
- `major` / breaking: any incompatible public contract change. Require `!` in the Conventional Commit type/scope or a `BREAKING CHANGE:` footer.

For pre-1.0 projects, inspect Release Please config for options such as `bump-minor-pre-major` or `bump-patch-for-minor-pre-major`. If absent, state Release Please's default behavior instead of assuming a custom pre-1.0 policy.

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

If the repo has its own contract docs, use those as the authoritative checklist and add only relevant domain-specific items.

## Workflow

1. Read release context.
   - Read local agent instructions, release docs, Release Please config/manifest, package metadata, and release workflows.
   - Identify the configured package(s), release type(s), tag format, changelog path, publishing workflow, and releasable/changelog commit types.

2. Determine the comparison range.
   - Find the last Release Please-managed release tag or manifest version.
   - For manifest or monorepo setups, determine the range per configured component/package, including path, version, tag format, and changelog.
   - Inspect commits and diff since the relevant release for each component/package.
   - If history is unavailable or shallow, say so and use the best available evidence.

3. Classify changes.
   - Group commits into fixes, features, breaking changes, chores/docs, and non-release changes.
   - For manifest or monorepo setups, classify release impact per configured component/package rather than collapsing everything into one bump.
   - Check whether commit messages match the intended Release Please classification and the repository's configured releasable commit behavior.
   - Separately inspect the diff for hidden breaking changes not reflected in commit messages.

4. Prepare or review release inputs.
   - For normal releases, recommend exact Conventional Commit or squash-merge messages.
   - For breaking changes, include a concise migration note suitable for a `BREAKING CHANGE:` footer.
   - For forced versions, mention `Release-As: x.y.z` only when the user explicitly needs an override.
   - Do not edit generated release PR files unless reviewing an already generated Release Please PR or doing an explicit manual fallback.

5. Validate.
   - Run documented typecheck/test/build commands and safe release validation checks.
   - For release commands, prefer read-only checks or documented dry-runs.
   - Do not push, tag, publish, create GitHub releases, merge release PRs, or run irreversible release automation unless the user explicitly authorizes that operation.
   - Before any authorized side-effecting release step, state the exact version number(s) to be released and ask the user to confirm those version(s).
   - Confirm CI gates, publishing permissions, and tag/version consistency when the task involves release automation.

6. Handle authorized release automation.
   - If the user explicitly authorizes running Release Please automation, wait for the relevant GitHub Actions workflow to finish and for the Release Please PR to be created or updated.
   - Review the resulting PR before recommending merge: version bump, changelog, release notes, component/package scope, breaking-change notes, CI status, and any workflows that would run after merge.
   - Report whether the PR appears merge-ready, with blockers and risks.
   - Stop before merge unless the user explicitly authorizes merging that specific PR and confirms the exact version number(s) to release after the review summary.

7. Report.
   - Recommended SemVer bump.
   - Evidence for the classification.
   - Required commit message changes or release PR review notes.
   - Breaking-change migration notes, if any.
   - Validation results and any blockers.

## Prompt Templates To Offer Users

When users ask how to work with agents, recommend prompts like:

```text
Prepare this repository for a Release Please-managed release.
Do not manually bump versions or edit generated changelog/manifest files unless I explicitly ask for manual fallback.
Inspect changes since the relevant Release Please release tag, per component/package when configured, and classify the release impact as patch, minor, or major/breaking.
Audit public contracts: CLI/API, output formats, exit codes, runtime requirements, package names, published files, artifact names, and documented behavior.
Ensure the intended classification is represented by Conventional Commit messages and supported by the repository's Release Please config.
Run documented read-only validation commands and dry-runs only unless I authorize side-effecting release operations.
If I explicitly authorize Release Please automation, wait for GitHub Actions and the Release Please PR, then review the PR contents and CI status.
Before any side-effecting release step, ask me to confirm the exact version number(s) that will be released.
Do not merge the release PR unless I explicitly authorize merging that specific PR and confirm the exact release version number(s) after your review summary.
Report the recommended SemVer bump, evidence, missing docs/tests, breaking-change migration notes, validation results, and PR merge-readiness when applicable.
```

For normal implementation tasks, suggest adding:

```text
After implementation, classify the Release Please impact and propose the Conventional Commit message. Call out possible breaking changes instead of hiding them as feat/fix.
```

## Response Style

Be concise and decision-oriented. Lead with the recommended release classification, then the evidence and next actions. When uncertain, ask focused questions about product intent or compatibility guarantees rather than guessing.
