# Check sjskills automatically without assigning work to the agent

## Outcome

On startup and resume, the hook runs a deterministic check of the installed
sjskills CLI, configured skill scopes, and this plugin's published version.
Healthy checks stay out of the conversation. Findings requiring attention produce
one short diagnostic with a next step. The hook never assigns maintenance work to
the agent and never updates, installs, synchronizes, quarantines, or restores.

The user selected **only check and report** on 2026-09-15. This replaces the
previous automatic-maintenance target. This request authorizes detailed planning;
implementation, publication, installation, and machine reconciliation have not
been requested by this planning task.

## Current state

- Planning complete; implementation and runtime acceptance have not started.
- [Hook registration](../plugins/sjskills-maintenance/hooks/hooks.json) matches
  startup/resume and gives the Node script five seconds. That timeout currently
  bounds prompt preparation, not the agent's later maintenance.
- [Hook implementation](../plugins/sjskills-maintenance/scripts/session-start.cjs)
  copies workflow snapshots and emits maintenance prose through
  `hookSpecificOutput.additionalContext`. It does not invoke sjskills.
- [Maintenance instructions](../plugins/sjskills-maintenance/references/maintenance.md)
  ask the agent to check/update the plugin, check/update the CLI, and review/apply
  skill plans. Only plugin update checking has a hook-level daily gate.
- [Status collection](../cmd/sjskills/status.go) already checks the CLI, fixed
  global baseline, and nearest configured project concurrently. Its existing
  JSON envelope exposes findings, configuration, errors, and evidence freshness.
- [Skill status](../internal/sjskills/status.go) already uses a 24-hour upstream
  freshness interval, a 15-minute failed-refresh cooldown, and a 30-second
  foreground budget. Local inventory is inspected anew; cached upstream evidence
  does not establish local equality. The
  [performance record](../tasks/sjskills-status-performance.md) reports cold
  checks taking seconds and warm native checks taking tens of milliseconds.
- [CLI evidence](../internal/sjskills/cli_status.go) distinguishes equal, ahead,
  update available, unavailable, and uncomparable versions. Exit zero alone is
  insufficient to establish a healthy check.
- On 2026-09-15, the installed Windows hook script and repository script had
  matching SHA-256 hashes. This establishes the inspected behavior, not a fresh
  release comparison. Existing tests assert maintenance-context injection.

## Next action

When implementation is requested, reproduce the current startup output using the
registered shell command in a temporary plugin installation, then add failing
behavior tests for the check-only contract below. Keep this item queued until
implementation starts.

## Selected design

### One check path

Keep the existing plugin identity and platform entry points. Replace the body of
the hook with a small executable adapter that runs the installed native
`sjskills --json status` once in the event's absolute working directory, validates
the result, and renders the supported hook response. Reuse CLI status ownership
of version selection, configured scopes, authenticated fetching, local inventory,
and cache safety. Do not duplicate reconciliation logic in JavaScript.

Do not add a new sjskills subcommand or change the public status exit contract
just to make the hook work. A Go change is warranted only if a reproduced status
gap prevents this contract; document that gap and its regression before extending
the CLI. Do not fall back to the source-building wrapper, install a missing CLI,
change PATH, or load skill instructions when the native executable is absent.

Keep plugin-version observation in the adapter because it is Codex-specific.
Compare the installed manifest with the manifest at a resolved published `main`
commit of `sjunepark/agent-scripts`, including the cachebuster. Inspect configured
marketplace provenance using read-only metadata; never refresh the marketplace
snapshot merely to check it. A custom source/ref, malformed manifest, disabled
plugin, or unavailable source must not trigger repair. Respect disabled hooks.
Fetch metadata as bounded data, never execute remote content. Cache plugin
observations for 24 hours and failed attempts for 15 minutes, separately from
CLI-owned caches; changing installed version/source invalidates that observation.

### Result and reporting contract

The adapter classifies results internally; these names do not require a new
public CLI interface. Use stable ordering and bounded text, not raw subprocess
output or remote instructions.

| Result | Meaning | Hook-visible output |
| --- | --- | --- |
| `ok` | All applicable checks have fresh evidence and no actionable findings | No conversation context or message; a short execution status label is sufficient |
| `attention` | Update available, missing/extra/changed skill, conflict, or configuration requiring intervention | One concise factual line naming affected scopes and a manual next step |
| `unavailable` | Missing prerequisite, unsupported/incomplete JSON, stale/unavailable evidence, timeout, or failed inspection | One concise line saying verification could not complete and why |

