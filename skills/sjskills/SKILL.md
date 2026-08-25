---
name: sjskills
description: "Operate sjskills to initialize project manifests, inspect or reconcile project skill state, recover quarantined changes, or conduct an explicitly authorized rollout of the fixed global baseline. Explicit invocation only."
---

# sjskills

Use `sjskills` to reconcile the exact skill state selected by a project's
committed `sjskills.toml` or by the tool's fixed global baseline. The tool owns
verified placement, provenance, quarantine, and restore; it does not publish
source changes or perform ad hoc catalog installs.

Keep four states distinct:

1. A source repository's local catalog is editable but not yet published.
2. The registry identifies the published remote content reconciliation can
   consume.
3. A project's committed `sjskills.toml` selects project profiles and direct
   declarations.
4. The registry defines one fixed, machine-independent global baseline.

Prefer `sjskills` on `PATH`. In the tool's source repository, use
`bin/sjskills` when the command is not installed. Do not install the command or
change `PATH` unless the user asks.

## Classify the request

- **Explain or inspect:** Use `profiles`, `plan`, or `plan --global`. These are
  read-only with respect to managed roots, although planning may fetch and
  materialize remote expected content temporarily.
- **Adopt a project:** Create `sjskills.toml` only when the user asks to adopt
  or initialize managed project skills. List available profiles first and use
  only profiles the user selected.
- **Reconcile a project:** Review `plan`, run `apply` only when the user asked
  to install, bootstrap, reconcile, or sync the project, then run `plan` again.
- **Recover project state:** Restore only the exact quarantine identifier the
  user named, and only after confirming every modeled destination is absent.
- **Inspect or roll out global state:** Default to `plan --global`. For global
  apply or restore, read [references/global-rollout.md](references/global-rollout.md)
  and require its separate evidence and authorization.

Use the ordinary Skills CLI workflow for direct `bunx skills` discovery or
ad hoc installs. Use the repository's plugin workflow for Codex plugins. Local
catalog validation and publication are not reconciliation.

## Project workflow

1. Resolve the command and project root. If no `sjskills.toml` exists, stop
   unless the user requested initialization.
2. For a new manifest, run `sjskills profiles`, obtain the user's profile
   choices, and run `sjskills init <profile> ...`. Do not infer profiles from
   installed copies. `init` must not overwrite an existing manifest.
3. Read the manifest and run `sjskills plan`. Summarize installs, updates,
   quarantines, unchanged placements, manual or workflow-managed entries,
   warnings, and blocks.
   Distinguish desired-state drift or conflict from a failed command, network
   request, or materialization; an operational failure is not a valid plan.
4. Stop before apply when the plan reports an unmanaged desired path, a locally
   modified managed tree, malformed or untrusted provenance, an unsafe
   filesystem boundary, failed materialization, or another conflict. Preserve
   unknown and legacy entries.
5. If the user authorized project reconciliation, run `sjskills apply` and
   keep interactive confirmation unless non-interactive execution was
   explicitly requested. `--yes` suppresses a prompt; it does not grant
   authority.
6. Run `sjskills plan` again. Do not claim exact state while changes or blocks
   remain. Retain every reported quarantine identifier through a normal work
   cycle.

When intended content changed in a registry-backed source, reconciliation can
only consume the published remote content. Do not apply a local-only change.
If publication is requested, follow the source repository's instructions for
validation, review, commit, and publication; those mechanics are outside this
skill. Before reconciling, resolve the registry source and verify that its
pinned remote ref contains each intended skill tree. A local or feature-branch
commit is insufficient when the registry points elsewhere.

## Restore

Treat restore as a separate mutation. Require the user to name the quarantine
identifier, inspect its current state, and confirm that all destinations are
absent. Run:

```text
sjskills restore <quarantine-id>
```

Do not move, delete, or overwrite active destinations to make restore succeed.
If content or provenance changed, preserve both sides and report the conflict.
Run `sjskills plan` after a successful restore.

## Global boundary

`sjskills plan --global` is the ordinary global operation. The baseline is
fixed and machine-independent; do not select a profile, infer one from the
hostname, or synthesize Pi-specific copies.

The `--approved-plan` and `--approved-plan-sha256` flags bind global apply to
reviewed evidence but do not authorize mutation. Never run global apply or
restore from general setup, validation, cleanup, or sync authority. Use the
global rollout reference and stop if any required evidence or exact approval is
missing.

## Completion

Report:

- project or fixed-global scope and whether work was read-only or mutating;
- the selected profiles or direct manifest declarations when relevant;
- plan and final-plan results, including preserved unknown, legacy, or blocked
  state;
- published-ref evidence when remote-backed local changes were involved;
- every quarantine identifier retained or restored; and
- any mutation still awaiting explicit authority.
