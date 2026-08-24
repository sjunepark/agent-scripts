# Roll out the fixed global baseline

Status: proposed; not authorized

## Boundary

This is the separate real-machine rollout required by the `sjskills` v1
delivery contract. Creating or reviewing this plan does not authorize
`apply --global`, restore, migration cleanup, moving active placements, or
any mutation under a real home.

The rollout targets only the fixed registry v4 baseline in
`~/.agents/skills` and `~/.claude/skills`, plus reconciler-owned state at
`~/.agents/.global-skill-state.json` and
`~/.agents/.sjskills-global/`. Former profile skills, Pi copies, unknown
entries, plugin caches, vendor metadata, and the legacy
`~/.skill-quarantine` tree remain preserved.

## Required authorization evidence

For each machine, collect a path-free read-only artifact from the exact
published commit proposed for rollout:

```bash
git fetch origin main
git rev-parse origin/main
bin/sjskills --json plan --global > /tmp/sjskills-global-plan.json
shasum -a 256 /tmp/sjskills-global-plan.json
```

On Windows PowerShell, use an operator-owned temporary directory and the
equivalent commands:

```powershell
git fetch origin main
git rev-parse origin/main
go build -o $env:TEMP\sjskills.exe ./cmd/sjskills
& $env:TEMP\sjskills.exe --json plan --global > $env:TEMP\sjskills-global-plan.json
(Get-FileHash -Algorithm SHA256 $env:TEMP\sjskills-global-plan.json).Hash
```

Review the complete JSON, including every operation, warning, expected-content
hash, and materialization result. Stop if it contains a blocked placement,
untrusted provenance, an unmanaged desired path, a modified managed tree, an
unsafe filesystem boundary, or any operation beyond the expected fixed
baseline.

Authorization must identify:

- machine;
- exact repository commit;
- SHA-256 of the reviewed plan artifact;
- allowed install, update, and quarantine counts;
- whether trusted legacy state migration is expected;
- the operator and approval time.

Immediately before mutation, rerun the plan into a second artifact and require
byte-for-byte equality with the approved artifact:

```bash
bin/sjskills --json plan --global > /tmp/sjskills-global-plan.recheck.json
cmp /tmp/sjskills-global-plan.json /tmp/sjskills-global-plan.recheck.json
```

On Windows PowerShell:

```powershell
& $env:TEMP\sjskills.exe --json plan --global > $env:TEMP\sjskills-global-plan.recheck.json
if (Compare-Object `
    (Get-Content -AsByteStream $env:TEMP\sjskills-global-plan.json) `
    (Get-Content -AsByteStream $env:TEMP\sjskills-global-plan.recheck.json)) {
  throw "global plan changed after approval"
}
```

Any difference voids authorization. Review and authorize the new artifact
instead of widening the old approval.

## Authorized execution

Only after the evidence above receives explicit approval:

```bash
bin/sjskills apply --global
bin/sjskills --json plan --global > /tmp/sjskills-global-plan.after.json
```

On Windows PowerShell, use the already reviewed executable:

```powershell
& $env:TEMP\sjskills.exe apply --global
& $env:TEMP\sjskills.exe --json plan --global > $env:TEMP\sjskills-global-plan.after.json
```

Use the interactive confirmation so the operator sees the recomputed mutation
counts. `sjskills` retains that one verified materialization session through
the locked transaction and revalidates filesystem and provenance evidence
before each move or publication. Do not substitute the legacy audit wrapper,
Skills CLI direct installs, `--all`, or manual root copying.

Success requires the post-apply plan to contain no install, update, quarantine,
or blocked operations. Retain every reported quarantine identifier and the
before/after artifacts until the machine completes a normal work cycle.

## Failure and recovery

Stop on conflict, recovery-required status, changed counts, or partial-failure
evidence. Do not rerun blindly and do not delete or overwrite either active or
quarantined content.

A committed quarantine can be restored only when every destination is absent:

```bash
bin/sjskills restore --global <quarantine-id>
```

On Windows PowerShell, run
`& $env:TEMP\sjskills.exe restore --global <quarantine-id>`.

Moving an active replacement aside is a separate real-home mutation and needs
its own reviewed target list and explicit approval. Restore refuses overwrite,
re-proves the whole quarantine, restores provenance atomically, and leaves
ambiguous content in recovery-required state.

Former profile placements and legacy Pi copies are intentionally report-only.
Any later cleanup needs a new exact-source plan and authorization; it is not
part of baseline rollout.
