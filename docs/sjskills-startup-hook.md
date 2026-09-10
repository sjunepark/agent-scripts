# Automatic sjskills maintenance plugin

`plugins/sjskills-maintenance/` packages the startup hook and its workflow
references. The agent performs maintenance in the first turn of each new or
resumed session: check this plugin against published `main` daily, update the
stable sjskills CLI, and review/apply configured global and project skill plans.
The hook caches workflow references and emits the task; it does not start an
unattended updater or another agent.
See [PROGRESS.md](../PROGRESS.md) for catalog registration, publication, and
machine activation status. Source implementation does not imply availability
from the remote marketplace.

## Install and manage on each host

Once this plugin and its catalog entry are published to `main`, use the existing
remote-backed `personal` marketplace. Inspect its source first:

```text
codex plugin marketplace list --json
codex plugin list --marketplace personal --available --json
```

If not yet configured, register it:

```text
codex plugin marketplace add https://github.com/sjunepark/agent-scripts.git --ref main
```

An existing marketplace called `personal` must refer to that Git repository and
`main`; do not silently replace another marketplace using the same name. Then:

```text
codex plugin marketplace upgrade personal
codex plugin add sjskills-maintenance@personal
```

Open `/hooks` and review/trust the new definition, then start a new session.
Enabling this plugin supplies standing authority for published-main plugin
updates, stable CLI updates, and configured skill reconciliation. Skill sync can
quarantine undeclared copies and preserve managed local edits in quarantine
before replacement. Unmanaged desired copies or untrusted provenance block sync. It follows the [sjskills workflow](../skills/sjskills/SKILL.md).

If the former standalone hook was installed, disable its user-level
`~/.codex/hooks.json` entry in `/hooks` before enabling the plugin entry.
Identify it by the command referencing `hooks/sjskills/session-start.cjs`;
both entries may display the same status message. Preserve unrelated hooks.
Codex runs matching hooks from all sources, so leaving both enabled would
request maintenance twice. The standalone installer has been retired.

The plugin's installed cache belongs to Codex. Do not symlink it to a working
checkout or manage it through chezmoi. Runtime paths come from `PLUGIN_ROOT`
and `PLUGIN_DATA`; Node must be available on the hook's PATH. To disable
maintenance, disable its hook through `/hooks` or remove the plugin:

```text
codex plugin remove sjskills-maintenance@personal
```

## Updates and inspection

- Each startup checks the plugin's persisted schedule. A successful remote
  check remains current for 24 hours; a failed attempt waits 15 minutes before
  retrying. Scheduling a task does not count as a completed check. No sessions
  means no background polling.
- When due, the agent verifies remote repository/ref provenance, refreshes only
  `personal`, reads the published plugin manifest from the refreshed snapshot,
  and compares its full version with the installed version. Cachebuster suffixes
  count. `plugin list` may show only the installed version for an installed
  plugin, so that output alone cannot establish remote freshness.
- A different published version triggers installation of only
  `sjskills-maintenance@personal`. Installed-version verification precedes
  recording success. Other plugins are not reinstalled. A local development
  marketplace or unexpected source/ref is reported and left unchanged.
- The task requires an exclusive lock around each plugin update and each CLI
  update. Skill sync uses a stable executable copy and existing reconciliation
  locks and evidence checks. Interrupted locks are reported, never stolen.
- `PLUGIN_DATA/update-check.json` stores the last completed check timestamp and
  outcome; `PLUGIN_DATA/update.lock` is the agent-owned update lock. The helper
  at `scripts/update-check.cjs` inside the installed plugin accepts
  `status PLUGIN_DATA` for read-only inspection and `success PLUGIN_DATA` or
  `failure PLUGIN_DATA` to record a completed attempt. These records describe
  update checks, not hook trust or verified skill state.
- The hook preserves content-addressed copies of its workflow references and
  update helper under `PLUGIN_DATA/workflows/`. Codex removes old plugin caches
  during reinstall; these snapshots keep active and suspended sessions' paths
  valid. Snapshots are retained rather than automatically pruned while sessions
  may still reference them. They contain code/workflow text, not credentials or
  session transcripts.
- New plugin code and installed skill changes may need a new session. The
  current agent continues its loaded workflow without recursively restarting.
  New or changed hook definitions may require `/hooks` trust review. Updates
  never write trust hashes or bypass approval settings.

For a manual refresh use the same marketplace upgrade and plugin add commands
above. Inspect installed version with `codex plugin list --marketplace personal
--json`, CLI/skill state with `sjskills status`, and activation with `/hooks`.

## Coverage and failure behavior

Install per host and per custom `CODEX_HOME`. Codex desktop and CLI sessions
sharing that configuration share the plugin; remote sessions use their remote
host. Separate ChatGPT or cloud conversations are covered only when their
runtime loads these hooks and can access the host's tools and files.

Supported sjskills targets are Windows x64 and macOS Intel/Apple silicon.
Linux and other unsupported targets receive a skip diagnostic. Only `startup`
and `resume` match; compaction, context clearing, and subagents do not trigger
another sync. A missing project manifest means global-only reconciliation,
without project initialization.

Changes and actionable failures receive a short report; verified unchanged
state is quiet. Missing prerequisites, network failures, permission limits,
and update conflicts do not prevent the user's original task from proceeding.
The task needs Codex CLI for plugin updates, platform release-installer tools,
Bun/Git for skill materialization, and access to the relevant installation and
skill roots. Neither the hook nor its update authority bypasses the sandbox.

## Maintain and publish the plugin

Edit the plugin source in this repo. The bundle under
`references/sjskills/` is generated from the canonical skill; refresh and check it:

```text
node scripts/sync-sjskills-plugin
node scripts/sync-sjskills-plugin --check
```

Update the plugin cachebuster with the plugin-creator helper, validate the
manifest, and reinstall the resulting plugin in an isolated local marketplace
before testing. Run `node --test scripts/sjskills-hook.test.js`. The tests cover
bundled resources, native shell invocation, state intervals, and failure behavior;
they do not prove a trusted agent completed a real-machine update and sync.

Commit and publish reviewed source plus its marketplace entry to `main` before
updating ongoing installations. Every change to plugin code or bundled workflow
needs a new cachebuster. The auto-update task adopts published content; it does
not commit, push, or consume uncommitted development changes. Follow the
[repo marketplace workflow](settings-sync.md#codex-plugins).

Protocol, plugin paths, and trust behavior follow the
[official hooks documentation](https://learn.chatgpt.com/docs/hooks).
