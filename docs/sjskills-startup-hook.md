# sjskills startup checks

The source plugin `sjskills-maintenance@personal` runs checks directly on startup
and resume. Healthy results produce no conversation output. Actionable findings
or incomplete verification produce one short UI diagnostic with a manual next
step. The hook supplies no agent instructions or maintenance authority.

The [delivery plan](../plans/sjskills-startup-check.md) records validation and
publication status. Installing or activating this replacement on real hosts is
separate from source delivery.

## Checks and boundaries

- Run the installed native `sjskills --json status` once in the session directory.
  The CLI owns release comparison, configured project discovery, global status,
  authenticated fetching, local inventory, and its existing caches. Missing
  project configuration skips only that scope. Missing native binaries and
  source-building wrappers produce an incomplete check; nothing is installed.
- Inspect this plugin's enabled installed entry with read-only Codex metadata.
  Require the personal marketplace's Git source to be
  `https://github.com/sjunepark/agent-scripts.git`, with ref `main`. The adapter
  reads the ordinary `[marketplaces.personal]` table emitted by Codex to verify
  the ref that listings omit. Unrecognized configuration forms fail closed.
- Resolve published `main` to a commit and fetch that commit's plugin manifest
  as bounded data. Compare the complete version, including its cachebuster.
  Do not refresh the marketplace or execute downloaded content.
- Never update binaries/plugins, change marketplace settings, synchronize skills,
  quarantine, restore, log in, or change trust. Existing CLI temporary
  materialization and disposable status caches remain permitted.

Fresh, complete evidence is required for silence. Exit zero, stale empty
findings, missing scopes, disabled inspection, and development versions do not
establish health. Known findings remain visible when another check fails.
Diagnostics use fixed text and omit skill names, private sources, credentials,
raw subprocess output, and remote text. Repeated unresolved findings remain
visible on each applicable invocation.

## Timing and process ownership

The adapter has a 35-second deadline inside a 40-second hook timeout. It bounds
stdin and subprocess output to 1 MiB and each metadata response to 128 KiB.
CLI status and plugin observation run concurrently. There is no background
worker or later agent recovery task.

On macOS, children run in an owned process group. On Windows, hidden PowerShell
supervisors join kill-on-close Job Objects before launching the native programs;
this prevents a departed parent from leaving descendants behind. Environments
that disallow that supervision report incomplete checks. Node.js and native
Codex/sjskills executables must be available; status materialization uses the
CLI's existing prerequisites. Windows supervision adds startup overhead.

Upstream plugin observations remain fresh for 24 hours; failed requests have a
15-minute cooldown. Changed installed version/source invalidates reuse, and a
backward clock forces a refresh. The CLI still inspects local skills each time.
The plugin owns only `PLUGIN_DATA/plugin-observation.json` and its exclusive
`plugin-observation.lock`. It never steals a lock. Unsafe paths fail closed;
a denied cache write does not invalidate a successful fresh read.

Old `update-check.json`, `update.lock`, and `workflows/` snapshots are neither
read nor removed. Suspended sessions may still hold their paths; cleanup is a
separate task. Reinstalling cannot retract old instructions from an active
conversation.

## Manual installation and updates

After reviewed source is published to `main`, inspect the existing marketplace:

```text
codex plugin marketplace list --json
codex plugin list --marketplace personal --available --json
```

An existing `personal` marketplace must have the Git source and `main` ref above.
If absent, register it explicitly:

```text
codex plugin marketplace add https://github.com/sjunepark/agent-scripts.git --ref main
```

For a requested plugin update:

```text
codex plugin marketplace upgrade personal
codex plugin add sjskills-maintenance@personal
```

Review changed hook trust through `/hooks` and start a fresh session. Preserve
disabled hooks and unrelated plugins. If a legacy standalone entry references
`hooks/sjskills/session-start.cjs`, disable that duplicate through `/hooks`.
Do not write trust hashes or bypass review. Installations are host-local and
respect custom `CODEX_HOME`; source validation is not host activation.

Inspect CLI/skill state with `sjskills status`. CLI updates follow the
[release guide](sjskills-releases.md); explicitly requested reconciliation follows
the canonical [sjskills workflow](../skills/sjskills/SKILL.md). Enabling the
startup check grants neither operation. Disable its hook through `/hooks` or
remove the plugin with `codex plugin remove sjskills-maintenance@personal`.

## Source validation

Update the plugin cachebuster, validate its manifest, and install/reinstall in
an isolated temporary Codex configuration before testing the package. Run:

```text
node --test scripts/sjskills-hook.test.js
```

Tests compile a controlled native CLI fixture with Go and use temporary homes,
command logs, protected-path assertions, and controlled responses. Native CI
covers Windows x64 and macOS Intel/Apple silicon. Linux CI exercises the adapter
logic and unsupported-target response. Manually dispatching `Checks` also runs
native artifact acceptance. Actual fresh-session presentation and hook trust
remain host-rollout acceptance, beyond source-only tests.

The response envelope follows the [official hook contract](https://learn.chatgpt.com/docs/hooks).
Windows supervision follows [Job Object lifetime rules](https://learn.microsoft.com/en-us/windows/win32/procthread/job-objects).
