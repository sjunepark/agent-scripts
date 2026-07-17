# Release Impact Checklist

Use this checklist after inspecting Release Please configuration and actual diffs. Classify per package or component rather than collapsing a manifest release into one bump.

## SemVer Classification

- `patch`: bug fixes, plus configured patch-level docs/refactor/chore/dependency changes that do not change public behavior.
- `minor`: backward-compatible user-facing behavior, new commands/options/APIs, broader support, or additive output fields consumers can ignore safely.
- `major` / breaking: incompatible public contract changes. Require `!` in the Conventional Commit type/scope or a `BREAKING CHANGE:` footer.

For pre-1.0 projects, inspect options such as `bump-minor-pre-major` and `bump-patch-for-minor-pre-major`; if absent, state the configured or default behavior instead of guessing.

## Breaking Change Audit

Use repository contract docs as the authoritative checklist when present. Otherwise inspect:

- CLI command, flag, argument, environment variable, config file, or default behavior removal/rename.
- Output format changes, especially JSON shape, field meaning, ordering guarantees, stdout/stderr behavior, or machine-readable envelopes.
- Exit-code or error-classification changes used by scripts.
- Public API signature, type, schema, route, protocol, event, database migration, or file format incompatibility.
- Runtime/support policy changes such as minimum Node/Python/OS/browser version or dropped binary/platform target.
- Package identity changes: package name, executable/bin name, import path, published files, artifact names, tag convention, or registry.
- Security/auth/permission changes that require users to reconfigure existing deployments.
- Data deletion, migration, or one-way state changes that affect existing users.

Call out possible breaking changes as decision points and ask when compatibility intent is ambiguous.
