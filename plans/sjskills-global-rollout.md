# Global reconciliation rollout

## Outcome

Global skill reconciliation runs safely on any machine from a configured-sync
request, bound to agent-reviewed plan evidence, without a separate human
approval ceremony.

## Current state

Complete. Evidence binding is implemented and configured sync uses request
authority. On 2026-09-30, two configured global syncs on `DESKTOP-V33SCUA` followed
this contract with verified sjskills v1.3.0 and ended with exact final global
plans. Per-machine outcomes belong with their retained execution artifacts, not
with this plan.

## Next action

None — complete.

## Delivered contract

The CLI requires the complete reviewed global plan and its SHA-256, recomputes
stable warnings, operations, and current/expected content, and retains that
verified materialization through mutation. Missing evidence, changed content,
unsafe boundaries, and provenance conflicts fail closed. `--yes` only suppresses
the interactive prompt.

The user's configured-sync request supplies authority. The agent reviews the
plan, records executable and plan evidence, and executes within the selected
scope without a separate human hash/count approval. Inspection and repository
validation do not authorize real-home mutation; restore requires a named
quarantine request.

## Operational ownership

- The [sjskills entry point](../skills/sjskills/SKILL.md) owns scope selection and
  interpretation of sync, inspection, and restore requests.
- Its [global procedure](../skills/sjskills/references/global-rollout.md) owns
  executable verification, plan review, execution, drift handling, and recovery.
- The [registry contract](../docs/skill-registry.md) owns desired sets,
  provenance, managed-root boundaries, and evidence binding.
- [Release documentation](../docs/sjskills-releases.md) owns verified binary
  distribution. A rebuilding source wrapper is not a retained rollout binary.

Store each machine's before/after plans, executable and plan hashes, and
quarantine identifiers together through a normal work cycle. Keep built-in
skills, plugin caches, protected vendor/backup locations, and legacy Pi copies
outside reconciliation. Do not use this plan as an additional approval gate.
