---
name: harmonize-docs
description: "Harmonize active repository documentation and plans into one current, coherent system. Use only when the user explicitly invokes $harmonize-docs."
---

# Harmonize Docs

Treat the repository's documentation as one system. Replace sediment with a
coherent current state: each durable fact has one canonical home, each active
plan reflects reality and remaining intent, and the documents route readers
without contradiction.

## Coordination

Leverage subagents when they are available, choosing the decomposition
dynamically from the repository's shape and the work discovered. The
coordinating agent owns exhaustive coverage, conflict resolution, and the
integrated result. When subagents are unavailable, perform the same work
directly.

## Workflow

### 1. Establish the documentation boundary

- Read the effective repository instructions and inspect working-tree state.
- Discover documentation through repository conventions, navigation files,
  links, and common documentation and planning names or extensions.
- Include all active, human-authored documentation: overview and contributor
  docs, plans and progress trackers, decisions, runbooks, instructions, and
  subsystem or tool documentation.
- Include documentation-specific navigation and configuration files needed to
  keep the documentation reachable and valid, even when they are not prose.
- Treat archives, immutable historical records, generated output, vendored
  material, dependencies, and build artifacts as outside the default boundary.
  Honor any narrower or broader scope in the invocation.
- Preserve unrelated working-tree changes as intentional.

Complete this step when every plausible documentation file is classified as
in scope or outside it, with no active document left undiscovered.

### 2. Reconstruct truth by claim type

- Read every in-scope document completely. Use source files as evidence, not as
  additional editing scope.
- Verify claims about current behavior and status against code, configuration,
  tests, commands, and repository state.
- Recover future intent from the latest confirmed user direction and active
  plans. Distinguish unfinished intent from a false current-state claim.
- Recover decisions and rationale from decision records and confirmed context.
  Mark genuinely unresolved intent instead of inventing it.
- Classify material claims as verified current state, intended future state,
  durable rationale, obsolete or duplicate material, or unresolved conflict.

Complete this step when every in-scope document has been read, every material
claim has an evidence-backed classification, and every conflict is visible.

### 3. Design the target documentation system

- Give each document a clear purpose, audience, scope, and lifecycle.
- Give each durable fact one canonical home. Let other documents link to it or
  carry only the local context their readers need.
- Treat tool-recognized instruction files as configuration: verify their
  filenames, scope, discovery, and precedence before changing their topology,
  and preserve the intended effective hierarchy.
- Assign every in-scope document a disposition: keep, rewrite, merge, split,
  move, delete, or replace. Create a document only when a distinct durable
  responsibility has no suitable home.
- Merge overlapping documents. Split a sprawling document when its parts have
  distinct audiences, ownership, scopes, or lifecycles. Keep parent documents
  concise and route detail to focused child documents.
- Shape active plans around current state, remaining work, blockers, decisions,
  and the next useful action. Compress completed history once its durable
  lessons and decisions have a proper home.

Complete this step when every in-scope document has one disposition and every
durable claim has one destination or an explicit reason for removal, with no
planned orphan or duplicate source of truth.

### 4. Apply one harmonious rewrite

- Apply the target topology directly, including justified file creation,
  merging, splitting, moving, and deletion.
- Rewrite each document as a coherent whole. Integrate new information at its
  natural location and reshape surrounding material so the result does not read
  like a sequence of appended updates.
- Remove stale statements, superseded plans, duplicate explanations, empty
  scaffolding, transcript-like discussion, and incidental detail that lacks
  durable reader value.
- Preserve useful intent, rationale, constraints, and unresolved questions even
  when their original wording or file no longer belongs.
- Update tables of contents, indexes, cross-references, links, paths, commands,
  terminology, and parent-child routing for the new topology.
- Keep non-documentation implementation files read-only. Documentation-specific
  navigation and configuration remain within the boundary from step 1. Leave
  staging, commits, and pushes to a separate explicit request.

Complete this step when every planned disposition is applied, each surviving
claim is in its canonical home, and all affected navigation follows the new
structure.

### 5. Reconcile the finished system

- Re-read the final documents together rather than reviewing only the diff.
- Search for old paths, renamed terms, stale commands, duplicated claims,
  conflicting statuses, broken references, and append-style sediment.
- Review durable documentation for hard-coded counts and other volatile facts,
  such as numbers of files or tests. Replace them with stable invariants or
  pointers to the source of truth.
- Recheck current-state claims against repository evidence and active-plan
  claims against confirmed intent.
- Run available documentation, link, example, and repository validation that is
  relevant to the changed files. Inspect the final diff and working-tree state.
- Account for every initially discovered file and every earlier conflict.
  Report anything that cannot be verified and keep uncertainty explicit in the
  appropriate document when readers need it.

Complete this step only when all in-scope files and material claims are
accounted for, all verifiable references and current-state claims have been
checked, no conflict is concealed, and validation passes or its limitation is
recorded.

### 6. Report the harmonization

Summarize topology changes, meaningful corrections, canonical ownership
decisions, unresolved uncertainties, and validation results. Keep the report
short and point to the changed documents instead of repeating their content.

Complete the run when the user can see what changed, where authoritative
information now lives, and what—if anything—still needs a decision.
