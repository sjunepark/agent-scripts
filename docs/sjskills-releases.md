# sjskills releases

Status: tooling is implemented and locally checked; remote validation and
activation are pending. No binary release has been published. See the
[delivery record](../plans/sjskills-release.md) for remaining work. The checkout
wrapper remains usable until the first release is available.

## Distribution contract

Follow [mytech's standalone CLI guidance](https://github.com/sjunepark/mytech/blob/main/practices/standalone-cli-distribution.md).
GitHub Releases are the canonical binary source. `packaging/targets.json` owns
both the native verification matrix and installer target selection: macOS Intel,
macOS Apple silicon, and Windows x64. Linux reconciliation is unsupported.

`internal/sjskills/VERSION` is embedded in the executable. Tags use
`sjskills-vX.Y.Z`; archives use `sjskills_X.Y.Z_OS_ARCH.tar.gz` or `.zip`.
Each archive contains only the executable and `release.json`, which records its
version, target, and source commit. `SHA256SUMS` covers every archive and both
installers. Checksums detect corruption; publisher authenticity depends on the
GitHub HTTPS channel and immutable release identity.

The installers require no checkout, Go, or GitHub CLI. macOS uses system shell,
`curl`, `tar`, and `shasum`; Windows uses Windows PowerShell 5.1 or newer and
.NET. `plan` and `apply` still require Bun (`bunx`) and access to the existing pinned Skills CLI and Git
sources. Help and version perform no status checks. Profiles and manifest
initialization retain their primary behavior without Bun; their incidental
notices may report unavailable upstream evidence. Bare/named `status` also
succeeds with explicit unavailable skill evidence when no cached evidence or Bun
is available. CLI release comparison independently uses public HTTPS metadata
and needs no external command. `--no-status-check` skips those checks entirely. See the
[status contract](skill-registry.md#automatic-status-evidence) for latency and
stream/exit semantics.

## Install and update

After a release is published, select its numeric version from the repository's
[releases](https://github.com/sjunepark/agent-scripts/releases). Download the
installer from that exact release and run it with the same version.

macOS (replace `X.Y.Z` with the selected version):

```sh
version=X.Y.Z
curl --fail --location --proto '=https' --proto-redir '=https' \
  "https://github.com/sjunepark/agent-scripts/releases/download/sjskills-v$version/install.sh" \
  -o install.sh
sh install.sh "$version"
```

Windows, in Windows PowerShell 5.1 or newer:

```powershell
$version = 'X.Y.Z'
Invoke-WebRequest "https://github.com/sjunepark/agent-scripts/releases/download/sjskills-v$version/install.ps1" -OutFile install.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./install.ps1 -Version $version
```

Add `~/.local/bin` on macOS or `%LOCALAPPDATA%\sjskills\bin` on Windows to PATH.
Place it before the checkout's `bin/` so the development wrapper does not shadow
the installed executable and rebuild the CLI on every invocation.
Use the same installer and destination to update or reinstall a chosen version.
Status checks offer an advisory update notice linking to the selected stable
release; installing that update remains an explicit operation. A failed download, checksum, or
binary identity check leaves an existing executable untouched. Symlink and
directory destinations are refused; remove an old checkout symlink explicitly
before migrating that path, or select another directory.

For offline installation, download the selected archive and `SHA256SUMS` along
with the installer. Use `sh install.sh X.Y.Z INSTALL_DIR ARCHIVE_DIR` or
`powershell -ExecutionPolicy Bypass -File ./install.ps1 -Version X.Y.Z -InstallDir INSTALL_DIR -ArchiveDir ARCHIVE_DIR`.
Installation verifies the archive before staging and replacing the executable.
The installer does not modify PATH or install Bun.

## Verification and publication

Run the same repository-owned checks locally and in CI:

```sh
go vet ./...
python3 scripts/release_test.py
node --test scripts/lib/skill-registry.test.js scripts/audit-global-skills.test.js
scripts/validate-skills
python3 scripts/release.py build --output .tmp/release-check
python3 scripts/release.py verify --output .tmp/release-check
```

The output directory must be empty. On a supported native target, also run:

```sh
go test ./...
python3 scripts/test-release.py --output .tmp/release-check
```

The consumer checks install the exact archives into a temporary location,
exercise help/version, profiles/init, and bare/named/JSON/disabled status from an
unrelated directory with an empty PATH and isolated home/cache/staging roots,
and verify reinstallation and failure preservation. Status checks cover both
unconfigured and configured projects without Bun, plus CLI update comparison
from seeded release metadata with network access blocked. Linux runs source checks and
cross-builds, not reconciliation tests. Development PRs targeting `dev` and
manual source checks run on Linux only.
Integration PRs to `main`, merge-queue runs targeting `main`, and releases
exercise every release target. Release support does not enable macOS or Windows
jobs on development PRs. Set the `required` job
from `Checks` as the required branch check; there is no duplicate post-merge
full matrix.

The `Release sjskills` workflow verifies an existing tag and defaults to
verification only. Its publication mode is restricted to the `main` workflow
ref, requires repository release immutability to be enabled, and publishes only
after every native consumer job succeeds. It checks tag/version/checkout
identity, refuses tracked source changes, uploads to a draft, verifies server
asset SHA-256 identities, and only then publishes. Existing published releases
and mismatched draft assets are never overwritten; corrections use new versions.

Enable repository release immutability before first publication. Choose a
version/tag owner before activating release automation; the delivery record
tracks that decision and remote setup.
