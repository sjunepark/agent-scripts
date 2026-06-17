---
name: progress-committer
description: "Update stale progress documentation relevant to completed work, then commit only the intended changes. Use when the user asks to update plans, TODOs, or roadmap notes before committing finished work."
---

# Progress Committer

Update relevant progress docs to match completed work, then create a focused commit. Keep unrelated working-tree changes untouched.

## Workflow

1. Inspect repository state first.
   - Run `git status --short` and identify the current branch.
   - Inspect staged and unstaged diffs.
   - Separate intended task changes from unrelated user work.

2. Identify relevant progress docs.
   - Look for planning docs tied to the changed area: `PLAN.md`, `TODO.md`, `ROADMAP.md`, docs plans, task plans, or colocated feature plans.
   - Update only docs relevant to the completed work.
   - Do not create a progress doc just for this skill.
   - If several docs are plausible and the right one is unclear, ask before editing.

3. Update stale progress information.
   - Mark or remove items that are actually complete.
   - Adjust next steps that changed because of implementation or review.
   - Remove obsolete notes instead of leaving stale checkboxes.
   - Keep future work that remains true.
   - Do not rewrite broad roadmap direction unless the diff clearly justifies it.

4. Review and validate.
   - Review the final diff, including progress-doc updates, for accidental unrelated edits.
   - Run the repository's relevant cheap validation command if configured and appropriate.
   - If validation already ran in the session and still applies, mention that instead of rerunning expensive checks.

5. Commit.
   - Stage only intended task files and relevant progress-doc updates.
   - Do not stage unrelated changes.
   - Use a clear normal project commit message. Do not mention this skill or any old prompt template.
   - Include validation results in the commit body when useful.
   - Do not push unless explicitly asked.

## Final Response

Report:

- commit hash and subject
- planning docs updated, or that none were relevant
- validation run or why skipped
- remaining uncommitted changes and why they were left out
