---
name: harmonize-docs
description: "Reconcile repository documentation and plans, either repository-wide or within a requested change or topic. Explicit invocation only."
---

# Harmonize Docs

Make the documentation a coherent account of implementation, intent, and remaining
work. Give each durable fact and mutable status one canonical owner; link from
other reader paths. Prefer pruning stale or duplicate material, while preserving
detail needed for understanding, rationale, operations, or decisions.

## Scope

With no scope, cover all active repository documentation. An explicit change,
path, topic, plan, subsystem, or comparison bounds the run to that seed and its
necessary documentation consequences: owning documents, affected plans and status
records, navigation, and cross-references. Do not inventory the whole repository
for a scoped run.

For `changes`, use the established change set, or current staged, unstaged, and
untracked changes if none is named. If no meaningful change set exists, resolve
the boundary rather than substituting unrelated history.

Include documentation-specific navigation and configuration. Source code supplies
evidence but stays outside the edit boundary. Archives, generated output, vendored
files, and dependencies are excluded; preserve immutable history while correcting
active status framing or routes to it. Record credible unrelated conflicts or
structural concerns for follow-up without expanding scope.

## Separate the claims

Keep these states independently legible:

- **Implementation reality:** behavior and operational boundaries verified in
  code, configuration, tests, commands, or releases.
- **Target design:** selected, proposed, or unresolved future behavior supported
  by the latest confirmed direction.
- **Delivery status:** completed, active, blocked, deferred, and next work in the
  repository's existing planning convention.

Accepted design is not implemented behavior. Source implementation alone does not
prove validation, publication, support, authorization, or production readiness.
Use clear document framing, tense, or a concise status map for these distinctions;
avoid repeating qualifications in every sentence.

## Reconcile the documents

Read each in-scope document and verify its material claims against the appropriate
evidence. Recover decisions and useful rationale; expose unresolved conflicts
rather than inventing intent. Independent areas may be delegated, with the
coordinator owning coverage and the integrated result.

Choose a purpose, audience, scope, and lifecycle for each document, then assess
the in-scope structure as a whole. Accurate, noncontradictory documents can still
route readers poorly, so consistent claims do not settle this:

- **Reader paths:** each audience reaches what it needs from a suitable entry
  point without reading material meant for another audience or reconstructing
  the route. A user-facing entry point that also carries maintainer procedures
  is friction even when every section is accurate and the repository is small.
- **Document boundaries:** mixed purposes or audiences, overlapping ownership,
  scattered explanations, or mismatched lifecycles.
- **Net benefit:** a change must materially improve discovery, comprehension, or
  maintenance enough to justify changed paths, links, and navigation.

Reach a proportionate change-or-retain decision grounded in concrete reader
friction; retaining a structure that works is a valid outcome. Rewrite, merge,
split, move, or remove material when that decision warrants it, and create a
document only for a distinct durable responsibility. Apply justified in-scope
changes without a separate approval round; ask only about a consequential choice
the evidence cannot resolve. Preserve tool-recognized instruction discovery and
precedence when changing those files.

For creating, rewriting, splitting, merging, or moving `ARCHITECTURE.md`, read
the [architecture documentation guide](guides/architecture-docs.md). Keep target
architecture visibly framed, and let plans own delivery order and remaining work
instead of making architecture a second roadmap.

Apply the coherent result directly. Integrate updates where they belong, prune
superseded statements and transcript-like history, and repair affected links,
commands, terminology, and navigation. Preserve useful intent even when its
original wording or location no longer fits. Stage, commit, and push only when
separately requested.

## Completion

Read the final documents together. Check affected links, paths, examples, and
commands; search for stale terms and conflicting claims. Replace volatile counts
with stable facts or source pointers where the counts serve no durable purpose.

Check transitions in both directions: completed work reaches current-state docs
and progress records, while unfinished design stays visibly unimplemented.
Direct-entry overviews must route to the same authoritative state as detailed
plans. Account for every in-scope document and conflict, run relevant available
checks, and fix failures caused by the rewrite.

Report meaningful corrections, the structural disposition (each topology change
and the reader problem it solves, or briefly why the structure was retained when
that is material), validation, and unresolved limits. Finish when the scoped documents agree and readers can find
what exists, what is intended, and what remains.
