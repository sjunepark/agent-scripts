# macOS storage cleanup catalog

Use this reference during every macOS storage-cleanup task. Resolve paths and
commands from the installed system, app, or tool; locations and flags can vary by
version and configuration. Inspection commands must be genuinely read-only.

## macOS storage semantics

### Capacity, APFS, and Storage categories

- Start with System Settings **General → Storage**, `df`, and read-only
  `diskutil` information. APFS volumes in one container share free space, and
  snapshots or purgeable content can make category totals differ from filesystem
  totals. Do not sum shared-volume capacity or promise that deleting an apparent
  size will increase free bytes by the same amount.
- macOS Storage **Documents** is a broad category of user and developer data; it
  does not mean the `~/Documents` directory. **System Data** aggregates data that
  has no more specific category. Do not treat either label as a deletable path.
- Use targeted size inspection under user-owned roots. Stay on the selected
  filesystem and do not follow symlinks. Treat permission-denied, privacy-blocked,
  and timed-out paths as uninspected.

### User files, Trash, and archives

- Inspect Downloads, Desktop, Documents, media, disk images, installers, exports,
  archives, and backup files item by item. Age and extension do not prove that a
  file is reproducible or duplicated.
- Moving an item to Trash changes its location and may disrupt an app or project.
  It frees no local space until permanently emptied. Preview the exact items
  separately from any later permanent removal, and preserve unrelated Trash.
- Prefer an approved move to a verified archive or external destination when the
  user wants retention. Confirm the copy before proposing deletion of the source.

### iCloud, Photos, and other sync providers

- Verify sync is complete and identify whether an action propagates to the cloud
  or other devices. A synced folder is not an independent backup.
- Finder **Remove Download**, Dropbox **Make online-only**, and equivalent
  provider actions retain a remote copy but remove local availability. They are
  state changes requiring exact consent and network-dependent recovery.
- Delete iCloud Drive or Photos items only through their owning UI after
  explaining cross-device deletion and recovery windows. Never modify a Photos
  library package internally.
- Do not act while a provider is paused, conflicted, offline, or still uploading.

### Time Machine, device backups, logs, and protected data

- Inspect local Time Machine snapshots with the installed `tmutil` or `diskutil`
  read-only listing interface. macOS normally manages their retention and counts
  their space as available. Never broad-thin snapshots or delete external backup
  history merely to improve an apparent category.
- Manage iPhone and iPad backups through Finder or the Apple storage UI. A backup
  may be the only restore point; timestamp alone is insufficient deletion evidence.
- Do not wipe unified logging stores, `/private/var`, `/Library`, or macOS-managed
  caches manually. Logs may be required diagnostic evidence.
- Do not inspect or modify other users' data without their authority. Never seek
  credentials, recovery keys, or privacy bypasses for cleanup.

