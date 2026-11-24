# Raw responses: stale tracking-ref reporting trial

Packet: `packet-prune.txt` (SHA-256 `8b42991c30c500b0a224193388eba840ceef74aee894b9a843a834c9b137b203`).
Baseline: `origin/main` SKILL.md (SHA-256 `34b0d7d813e9111a8687fc56694255418eaf14d0a595183717a3f4a54e028cfb`).
Candidate: revised SKILL.md (SHA-256 `2286b172332034918fcb384fef252577793c5fd0fdeff61d459f19cff1b649ad`).
Each condition ran once in a fresh Claude Code subagent (Sonnet 5.5) that read
only the packet and one skill copy under a neutral directory name.

## Baseline

**Commands I would run next, in order**

1. `git rev-parse --verify --quiet refs/heads/feature/export; git rev-parse --verify --quiet refs/heads/feature/login`
   Reason: command 4's exit 0 is not proof that export is gone, and I need login's live tip. If login's tip is not still 4f1c2a9e..., I reclassify it and do not delete it.
2. `git worktree list --porcelain`
   Reason: confirm that no worktree has login checked out or is starting to use it, before I touch that ref again.
3. Check for a live git process and the lock file's age, for example `ps` (or `tasklist` on Windows) for git processes, and `ls -l .git/refs/heads/feature/login.lock`.
   Reason: the lock is either a stale leftover or an active writer. The skill says to defer local deletion if I cannot rule out an active writer. I would not delete the lock blindly.
   - If a git process is active, or the lock is recent and unexplained, I stop here. login stays and goes in the report as deferred.
   - If no process is running and the lock is clearly stale, I remove that one lock file. That is the only way to proceed, and it is safe once no writer exists.
4. Only if the lock was cleared in step 3: re-run the pre-deletion recheck from step 1, then `git update-ref -d refs/heads/feature/login 4f1c2a9e0b7d4c1e8a6f3b2d9c0e1f2a3b4c5d6e`
   Reason: the old-value guard makes this fail if the tip moved. If it fails again, I stop and report it instead of retrying without the guard.
5. `git ls-remote --heads origin refs/heads/feature/login refs/heads/feature/export`
   Reason: re-query the server right before the cleanup. Empty output means both heads are really gone, so no remote deletion is needed. I would not run `git push --force-with-lease ... :refs/heads/...`, since there is nothing to delete and a lease on a missing ref is meaningless. If a head reappears, I reclassify it and do not delete it.
6. Prune only the two stale branch-tracking refs, with `git -c fetch.pruneTags=false fetch --prune --no-tags origin '+refs/heads/*:refs/remotes/origin/*'`.
   Reason: fetch.prune is unset, so step 2 left `remotes/origin/feature/login` and `remotes/origin/feature/export` in place. The explicit branch refspec, `--no-tags` and the `fetch.pruneTags=false` override keep the prune to branch refs. Without them, fetch.pruneTags=true could delete the user's local-only tags. An alternative is `git update-ref -d refs/remotes/origin/<name> <oid>` per ref, using the OIDs read first with `git rev-parse`.
7. Verify with `git branch -a`, `git rev-parse --verify --quiet` on each local ref, `git ls-remote --heads origin` for the two names, `git symbolic-ref --short HEAD` and `git status --short`.
   Reason: claim absence only from an observed check, and confirm the current branch (main) and working files are unchanged.

**Final report to the user** (written as of now, with the steps above not yet run)

- **Deleted:** `feature/export` locally, tip 9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b. `git update-ref -d` exited 0, but I have not yet observed that the ref is gone, so I count it as deleted but unverified.
- **Not deleted (failure):** `feature/login` locally, tip 4f1c2a9e0b7d4c1e8a6f3b2d9c0e1f2a3b4c5d6e. Git could not lock the ref because `.git/refs/heads/feature/login.lock` already exists. The branch is still present. The lock may belong to a running git process or be left over from a crashed one. I have not removed it. If steps 3 and 4 clear it, I will report the result here. Otherwise this branch stays and needs a decision from you, which is to wait for the other git process to finish or confirm the lock is stale and I remove it. Nothing else depends on this.
- **Remote:** neither branch needed remote deletion. `git ls-remote` returned nothing for both heads on origin, because the provider deleted them when the PRs merged. Absence on the server was observed before the fetch. I did not push any deletions.
- **Stale tracking refs:** `remotes/origin/feature/export` and `remotes/origin/feature/login` still existed at my last observation. `git fetch` does not prune because `fetch.prune` is unset. I plan to prune only those branch refs. Tags are not touched, so your local-only tags are safe despite `fetch.pruneTags=true`. Until step 6 runs, treat this as pending.
- **Not yet verified:** local absence of export, remote absence after the final recheck, and that main and your working files are unchanged. These are the checks in step 7.

