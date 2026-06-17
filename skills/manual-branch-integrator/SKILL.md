---
name: manual-branch-integrator
description: "Manually integrate Git branch work without blind mechanical merges. Use when merging, dry-planning a merge, transplanting or refactoring branch work, resolving conflicts, preserving source-branch intent in a clean current-branch structure, or auditing a completed merge for lost intent, unsafe assumptions, and integration quality."
---

# Manual Branch Integrator

Integrate branch intent, not just patches. Prefer Git's merge machinery as a draft, then edit, refactor, audit, validate, and report the deliberate result.

Use clear branch terms throughout the task: the current `HEAD` branch is the destination branch, and `<source>` is the branch being integrated into it. If the user says "target branch" ambiguously, confirm whether they mean the source branch to merge or the destination branch to receive the work.

If the user asks for a dry run, merge plan, planning-only merge, or merge audit/review, use the specialized modes below instead of the default real integration workflow.

## Dry Planning Mode

Use dry planning mode when the user asks to analyze or plan a branch merge without changing repository state.

- Do not run `git merge`, edit files, stage changes, commit, or otherwise mutate the working tree.
- Use read-only git inspection commands to compare destination `HEAD`, source `<source>`, and their merge base.
- Record destination `HEAD`, source tip, and merge base.
- Inspect likely conflicts, overlapping edits, deleted or renamed files, dependency/config/schema interactions, tests/docs impact, and areas needing user decisions.
- Propose one or more integration strategies with concrete tradeoffs.
- List validation that should run during the real merge, but do not run expensive or mutating validation.
- Stop after asking the user how to continue. Do not proceed into an actual merge until explicitly instructed.

Dry planning output should include destination branch, source branch, merge base, likely changed areas, likely conflicts or decisions, recommended strategy, risks, and the exact decision needed before starting the merge.

## Merge Audit Mode

Use audit mode when the user asks to review a completed or in-progress merge result, merge-review a branch, or check whether source intent was preserved.

First establish merge context from evidence:

1. Identify the destination branch and working tree state.
2. Determine whether the merge is uncommitted, staged, or already committed.
3. Identify the source branch or merge parent from the user request, `MERGE_HEAD`, `ORIG_HEAD`, merge commits, reflog, or branch history.
4. Compare the destination before the merge, source branch changes, final merge result, and conflict resolutions or manual edits.
5. If the target, source, or base cannot be determined confidently, ask for clarification before judging intent.

Audit these concerns:

- History integrity: missing merge parents, missing or mismatched `MERGE_HEAD`, source tips not reachable after an intended whole-branch merge, or squash/cherry-pick/copied-patch results when a real merge was intended.
- Lost or partially integrated work: commits whose intent is absent, tests/docs/config/migrations/assets omitted, deletions ignored, stale references after renames, or tests removed/weakened without evidence.
- Unsafe assumptions: product behavior, API shape, naming, data model, permissions, defaults, feature flags, user flows, or compatibility policy chosen without clear evidence.
- Append-only integration: duplicate helpers, types, config fields, routes, commands, validation paths, competing sources of truth, or parallel abstractions that should be unified.
- Missed refactors: places where combining branches created an opportunity for one clearer owner, model, or execution path.
- Cross-layer side effects: imports, exports, public APIs, CLI commands, routes, hooks, schemas, generated types, fixtures, env/config defaults, auth, privacy, serialization, logging, tests, docs, examples, changelogs, and roadmap/TODO files.

Run relevant existing validation when practical. Prefer targeted checks for the merged area plus the broadest cheap safety check available. If validation cannot run, explain why and what should run next.

Use this output shape:

```markdown
## Merge-review verdict

Pass | Needs fixes | Blocked on decision

## Merge context

- Destination:
- Source / merge parent:
- Base or pre-merge ref:
- Review scope:

## History and merge-shape issues

## What was preserved

## Possible omissions or regressions

## Questionable assumptions

## Append-only or interaction problems

## Refactor opportunities

## Validation

## Applied low-risk fixes, if any

## Decisions needed, if any

## Recommended next step
```

Separate confirmed defects from risks. If the merge is sound, explain why it passed.

## Workflow

