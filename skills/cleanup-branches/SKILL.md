---
name: cleanup-branches
description: "Clean completed local and remote Git branches using integration and pull-request evidence. Explicit invocation only."
disable-model-invocation: true
---

# Cleanup Branches

Requires Git and repository access; use authenticated `gh` for GitHub evidence.

Clean completed local and remote branches in the requested repository. A cleanup
request authorizes both scopes unless the user narrows it. Choose and perform
routine deletions from evidence without a separate approval round. Ask only when
an unresolved fact or retention decision materially changes what can be removed;
complete independent cleanup first. Honor requests limited to inspection.

## Establish the live state

- Resolve the repository, fetch and push destinations, default branch, and
  durable integration branches from configuration, repository instructions, and
  PR history. Do not assume `origin`, `main`, or that every configured remote is
  a deletion target; distinguish the project's repository from upstreams and
  forks. Resolve ambiguous ownership before deleting from that remote.
- Inspect local branches, upstreams, worktrees, and working-tree status. Refresh
  relevant remote branch refs and inspect live remote tips. Scope fetching and
  pruning to branch refs; do not prune tags or other namespaces. Read each
  target remote's fetch refspec for its tracking namespace rather than assuming
  `refs/remotes/<remote>/`. A plain `git fetch` prunes nothing unless
  `fetch.prune` or `remote.<name>.prune` is set, and `git branch -a` lists local
  heads beside these cached tracking refs, so neither shows server state.
- For GitHub, use `gh` to obtain complete relevant PR history, open PR heads and
  bases, and branch protection/ruleset evidence. Match repository identity as
  well as branch name. Follow pagination; a failed or partial lookup is unknown,
  not evidence that no PR or protection exists. For another host, use its
  equivalent API. Continue only deletions whose required evidence is available.

## Decide what is finished

Preserve the current branch, branches checked out in any worktree, default and
protected branches, established long-lived branches such as `dev` or release
lines, intentional backups, and branches serving as an open PR's head or base.
Apply these retention checks to corresponding remote branches too. Do not switch
branches, remove worktrees, stash, reset, or discard working files to enable
cleanup; dirty files need not block deletion of unrelated completed branches.

Evaluate each local and remote tip independently. A branch is eligible when it
passes the retention checks and either:

- Its exact tip is an ancestor of a freshly verified durable integration tip.
  A PR is not required for work demonstrably integrated this way.
- For a squash/rebase merge, a merged PR in the correct repository identifies
  that exact final head, and its resulting merge commit is reachable from the
  durable integration tip. Verify the current branch has no additional commits.
  Trace stacked merges through to durable integration; merging into a temporary
  branch alone does not establish completion. If this proof is unavailable,
  preserve the branch and explain the missing evidence.

Age, a missing upstream, a closed unmerged PR, a matching name, or a successful
`git branch -d` alone does not establish completion. Preserve branches with
unaccounted-for commits. Infer routine retention from repository evidence;
batch only consequential remaining questions with a recommended disposition.

## Delete and verify

1. Record each selected ref, repository/remote, full tip OID, and integration
   evidence before deletion. Briefly state the intended cleanup and proceed
   within the existing request; do not turn the record into an approval gate.
2. Recheck tips, worktree occupancy, and relevant PR/protection state immediately
   before mutation. Reclassify changed branches; do not delete using stale proof.
3. After verifying worktree retention checks, delete the exact local tip with
   `git update-ref -d "refs/heads/$branch" "$oid"`. This fails if another process
   moved the ref; refresh and reassess instead of dropping the expected OID.
   This plumbing command does not enforce worktree occupancy: coordinate with
   active local writers so no worktree can start using the branch during deletion.
   If that cannot be established, defer that local deletion. Remove its branch
   configuration only after successful deletion and while recreation is excluded.
4. Delete remote heads with an explicit expected-tip lease:
   `git push --force-with-lease="refs/heads/$branch:$oid" "$remote" ":refs/heads/$branch"`.
   Populate these variables from the verified record and quote ref names. A lease
   rejection means the evidence changed: refresh and reassess, never retry with
   an unguarded deletion or bypass server protection.
5. Verify three refs per branch independently; each is a separate operation
   and outcome:
   - **Server head:** query the server or provider (`git ls-remote --heads` or
     its API) for absence.
   - **Tracking ref:** prune obsolete branch-tracking refs, then check the
     tracking ref's absence locally. Substitute the remote's resolved mapping
     in `git fetch "$remote" --prune --no-prune-tags "+refs/heads/*:refs/remotes/$remote/*"`;
     `--no-prune-tags` is required because `fetch.pruneTags` or
     `remote.<name>.pruneTags` enables tag pruning even with a branch refspec.
     Pruning removes cached local refs, never server branches.
   - **Local head:** check `refs/heads/$branch` absence, including lock,
     permission, or expected-OID failures and intentional deferrals.

   Confirm the current branch and working files are unchanged. Report deleted
   branches with their tip OIDs, retained/uncertain branches with reasons, and
   failures separately. Keep the ordinary success report concise; when a
   branch's three outcomes differ, report each and explain the entry that stays
   in `git branch -a`. Continue independent eligible deletions after an
   isolated failure. A pruning failure does not undo separately verified server
   deletion, and a vanished tracking ref does not prove it. Claim verified
   absence only from an observed follow-up check, never from command success or
   an assumed result. If that check is unavailable or fails, report the
   command's outcome and mark absence as unverified.