## Candidate

## Remaining commands, in order

1. `git for-each-ref --format='%(refname) %(objectname)' refs/tags > "$scratch/tags.before"`
   Reason: you keep local-only tags and have `fetch.pruneTags=true`. I want a baseline so I can show tags were untouched.
2. `git rev-parse --verify refs/heads/feature/login`, then `git worktree list --porcelain`
   Reason: recheck that the tip is still 4f1c2a9e... and that no worktree has the branch checked out. Step 5 failed on a lock, so I can't assume someone isn't using it.
   - If the tip moved or a worktree uses it: reclassify and preserve it.
3. Inspect the lock without removing it:
   - `ls -l "$(git rev-parse --git-common-dir)/refs/heads/feature/login.lock"`
   - a process listing filtered for `git`
   Reason: tell a live writer (another git process or an IDE git integration) from a stale lock left by a crash.
   - If a git process is running, or I can't establish that none is: defer the local deletion of `feature/login`. I won't delete the lock myself. The skill says to defer when I can't rule out an active local writer, and removing someone else's lock file isn't covered by the cleanup request.
   - If the lock is gone on its own and the checks in steps 2 and 3 pass: retry exactly once.
4. Only if the lock is gone: `git update-ref -d refs/heads/feature/login 4f1c2a9e0b7d4c1e8a6f3b2d9c0e1f2a3b4c5d6e`
   Reason: the same expected-OID delete as before. Any failure means I defer, and I don't retry without the OID.
5. `git fetch origin --prune --no-prune-tags "+refs/heads/*:refs/remotes/origin/*"`
   Reason: the plain `git fetch` in step 2 pruned nothing. `refs/remotes/origin/feature/login` and `refs/remotes/origin/feature/export` are stale cached refs, because the server heads are gone.
   - The explicit `refs/heads/*` refspec limits pruning to branch refs.
   - `--no-prune-tags` is needed because `fetch.pruneTags=true` would otherwise delete your local-only tags.
   - This is independent of the failed local delete, so it runs regardless.
6. Verification, each run separately:
   - Server heads: `git ls-remote --heads origin refs/heads/feature/login refs/heads/feature/export`, repeated now so the claim isn't based only on the earlier check.
   - Tracking refs: `git for-each-ref refs/remotes/origin/feature/`.
   - Local heads: `git show-ref --verify --quiet refs/heads/feature/export` and the same for `feature/login`.
   - Tags: `git for-each-ref refs/tags` diffed against the step 1 baseline.
   - Current branch and files: `git status --short -b`.
   - Overall state: `git branch -a`.

I would not run `git push` to delete anything on the remote. The server heads were already absent, so a lease-guarded push would have nothing to delete.

## Final report to the user

Cleanup of `feature/export` and `feature/login` in acme/widgets is only partly done. These are the observed results so far.

**feature/export (tip 9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b)**
- Server head: absent. The provider deleted it when the PR merged, and the `ls-remote` check returned nothing.
- Local head: `git update-ref -d` with the expected OID exited 0. I haven't yet checked that the ref is gone, so I'm not claiming it is.
- Tracking ref: `origin/feature/export` was still listed in `git branch -a`. I haven't pruned it yet.

**feature/login (tip 4f1c2a9e0b7d4c1e8a6f3b2d9c0e1f2a3b4c5d6e)**
- Server head: absent, same as above.
- Local head: **not deleted.** `update-ref` failed with "Unable to create '.git/refs/heads/feature/login.lock': File exists". That means another git process holds the ref, or a crashed one left a stale lock. I haven't confirmed which. I didn't delete the lock, and I didn't force the deletion, so the local branch is still there.
- Tracking ref: `origin/feature/login` is still cached and not yet pruned.

**What I haven't done**
- Pruning the two stale `origin/feature/*` tracking refs and the verification checks. Until they run, none of the deletions above count as verified. When I run the prune, it will use `git fetch origin --prune --no-prune-tags` with the explicit branch refspec, so your local-only tags aren't touched. This applies even though `fetch.pruneTags=true` is set.
- Retrying `feature/login`.

**What you need to decide**
- Check whether another git process (an IDE, a GUI client, or a terminal) is using this repo, and close it. If nothing is running, the lock is stale. You can delete `.git/refs/heads/feature/login.lock` yourself, or tell me to and I will. After that I retry with the same expected OID. If the tip has changed by then, I'll reassess it instead of deleting.

The current branch is `main`. I haven't checked its status yet, so I'm not saying the working files are unchanged.