1. Preflight the repository.
   - Confirm the current destination branch, source branch, and whether the user wants whole-branch integration or partial adoption.
   - Check state before modifying anything:

     ```bash
     git status --short --branch
     git merge-base HEAD <source>
     git log --oneline --decorate --left-right --cherry-pick HEAD...<source>
     ```

   - Refuse or ask before proceeding if unrelated uncommitted changes are present, if a merge/rebase is already in progress, or if the source branch is ambiguous.
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
   - Default for whole-branch integration:

     ```bash
     git merge --no-commit --no-ff <source>
     ```

   - Treat the merge result as a draft, not an answer. Do not commit until conflicts are resolved, residual differences are classified, deletion greps are complete, and validation has run.
   - After starting a real merge, confirm `MERGE_HEAD` exists and matches the intended source tip. If it does not, stop and resolve the merge-shape problem before editing conflicts.
   - Resolve conflicts by understanding both branches. Preserve current-branch behavior unless the source branch intentionally changes it. Refactor append-style conflict results into one coherent current-branch structure.
   - Ask before choosing between two plausible product behaviors.
   - Use pure manual transplant only when explicitly appropriate: partial adoption, intentionally linear history, source branch work that should not be recorded as merged, or a merge draft that would import broad unrelated history. Document why the merge-as-draft default was not used.
   - Do not integrate by cherry-picking, copying patches, squashing, or creating an ordinary independent commit unless the user explicitly requests non-convergent history or says the source branch should not be recorded as merged.

4. Map renamed or refactored paths.
   - Create a path map when the current branch renamed, split, or reorganized source files.
   - Apply source behavior into the current structure instead of resurrecting obsolete files.

   ```text
   source: app/renderer/workflow/ui/canvas/NodeSessionDialog.svelte
   current: app/renderer/workflow/ui/canvas/CanvasNodeDetailsDialog.svelte
   reason: current branch extracted a shared dialog frame and renamed the details dialog
   ```

5. Integrate deliberately.
   - Implement each ledger item in its chosen destination.
   - Keep the current branch's architecture and naming where it is intentional.
   - Avoid compatibility layers, duplicate paths, and appended code unless they are required by real callers or the user explicitly wants a staged transition.
   - If the source branch deliberately removed behavior, remove it from the current structure too unless the user approves keeping it.

6. Audit residual differences against the source branch.
   - After integration, compare the result to the source branch:

     ```bash
     git diff --stat <source>
     git diff --name-status <source>
     git diff <source> -- <relevant-or-mapped-files>
     ```

   - Classify every meaningful residual difference as one of:
     - current-branch structure intentionally preserved
     - source behavior integrated under a different path
     - missed source behavior to fix before completion
     - intentionally rejected source behavior, with reason
   - Do not report completion while any meaningful residual difference is unclassified.

7. Run deletion grep audits.
   - For every behavior removed by the source branch, grep the current tree for names, labels, tests, helpers, fields, and distinctive strings that would reveal a missed deletion.
   - Example:

     ```bash
     rg "item\.summary|messageSummary|chars|blocks ·" app tests
     ```

   - Record whether each pattern is absent, still present for an intentional reason, or a missed source deletion that must be fixed before completion.

8. Validate.
   - Run relevant checks from existing repo scripts. Prefer targeted tests first, then broader checks when practical.
   - In JavaScript/TypeScript repos, likely commands include:

     ```bash
     bun run check
     bun run test
     bun run lint
     bun run build
     ```

   - If broad checks fail because of unrelated existing issues, run targeted validation and clearly separate unrelated failures from integration failures.

9. Final report.
   - Include:
     - destination branch, source branch, and merge base
     - whether merge-as-draft or manual transplant was used, and why
     - convergence status for real whole-branch integrations: before the merge commit, verify `MERGE_HEAD` and report convergence as pending; after the merge commit, verify with `git merge-base --is-ancestor <source> HEAD`
     - source commit-intent ledger summary
     - where each source intent landed
     - path mappings used
     - residual differences from the source branch and the classification for each meaningful difference
     - deletion grep audit commands and results
     - files changed
     - validation commands and outcomes
     - remaining risks or user decisions

## Guardrails

- Do not commit a mechanical merge result just because conflicts are resolved.
- Do not lose source-branch deletions during a hand transplant.
- Do not restore obsolete files merely to make diffs look closer to the source branch when current-branch structure intentionally changed.
- Do not claim exact patch equivalence is required when behavior legitimately landed under different paths.
- Do not hide unresolved product choices; ask the user.
- Do not move, force-update, or delete the source branch automatically. Convergence means the source branch tip is reachable from the destination `HEAD`; aligning branch refs requires explicit approval.
