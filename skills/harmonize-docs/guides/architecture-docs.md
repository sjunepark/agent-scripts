# Architecture Documentation

Use this guide when a harmonization creates, rewrites, splits, merges, or moves
an `ARCHITECTURE.md` file. The document should give contributors a fast,
durable model of system shape, starting points, flows, and constraints without
becoming an implementation catalog.

## Choose the topology

Keep a root `ARCHITECTURE.md` as the system map. It should explain the system's
purpose and boundary, identify the major subsystems and their relationships,
trace one or two cross-system flows, and route readers to important entry
points and any nested architecture documents.

Create a nested `ARCHITECTURE.md` only when several of these are true:

- the subtree has its own entry point, runtime, deployment boundary, lifecycle,
  or event flow;
- it owns important state, integrations, or invariants;
- contributors regularly need to understand it independently; or
- enough stable detail exists that keeping it in the parent would obscure the
  system map.

Let the parent explain why the subsystem exists and how it relates to peers.
Let the child explain its internal shape, dominant flow, and local invariants.
Link downward with a one-line scope description and upward with the child's
place in the larger system. Do not duplicate component descriptions or hide a
cross-system invariant only in a child document. When the split is uncertain,
keep one document until another file clearly reduces confusion.

## Recover stable evidence

Inspect entry points, package boundaries, long-lived directories, public
interfaces, background jobs, storage seams, and external integrations. Trace
one or two representative flows end to end. Capture responsibilities,
interactions, ownership, and invariants before drafting. Use implementation as
evidence while keeping it outside a documentation-only edit boundary.

## Select only useful sections

Choose headings that fit the system. Common durable sections are:

- **Purpose and boundaries:** what belongs inside and outside, and the primary
  callers or actors when that clarifies the boundary.
- **Subsystem map:** major pieces, responsibilities, and relationships at the
  highest useful level.
- **Runtime flow:** one representative request, job, startup, or event path,
  including control, state changes, and outgoing side effects.
- **Code map:** a short set of concrete `start here` paths with why each matters,
  not a directory inventory.
- **Invariants and constraints:** precise ownership, sequencing, concurrency,
  isolation, or lifecycle rules contributors must preserve.
- **Dependencies and related decisions:** only integrations and ADRs that
  materially shape the architecture.

Use a compact ASCII sketch only when it shortens the explanation. Omit Mermaid,
class-by-class catalogs, setup instructions, coding standards, tutorials, and
runbooks unless a design constraint depends on them. Link to ADRs for durable
rationale rather than replaying decision history.

## Reconcile for contributor use

Verify every remaining claim, link, path, and parent-child route. Remove facts
that would become stale after an ordinary refactor. Prefer repository terms and
searchable headings. The finished architecture documentation should let a new
contributor answer three questions quickly: what are the major parts, where
should I start tracing behavior, and what must not be broken?
