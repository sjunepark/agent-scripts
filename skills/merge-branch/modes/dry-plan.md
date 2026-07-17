# Dry Planning Mode

Use this mode to analyze or plan a branch merge without changing repository state.

- Do not run `git merge`, edit files, stage changes, commit, or otherwise mutate the working tree.
- Use read-only git inspection commands to compare destination `HEAD`, source `<source>`, and their merge base.
- Inspect likely conflicts, overlapping edits, deleted or renamed files, dependency/config/schema interactions, tests/docs impact, and areas needing user decisions.
- Propose one or more integration strategies with concrete tradeoffs.
- List validation that should run during the real merge, but do not run expensive or mutating validation.
- Stop after asking the user how to continue. Do not proceed into an actual merge until explicitly instructed.

Include the destination branch, source branch, merge base, likely changed areas, likely conflicts or decisions, recommended strategy, risks, and the exact decision needed before starting the merge.
