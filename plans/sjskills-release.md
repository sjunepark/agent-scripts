# sjskills standalone release alignment

Status: implementation and local review complete; release ownership and remote
activation pending. No push, tag, or publication performed.

The [release guide](../docs/sjskills-releases.md) owns the distribution contract
and installation procedures. Initial targets are macOS amd64/arm64 and Windows
amd64; Linux reconciliation is unsupported and outside this change.

## Delivered

- One target definition for archives, generated installers, and native CI.
- Embedded version file; deterministic archives and SHA-256 manifest.
- Installers that verify and stage before replacing; reinstall upgrades.
- Linux-only development CI; native checks only for main integration and release.
  Publication requires immutable releases,
  matching tag/version/commit, successful native tests, and verified draft assets.
- Bounded subagent review completed. Fixed timestamp nondeterminism that prevented
  partial draft retries; regression tests cover it and publication boundaries.
- Documentation harmonized, including the existing rollout platform claim.
- CI follow-up review found no actionable issues; latest strict-sync source and
  registry behavior remain unchanged from `origin/main`.

## Validation

Passed locally: Go tests and vet, Node registry tests, skill validation,
Actionlint, release regression tests, cross-builds for all three targets, and
macOS arm64 installation/consumer/failure-preservation checks. Downloads are
controlled in local installer tests; live HTTPS installation and macOS Intel /
Windows native runs still require remote execution.

## Remaining

- User choice pending: Release Please-managed versions/tags or manual version
  tags. The reusable release pipeline is ready for either owner.
- Push through the normal review flow; run hosted checks at their intended boundaries.
- Enable GitHub release immutability (read-only inspection found it disabled)
  and configure the `Checks` required job in branch protection.
- Select and authorize the exact first release version before publishing.

The feature branch was fast-forwarded from `bc9fac9` to the latest published
`a61eb73` before committing, preserving strict-sync and registry corrections.
The CI correction follows the committed mytech Linux-only dev policy; local
mytech edits were left untouched.

Next: resolve release ownership and complete its configuration. No real-home
skill reconciliation or global rollout is part of this work.
