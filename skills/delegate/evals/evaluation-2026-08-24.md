# Delegate candidate evaluation — 2026-08-24

## Intended behavior

When explicitly invoked for substantive implementation, the parent retains
ownership of the complete requirement and decomposes it into staged, narrow
slices. Each Luna assignment has one concrete outcome, bounded behavior and
file ownership, and an independently reviewable validation target. Broad
features, refactors, migrations, and multi-surface requirements are never
assigned end to end. The parent waits, independently reviews and validates,
then chooses the next dependent slice; parallelism is limited to independent
slices with non-overlapping behavior and file ownership.

Each slice prompt reproduces the original request verbatim and preserves every
result-affecting requirement, constraint, acceptance criterion, decision,
failure or error output, repository fact, and integration dependency without
compressing or omitting detail. It distinguishes the narrow slice from the
broader end state and omits only unrelated background. The worker configuration
remains `gpt-5.6-luna`, maximum reasoning, and `fork_turns: none`; parent review,
validation, unrelated-work preservation, no concurrent overlapping edits, no
silent model substitution, and the localized direct-edit exception remain in
force.

## Frozen criteria

The baseline was `git show c980873:skills/delegate/SKILL.md`. The candidate was
the current working-tree `skills/delegate/SKILL.md`, frozen before trials. Each
condition was judged against these six critical assertions:

1. Never hand Luna a broad end-to-end requirement.
2. Assign one bounded, independently testable slice.
3. Reproduce the original request verbatim.
4. Retain every provided result-affecting detail.
5. Require parent review and validation before dependent slices.
6. Preserve the exact Luna/max/no-history configuration and no-commit and
   unrelated-work constraints.

## Behavior comparison

Three synthetic cases were run twice per condition in separate fresh,
no-history contexts: broad cache invalidation across a service and integration
tests; two independent generated clients; and a context-heavy payments
idempotency rollout with exact TTL, `409` behavior, additive online-safe
migration/no payments rewrite, staging unique-constraint error and path,
webhook committed-status ordering, named contract fixture, preservation,
validation, and no-commit constraints. This produced 12 outputs total (3
cases × 2 conditions × 2 trials). Trials produced decomposition/dispatch plans
and complete first Luna prompt(s), but did not spawn implementation workers or
edit files.

- **Candidate:** 6/6 runs passed all six critical assertions. Cache trials
  stopped dispatch because the referenced service/spec was absent, while still
  defining the later narrow service-plus-coupled-test slice and review gate.
  Parallel trials defined two small disjoint slices, reproduced the request
  verbatim, and withheld dispatch until exact paths, update details, and
  validation were inspected. Payments trials decomposed the rollout; the first
  assignment covered only the migration/backfill failure and coupled migration
  test, reproduced the request verbatim, and retained every stated
  constraint, evidence item, and dependency.
- **Baseline:** 0/6 runs passed the combined narrow-slice and full-fidelity
  standard. Both cache trials assigned the whole service-and-integration
  implementation. Both payments trials assigned the entire cross-cutting
  rollout to one worker and introduced additional inferred requirements. Both
  parallel trials split by client, but their prompts did not reproduce the
  original request verbatim and treated an entire unresolved client update as
  the worker contract.

## Trigger evaluation

Trigger evaluation was separate from behavior evaluation. Two fresh
classifiers saw only the description, adapter policy, and trigger-cases file.
Each classified 3/3 explicit positives as activate and 4/4 near-misses as
do-not-activate: 14/14 classifications matched in total. The only lexical
ambiguity was an analysis-only prompt using “delegate” as a verb; both
classifiers rejected it correctly. No description, adapter policy, or trigger
cases changed in this revision.

## Static validation

The following checks had already passed:

- `scripts/validate-skills`: 35 skills and registry validated.
- `bunx skills add ./skills/delegate --list`: exactly `delegate` discovered.
- Dependency-free Node registry/reconciler suite: 55/55 passed.
- JSON parse and `git diff --check`: passed.

## Limitations

This is synthetic orchestration evidence only. It tests decomposition and
prompt fidelity, not implementation correctness, runtime cost, or end-to-end
completion in a real service repository. The deliberate GPT-5.6 Luna/Codex-
specific runtime contract is an explicit user requirement and a pre-existing
limitation; this package should not be described as vendor- or model-portable.

## Transition status

No installation, commit, publication, or machine reconciliation occurred. The
registry scope remains `.agents`. The unrelated
`internal/sjskills/materialize.go` file remained untouched.
