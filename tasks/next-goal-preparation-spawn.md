# Prepare and spawn the next goal

## Outcome

Extend `next-goal` with `prepare` and `spawn [model] [reasoning]` while retaining
the existing prompt-only default. Preparation completes and reviews planning,
commits its authorized scope, and hands off a closed contract. Spawn creates a
new Codex task from the prepared state and verifies native goal startup.

## Current state

- Runtime, metadata, usage documentation, and companion goal-contract ownership
  are implemented. Structural validation and local skill-source discovery pass.
- Selected presets are Astra/medium and Luna/max, with an optional reasoning
  override. Omitted model settings use the configured task defaults.
- Independent bounded review found no behavioral issues. The candidate passed
  six paired behavior trials, three focused boundary checks, and six invocation
  checks. The disposable Git trial committed only the plan and preserved
  unrelated staging. See the
  [evaluation report](../skills/next-goal/evals/prepare-spawn-report.md) for frozen
  criteria, raw evidence, and baseline results. Task APIs and native goals were
  simulated; live Codex integration remains unverified by this evaluation.
- Validation covers source behavior; installed state is tracked by `sjskills`
  on each machine after verifying the published source.

## Next action

None — requested source implementation and evaluation complete.