If some checks have findings and others fail, retain both facts in the one-line
summary; never replace known findings with a generic success or failure.
Examples: `sjskills: CLI update available; project skills differ. Run sjskills status.`
and `sjskills: check incomplete (private-source authentication unavailable). Run sjskills status.`
Use fixed diagnostic templates that do not leak tokens, repository content, or
unbounded tool output. Findings are reports, never requests for the agent to run
the suggested command. No `additionalContext` maintenance task in any branch.

Missing project configuration means the project scope is inapplicable, while
global and CLI checks still run. Do not recommend initialization on every session.
Unsupported release targets receive one short skip diagnostic and no install
attempt. An installed stable version ahead of the latest release is not an update
request; an uncomparable development version must not be presented as verified.
Unknown result values, absent applicable scopes, disabled inspection, malformed
JSON, and stale empty findings never count as `ok`.

Preserve fresh findings on each invocation, including repeated actionable
findings. Do not introduce another notification database or suppress unresolved
problems across unrelated sessions. Successful checks have no narrative.

### Execution and state boundaries

- Managed skill roots, manifests, ownership records, quarantines, binaries,
  plugin installations, marketplace sources/snapshots, credentials, and trust
  settings remain unchanged. Existing temporary materialization and disposable
  status-cache writes are permitted; "check-only" is not a ban on all I/O.
- Preserve credential boundaries in the existing status service. No login,
  account switching, setup commands, or fallback to public access after a
  private-source failure. Partial scope failure does not hide another scope.
- Invoke programs using argument arrays and resolved executable paths. Treat
  event paths, names, and returned metadata as data. Preserve Windows hidden
  execution and support spaces, apostrophes, shell metacharacters, and Unicode.
- Use one bounded synchronous invocation rather than a daemon, scheduled job,
  detached updater, or agent follow-up. Target a 35-second adapter deadline and
  a 40-second hook timeout, subject to native cancellation tests and supported
  hook schema verification. Metadata checks share the adapter deadline. Cold
  checks may delay startup; warm checks reuse evidence. Measure both and document
  the observed tradeoff rather than promising instantaneous fresh remote checks.
- Bound input/output and network reads. On timeout, cancel all owned work,
  including spawned materialization descendants; verify cleanup on Windows and
  macOS. Hook failure must not block the user's task or schedule agent recovery.
- Reuse CLI cache locks. For plugin observation caching, acquire only an
  exclusive owned lock, use safe atomic writes, and never remove another run's
  lock. Contention with no usable evidence is unavailable, not success. Cache
  write denial may degrade caching but must not discard otherwise valid fresh
  read results; required read failures remain unavailable.

## Implementation stages

### A. Reproduce and pin the behavior contract

- [ ] Capture the current registered hook's stdout, exit code, duration, and
  injected maintenance context using an isolated home/plugin data directory.
- [ ] Extend [hook tests](../scripts/sjskills-hook.test.js) with controlled native
  CLI fixtures and result cases from the acceptance matrix. First demonstrate
  that the old implementation fails the no-agent-work requirement.
- [ ] Confirm the minimum released CLI JSON fields against an actual compatible
  release. Validate required fields and scope completeness; tolerate harmless
  additive fields. Older/incompatible output gets an actionable diagnostic.
- [ ] Verify current official hook output, timeout, and failure behavior before
  selecting the final response envelope. The existing `systemMessage` use is
  an implementation clue, not proof of how the target client presents it.

Exit: reproducible evidence and failing tests identify the user-visible problem
without running automatic updates or applying plans to a real home.

### B. Execute checks directly and remove the agent workflow

- [ ] Implement CLI execution, JSON validation, result classification, and one-line
  reporting in the plugin, reusing the existing status engine.
- [ ] Implement read-only plugin manifest observation with bounded fetching,
  provenance validation, cache freshness, and explicit unavailable results.
- [ ] Update hook status text and timeout; remove the context-injection contract.
- [ ] Remove the maintenance prompt, bundled sjskills workflow copies, snapshot
  creation, update-success bookkeeping, and agent update/apply lock handling.
  Search actual callers before deleting [the bundle generator](../scripts/sync-sjskills-plugin)
  and its test/CI/documentation references. Keep the canonical sjskills skill.
- [ ] Replace tests for instruction delivery with tests for process execution,
  checked results, no mutating commands, and no injected agent task.

Exit: a real shell invocation performs the check itself and emits the specified
result without an agent turn, installer, `apply`, `restore`, or plugin update.

### C. Validate reliability and native behavior

- [ ] Run `node --test scripts/sjskills-hook.test.js` with temporary homes,
  caches, credentials, and controlled network responses.
- [ ] Exercise the registered Windows command through PowerShell and cmd.exe,
  and the macOS shell command on Intel and Apple silicon. Preserve Linux CI
  testing of supported-platform logic and the unsupported-target response.
