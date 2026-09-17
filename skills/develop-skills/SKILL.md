---
name: develop-skills
description: "Author, audit, merge, and evaluate Agent Skills and instruction stacks. Explicit invocation only."
---

# Develop Skills

Produce a focused skill or instruction change that improves the requested work.
Use task evidence and concrete constraints; retain only guidance that changes a
decision, supplies non-obvious knowledge, or makes completion verifiable.

## Choose the relevant guidance

Read the route needed for the requested operation, not every workflow:

- [Create or revise](workflows/create-or-revise.md) for skill authoring.
- [Audit instructions](workflows/audit-instructions.md) for instruction-stack or
  catalog audits, including model migrations.
- [Merge skills](workflows/merge.md) for consolidation and predecessor retirement.
- [Evaluate a skill](workflows/evaluate.md) for validation and behavioral or
  selection evidence. For edits, use the checks relevant to what changed.
- [Authoring rubric](rubrics/authoring.md) for reviewing the candidate's scope,
  instruction quality, and completion boundaries.
- [Portability contract](references/portability.md) when creating a package or
  changing its structure, metadata, dependencies, or cross-client assumptions.
- [Codex invocation policy](guides/codex-invocation.md) when creating or changing
  Codex adapter settings.
- [Refresh upstream guidance](workflows/refresh-guidance.md) only for an explicit
  request to reconcile this skill with authoring sources.

## Constraints and completion

Preserve existing invocation policy unless a change is requested. This catalog's
authoring preference is explicit invocation for new skills; implicit discovery
needs evidence of recurring use, reliable matching, and useful activation within
the intended installation scope. Installation reach is a separate choice.

Honor the user's operation and existing authorization. A read-only audit stays
read-only; a request to revise authorizes necessary in-scope source edits.
Publication, installation, and removal of active predecessors must be covered by
the requested transition. Resolve only missing authority, not permissions already
given, and continue independent work.

Keep runtime guidance independent of incidental host or model assumptions.
Skills targeting a named tool may declare its actual prerequisites. Preserve
licensing and attribution for retained third-party material.

Finish the authorized change through validation and fixes, without stopping at
a first draft. Report changed or deliberately retained behavior, checks and
observations, and concrete limits. Static checks, simulated decisions, and live
execution support different claims; identify which evidence you obtained.
