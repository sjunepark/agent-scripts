# Check sjskills automatically without assigning work to the agent

## Outcome

On startup and resume, the hook runs a deterministic check of the installed
sjskills CLI, configured skill scopes, and this plugin's published version.
Healthy checks stay out of the conversation. Findings requiring attention produce
one short diagnostic with a next step. The hook never assigns maintenance work to
the agent and never updates, installs, synchronizes, quarantines, or restores.

The user selected **only check and report** on 2026-09-15. This replaces the
previous automatic-maintenance target. Stages A–D were delivered through the
[closed goal](../goals/sjskills-startup-check.md) and PR lifecycle. Real-host
activation remains stage E work requiring separate authority; configured
reconciliation remains excluded.

## Current state

- Stages A–B are implemented: native checks, strict partial-result reporting,
  read-only plugin observation, and removal of the injected maintenance path.
- Stage C: Windows shell, subprocess, cache, provenance, transport, and released
  CLI compatibility checks pass. Final hosted native acceptance passed on all three targets in
  [run 34927077118](https://github.com/sjunepark/agent-scripts/actions/runs/34927077118).
- Stage D: manifest/cachebuster and operator documentation describe check-only
  behavior. Isolated install/reinstall passed: healthy packaged Windows hook exited 0
  with zero stdout bytes in 5,321 ms. Scoped documentation harmonization and
  repository source checks passed.
- The old registered installed command was reproduced under a temporary
  CODEX_HOME: exit 0, 1,147 ms, and 9,463 injected context characters. A
  no-agent-context assertion failed against that output before replacement.
- Immutable Windows v1.3.0 archive SHA-256 matched SHA256SUMS. Isolated status
  confirmed configured/unconfigured JSON compatibility. Cold checks took
  9,680–11,040 ms; warm checks took 55–91 ms. GitHub HTTP 403 and a project
  refresh failure remained incomplete while fresh global findings survived.
- A native parent-exit reproduction proved taskkill could miss a surviving
  descendant. Windows Job Object supervision now owns that lifecycle; tests
  cover ordinary cancellation and descendants outliving the native parent.
- Bounded independent code review found one mixed-validity reporting defect;
  supported findings now remain visible alongside incomplete verification.
- Actual host installation, hook trust, and fresh-session activation have not
  been performed and remain stage E.

## Next action

Stages A–D are complete. [PR #24](https://github.com/sjunepark/agent-scripts/pull/24)
merged as `751d93f` to `codex/sjskills-startup-integration`. Stage E remains
unstarted and requires separate authority, including any promotion to `main`
and host activation.

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

- [x] Capture the current registered hook's stdout, exit code, duration, and
  injected maintenance context using an isolated home/plugin data directory.
- [x] Extend [hook tests](../scripts/sjskills-hook.test.js) with controlled native
  CLI fixtures and result cases from the acceptance matrix. First demonstrate
  that the old implementation fails the no-agent-work requirement.
- [x] Confirm the minimum released CLI JSON fields against an actual compatible
  release. Validate required fields and scope completeness; tolerate harmless
  additive fields. Older/incompatible output gets an actionable diagnostic.
- [x] Verify current official hook output, timeout, and failure behavior before
  selecting the final response envelope. The existing `systemMessage` use is
  an implementation clue, not proof of how the target client presents it.

Exit: reproducible evidence and failing tests identify the user-visible problem
without running automatic updates or applying plans to a real home.

### B. Execute checks directly and remove the agent workflow

- [x] Implement CLI execution, JSON validation, result classification, and one-line
  reporting in the plugin, reusing the existing status engine.
- [x] Implement read-only plugin manifest observation with bounded fetching,
  provenance validation, cache freshness, and explicit unavailable results.
- [x] Update hook status text and timeout; remove the context-injection contract.
- [x] Remove the maintenance prompt, bundled sjskills workflow copies, snapshot
  creation, update-success bookkeeping, and agent update/apply lock handling.
  Search actual callers before deleting the retired `scripts/sync-sjskills-plugin` bundle generator
  and its test/CI/documentation references. Keep the canonical sjskills skill.
- [x] Replace tests for instruction delivery with tests for process execution,
  checked results, no mutating commands, and no injected agent task.

Exit: a real shell invocation performs the check itself and emits the specified
result without an agent turn, installer, `apply`, `restore`, or plugin update.

### C. Validate reliability and native behavior

- [x] Run `node --test scripts/sjskills-hook.test.js` with temporary homes,
  caches, credentials, and controlled network responses.
- [x] Exercise the registered Windows command through PowerShell and cmd.exe,
  and the macOS shell command on Intel and Apple silicon. Preserve Linux CI
  testing of supported-platform logic and the unsupported-target response.
- [x] Run a compatible real native CLI against isolated configured and
  unconfigured projects. Capture cold/warm timings, nonzero exits, cancellation,
  process cleanup, network/auth failures, and parallel startup results.
- [x] Inspect before/after filesystem and command logs to prove managed roots,
  executables, installs, credentials, and trust remain unchanged.
- [x] Run one bounded code review; fix actionable in-scope defects and run the
  relevant checks. If Go changes prove necessary, also use the Go skill and run
  focused status tests plus required Go checks.

Exit: the acceptance matrix passes, native limitations are explicit, and a
healthy check cannot be confused with stale, partial, or failed verification.

### D. Package and document the replacement

- [x] Update manifest descriptions and display text to check/report behavior,
  preserving the install identifier `sjskills-maintenance@personal`.
- [x] Refresh the plugin cachebuster and validate the manifest. Install/reinstall
  in an isolated temporary Codex configuration and invoke the packaged hook.
- [x] Update [the operator guide](../docs/sjskills-startup-hook.md),
  [settings guide](../docs/settings-sync.md), affected CI/test references, and
  progress status. Remove ongoing standing-maintenance claims from runtime
  instructions; retain manual CLI update and explicit sjskills sync workflows.
- [x] Run scoped documentation harmonization. Record source validation separately
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
- Released-CLI compatibility, read-only marketplace metadata, and native
  timing/cancellation have source acceptance evidence. Exact client presentation
  in a fresh trusted session remains stage E rollout acceptance.
