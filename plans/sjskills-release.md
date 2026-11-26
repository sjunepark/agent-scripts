# sjskills standalone release delivery

Status: `sjskills-v1.7.1` published as an immutable release on 2026-11-26 (KST).
Hosted verification is working; unattended hosted publication still needs a
suitable GitHub identity.

## Published releases

- [sjskills-v1.7.1](https://github.com/sjunepark/agent-scripts/releases/tag/sjskills-v1.7.1)
  points to `b270f3f8591294ddc0f4834a8a1e236fdfdb0895` and is immutable.
  Apply now installs over a committed provenance record whose placement is
  absent, as in a fresh checkout that git-ignores its target root, instead of
  failing with `transaction recovery evidence is invalid` (merged PR #59).
  The state format is unchanged from 1.7.0.
- The [v1.7.1 release run](https://github.com/sjunepark/agent-scripts/actions/runs/45735680951)
  passed native reconciliation and installer checks on macOS amd64, macOS arm64,
  and Windows amd64 with `publish=false`.
- Publication used the exact CI artifacts, verified locally with
  `scripts/release.py verify --tag sjskills-v1.7.1` from a clean checkout of the
  tag. The authenticated session created draft ID `458839746`, uploaded all six
  assets by release ID, matched every server-side digest, then published it.
- The published installer updated `%LOCALAPPDATA%\sjskills\bin\sjskills.exe`
  on the Windows development machine to v1.7.1. Its binary matches the verified
  windows/amd64 archive, with SHA-256
  `4380cf36f7cf3de54f7132131e0824ab59d2ea5601affe8c8639ef9fc8d649c5`.

- [sjskills-v1.7.0](https://github.com/sjunepark/agent-scripts/releases/tag/sjskills-v1.7.0)
  points to `1793abf5b858935fce5de47e7093b45bbf4d537b` and is immutable.
  It records skill trees as `tree-sha256-v3`, which ignores the executable bit,
  and forces `core.symlinks=false`, so committed placements hash the same on
  every host (merged PR #50). The first apply rewrites verified v2 records.
  Version 1.6.0 accepts only v2 records and cannot read state written by 1.7.0.
- The [v1.7.0 release run](https://github.com/sjunepark/agent-scripts/actions/runs/45297560301)
  passed native reconciliation and installer checks on macOS amd64, macOS arm64,
  and Windows amd64 with `publish=false`.
- Publication used the exact CI artifacts. Tag/commit identity, `SHA256SUMS`,
  and archive contents were verified locally with `scripts/release.py`.
  `scripts/publish-release.py` created draft ID `456164397` and uploaded all
  six assets, then stopped at its by-tag lookup, which cannot see drafts. Every
  server-side digest was matched by release ID before the draft was published.
- The published installer updated `%LOCALAPPDATA%\sjskills\bin\sjskills.exe`
  on the Windows development machine to v1.7.0. Its binary matches the verified
  windows/amd64 archive, with SHA-256
  `eb8a4bd30ac45a9df2185cfd062c3efeaf7cd455b10917a66fc4ac1ee9128b93`.

- [sjskills-v1.6.0](https://github.com/sjunepark/agent-scripts/releases/tag/sjskills-v1.6.0)
  points to `0975d3d4dc4fc76ffbdee120c3ca32491ad6a162` and is immutable.
  It reads registry version 5, whose sources may declare `formerIdentities`
  so installs from a renamed repository migrate (merged PR #49), and includes
  host-independent skill checkouts (merged PR #48). Version 1.5.0 fails
  closed on the version 5 registry.
- The [v1.6.0 release run](https://github.com/sjunepark/agent-scripts/actions/runs/45293197112)
  passed native reconciliation and installer checks on macOS amd64, macOS arm64,
  and Windows amd64 with `publish=false`.
- Publication used the exact CI artifacts. Tag/commit identity, `SHA256SUMS`,
  and archive contents were verified locally with `scripts/release.py`.
  `scripts/publish-release.py` created draft ID `456144306` and uploaded all
  six assets, then stopped at its by-tag lookup, which cannot see drafts. Every
  server-side digest was matched by release ID before the draft was published.
- The published installer updated `%LOCALAPPDATA%\sjskills\bin\sjskills.exe`
  on the Windows development machine to v1.6.0. Its binary matches the verified
  windows/amd64 archive, with SHA-256
  `0115936d534f6ac71cd05e1cee9352ecf37663bcd2bcf26ac2ef65ffd1c6c196`.

- [sjskills-v1.5.0](https://github.com/sjunepark/agent-scripts/releases/tag/sjskills-v1.5.0)
  points to `67c3f60ca4d2d12fad637ed98638c7a73c40c7e0` and is immutable.
  It adds local-path `[[direct]]` sources (merged PR #29); older binaries
  reject `local:` provenance.
- The [v1.5.0 release run](https://github.com/sjunepark/agent-scripts/actions/runs/37399587178)
  passed native reconciliation and installer checks on macOS amd64, macOS arm64,
  and Windows amd64 with `publish=false`.
- Publication used the exact CI artifacts. Tag/commit, `SHA256SUMS`, and
  archive `release.json` identities were verified locally; the authenticated
  session created draft ID `404224031`, uploaded all six assets, matched every
  server-side digest by release ID, then published it.
- The published installer updated `%LOCALAPPDATA%\sjskills\bin\sjskills.exe`
  on the Windows development machine to v1.5.0. Its binary matches the verified
  windows/amd64 archive, with SHA-256
  `7d30bbd3777404349f651c43a5f1e95baeb6dfbecd27b2f2e08ad797a0061b0f`.

- [sjskills-v1.4.0](https://github.com/sjunepark/agent-scripts/releases/tag/sjskills-v1.4.0)
  points to `830304768fdd1fc20798b1b2a55118c1619d10b1` and is immutable.
  It reads the skill registry from the published source (merged PR #26).
- The [v1.4.0 release run](https://github.com/sjunepark/agent-scripts/actions/runs/36675721697)
  passed native reconciliation and installer checks on macOS amd64, macOS arm64,
  and Windows amd64 with `publish=false`.
- Publication used the exact CI artifacts. Tag/commit and archive contents were
  verified locally; the authenticated session created draft ID `399757443`,
  uploaded all six assets, matched every server-side digest by release ID, then
  published it. The published release is immutable.
- The published installer updated `%LOCALAPPDATA%\sjskills\bin\sjskills.exe`
  on the Windows development machine to v1.4.0. Its binary matches the verified
  windows/amd64 archive, with SHA-256
  `8a02f00e273730f4831aa695ebbb1f7c50694b5f1c9ad387524bb4857be090f8`.

- [sjskills-v1.3.0](https://github.com/sjunepark/agent-scripts/releases/tag/sjskills-v1.3.0)
  points to `f4e48e22252651229f6fc3da5ce40f60d5518361` and is immutable.
  It releases authenticated GitHub skill selections from merged PR #23.
- The [v1.3.0 release run](https://github.com/sjunepark/agent-scripts/actions/runs/34492463375)
  passed native reconciliation and installer checks on macOS amd64, macOS arm64,
  and Windows amd64. [Tagged source checks](https://github.com/sjunepark/agent-scripts/actions/runs/34492467041)
  also passed.
- Publication used the exact CI artifacts. Tag/commit, archive contents, and
  all six server-side asset digests were verified before publication and the
  immutable published state was verified afterward. The authenticated session
  retained release ID `386384350` through publication to avoid draft lookup by tag.
- This release operation did not update installed binaries or reconcile managed
  skills.

- [sjskills-v1.2.0](https://github.com/sjunepark/agent-scripts/releases/tag/sjskills-v1.2.0)
  points to `4aad1dcbd0b787598cff6ab9f7f1df838de85ad8` and is immutable.
- The [v1.2.0 release run](https://github.com/sjunepark/agent-scripts/actions/runs/34447921460)
  passed native reconciliation and installer checks on macOS amd64, macOS arm64,
  and Windows amd64. [Hosted source checks](https://github.com/sjunepark/agent-scripts/actions/runs/34447924052)
  also passed against that commit.
- Publication used the exact CI artifacts. Tag/commit, archive contents,
  immutability, and all server-side asset digests were verified. The authenticated
  administrator session retained release ID `386067266` through draft creation,
  asset verification, and publication instead of relying on draft lookup by tag.
- The published installer updated `~/.local/bin/sjskills` to v1.2.0. Its binary
  matches the verified macOS arm64 archive, with SHA-256
  `59f9cfb042f61b5e760f038eb9a6881d85f78c5a8c6e90eb6e023f749911afab`.
  No managed skill reconciliation was requested or performed.

- [sjskills-v1.1.0](https://github.com/sjunepark/agent-scripts/releases/tag/sjskills-v1.1.0)
  points to `79baae2bf3ba1be459552b5fe3ea6e3cb3db9ead` and is immutable.
- The [v1.1.0 release run](https://github.com/sjunepark/agent-scripts/actions/runs/34416271218)
  built all archives and passed native reconciliation and installer/consumer
  checks on macOS amd64, macOS arm64, and Windows amd64.
- Publication used those exact CI artifacts: three archives, both installers,
  and `SHA256SUMS`. Tag/commit and archive identity were revalidated locally;
  every uploaded asset digest matched before publishing through the
  authenticated administrator session.

- [sjskills-v1.0.0](https://github.com/sjunepark/agent-scripts/releases/tag/sjskills-v1.0.0)
  points to `7e5e2ba91581ee903c6eef069ad812f9180e94fe` and is immutable.
- The [release run](https://github.com/sjunepark/agent-scripts/actions/runs/34181179768)
  built all archives and passed native reconciliation and installer/consumer
  checks on macOS amd64, macOS arm64, and Windows amd64. Hosted
  [source checks](https://github.com/sjunepark/agent-scripts/actions/runs/34181181919)
  also passed.
- Publication used those exact CI artifacts: three archives, both installers,
  and `SHA256SUMS`. Tag/commit and archive identity were revalidated locally;
  every uploaded asset digest matched before publishing. No installed binary
  or managed skill state was changed.

The [release guide](../docs/sjskills-releases.md) owns distribution, installation,
and verification procedures.

## Remaining automation work

The workflow's default `GITHUB_TOKEN` cannot read the repository-administration
endpoint used to check release immutability; the v1.1.0 publication job failed
with HTTP 403 after all native jobs passed. Immutability was enabled and
verified through the user's authenticated administrator session, which
completed publication.
Configure a suitable publisher identity before using hosted `publish=true` again;
verification-only runs remain usable without additional permissions.

During v1.1.0 authenticated publication, the draft's by-tag API lookup returned
HTTP 404 after successful creation/upload. Listing releases and fetching draft
ID `385905486` worked. Verification used that ID, then the release ID API
published it. Harden draft lookup to retain or resolve the draft ID before
future releases.

Ongoing automated version ownership and the previously recommended required
branch check remain separate repository configuration work; this release did
not configure either.
