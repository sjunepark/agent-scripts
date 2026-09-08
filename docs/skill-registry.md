# Skill registry

`skill-registry.json` is the version 4 desired-state contract consumed by
`sjskills`. It classifies this repository's published skills and deliberately
recommended external skills without duplicating a skill declaration across
global and project policy.

## Desired sets

`global.baseline` is one fixed, machine-independent set. It is the only
selection resolved by `sjskills plan --global`; global machine profiles and
hostname inference do not exist.

`profiles` contains composable project selections such as `dev`, `go`,
`rust`, and `kicpa`. A project's committed `sjskills.toml` selects those
profiles and may add direct third-party declarations. Profile membership does
not make a skill global.

Every selected skill name must be unique across the global baseline, selected
project profiles, and direct project declarations. Duplicate names,
contradictory sources, and overlap between centrally declared desired sets fail
before materialization or managed-root writes.

The default targets are `.agents` and `.claude`, mapping to the
`.agents/skills` and `.claude/skills` roots in the selected project or
home. `targetExceptions` records the small set of skills that support only
one target. The reconciler never synthesizes a Pi placement and never invokes
the Skills CLI with `--all`.

## Skill and source records

Each `skills` entry records:

- `name`: the portable skill identity.
- `source`: a key in `sources`.
- `manager`: `skills-cli`, `manual`, `workflow`, or `none`.
- `mode`: `copy` for Skills CLI-managed entries.
- `fullDepth`: an optional Skills CLI materialization requirement.
- `workflow`: the provisioning workflow for workflow-managed entries.

A `repository` source names the published GitHub `skills/` catalog.
An `external` source names a deliberately tracked upstream or a manual
boundary. Skills CLI-managed sources must be public Git shorthand
(`owner/repo[/path]`) or credential-free HTTPS. Local paths, embedded
credentials, URL query strings, npm specifiers, and other schemes are rejected
for that manager.

`manual` entries remain externally owned. `workflow` entries are provisioned
by the named project workflow. `none` entries are catalog-only: they are
classified for discovery but never selected by a desired set.

## Ownership and reconciliation

The registry owns desired classification; each `SKILL.md` owns skill
behavior. Installed trees and reconciler provenance are derived state, not
additional configuration.

`sjskills` invokes the exactly pinned Skills CLI only inside isolated
temporary homes, verifies one staged tree per desired skill, and owns final
placement itself. Skills with the same source and `fullDepth` option share one
fetch; every requested tree must still be present and verified. Materializer
process groups on Unix and jobs on Windows stop descendants before staging
cleanup, including after a parent exits. Byte equality and Skills CLI lock
metadata do not grant ownership. A desired placement is updated only when trusted reconciler
provenance still matches its source and current tree hash.

Both scopes strictly reconcile their `.agents/skills` and `.claude/skills`
roots to the selected desired set. Every verifiable undeclared directory moves
into recoverable quarantine, including unknown, former-profile, and locally
modified copies. Removal is bound to the observed current tree hash. Restore
retains ownership only when prior provenance matched those bytes; restoring an
unknown or modified copy does not grant ownership.

On Windows, identity checks around publication, quarantine, and recovery use
handle-based file information captured before a move or content check. Long
drive and UNC paths use extended-length forms for the native handle open. Plain
`os.Lstat` defers file-ID lookup on Windows, so comparing its result after a
rename can incorrectly report a conflict or resolve a reused path. Identity
checks remain separate from content hashes; identical bytes do not establish
ownership of a replacement directory.

An unknown or locally modified copy at a desired path, malformed provenance,
an unsafe root, or an unverifiable extra blocks apply. Extras are not silently
preserved as exact state. Declared manual and workflow entries retain their
external provisioning boundaries. Built-in skills, plugin caches, legacy Pi
copies, and legacy provenance outside the managed roots remain outside strict
sync.

Global provenance is stored at
`~/.agents/.global-skill-state.json`. Private global locks, journals,
recovery data, and manifest-backed quarantine live under
`~/.agents/.sjskills-global/`. The prior `~/.skill-quarantine` location is
protected and never reused by `sjskills`.

## Automatic status evidence

`sjskills` and `sjskills status` share one explicit report of the nearest project
and fixed global baseline. Project discovery follows the nearest ancestor
`sjskills.toml`, independently of Git boundaries, and displays its resolved root.
An inspectable directory without a manifest gets optional setup guidance through
`profiles` and `init`; no manifest or project cache entry is created. Invalid or
unreadable configuration and invalid starting directories are unavailable
inspection, with review/repair guidance. Each scope survives failure of its peer.
Absent global installation roots ordinarily mean missing skills, not missing setup.

Explicit human reports go to stdout, project before global. Fresh empty findings
say “no drift detected”; stale empty findings say “no drift detected against stale
evidence.” Reports include cache observation age when applicable and review
commands for drift or unavailable/stale evidence. A produced report exits 0 for
all advisory outcomes, including missing setup. Invalid command/flag/argument
invocations exit 64 without checking; cancellation uses execution-failure exit 2
and a diagnostic on stderr in human mode. JSON encodes failures in its envelope.

