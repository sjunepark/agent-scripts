# Concise Progress Handoff Migration Note

This note archives the useful part of the old Pi prompt ticket
`prompts/tickets/2026-06-16-concise-progress-handoff.md`.

## Issue

The old progress handoff prompt tended to produce plan documents that were
longer than the active implementation state justified. Common failure modes:

- validation sections listed every plausible command or stale file-specific
  command lines
- predictable sections such as `Next command`, empty progress notes, and generic
  open questions added low-signal text
- handoff summaries, decisions, scope, and risks often repeated the same facts
- implementation plans became exhaustive task trees instead of the next
  reviewable slice
- historical validation and progress logs stayed in generated plans after they
  no longer helped continuation

## Migrated Behavior

The behavior now lives in `skills/handoff-plan/SKILL.md` and
`skills/plan-runner/SKILL.md`.

The migrated handoff behavior:

- uses a lean default plan shape: `Purpose`, `Current state`, and
  `Next implementation slice`
- makes `Validation`, `Files to inspect first`, `Open questions`, scope,
  risks, decisions, and progress logs optional
- targets compact ordinary handoffs around 80-140 lines
- omits a dedicated `Next command` section by default
- includes concrete next commands only when they are real unblockers
- keeps progress logs compact and collapses stale history into current state

## Non-Goals

- Do not rewrite old generated plan files solely because this migration happened.
- Do not remove the ability to write longer reference-heavy handoffs when the
  task genuinely needs them.
