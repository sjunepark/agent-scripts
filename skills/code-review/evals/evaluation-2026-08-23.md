# Default four-lens review — 2026-08-23

## Contract

A bare `$code-review` invocation applies the implementation, system, design,
and diet lenses without additional lens-selection guidance. Each lens stays
bounded to the review target and affected consumers, treats inapplicable
dimensions as checked, and does not manufacture findings.

## Frozen assertions

- **Critical:** every review applies all four lenses.
- **Critical:** a routine local change does not expand into a whole-repository
  audit or force system, design, or diet findings.
- **Objective:** `SKILL.md` directly routes to all four lens resources and the
  OpenAI default prompt names the same four-lens default.
- **Subjective:** the workflow preserves useful depth for material technical
  changes while keeping trivial changes proportionate.
- **Safety:** edit policy and the Bucket I / Bucket II boundary remain
  unchanged.

## Paired behavior evidence

Fresh isolated agents evaluated the committed baseline and working-tree
candidate against the same synthetic scenarios. They received no authoring
conversation or expected conclusion beyond the frozen assertions.

- Baseline: the `code-review` tree at `0ec6551`.
- Candidate: the 2026-08-23 working-tree revision evaluated in this report.

### Routine local edit

The scenario changed one exported greeting string and invoked only
`$code-review`.

- The baseline required only the implementation gate. Its system, design, and
  diet lenses remained conditional, and the design instructions explicitly
  retained implementation-only review for a routine local edit.
- The candidate required all four lenses. Implementation checked the literal
  behavior and exported contract; system and design found no applicable
  cross-cutting or structural consequence; diet completed its brief check with
  no finding. The agent kept the review to the supplied patch.

Result: the baseline failed the universal-lens assertion; the candidate passed
all assertions.

### Material technical change

The scenario added a direct Redis dependency, a queue adapter used by two
callers, and retry configuration, again invoking only `$code-review`.

- The baseline loaded all four lenses because the patch happened to satisfy
  each lens's former trigger conditions.
- The candidate loaded all four because they are universal. It bounded system
  review to dependency, operational, compatibility, and reversal consequences;
  design to the adapter contract, two callers, and dependency direction; diet
  to whether the adapter and retry options earn their present cost; and
  implementation to integration behavior, failures, tests, and validation.

Result: both conditions preserved depth for a material change, while the
candidate removed dependence on the caller's wording or the executor's trigger
inference.

## Static validation

- `scripts/validate-skills`: passed for all published skills and
  `skill-registry.json`.
- `bunx skills add ./skills/code-review --list`: discovered exactly
  `code-review`.
- `git diff --check`: passed.
- Direct inspection confirmed portable frontmatter, client metadata isolated in
  `agents/openai.yaml`, and exact relative pointers to every runtime lens.

The trigger description did not change, so trigger classification was not
retuned or rescored. The evaluated workflow is explicitly invoked; publication,
installation, and implicit activation remain separate operations.

## Residual uncertainty

Each condition ran once per scenario and used synthetic patches. The evidence
demonstrates routing and proportional reasoning, not statistical reliability
across models or large real-world repositories.