- [ ] Run a compatible real native CLI against isolated configured and
  unconfigured projects. Capture cold/warm timings, nonzero exits, cancellation,
  process cleanup, network/auth failures, and parallel startup results.
- [ ] Inspect before/after filesystem and command logs to prove managed roots,
  executables, installs, credentials, and trust remain unchanged.
- [ ] Run one bounded code review; fix actionable in-scope defects and run the
  relevant checks. If Go changes prove necessary, also use the Go skill and run
  focused status tests plus required Go checks.

Exit: the acceptance matrix passes, native limitations are explicit, and a
healthy check cannot be confused with stale, partial, or failed verification.

### D. Package and document the replacement

- [ ] Update manifest descriptions and display text to check/report behavior,
  preserving the install identifier `sjskills-maintenance@personal`.
- [ ] Refresh the plugin cachebuster and validate the manifest. Install/reinstall
  in an isolated temporary Codex configuration and invoke the packaged hook.
- [ ] Update [the operator guide](../docs/sjskills-startup-hook.md),
  [settings guide](../docs/settings-sync.md), affected CI/test references, and
  progress status. Remove ongoing standing-maintenance claims from runtime
  instructions; retain manual CLI update and explicit sjskills sync workflows.
- [ ] Run scoped documentation harmonization. Record source validation separately
  from publication, hook trust, and actual host activation.

Exit: packaged code and documentation describe one check-only path; no shipped
prompt still requests unattended mutation or routine agent maintenance.

### E. Publish and activate when separately authorized

- [ ] Publish reviewed source to the configured remote `main` before updating
  ongoing installations. If the CLI was unchanged, no new CLI release is needed.
- [ ] Reinstall only this plugin from the remote-backed personal marketplace on
  each explicitly selected host. Do not use the old self-update workflow to
  deliver the replacement or silently enable a disabled hook.
- [ ] Review changed hook trust through `/hooks` as required; never write trust
  hashes. Check for a duplicate legacy standalone hook before host acceptance.
- [ ] Start a fresh trusted session and verify that a healthy check adds no
  maintenance instructions or agent tool calls. Verify an actionable fixture
  yields only the short report. Record host/plugin version and native evidence.

Exit: selected hosts actually execute the published check-only hook. Source-only
completion does not satisfy activation. If activation fails, disable the affected
hook rather than silently reverting to automatic maintenance.

## Acceptance matrix

| Scenario | Required evidence |
| --- | --- |
| Fresh CLI/global/project, no findings | Silent conversation; no maintenance context or agent calls |
| No project manifest | Global/CLI still checked; no initialization or repetitive setup warning |
| CLI/plugin update available | One-line attention report; zero installation/marketplace refresh calls |
| Missing, extra, modified, conflicted skills | Attention with affected scope; zero apply/quarantine/restore calls |
| CLI ahead or development version | Ahead is not downgraded; uncomparable is not verified healthy |
| Exit zero with unavailable/stale evidence | Incomplete check reported, never `ok` |
| Missing CLI, incompatible JSON, omitted scope, disabled status | Actionable incomplete-check report; no fallback build/install |
| Private auth failure or network outage | No login or credential output; independent results retained |
| Cache cold/warm, expiry, failed-refresh cooldown, clock rollback | Correct evidence reuse; local changes detected while upstream evidence is cached |
| Concurrent sessions or unsafe/unwritable cache | No stolen locks, corrupted state, false success, or protected-root writes |
| Timeout or cancellation | Owned child processes cleaned up; bounded hook returns without agent recovery |
| Startup/resume and ignored events | Exactly one check for each matching event; none for compact/clear/subagent events |
| Installed package and native shells | Correct quoting and stdout envelope; no visible Windows helper window |
| Fresh trusted session after rollout | Actual client presentation and absence of agent maintenance verified |

## Compatibility and removal decisions

- Retain the plugin identifier, supported native targets, existing status API,
  cache/provenance protections, and explicit manual reconciliation workflows.
- Remove the prompt-driven maintenance path and its generated source bundle.
  Its additional machinery exists to support agent-driven mutation, which is
  outside the selected target.
- Leave existing plugin-data workflow snapshots and old update locks/state
  untouched on disk. The replacement never reads them as current evidence and
  does not run an automatic migration or cleanup. Suspended old sessions can
  still reference snapshots; deleting those is a separate cleanup task.
- A fresh session is required to verify removal of already-injected instructions.
  Reinstalling code cannot retract context from an active old conversation.
- The only unresolved implementation evidence is exact client presentation,
  released-CLI compatibility, read-only marketplace metadata access, and native
  timing/cancellation. These are validation tasks, not unanswered product choices.
