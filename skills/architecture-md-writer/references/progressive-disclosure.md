# Progressive Disclosure

Use a layered architecture doc set when one file would otherwise have to choose between being too vague and too detailed.

## Root document

Use the repo-root `ARCHITECTURE.md` to:

- Explain the overall system purpose and boundaries.
- Map the major subsystems and how they interact.
- Trace one or two cross-system flows.
- Point readers to the most important top-level directories and entry points.
- Link to nested `ARCHITECTURE.md` files for subsystems that need more depth.

## Nested module documents

Create `ARCHITECTURE.md` inside a subtree when several of these are true:

- The subtree has its own entry points, runtime, or deployment boundary.
- The subtree has a distinct lifecycle or event flow.
- The subtree owns important data, integrations, or invariants.
- Contributors often need to understand the subtree in isolation.
- Explaining the subtree well would make the parent document too long or too noisy.

Nested docs work well for apps in a monorepo, major backend services, complex UI shells, compilers, data pipelines, storage layers, and plugin systems.

## Parent-child contract

Keep the split clean:

- Let the parent describe why the subsystem exists and how it relates to peers.
- Let the child describe the subsystem's internal shape, execution flow, and local invariants.
- Link downward from the parent and upward from the child.
- Avoid copying the same component descriptions into both places.

Start module docs with a short statement of where the module sits in the larger system.

## Practical structure

Common pattern:

- Root `ARCHITECTURE.md`
- `apps/foo/ARCHITECTURE.md`
- `services/bar/ARCHITECTURE.md`
- `packages/baz/ARCHITECTURE.md`

In the root doc, include a short section such as `Subsystem Architecture` or place links directly in the code map. Keep each link annotated with the subsystem's role so readers know where to go next.

## Failure modes

- Creating nested docs for every small library.
- Hiding key cross-system invariants only in module docs.

When unsure, keep one document until the extra file clearly reduces confusion.
