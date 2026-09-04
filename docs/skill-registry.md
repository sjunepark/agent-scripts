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
placement itself. Byte equality and Skills CLI lock metadata do not grant
ownership. A placement is updated or removed only when trusted reconciler
provenance still matches its source and current tree hash.

On Windows, identity checks around publication, quarantine, and recovery use
handle-based file information captured before a move or content check. Long
drive and UNC paths use extended-length forms for the native handle open. Plain
`os.Lstat` defers file-ID lookup on Windows, so comparing its result after a
rename can incorrectly report a conflict or resolve a reused path. Identity
checks remain separate from content hashes; identical bytes do not establish
ownership of a replacement directory.

Unknown entries are reported and preserved. An unknown entry at a desired path,
a locally modified managed tree, malformed provenance, or an unsafe filesystem
boundary blocks the affected operation. Former global-profile placements,
legacy Pi copies, and stale legacy provenance are migration evidence only; v1
does not automatically adopt or remove them.

Global provenance is stored at
`~/.agents/.global-skill-state.json`. Private global locks, journals,
recovery data, and manifest-backed quarantine live under
`~/.agents/.sjskills-global/`. The prior `~/.skill-quarantine` location is
protected and never reused by `sjskills`.

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
changes, not repository validation. They require a separate reviewed,
evidence-bound rollout plan and explicit authorization.
