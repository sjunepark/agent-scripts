# Section Patterns

Use only sections that explain stable structure.

## Durable sections

### Purpose and boundaries

- State what the system or module does.
- State what is inside the boundary and what is explicitly outside it.
- Name the primary actors or callers when that clarifies the boundary.

### Component map

- List the major pieces and their responsibilities.
- Show the most important relationships between them.
- Prefer a small ASCII sketch when relationships are easier to scan than prose.
- Stop at component or subsystem level unless lower detail is necessary to explain a constraint.

### Runtime flow

- Trace one representative request, job, startup path, or event flow.
- Show where control moves, where state changes, and where side effects leave the boundary.
- Focus on the dominant path, not every branch.

### Code map

- Point to key directories and "start here" files.
- Explain why each path matters.
- Avoid full tree dumps or long inventories of leaf modules.

### Invariants and constraints

- Record rules contributors must preserve.
- Prefer precise statements such as "handlers stay thin; business rules live in domain services" over abstract values.
- Call out data ownership, sequencing rules, concurrency assumptions, and isolation boundaries when relevant.

### Dependencies and integrations

- Name critical external systems and the internal abstractions that isolate them.
- Explain the dependency only when it materially shapes architecture.

### Related ADRs

- Link major decisions whose rationale would clutter the overview.
- Prefer a short label plus path.
