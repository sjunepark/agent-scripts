# Architecture Documentation

Use this guide when a harmonization creates, rewrites, splits, merges, or moves
an `ARCHITECTURE.md` file. The finished document should let a new contributor
quickly answer three questions: what the major parts are, where to start tracing
behavior, and what must not be broken. It should not become an implementation
catalog.

## Topology

Keep a root `ARCHITECTURE.md` as the system map. Add a nested one only when the
subtree has several of these traits:

- its own entry point, runtime, deployment boundary, lifecycle, or event flow;
- ownership of important state, integrations, or invariants;
- contributors who regularly need to understand it on its own; or
- enough stable detail that keeping it in the parent would obscure the map.

The parent explains why the subsystem exists and how it relates to its peers.
The child explains its internal shape, dominant flow, and local invariants. Link
down with a one-line scope description and up with the child's place in the
larger system. Do not duplicate component descriptions, and do not hide a
cross-system invariant in a child document alone. If the split is uncertain,
keep one document.

## Required content

At the appropriate system or subsystem scope, each architecture document must
contain these, backed by evidence from the code:

- **Representative runtime flow:** one or two request, job, startup, or event
  paths, traced end to end, with control, state changes, and side effects.
- **Code map:** a short list of concrete start-here paths, each with why it
  matters. Not a directory inventory.
- **Invariants and constraints:** the ownership, sequencing, concurrency,
  isolation, or lifecycle rules that contributors must preserve.

Add purpose and boundaries, a subsystem map, and the dependencies or ADRs that
shape the architecture when they help. Use a compact ASCII sketch only when it
shortens the explanation. Leave out Mermaid, class-by-class catalogs, setup
instructions, coding standards, tutorials, and runbooks unless a design
constraint depends on them. Link to ADRs for rationale instead of replaying the
decision history. Remove facts that an ordinary refactor would make stale.
