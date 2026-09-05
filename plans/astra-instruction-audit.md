# Instruction audit

## Scope and decisions

- Update shared instruction sources and the published `skills/` catalog using
  the revised `develop-skills` workflow. Preserve intentional authority boundaries,
  current activation policies, and pre-existing reconciler changes.
- Apply the official [Astra prompting guidance](https://developers.openai.com/api/docs/guides/latest-model?model=gpt-6-astra)
  reviewed on 2026-09-05 as portable decision rules; no model-specific package
  format or universal prompt block is required.
- The initial audit authorized source revisions. The subsequent user request
  authorizes committing and pushing all current changes, followed by `sjskills`
  sync. Global rollout retains its exact-evidence approval boundary. Global
  instruction symlink targets update in place; installed skill copies may remain
  different until reconciliation completes.
- Immutable pre-edit baseline for this run:
  `/var/folders/16/lvgrynkd1n93dqjg140gs5wm0000gn/T/astra-instruction-baseline-kfv8j94x`.

## Current state

Local source work is complete. Updated all three shared instruction sources,
repository maintenance guidance, and `develop-skills`; used the revised audit
workflow across all 31 published skills. Eighteen skills received targeted
corrections and thirteen were retained with evidence. Existing activation
policies and unrelated reconciler work are preserved.

One bounded independent code-review pass and scoped documentation harmonization
are complete. Shared defaults own follow-through, delegation, and validation
expectations; skills own their specific authority and completion conditions.

## Validation

Catalog validation and local-source discovery passed for all 31 skills;
whitespace checks passed. Independent authoring and decision simulations,
three trigger probes, and disposable Git merge fixtures exercised the changed
decisions and retained boundaries. Coverage, raw evidence, review corrections,
and limitations are in the
[evaluation record](../skills/develop-skills/evals/evaluation-2026-09-05.md).

All three personal instruction symlinks resolve to the intended updated source
files. Installed skill copies and active-session loading remain unverified;
no installation or publication occurred.

## Remaining rollout

Publish the reviewed changes, verify that the registry's configured remote ref
contains the intended trees, then inspect the applicable `sjskills` plans.
Apply only within resolved scope and existing authority; global rollout needs
its exact-evidence approval. Preserve the distinction between published source,
installed copies, and a session's loaded catalog.