Primary references: [Apple Storage settings](https://support.apple.com/guide/mac-help/change-storage-settings-mchl3d437fbc/mac),
[freeing storage](https://support.apple.com/en-us/102624),
[optimizing storage](https://support.apple.com/guide/mac-help/optimize-storage-space-sysp4ee93ca4/mac),
[Time Machine local snapshots](https://support.apple.com/en-us/102154),
[iCloud Drive](https://support.apple.com/guide/mac-help/mchl1a02d711/mac), and
[privacy access](https://support.apple.com/guide/mac-help/mchld5a35146/mac).

## Developer caches and generated artifacts

For every entry, first resolve the live location, measure it, inspect active
processes, and read the installed command's help. Keep cache cleanup separate
from project cleanup.

| Category | Read-only discovery and inspection | Supported mutation to preview | Data effect and guardrails |
| --- | --- | --- | --- |
| npm cache | `npm config get cache`, then size/list the resolved path | `npm cache clean --force`; newer versions may support narrower npx cleanup | Removes registry cache and causes redownloads. `npm cache verify` can garbage-collect data, so classify it as a state change rather than inspection. Never absorb `node_modules` or project pruning into this row. |
| npm project artifacts | Measure exact `node_modules` directories; inspect the project and lockfile | `npm prune` only for the named project after preview | Removes extraneous packages and can remove manually added content. Deleting and reinstalling a dependency tree is a separate project action. |
| pnpm store | `pnpm store path`, `pnpm store status`, then size the resolved store | `pnpm store prune` | Removes unreferenced packages from the shared store; old branches may redownload. Do not run while a store server or package operation is active. |
| Yarn | Detect Yarn generation; use `yarn config get cacheFolder` for modern Yarn or `yarn cache dir` for Yarn 1 | Version-appropriate `yarn cache clean` with exact flags | Modern `.yarn/cache` may be committed Zero-Install project data or an offline mirror. Never assume it is a disposable global cache. |
| Bun | `bun pm cache`, then size the returned path; honor configured cache variables | `bun pm cache rm` | Clears the global package cache and causes redownloads. Do not replace the supported command with broad deletion or include project `node_modules`. |
| `uv` | `uv cache dir`, then use documented read-only size output or size the resolved path | `uv cache clean [PACKAGE]` clears all or selected cache entries; `uv cache prune` removes unused entries and centralized project environments | Centralized environments removed by `prune` are recreated as needed; both operations can cause redownloads or rebuilds. Confirm installed help and never manually edit cache internals. |
| pip | `python -m pip cache dir`, `info`, or `list` | `python -m pip cache remove <pattern>` or `purge` | Removes wheel/download cache, not environments; rebuilding wheels can be expensive. Use the interpreter whose cache is being inspected. |
| Cargo project output | Resolve `CARGO_TARGET_DIR`; use `cargo clean --dry-run --verbose` where supported | `cargo clean` with the named workspace/profile selectors | Deletes generated build output and causes recompilation. Do not include source, untracked files, Cargo home, or another workspace. |
| Cargo and rustup homes | Resolve `CARGO_HOME` and `RUSTUP_HOME`; list installed toolchains/targets | Exact `rustup toolchain uninstall` or `rustup target remove` | A Cargo home also contains installed binaries, configuration, and credentials. Cargo cache internals are unstable; never recursively delete the home. |
| Go caches | `go env GOCACHE GOMODCACHE GOPATH`, then size exact returned paths | `go clean` with separately previewed cache flags | Build/test/fuzz caches are distinct from the module cache. `-modcache` removes downloaded module sources globally and has a larger recovery cost. |
| Gradle | Resolve `GRADLE_USER_HOME`; measure global and exact project `.gradle`/`build` paths | Project `clean`; let Gradle manage global cache retention | Stop named daemons before a confirmed mutation. Global cache formats and retention change; do not recursively delete `~/.gradle`. |
| Maven | Resolve `settings.localRepository`; measure exact project `target` and local repository | Project `clean` or scoped dependency-plugin purge | The local repository can contain locally installed artifacts that cannot be redownloaded. Never recursively delete `.m2` as generic cache. |
| Homebrew | `brew --cache`, size that path, and `brew cleanup -n` | `brew cleanup` with exact previewed age/package flags | Removes downloads and old installed formula versions. Keep `autoremove`, uninstall, and aggressive all-cache pruning as separate actions. |
| Playwright | Resolve an already-installed Playwright executable or project package, use its `install --list`, and measure its documented browser-binary cache; otherwise mark this category unavailable | Version-supported `playwright uninstall` for the exact installation | Never use `npx` for discovery because it may download a missing package and mutate npm's cache. `--all` can affect every Playwright installation. Persistent browser profiles, cookies, and sessions are user state, not browser binaries. |

Official references: [npm cache](https://docs.npmjs.com/cli/v11/commands/npm-cache),
[pnpm store](https://pnpm.io/cli/store), [Yarn caching](https://yarnpkg.com/features/caching),
[Bun package-manager commands](https://bun.sh/docs/pm/cli/pm),
[`uv` caching](https://docs.astral.sh/uv/concepts/cache/),
[pip cache](https://pip.pypa.io/en/stable/cli/pip_cache/),
[Cargo clean](https://doc.rust-lang.org/cargo/commands/cargo-clean.html),
[Go clean](https://go.dev/cmd/go/#hdr-Remove_object_files_and_cached_files),
[Gradle directories](https://docs.gradle.org/current/userguide/directory_layout.html),
[Maven local repositories](https://maven.apache.org/repositories/local.html),
[Homebrew's manual](https://docs.brew.sh/Manpage), and
[Playwright browser management](https://playwright.dev/docs/browsers).

## IDEs, SDKs, containers, and applications

| Category | Read-only discovery and inspection | Supported mutation to preview | Data effect and guardrails |
| --- | --- | --- | --- |
| JetBrains IDEs | Measure exact product/version cache and log directories; use the IDE's special-files view to distinguish cache, config, plugins, and Local History | IDE cache invalidation/restart or exact leftover-version cleanup | Quit the exact IDE first. Index caches rebuild; Local History, settings, plugins, and backups are user state. Never turn a JetBrains-only approval into other cleanup. |
| Xcode and simulators | Measure DerivedData, SourcePackages, Archives, device support, and CoreSimulator separately; list simulators/runtimes | Exact project clean, Organizer archive removal, or deletion of exact unavailable simulator items | Derived data rebuilds. Archives contain release/debug symbols; simulator deletion destroys device/app state. Do not broad-delete Xcode roots. |
| Android SDK and AVDs | Use `sdkmanager --list`, `avdmanager list avd`, and exact path sizes | Uninstall an exact SDK package or delete an exact AVD | AVD deletion destroys emulator state. Do not remove SDK or AVD roots blindly. |
| Docker Desktop | `docker system df -v`, container/image/volume lists, and Docker Desktop settings | Targeted image, container, build-cache, network, or system prune with every flag/filter shown | Volumes can contain databases and are not pruned by default. `--volumes` materially expands scope. Never delete `Docker.raw` or `Docker.qcow2` directly. |
| Colima and similar runtimes | Use runtime status plus the container engine's inventory | Targeted engine prune; runtime-specific disk reclaim only after current help review | Deleting the runtime data disk is reset/data-erasure work, not ordinary cleanup. Do not remove VM disk files directly. |
| Git repositories and worktrees | `git count-objects -vH`, `git worktree list --verbose`, and dry-run pruning where supported | Repository-scoped maintenance or exact stale-worktree metadata cleanup | Preserve worktrees, untracked changes, reflogs, LFS objects, and object databases. Avoid aggressive pruning during concurrent Git writes. |
| Browsers | Use the browser's storage or browsing-data UI and separate cached files from cookies, site data, downloads, and profiles | Clear only the exact displayed cache category/time range | Cookies and site data can sign the user out or propagate through sync. Profiles, credentials, history, and downloads are not cache. |
| Media apps | Use the app's storage UI and separate cache from offline downloads | App-supported clear-cache or exact offline-download removal | Offline downloads require redownload; local-only media may be irreplaceable. Preserve accounts, playlists, and libraries. |
| Virtual machines | Use the manager's disk-reclaim view with the VM shut down | Manager-supported reclaim, exact snapshot removal, archive, or deletion | VM disks and snapshots are user data and rollback state. Require an independent backup; never delete disk bundles directly. |

Official references: [JetBrains directories](https://www.jetbrains.com/help/idea/directories-used-by-the-ide-to-store-settings-caches-plugins-and-logs.html),
[Xcode archives](https://developer.apple.com/documentation/xcode/distributing-your-app-for-beta-testing-and-releases),
[Android SDK manager](https://developer.android.com/tools/sdkmanager),
[Docker pruning](https://docs.docker.com/engine/manage-resources/pruning/),
[Docker Desktop storage settings](https://docs.docker.com/desktop/settings-and-maintenance/settings/),
[Git worktrees](https://git-scm.com/docs/git-worktree.html),
[Safari website data](https://support.apple.com/guide/safari/sfri11471/mac), and
[Dropbox online-only files](https://help.dropbox.com/sync/online-only-mac).

## Refuse or hand off

Do not turn storage cleanup into any of these operations:

- device reset, secure erasure, repartitioning, filesystem repair, or SIP changes;
- malware containment, forensic log removal, or security-control weakening;
- deletion of cloud originals, Photos library internals, the only backup, or
  another user's data;
- application uninstall, account removal, or bulk “debloating” unless separately
  requested and handled by an appropriate workflow;
- Codex task-history, state-database, or runtime cleanup when a dedicated
  lifecycle-aware workflow is available.

When ownership, supported lifecycle, sync effect, or recovery is unknown, stop
at the measured finding and name the missing evidence.
