# Global rollout

Read this reference only for a requested `sjskills apply --global` or
`sjskills restore --global`. Planning alone does not need this procedure.

## Authorization boundary

A global rollout changes managed skill roots in a real home. Reviewing a plan,
asking for machine setup, or requesting a clean audit does not authorize apply,
restore, migration cleanup, or movement of active placements.

Global apply requires all of the following:

- the named machine;
- the exact published repository commit used to build the executable;
- the SHA-256 of that executable;
- the SHA-256 of the complete reviewed JSON plan artifact;
- the approved counts of installs, updates, quarantines, and provenance
  migrations; and
- an identified operator and approval time.

The approval must name these exact values. The CLI's required plan and digest
flags verify evidence; they do not supply authorization.

## Produce review evidence

Use a dedicated clean checkout of the exact published commit. Build one
executable and reuse that same file for plan, recheck, apply, and final
inspection. Do not use a wrapper that rebuilds the command on each invocation.

Create a complete JSON plan with:

```text
<reviewed-sjskills> --json plan --global > plan.json
```

Compute SHA-256 values for the executable and `plan.json`. Review every
operation, warning, current-state fact, expected-content hash, and
materialization result. Stop on blocked placements, untrusted provenance,
unmanaged desired paths, modified managed trees, unsafe filesystem boundaries,
or operations beyond the expected baseline.

## Recheck and apply

Immediately before mutation, prove again that the checkout, executable hash,
and plan bytes match the approved evidence. Any change voids the approval and
requires a newly reviewed artifact.

After the exact evidence receives explicit authorization, run the already
reviewed executable with the approved artifact and digest:

```text
<reviewed-sjskills> apply --global \
  --approved-plan plan.json \
  --approved-plan-sha256 <approved-sha256>
<reviewed-sjskills> --json plan --global > plan.after.json
```

Keep interactive confirmation so the operator sees the recomputed mutation
counts. Do not substitute a new build, the legacy audit wrapper, direct Skills
CLI installs, `--all`, or manual root copying.

Success requires the final plan to contain no install, update, quarantine, or
blocked operation. Retain the executable hash, before and after plans, and
every quarantine identifier until the machine completes a normal work cycle.

## Failure and recovery

Stop on conflict, changed counts, recovery-required status, or partial-failure
evidence. Do not rerun blindly or delete, overwrite, or move active or
quarantined content.

Global restore requires new explicit approval for the exact quarantine
identifier and current state. Confirm every destination is absent, then run:

```text
<reviewed-sjskills> restore --global <quarantine-id>
```

Moving an active replacement aside is a separate mutation requiring its own
reviewed target list and authorization. Preserve former-profile placements,
legacy Pi copies, and unknown entries unless a separately authorized cleanup
governs them.
