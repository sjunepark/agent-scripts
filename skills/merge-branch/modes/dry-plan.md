# Dry Planning Mode

Use this mode to analyze or plan a branch merge without changing repository state.

- Do not run `git merge`, edit files, stage changes, commit, or otherwise mutate the working tree.
- Use read-only git inspection commands to compare destination `HEAD`, source `<source>`, and their merge base.
- Inspect likely conflicts, overlapping edits, deleted or renamed files, dependency/config/schema interactions, tests/docs impact, and areas needing user decisions.
- Propose one or more integration strategies with concrete tradeoffs.
- List validation that should run during the real merge, but do not run expensive or mutating validation.
- For a plan-only request, finish with the plan. For an explicitly combined assessment-and-integration request, return to the entry point and continue the authorized integration once any consequential decisions are resolved.

Include the destination branch, source branch, merge base, likely changed areas, likely conflicts or decisions, recommended strategy, risks, and any exact decision still needed before starting the merge.
