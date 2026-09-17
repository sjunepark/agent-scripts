# Diet Lens

Find concepts, states, branches, and maintenance obligations that current
behavior does not justify. Fewer lines alone is not the goal. For a small change
without a complexity signal, conclude after a brief check.

## Evaluate the cost

Separate domain, contract, lifecycle, security, and performance requirements from
incidental machinery. For a candidate simplification, establish what depends on
it now, what cost it removes, and the concrete replacement shape.

Useful signals include pass-through wrappers, generic code with one real path,
unused fields or options, shared helpers needing flags for diverging cases,
translation layers that preserve only genericity, and side channels appended
instead of integrating with the existing design. Check indirect uses before
calling something unused.

Compatibility layers, aliases, migration fields, fallbacks, and dual paths need
a current rollout, external contract, or documented migration window. Identify
public, persisted, and cross-subsystem consumers before recommending removal.
Uncertain constraints belong in Bucket II.

Prefer deletion of unused machinery, inlining shallow hops, making generic code
specific, or splitting diverging cases before proposing broad redesign. Avoid
premature shared abstractions; small duplication can preserve clearer ownership.
A longer direct implementation can create fewer obligations.

## Findings

Do not penalize explicit code, single-use helpers with meaningful names, wrappers
that isolate side effects, or objects that make invariants visible. Require an
observable cost and a viable simpler shape; a clean conclusion is valid.

Use Bucket I only for narrow, evidenced safe changes under the edit policy.
Tradeoffs, public contracts, migrations, and uncertain uses belong in Bucket II.
State what becomes simpler and what behavior or compatibility must be retained.
Keep significant justified complexity as-is, with its reason when useful.
