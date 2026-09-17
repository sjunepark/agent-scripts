# Integration Mode

Integrate the requested source behavior into the destination's current structure,
including intentional removals, and verify the intended Git history shape.

## Establish the integration

Identify the destination, source tip, merge base, whole-branch or partial scope,
and authorized committed or uncommitted endpoint. Useful initial evidence:

```bash
git status --short --branch
git merge-base HEAD <source>
git log --oneline --decorate --left-right --cherry-pick HEAD...<source>
```

Inspect an operation already in progress. Continue a requested merge resolution
when `MERGE_HEAD` matches the intended source; do not start another merge or turn
an existing rebase into one. Resolve an unrelated or ambiguous operation before
mutation. Isolate unrelated uncommitted work without overwriting, staging, or
committing it; ask only when safe separation cannot be established.

Inspect source history and the net diff to account for behavior added, changed,
and removed, including tests, docs, configuration, and follow-up reversals.
Use `git log --reverse --oneline <base>..<source>` and individual commit patches
where the net diff does not establish intent. Keep a concise intent record for
material adaptations, omissions, and decisions; a row for every mechanical
commit is unnecessary. Map renamed or split paths when needed so integration
does not resurrect obsolete files.

## Preserve history and behavior

For a new whole-branch integration, default to
`git merge --no-commit --no-ff <source>`. Treat the merge as a draft until
conflicts, residual differences, and validation are resolved. If the source is
already an ancestor of `HEAD`, verify the requested behavior and report that no
new merge is needed. Otherwise, for a new or resumed merge, require `MERGE_HEAD`
to match the intended source tip before editing conflicts or committing.

Use a manual transplant only for authorized partial adoption or intentionally
non-convergent history; explain the choice. Broad unrelated source history may
justify proposing a transplant, but does not itself authorize replacing an
intended whole-branch merge with cherry-picks, copied patches, squash, or an
ordinary independent commit.

Preserve source behavior and intentional deletions in the destination's current
architecture. Preserve destination behavior unless the source intentionally
changes it. Resolve duplicate paths and owners introduced by the integration.
Use evidence and delegated judgment for routine adaptations; ask about unresolved
product or compatibility choices. Record reasons for intentionally rejected
source behavior; keeping behavior the source deliberately removed requires the
user's approval.

## Verify and finish

Compare the result with the source, focusing on meaningful residual differences:

```bash
git diff --stat <source>
git diff --name-status <source>
git diff <source> -- <relevant-or-mapped-files>
```

Account for each material difference as preserved destination structure, source
behavior adapted elsewhere, an intentional rejection, or an omission to fix.
Check removed names, labels, fields, helpers, tests, and distinctive strings with
targeted searches where they can reveal missed deletions. Exact patch equality
is not required when behavior has moved into the destination's structure.

Run required checks and validation covering integrated behavior and affected
contracts. Separate pre-existing failures from integration failures and resolve
actionable review findings. Do not declare integration complete with unaccounted
source behavior or meaningful residual differences.

When committing is authorized, stage only integration changes and inspect the
full index before committing. Unrelated staged work must be isolated with the
user's authority, not silently unstaged or included. Preserve explicitly
uncommitted and conflict-resolution-only endpoints; publish only when authorized.

For an uncommitted whole-branch merge, report ancestry convergence as pending and
verify `MERGE_HEAD`. After committing, verify
`git merge-base --is-ancestor <source> HEAD`. For an authorized transplant, report
adopted behavior without requiring ancestry.

Report source, destination, merge base, method, material adaptations or rejections,
deletion and residual-difference evidence, validation, and the resulting commit
or uncommitted state. Include detailed intent records only where they explain
the result or an unresolved decision.
