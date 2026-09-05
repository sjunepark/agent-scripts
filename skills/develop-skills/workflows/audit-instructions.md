# Audit Instructions

Use for a model migration, instruction-stack audit, or catalog-wide revision.
The result is a scoped set of corrections supported by inspected rules and task
evidence. A skill that already meets the contract may remain unchanged.

## Establish scope and baseline

- Identify the requested operation: report, revise, evaluate, or install. Carry
  out all operations already authorized without requesting the same permission
  again. Treat publication and installation as separate when they are absent
  from the request.
- Inventory shared instructions, skill entry points, runtime resources, client
  metadata, and intended installation locations. Exclude fixtures and historical
  reports from live instructions; inspect them as evidence when relevant.
- Record the current file contents and pre-existing edits before revising. Keep
  an immutable baseline for comparisons and preserve unrelated work.
- When a model change motivates the audit, read the requested model's current
  official guidance. Distinguish explicit recommendations from local inferences;
  preserve package formats and invocation policies unless a change is justified.
- For installation or discovery problems, compare source, installed files, and
  active catalog metadata. A source edit does not establish that a running
  session uses it. Do not reinstall merely to complete a source-only audit.

## Inspect every skill in scope

Read its entry point and account for each runtime resource and adapter. Delegate
independent groups when available, with clear ownership and concise findings;
keep edits to shared instructions coordinated. For each skill, record either a
specific change with its evidence or a reason to retain it.

Check the decisions most affected by instruction interactions:

- **Authority:** Does a gate ask again despite an applicable user instruction or
  earlier approval? Preserve deliberate boundaries for destructive actions,
  external writes, sensitive access, and unrelated state.
- **Scope:** Can the skill distinguish research, planning, implementation, and
  publication? Do prerequisites block only the dependent action? Do examples or
  preferred tools accidentally replace the user's chosen outcome?
- **Questions:** Is the missing answer consequential and unavailable from context?
  Allow routine decisions within scope; retain questions needed for correctness
  or authorization. When a rule causes a pause, make its source and unresolved
  decision identifiable.
- **Delegation:** Are independent tasks bounded and useful? Preserve required
  isolation and review ownership without forcing parallelism for dependent work.
- **Validation:** Are required checks tied to changed behavior and completion?
  Remove unjustified repeated checks; preserve regression evidence and hard
  operational checks. Define when new evidence warrants another pass.
- **Output:** Does the format preserve required decisions, artifacts, evidence,
  and uncertainty without forcing empty sections or repetitive status reports?
- **Discovery:** Do the description, adapter policy, and trigger cases agree?
  Preserve existing activation intent; a catalog audit is not a policy migration.

## Correct the smallest owner

Put shared collaboration defaults in the applicable instruction file and
task-specific requirements beside the affected skill step. Do not append a
generic autonomy or model-tuning section to every skill. Resolve conflicting
clauses at their owners, including relevant resources and examples.

Keep true invariants exact. Replace an unconditional judgment rule with the
decision that matters, such as whether committing the intended files is already
authorized. Do not weaken a boundary merely because it can cause a pause.

## Validate and report coverage

Use the evaluation workflow from the entry point. Run catalog-wide static checks
and targeted independent behavior cases for changed decisions, including a
case that must proceed and a case that must still stop. Reuse a case across
skills only when it exercises the same changed contract; record that coverage.
Test trigger behavior separately when discovery instructions or metadata change.

Report reviewed skills, changed and retained decisions, checks, observed results,
and untested boundaries. Distinguish a simulated decision from live tool execution
and structural validation from measured behavior. Finish the authorized local
work with remaining rollout steps stated precisely.
