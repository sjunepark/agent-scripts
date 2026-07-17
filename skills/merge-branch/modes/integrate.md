# Integration Mode

## Workflow

1. Preflight the repository.
   - Confirm the current destination branch, source branch, and whether the user wants whole-branch integration or partial adoption.
   - Refuse or ask before proceeding when unrelated uncommitted changes are present, another merge or rebase is in progress, or the source branch is ambiguous.
   - Check state before modifying anything:

     ```bash
     git status --short --branch
     git merge-base HEAD <source>
     git log --oneline --decorate --left-right --cherry-pick HEAD...<source>
     ```

   - Record the merge base for the final report.

2. Build a source commit-intent ledger.
   - Inspect every source commit from the merge base to the source branch, in order:

     ```bash
     git log --reverse --oneline <base>..<source>
     git show --stat --patch <commit>
     ```

   - For each commit, record:
     - behavior added
     - behavior removed
     - files touched
     - tests or docs updated
     - diagnostics, temporary code, compatibility shims, or experiments later removed by follow-up commits
     - disposition: preserve, adapt to current structure, or intentionally reject with reason
   - Treat deletions as first-class source intent. Removed UI text, helpers, fields, tests, diagnostics, compatibility paths, and summaries must be integrated or explicitly rejected.

3. Choose the integration mode.
   - Default whole-branch integration to `git merge --no-commit --no-ff <source>`. Treat the result as a draft, verify `MERGE_HEAD`, and do not commit until conflicts, residual differences, deletion audits, and validation are complete.
   - After starting the merge, confirm `MERGE_HEAD` exists and matches the intended source tip. If it does not, stop and resolve the merge-shape problem before editing conflicts.
   - Use manual transplant only when partial adoption, intentionally linear or non-convergent history, or broad unrelated source history makes it appropriate. Explain why merge-as-draft was not used.
   - Ask before choosing between plausible product behaviors. Preserve destination architecture and naming where intentional, but preserve every source behavior or explicitly reject it with a reason.
   - Preserve destination behavior unless the source intentionally changes it. Refactor append-style conflict results into one coherent destination structure.
   - Do not integrate by cherry-picking, copying patches, squashing, or creating an ordinary independent commit unless the user explicitly requests non-convergent history or says the source branch should not be recorded as merged.

4. Map renamed or refactored paths.
   - Create a path map when the destination renamed, split, or reorganized source files.
   - Apply source behavior into the current structure instead of resurrecting obsolete files.

   ```text
   source: app/renderer/workflow/ui/canvas/NodeSessionDialog.svelte
   current: app/renderer/workflow/ui/canvas/CanvasNodeDetailsDialog.svelte
   reason: current branch extracted a shared dialog frame and renamed the details dialog
   ```

5. Integrate deliberately.
   - Implement each ledger item in its chosen destination.
   - Keep the destination architecture and naming where intentional.
   - Avoid compatibility layers, duplicate paths, and appended code unless required by real callers or explicitly requested for a staged transition.
   - If the source deliberately removed behavior, remove it from the destination structure too unless the user approves keeping it.

6. Audit residual differences against the source branch.
   - Compare the result to the source branch:

     ```bash
     git diff --stat <source>
     git diff --name-status <source>
     git diff <source> -- <relevant-or-mapped-files>
     ```

   - Classify every meaningful residual difference as one of:
     - destination structure intentionally preserved
     - source behavior integrated under a different path
     - missed source behavior to fix before completion
     - intentionally rejected source behavior, with reason
   - A residual difference is not itself a defect: do not restore obsolete files or demand exact patch equivalence when behavior legitimately landed under the destination structure.
   - Do not report completion while any meaningful residual difference is unclassified.

7. Run deletion grep audits.
   - For every behavior removed by the source branch, grep the destination tree for names, labels, tests, helpers, fields, and distinctive strings that would reveal a missed deletion.
   - Example:

     ```bash
     rg "item\.summary|messageSummary|chars|blocks ·" app tests
     ```

   - Record whether each pattern is absent, still present for an intentional reason, or a missed source deletion that must be fixed before completion.

8. Validate.
   - Run relevant checks from existing repo scripts. Prefer targeted tests first, then broader checks when practical.
   - If broad checks fail because of unrelated existing issues, run targeted validation and clearly separate unrelated failures from integration failures.

9. Report.
   - Include:
     - destination branch, source branch, and merge base
     - whether merge-as-draft or manual transplant was used, and why
     - convergence status for whole-branch integrations: before the merge commit, verify `MERGE_HEAD` and report convergence as pending; after the merge commit, verify with `git merge-base --is-ancestor <source> HEAD`
     - source commit-intent ledger summary
     - where each source intent landed
     - path mappings used
     - residual differences and the classification for each meaningful difference
     - deletion grep audit commands and results
     - files changed
     - validation commands and outcomes
     - remaining risks or user decisions
