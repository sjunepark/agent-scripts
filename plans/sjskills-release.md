# sjskills standalone release delivery

Status: first immutable release published on 2026-09-08. Hosted verification is
working; unattended hosted publication still needs a suitable GitHub identity.

## Published release

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
endpoint used to check release immutability; publication failed with HTTP 403
after all native jobs passed. Immutability was enabled and verified through the
user's authenticated administrator session, which completed publication.
Configure a suitable publisher identity before using hosted `publish=true` again;
verification-only runs remain usable without additional permissions.

During authenticated publication, the draft's by-tag API lookup returned HTTP
404 after successful creation/upload. Listing releases and fetching draft ID
`384416450` worked. Verification used that ID, then `gh release edit` published
it. Harden draft lookup to retain or resolve the draft ID before future releases.

The first release used a manually created tag matching the embedded `1.0.0`
version. Ongoing automated version ownership and the previously recommended
required branch check remain separate repository configuration work; this
release did not configure either.
