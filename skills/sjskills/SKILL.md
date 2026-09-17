---
name: sjskills
description: "Configure, inspect, or reconcile sjskills-managed project and global skills; restore named quarantines. Explicit invocation only."
---

# sjskills

Use `sjskills` to reconcile the exact skill state selected by a project's
committed `sjskills.toml` or by the tool's fixed global baseline. The tool owns
verified placement, provenance, quarantine, and restore; it does not publish
source changes or perform ad hoc catalog installs.

Strict sync covers only the selected global or project `.agents/skills` and
`.claude/skills` roots. Undeclared directories move into recoverable quarantine,
including unknown and locally modified copies. Built-in skills, plugin caches,
and legacy Pi copies are outside this boundary.

Previously managed desired copies with local edits are also preserved in
quarantine, then replaced with verified published content. Plans report these
as updates with reason `local-modification`. This requires trusted provenance
for the same source and a verifiable current tree; an unknown desired copy
still blocks reconciliation. Restoring edited bytes does not grant ownership.

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

Read [status inspection](references/status.md) for status commands, freshness,
advisories, and JSON interpretation. Read
[project selections](references/project-selection.md) when creating or changing
profiles or direct declarations, or resolving authenticated-source requirements.
Neither status nor configuration alone authorizes reconciliation.

## Classify the request

- **Sync configured state:** An unqualified request to sync or reconcile with
  this skill authorizes both the fixed global baseline and the current project's
  committed `sjskills.toml`, when present. Honor an explicit project-only,
  global-only, machine, or preservation limit. If no project manifest exists,
  reconcile only the global baseline and report that the project is unconfigured;
  do not initialize it or infer profiles from installed copies.
- **Explain or inspect:** Use `status` for both scopes, `profiles` for available
  selections, or `plan` / `plan --global` for a full selected-scope review. These
  preserve managed roots but may fetch expected content temporarily. A bare
  invocation without an action defaults to status inspection.
- **Configure a project:** Create or edit `sjskills.toml` when the user asks to
  adopt managed project skills or change the selection. Use selected profiles,
  direct declarations, or both using the project-selection reference. Preserve unrelated entries;
  configuration alone does not authorize apply or global changes.
- **Reconcile a project:** Review `plan`, run `apply` only when the user asked
  to install, bootstrap, reconcile, or sync the project, then run `plan` again.
- **Recover project state:** Restore only the exact quarantine identifier the
  user named, and only after confirming every modeled destination is absent.
- **Global reconciliation or restore:** Read
  [references/global-rollout.md](references/global-rollout.md) before global
  mutation. A sync request supplies reconciliation authority; prepare and review
  the required evidence yourself instead of asking the user to approve hashes.
  Restore still requires a request identifying the quarantine.

Use the ordinary Skills CLI workflow to discover source skill names or perform
requested ad hoc installs. For a project managed by `sjskills`, record external
skills in `[[direct]]` and reconcile through `sjskills`; undeclared ad hoc
installs in managed roots become quarantine candidates. Use the repository's
plugin workflow for Codex plugins. Local catalog validation and publication are
not reconciliation.

## Use configured authority

The request grants authority; the committed manifest and fixed baseline define
the desired set. Proceed with verified installs, updates (including quarantining
and replacing locally edited managed copies), provenance migration, and
recoverable quarantine of undeclared copies within that set's managed roots.
Summarize the reviewed operations before applying, without requiring another
user turn. Use `--yes` for an already-authorized apply unless the user requests
an interactive checkpoint. A flag or plan artifact does not create authority.

Honor an explicit request to keep local edits active or preserve particular
paths: stop the affected scope if its plan would replace or quarantine them.
Ordinary sync already authorizes preserving local edits in quarantine before
replacement; no override flag or additional confirmation is needed.

"Override and sync" does not resolve unmanaged desired copies, source mismatch,
corrupt provenance, unsafe boundaries, or failed materialization.
Stop the affected scope on those conflicts and explain
the concrete resolution needed. Do not force-adopt, manually replace roots,
change profiles, or include another machine to make a plan pass. Complete an
independent unblocked scope unless the user required an all-or-nothing result.

## Project workflow

1. Resolve the command and project root. If no `sjskills.toml` exists, stop
   unless the user requested adoption or initialization.
2. For a new manifest with profiles, run `sjskills profiles`, use the user's
   selected profiles, and run `sjskills init <profile> ...`; then add requested
   direct declarations. For direct-only adoption or an existing manifest, use
   the project-selection reference. Do not infer profiles from installed
   copies or overwrite an existing manifest with `init`.
3. Read the manifest and run `sjskills plan`. Summarize installs, updates,
   quarantines, unchanged placements, manual or workflow-managed entries,
   warnings, and blocks.
   Distinguish desired-state drift or conflict from a failed command, network
   request, or materialization; an operational failure is not a valid plan.
4. Review `local-modification` updates as preservation-and-replacement operations,
   including when the current tree already matches the published tree but
   differs from its recorded installation. Stop before apply for an unmanaged
   desired copy, source mismatch, malformed or untrusted provenance, an unsafe filesystem
   boundary, an unverifiable extra, failed materialization, or another conflict.
5. For authorized project reconciliation, run `sjskills apply --yes` after the
   plan passes review. Honor any explicit request to stop at a preview or keep
   interactive confirmation.
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
Restore does not grant ownership to unknown or locally modified copies. An
edited backup replaces the new installation's provenance with no ownership
record. Run `sjskills plan` afterward: restored undeclared skills are removal
drift; restored edited desired skills are unmanaged conflicts.

## Global boundary

Global reconciliation uses the fixed, machine-independent baseline. Profiles
select project skills; do not invent a global profile, infer one from the
hostname, or synthesize Pi-specific copies. An inspection request runs only
`sjskills plan --global`; configured sync continues through verified apply.

The `--approved-plan` and `--approved-plan-sha256` flags bind global apply to
the agent-reviewed evidence. Their names do not require a separate human
approval ceremony: create the artifact, inspect every operation and warning,
compute its digest, and use the same verified executable through apply and the
final plan. A sync request covers this work; an audit, repository validation,
or unrelated setup request does not. Restore and conflict resolution remain
separate from making the configured desired set exact.

## Completion

Report:

- project or fixed-global scope and whether work was read-only or mutating;
- the selected profiles or direct manifest declarations when relevant;
- plan and final-plan results, including remaining extras, blocks, and
  out-of-scope legacy state;
- published-ref evidence when remote-backed local changes were involved;
- every quarantine identifier retained or restored; and
- any mutation still awaiting explicit authority.
