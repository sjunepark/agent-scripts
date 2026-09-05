# Roll out the fixed global baseline

Status: proposed; approval binding delivered; rollout not authorized

## Boundary

This plan governs real-machine rollout of the `sjskills` v1 global baseline.
Reviewing it does not authorize `apply --global`, restore, migration cleanup,
moving active placements, or any other mutation under a real home.

The rollout strictly reconciles `~/.agents/skills` and `~/.claude/skills`
to the baseline, including quarantining undeclared copies regardless of
ownership or local edits. It also owns global provenance at
`~/.agents/.global-skill-state.json`, and reconciler-private state under
`~/.agents/.sjskills-global/`. Built-in skills, Pi copies, plugin caches,
vendor metadata, and the legacy `~/.skill-quarantine` tree remain outside
the rollout boundary.

## Approval binding

`apply --global` requires both the reviewed JSON plan artifact and its approved
SHA-256. It reads the artifact once, verifies the exact bytes against that
digest, accepts only a strict successful global plan, and then recomputes the
complete global plan. Stable warnings, operations, current-state evidence, and
expected-content evidence must all match before confirmation or mutation.

The recheck retains its verified materialization session and apply uses that
same content. A remote source ref cannot move between the binding check and the
copy boundary. Missing flags, a changed artifact, moved expected content, or
inventory drift fails closed and requires review of a new artifact.

This technical binding does not authorize a machine. The separate approval
below remains mandatory and must name the exact artifact digest.

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
plan_sha256=$(shasum -a 256 "$rollout_dir/plan.json" | awk '{print $1}')
printf '%s\n' "$plan_sha256"
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
$utf8NoBom = New-Object System.Text.UTF8Encoding($false)
function Write-SjskillsGlobalPlan([string] $Path) {
  $startInfo = New-Object System.Diagnostics.ProcessStartInfo
  $startInfo.FileName = $rolloutExe
  $startInfo.Arguments = "--json plan --global"
  $startInfo.UseShellExecute = $false
  $startInfo.CreateNoWindow = $true
  $startInfo.RedirectStandardOutput = $true
  $startInfo.RedirectStandardError = $true
  $startInfo.StandardOutputEncoding = $utf8NoBom
  $startInfo.StandardErrorEncoding = $utf8NoBom
  $process = New-Object System.Diagnostics.Process
  $process.StartInfo = $startInfo
  try {
    if (-not $process.Start()) { throw "global plan could not start" }
    $stdout = $process.StandardOutput.ReadToEnd()
    $stderr = $process.StandardError.ReadToEnd()
    $process.WaitForExit()
    if ($process.ExitCode -ne 0) { throw "global plan failed: $stderr" }
    [System.IO.File]::WriteAllText($Path, $stdout, $utf8NoBom)
  } finally {
    $process.Dispose()
  }
}
$planPath = Join-Path $rolloutDir "plan.json"
Write-SjskillsGlobalPlan $planPath
$planSHA256 = (Get-FileHash -Algorithm SHA256 $planPath).Hash.ToLowerInvariant()
$planSHA256
```

Review the complete JSON, including every operation, warning, expected-content
hash, and materialization result, including undeclared-skill quarantines. Stop
on blocked placements, untrusted provenance, unmanaged or modified desired
copies, unsafe filesystem boundaries, unverifiable extras, or placement
operations outside the two managed skill roots.

Authorization must identify:

- machine;
- exact published repository commit;
- SHA-256 of the reviewed executable;
- SHA-256 of the reviewed plan artifact;
- allowed install, update, quarantine, and provenance-migration counts;
- operator and approval time.

## Recheck the approved state

Immediately before mutation, re-prove the checkout and executable. An explicit
byte-identical recheck is useful operator evidence; global apply performs its
own mandatory recheck again and retains that recheck's expected content:

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
Write-SjskillsGlobalPlan $recheckPath
if ((Get-FileHash -Algorithm SHA256 $planPath).Hash -ne `
    (Get-FileHash -Algorithm SHA256 $recheckPath).Hash) {
  throw "global plan changed after approval"
}
```

The displayed commit and executable hash must equal the authorized values. Any
mismatch or plan difference voids authorization; review a new artifact instead
of widening the old approval. Do not substitute `plan.recheck.json` for the
approved `plan.json`; apply verifies the approved artifact's exact digest.

## Authorized execution

Only after the evidence above receives explicit approval, run the already
reviewed executable with the approved artifact and digest:

```bash
"$rollout_dir/sjskills" apply --global \
  --approved-plan "$rollout_dir/plan.json" \
  --approved-plan-sha256 "$plan_sha256"
"$rollout_dir/sjskills" --json plan --global > "$rollout_dir/plan.after.json"
```

On Windows PowerShell:

```powershell
& $rolloutExe apply --global `
  --approved-plan $planPath `
  --approved-plan-sha256 $planSHA256
$afterPath = Join-Path $rolloutDir "plan.after.json"
Write-SjskillsGlobalPlan $afterPath
```

Use interactive confirmation so the operator sees the recomputed mutation
counts. The CLI rejects the apply before prompting if the approval binding
fails. Do not substitute the legacy audit wrapper, a freshly rebuilt
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
its own reviewed target list and authorization. Legacy Pi copies remain
outside this rollout.

## Current state

The delivered approval binding includes the exact-artifact digest check,
complete plan recheck, retained-materialization boundary, and mutation-bound
session fingerprint required by this plan. Go, race, registry, skill, catalog,
vet, and diff validation passed against isolated temporary homes or read-only
paths. Bounded code review findings were resolved, and PR #15 merged into `dev` as
`6664543`. The subsequent strict-sync correction is implemented locally; its
validation and publication status are tracked in [PROGRESS.md](../PROGRESS.md).
No machine approval or real-home rollout has occurred.

## Next action

Complete review and publication of the strict-sync correction, then obtain
separate evidence-bound authorization for a named machine before any real-home
execution. That rollout is outside the implementation task.
