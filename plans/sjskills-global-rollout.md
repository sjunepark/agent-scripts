# Roll out the fixed global baseline

Status: proposed; not authorized; exact-content approval binding unresolved

## Boundary

This plan governs real-machine rollout of the `sjskills` v1 global baseline.
Reviewing it does not authorize `apply --global`, restore, migration cleanup,
moving active placements, or any other mutation under a real home.

The rollout owns only baseline placements in `~/.agents/skills` and
`~/.claude/skills`, global provenance at
`~/.agents/.global-skill-state.json`, and reconciler-private state under
`~/.agents/.sjskills-global/`. Former profile skills, Pi copies, unknown
entries, plugin caches, vendor metadata, and the legacy
`~/.skill-quarantine` tree remain preserved.

## Approval limitation

The current CLI does not accept a reviewed plan digest. `apply --global`
materializes remote sources again in a new session, so a source ref could move
after the byte-identical recheck and before apply. The executable and artifacts
below bind what was reviewed, but they do not cryptographically force apply to
use the reviewed expected-content hashes.

Do not describe this as exact-content authorization. Before rollout, either
add and review a CLI contract that binds apply to the approved expected-content
evidence, or obtain explicit authorization that names and accepts this weaker
remote-source race boundary. The first option is preferred for real-home
mutation.

## Produce review evidence

Use a dedicated clean checkout of the exact published commit proposed for the
machine. `bin/sjskills` builds the current checkout on every invocation, so it
is not suitable for binding approval by itself. Build one executable, hash it,
and reuse that exact file for plan, recheck, apply, and post-apply inspection.

On macOS or Linux:

```bash
git fetch origin main
rollout_commit=$(git rev-parse origin/main)
test "$(git rev-parse HEAD)" = "$rollout_commit"
test -z "$(git status --porcelain --untracked-files=all)"
rollout_dir=$(mktemp -d)
go build -trimpath -o "$rollout_dir/sjskills" ./cmd/sjskills
shasum -a 256 "$rollout_dir/sjskills"
"$rollout_dir/sjskills" --json plan --global > "$rollout_dir/plan.json"
shasum -a 256 "$rollout_dir/plan.json"
```

On Windows PowerShell, use a dedicated clean checkout and one temporary
executable:

```powershell
git fetch origin main
$rolloutCommit = git rev-parse origin/main
if ((git rev-parse HEAD) -ne $rolloutCommit) { throw "checkout is not the proposed published commit" }
if (git status --porcelain --untracked-files=all) { throw "checkout is not clean" }
$rolloutDir = Join-Path $env:TEMP ("sjskills-rollout-" + [guid]::NewGuid())
New-Item -ItemType Directory -Path $rolloutDir | Out-Null
$rolloutExe = Join-Path $rolloutDir "sjskills.exe"
go build -trimpath -o $rolloutExe ./cmd/sjskills
(Get-FileHash -Algorithm SHA256 $rolloutExe).Hash
$planPath = Join-Path $rolloutDir "plan.json"
& $rolloutExe --json plan --global |
  Out-File -FilePath $planPath -Encoding utf8
(Get-FileHash -Algorithm SHA256 $planPath).Hash
```

Review the complete JSON, including every operation, warning, expected-content
hash, and materialization result. Stop if it contains a blocked placement,
untrusted provenance, an unmanaged desired path, a modified managed tree, an
unsafe filesystem boundary, or an operation beyond the expected baseline.

Authorization must identify:

- machine;
- exact published repository commit;
- SHA-256 of the reviewed executable;
- SHA-256 of the reviewed plan artifact;
- allowed install, update, quarantine, and provenance-migration counts;
- operator and approval time.

## Recheck the approved state

Immediately before mutation, re-prove the checkout and executable, then use
that same executable to produce a byte-identical plan:

```bash
test "$(git rev-parse HEAD)" = "$rollout_commit"
test -z "$(git status --porcelain --untracked-files=all)"
shasum -a 256 "$rollout_dir/sjskills"
"$rollout_dir/sjskills" --json plan --global > "$rollout_dir/plan.recheck.json"
cmp "$rollout_dir/plan.json" "$rollout_dir/plan.recheck.json"
```

On Windows PowerShell:

```powershell
if ((git rev-parse HEAD) -ne $rolloutCommit) { throw "checkout changed after approval" }
if (git status --porcelain --untracked-files=all) { throw "checkout changed after approval" }
(Get-FileHash -Algorithm SHA256 $rolloutExe).Hash
$recheckPath = Join-Path $rolloutDir "plan.recheck.json"
& $rolloutExe --json plan --global |
  Out-File -FilePath $recheckPath -Encoding utf8
if ((Get-FileHash -Algorithm SHA256 $planPath).Hash -ne `
    (Get-FileHash -Algorithm SHA256 $recheckPath).Hash) {
  throw "global plan changed after approval"
}
```

The displayed commit and executable hash must equal the authorized values.
Any mismatch or plan difference voids authorization; review a new artifact
instead of widening the old approval.

## Authorized execution

Only after the evidence above receives explicit approval and the approval
limitation is resolved or expressly accepted, run the already reviewed
executable:

```bash
"$rollout_dir/sjskills" apply --global
"$rollout_dir/sjskills" --json plan --global > "$rollout_dir/plan.after.json"
```

On Windows PowerShell:

```powershell
& $rolloutExe apply --global
$afterPath = Join-Path $rolloutDir "plan.after.json"
& $rolloutExe --json plan --global |
  Out-File -FilePath $afterPath -Encoding utf8
```

Use interactive confirmation so the operator sees the recomputed mutation
counts. Do not substitute the legacy audit wrapper, a freshly rebuilt
executable, direct Skills CLI installs, `--all`, or manual root copying.

Success requires the post-apply plan to contain no install, update,
quarantine, or blocked operation. Retain every reported quarantine identifier,
the executable hash, and the before/after artifacts until the machine
completes a normal work cycle.

## Failure and recovery

Stop on conflict, recovery-required status, changed counts, or partial-failure
evidence. Do not rerun blindly and do not delete or overwrite active or
quarantined content.

A committed quarantine can be restored only when every destination is absent,
and restore needs its own explicit approval against the exact identifier and
current state:

```bash
"$rollout_dir/sjskills" restore --global <quarantine-id>
```

Moving an active replacement aside is a separate real-home mutation and needs
its own reviewed target list and authorization. Former-profile placements and
legacy Pi copies remain report-only; any later cleanup requires a new
exact-source plan and authorization.
