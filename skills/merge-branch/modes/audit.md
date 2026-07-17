# Merge Audit Mode

Use this mode to review a completed or in-progress merge result, merge-review a branch, or check whether source intent was preserved.

## Establish Context

1. Identify the destination branch and working tree state.
2. Determine whether the merge is uncommitted, staged, or already committed.
3. Identify the source branch or merge parent from the user request, `MERGE_HEAD`, `ORIG_HEAD`, merge commits, reflog, or branch history.
4. Compare the destination before the merge, source branch changes, final merge result, and conflict resolutions or manual edits.
5. If the destination, source, or base cannot be determined confidently, ask for clarification before judging intent.

## Audit Concerns

- History integrity: missing merge parents, missing or mismatched `MERGE_HEAD`, source tips not reachable after an intended whole-branch merge, or squash/cherry-pick/copied-patch results when a real merge was intended.
- Lost or partially integrated work: commits whose intent is absent, tests/docs/config/migrations/assets omitted, deletions ignored, stale references after renames, or tests removed/weakened without evidence.
- Unsafe assumptions: product behavior, API shape, naming, data model, permissions, defaults, feature flags, user flows, or compatibility policy chosen without clear evidence.
- Append-only integration: duplicate helpers, types, config fields, routes, commands, validation paths, competing sources of truth, or parallel abstractions that should be unified.
- Missed refactors: places where combining branches created an opportunity for one clearer owner, model, or execution path.
- Cross-layer side effects: imports, exports, public APIs, CLI commands, routes, hooks, schemas, generated types, fixtures, env/config defaults, auth, privacy, serialization, logging, tests, docs, examples, changelogs, and roadmap/TODO files.

Run relevant existing validation when practical. Prefer targeted checks for the merged area plus the broadest cheap safety check available. If validation cannot run, explain why and what should run next.

## Output

Use this shape:

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
