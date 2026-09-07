# Global reconciliation

Read this reference for global apply or restore. Planning alone does not need
this procedure. Use the entry point's request classification to resolve scope.

## Authority and evidence

A request to sync configured state or make the fixed global baseline exact
supplies authority for reconciliation on the requested machine. The agent owns
plan review and evidence preparation; do not require the user to approve a
machine/commit/hash/count packet or repeat an already-applicable request.
Inspection and repository validation remain read-only. A restore request must
identify its quarantine and does not authorize overwriting active destinations.

Retain the machine identity, executable version or published build commit,
executable SHA-256, complete before/after JSON plans, reviewed plan SHA-256, and
reported quarantine identifiers through a normal work cycle. These are
execution evidence, not extra user decisions.

## Prepare and review

Use one verified installed executable for plan, apply, and final inspection.
If the command is a wrapper that rebuilds on every invocation, build one binary
from a dedicated clean checkout of the intended published commit and reuse that
file. Do not install a new command or change PATH to perform reconciliation.

Create a complete JSON plan:

```text
<reviewed-sjskills> --json plan --global > plan.json
```

Compute SHA-256 values for the executable and `plan.json`. Review every
operation, warning, current-state fact, expected-content hash, and
materialization result, including quarantine of undeclared skills regardless of
ownership or local edits. Stop on blocked placements, untrusted provenance,
unmanaged or modified desired copies, unsafe filesystem boundaries,
unverifiable extras, or placement operations outside the two managed skill
roots. Honor preservation limits in the user's request even if strict sync
would otherwise quarantine those entries.

## Apply and verify

Before mutation, verify the executable and artifact still match the reviewed
hashes. Apply uses the complete reviewed artifact and recomputes the plan before
mutation; `--yes` only skips the redundant interactive prompt:

```text
<reviewed-sjskills> apply --global --yes \
  --approved-plan plan.json \
  --approved-plan-sha256 <reviewed-sha256>
<reviewed-sjskills> --json plan --global > plan.after.json
```

Omit `--yes` when the user requested an interactive checkpoint. Do not substitute
a rebuilding wrapper, direct Skills CLI installs, `--all`, or manual root copies.

Success requires the final plan to contain no install, update, quarantine, or
blocked operation. Retain the evidence and every reported quarantine identifier.
Built-in skills, plugin caches, vendor metadata, backups outside managed roots,
and legacy Pi copies remain outside this reconciliation.

## Drift and recovery

An evidence mismatch prevents mutation; it does not erase the user's task
request. Inspect fresh evidence and prepare a new reviewed artifact when the
changes remain within the configured scope and authority. Stop on repeated
drift, conflict, recovery-required state, or partial failure instead of retrying
blindly. Never weaken provenance or filesystem checks to force completion.

For an explicitly requested restore, inspect the named quarantine and verify
that every modeled destination is absent, then run:

```text
<reviewed-sjskills> restore --global <quarantine-id> --yes
```

Use the user's interactive preference when supplied. Moving an active
replacement aside is a separate mutation needing its own applicable authority;
do not infer it from sync or restore. Run a final plan and report any resulting
drift, including restored entries outside the baseline.
