# Codex Lifecycle Workarounds

These machine-level hooks mitigate known Codex lifecycle bugs without creating
personal plugins. `agent-scripts` owns their definitions and implementations;
an installer deploys that source into Codex's machine-owned runtime state.

## Ownership and installation

| Location | Owner | Purpose |
| --- | --- | --- |
| `codex-hooks/hooks.json` | `agent-scripts` | Canonical hook registrations |
| `codex-hooks/scripts/` | `agent-scripts` | Canonical hook implementations |
| `bin/install-codex-hooks` | `agent-scripts` | Install, merge, and drift-check interface |
| `~/.codex/hooks/` and `~/.codex/hooks.json` | Machine install | Executable copies and merged registrations |
| `~/.codex/config.toml` | Codex | Hook trust hashes and other runtime configuration |
| `~/.codex/log/` and plugin caches | Codex | Mutable logs and runtime state |

Install or update the module from the repository root:

```bash
bin/install-codex-hooks
```

The installer copies both scripts, replaces only registrations whose commands
belong to this module, preserves unrelated hooks, and creates
`~/.codex/hooks.json.agent-scripts-backup` before changing an existing
manifest. It is idempotent. Check for drift without writing anything:

```bash
bin/install-codex-hooks --check
```

After an install changes hook definitions, use `/hooks` to review and trust
the exact entries. Trust is deliberately not granted by the installer. Review
repository changes before updating: Codex trusts the installed command path,
so a later script-only update at an already trusted path may not prompt again.

## Chrome plugin `latest` link

### Failure and repair boundary

The ChatGPT Chrome extension uses a native-messaging manifest whose executable
path passes through the bundled Chrome plugin cache's `latest` entry. Some
Codex plugin installs create a versioned cache directory but do not create or
refresh that symlink. Chrome then reports that the ChatGPT app is unavailable
even though both the app and extension are installed.

The upstream tracker is
[openai/codex#31904](https://github.com/openai/codex/issues/31904). Check its
current state without relying on this document's snapshot:

```bash
gh issue view 31904 --repo openai/codex --json state,closedAt,url
```

[`codex-hooks/scripts/repair-codex-chrome-plugin-latest.sh`](../codex-hooks/scripts/repair-codex-chrome-plugin-latest.sh)
selects the highest dotted-numeric cache version containing both plugin
metadata and an executable macOS native host. It repairs only a missing,
broken, or stale `latest` symlink. It does not delete installed versions and
refuses to overwrite a real file or directory at that path. It also does not
rewrite Chrome's native-messaging manifest; an obsolete executable name in the
manifest is a separate installer failure that must not be hidden by this
workaround.

The installer registers the script for `startup|resume`. Its default invocation
self-repairs and logs only repairs or errors. Use the installed copy with
`--check` for a read-only diagnostic:

```bash
~/.codex/hooks/repair-codex-chrome-plugin-latest.sh --check
```

### Checking whether Codex fixed the installer

The session hook masks the installer defect, so test immediately after an
upgrade or reinstall and before starting or resuming another Codex task:

1. Temporarily disable the Chrome repair entry in `/hooks` or
   `~/.codex/hooks.json`.
2. Move `latest` aside as a recoverable backup, then reinstall or upgrade the
   bundled Chrome plugin through Codex.
3. Run the script with `--check`. A passing result means the installer created
   a symlink to the highest valid installed version without help from the
   workaround.
4. Confirm Chrome can connect through `com.openai.codexextension.json`, whose
   native-host executable must exist and be executable.
5. Repeat after a later plugin version replaces the current one. If the link is
   created and advanced without the hook in both cases, remove this workaround.

If the installer still fails, re-enable the hook and run the script without
`--check` to restore the link. Preserve any backup until the connection test
succeeds.

## Stale `node_repl` helpers

[`codex-hooks/scripts/cleanup-stale-node-repl.sh`](../codex-hooks/scripts/cleanup-stale-node-repl.sh)
is installed at `~/.codex/hooks/cleanup-stale-node-repl.sh` and mitigates
[openai/codex#26984](https://github.com/openai/codex/issues/26984). Abandoned
`node_repl` children can retain standard-I/O pipes after their owning session
ends, eventually pushing a Codex app-server toward its file-descriptor limit.
The hook intervenes only under descriptor pressure, selects old direct helper
children, and revalidates process identity immediately before signalling a
dedicated process group.

Check the upstream tracker directly before retiring the workaround:

```bash
gh issue view 26984 --repo openai/codex --json state,closedAt,url
```

To check whether the upstream lifecycle bug is fixed, disable this cleanup
entry temporarily and exercise several ordinary task start/end cycles that use
the JavaScript REPL. Compare app-server descendants and descriptor counts
before and after the tasks. The workaround can be removed when ended tasks no
longer leave persistent `node_repl` children or growing pipe descriptors, and
the upstream issue or release notes confirm the lifecycle fix. Re-enable the
hook if pressure begins accumulating again.
