---
name: harmonize-docs
description: "Reconcile repository documentation and plans, either repository-wide or within a requested change or topic. Explicit invocation only."
---

# Harmonize Docs

Make the documentation a coherent account of what is implemented, what is
intended, and what remains, with each durable fact and mutable status owned by
one document.

## Scope

With no scope, cover all active repository documentation. An explicit change,
path, topic, plan, subsystem, or comparison limits the run to that seed and the
documentation it affects: owning documents, affected plans and status records,
navigation, and cross-references. Do not inventory the whole repository for a
scoped run.

For `changes`, use the established change set. If none is named, use the
current staged, unstaged, and untracked changes. If there is no meaningful
change set, resolve the boundary rather than substituting unrelated history.

Documentation-specific navigation and configuration are in scope. Source code is
evidence and stays outside the edit boundary. Archives, generated output,
vendored files, and dependencies are excluded. Preserve immutable history while
correcting active status framing or the routes that lead to it. Record credible
unrelated conflicts or structural concerns as follow-up without expanding scope.

## Separate the claims

Keep these states independently legible:

- **Implementation reality:** behavior and operational boundaries verified in
  code, configuration, tests, commands, or releases.
- **Target design:** selected, proposed, or unresolved future behavior supported
  by the latest confirmed direction.
- **Delivery status:** completed, active, blocked, deferred, and next work, kept
  in the repository's existing planning convention.

Accepted design is not implemented behavior. Source implementation alone does
not prove validation, publication, support, authorization, or production
readiness. Mark these distinctions with document framing, tense, or a concise
status map rather than qualifying every sentence. When intent is unclear, expose
the conflict rather than inventing intent.

## Assess the structure

Accurate, noncontradictory documents can still route readers poorly, so
consistent claims do not settle the structure. Assess it as a whole:

- **Reader paths:** each audience reaches what it needs from a suitable entry
  point without reading material meant for another audience or reconstructing
  the route. In a user-facing entry point, maintainer procedures beyond a short
  pointer are friction even when every section is accurate and the repository
  is small.
- **Document boundaries:** mixed purposes or audiences, overlapping ownership,
  scattered explanations, or mismatched lifecycles.
- **Net benefit:** a change must improve discovery, comprehension, or
  maintenance enough to justify the changed paths, links, and navigation.
  Keeping a structure that works is a valid outcome. Create a document only for
  a distinct durable responsibility.

When you change instruction files such as `AGENTS.md` or `CLAUDE.md`, preserve
the filenames and locations their tools discover and the precedence between
them, and repoint any routing they contain to moved material.

For creating, rewriting, splitting, merging, or moving `ARCHITECTURE.md`, read
the [architecture documentation guide](guides/architecture-docs.md). Keep target
architecture visibly framed, and let plans own delivery order and remaining work
so architecture does not become a second roadmap.

Apply the result directly, including justified in-scope structural changes,
without a separate approval round. Verify each in-scope document's material
claims against code, configuration, or other authoritative evidence before
relying on or rewriting them. Prune transcript-like history along with
superseded statements, preserving rationale and intent that still guide
maintenance. Stage, commit, and push only when separately requested.

## Completion

Read the final documents together. Check affected links, paths, examples, and
commands, and search for stale terms and conflicting claims. Replace volatile
counts with stable facts or source pointers unless a count serves a durable
purpose.

Check transitions in both directions. Completed work should reach current-state
docs and progress records, and unfinished design should stay visibly
unimplemented. Overviews that readers enter directly must route to the same
authoritative state as detailed plans. Account for every in-scope document and
conflict, and run the relevant checks that are available.

Report meaningful corrections, validation, and unresolved limits. Also report
the structural disposition: each topology change and the reader problem it
solves, or briefly why the structure was kept when that matters. Finish when the
scoped documents agree and readers can find what exists, what is intended, and
what remains.