Successful `init`, `profiles`, `plan`, `apply`, and `restore` instead emit
incidental notices on stderr after primary output, omitting unconfigured projects
and fresh scopes without findings. Both paths share the reconciliation classifier
and inspect current files and provenance each time; cached findings are never
reused. “No drift detected” does not verify externally provisioned manual or
workflow tools. Protected locations retain their existing ownership boundaries.

Disposable complete expected-hash snapshots live under the platform user-cache
directory in `sjskills/status/`. Identity includes the canonical scope root,
embedded registry, desired sources, install options, targets, and format versions.
Snapshots refresh after 24 hours, with a shared 30-second foreground budget and
at most two concurrent scope refreshes. Failed attempts retain matching stale
evidence and a 15-minute retry cooldown; incompatible or incomplete evidence
cannot establish status. Per-scope locks, bounded reads, atomic replacement, and
bounded pruning of entries older than 30 days protect this disposable cache.
Deleting it causes a cold check without changing installed state or provenance.
SIGINT and SIGTERM cancel ongoing checks and permit cleanup; a repeated signal
restores the normal force-exit behavior when a command remains blocked on input.

A successful verified plan or apply supplies its scope's hashes only after
staging verification and cleanup. Notices inspect post-operation state, run after
mutation locks are released, and never run during confirmation. Human output
groups at most five names per category, with remaining counts and cached age;
fresh explicit plans suppress duplicate notices for their scope. Fresh scopes
with no findings are silent in incidental notices; explicit reports always show
checked scopes.

The optional JSON `advisories` field carries full findings, targets, reason codes,
observation time, freshness, and review commands. It is separate from stable
warnings and approval evidence. The artifact SHA-256 binds all bytes, including
advisories; only the subsequent fresh-plan semantic comparison ignores that field.
The strict loader validates advisory structure and still rejects unknown fields.
Older artifacts remain readable by the new executable, but older strict loaders
cannot read new artifacts containing advisories. Keep the same executable through
plan and apply. Advisory failures never change primary command success or grant
mutation authority.

Only status envelopes include the optional `status` result object:
`projectConfiguration` is `configured`, `not-configured`, `unavailable`, or
`skipped`; `projectRoot` is present when discovery resolved it. A configured or
unavailable project has a project advisory, including empty findings; an
unconfigured project has only the global advisory. Setup states never replace
advisory freshness. Status contains no plan or approval evidence. Reviewed global
apply rejects status envelopes and any `status` field on a plan, including null;
only advisories retain the existing limited semantic-comparison exclusion.

`--no-status-check` disables all status discovery, inventory, refresh, and cache
work. Explicit status then prints “Status checks disabled (--no-status-check).”
or emits `projectConfiguration: "skipped"` with no root or advisories in JSON.
For other commands the flag does not disable primary live verification. Root
flags work before or after a named command. Help and exact version requests
perform no status work.

## Validation and consumers

After changing the registry or published catalog, run:

```bash
scripts/validate-skills
node --test scripts/lib/skill-registry.test.js \
  scripts/audit-global-skills.test.js
go test ./...
```

Run the Go suite on Windows as well as Unix when changing reconciliation.
CLI integration tests build a native fake `bunx` in a temporary directory so
they do not fall through to a real Skills CLI on Windows. Unix permission-bit
assertions apply only on Unix; Windows uses native ACLs and synthesized mode
bits. All global mutation tests use isolated temporary homes.

`scripts/validate-skills` requires every `skills/*/SKILL.md` to have exactly
one repository-source record and rejects missing sources, unused sources,
invalid desired-set membership, unsupported target exceptions, and invalid
manager/source combinations.

Use `sjskills plan` in a project and `sjskills plan --global` for the fixed
baseline. Planning establishes current remote content in temporary storage but
does not change managed roots. `scripts/audit-global-skills` remains only as a
read-only compatibility wrapper for the global plan; its version 3 profile and
mutation arguments are retired.

Global apply requires `--approved-plan <plan.json>` together with
`--approved-plan-sha256 <digest>`. The command reads the artifact once, verifies
its approved digest and strict successful-global-plan shape, then rematerializes
and recomputes the complete plan. All stable warnings, operations, current and
expected evidence must match before confirmation or mutation. Apply uses that
same still-live verified materialization session, so a remote ref change cannot
replace reviewed expected content after the recheck. Missing evidence, artifact
substitution, content movement, or inventory drift fails closed.

Real-home global apply, restore, migration, and quarantine are operational
changes, not repository validation. The [sjskills skill](../skills/sjskills/SKILL.md)
defines request scope: a configured sync authorizes reconciliation, while the
agent prepares and reviews the required evidence. No separate human approval of
hashes or counts is required. Restore still needs its named-quarantine request;
sync never bypasses provenance, conflict, or filesystem checks.
