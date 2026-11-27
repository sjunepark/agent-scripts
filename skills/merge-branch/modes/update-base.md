# Update Base Mode

Switch the current worktree to its base or integration branch and fast-forward
it to its configured upstream, without discarding work or guessing across
ambiguous branches, remotes, or worktrees. This mode does not merge or rebase
the base into a feature branch.

## Protect local work

If the worktree is dirty, assess the target and upstream read-only, then report
the result or remaining ambiguity with a request to commit or stash local work.
Do not fetch or switch until the worktree is clean. Never stash automatically.

## Resolve the target

- Use a branch the user named. Otherwise select the worktree's established base
  or integration branch when repository evidence makes it unambiguous. An
  available `dev` branch is the usual default candidate, not permission to
  ignore conflicting evidence.
- When the target remains ambiguous, ask, listing likely candidates and their
  upstreams.
- Stay inside the current worktree. If the target is checked out in another
  worktree, ask for a different base branch for this worktree or tell the user
  to run the update in the owning worktree; never update the other worktree.

## Resolve the upstream

- Use the selected branch's configured upstream; do not assume a remote named
  `origin`.
- If there is no upstream, or the same branch name exists on several remotes
  without a clear choice, ask which remote branch is authoritative.

## Update

1. Fetch only after the target and authoritative remote are resolved.
2. Before switching, verify the local branch is an ancestor of the fetched
   upstream. If it has commits absent from the upstream (ahead or diverged),
   stop and report both sides; do not choose a reconciliation strategy, reset,
   force-switch, or rewrite commits.
3. Switch to the local branch. For a remote-only branch, create its local
   tracking branch without overwriting an existing ref.
4. Fast-forward only to the fetched upstream.
5. Verify the worktree is on the selected branch, tracks the resolved upstream,
   has `HEAD` equal to the fetched upstream tip, and remains clean.

Report the branch, upstream, and resulting commit. If any check fails, report
the exact state and do not claim completion.
