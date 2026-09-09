# sjskills standalone release delivery

Status: `sjskills-v1.1.0` published as an immutable release on 2026-09-10.
Hosted verification is working; unattended hosted publication still needs a
suitable GitHub identity.

## Published releases

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
