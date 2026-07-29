---
name: project-status
description: "Project status: reconcile plans, Git and implementation evidence, validation, and relevant GitHub work to report current state and the best-supported next action."
---

# Project Status

Produce a concise, evidence-backed briefing and recommend one next action. Keep
the run read-only: inspect local and remote state without editing code,
documentation, planning state, git state, or GitHub state. End after the
briefing; implementation requires a separate request.

Treat **triangulation** as the governing rule. Plans, roadmaps, TODOs, handoffs,
and status documents describe intent or remembered state; verify their material
claims against independent repository evidence before presenting them as
current reality.

## 1. Establish the Snapshot

1. Read applicable instruction files and honor any user-named scope. Use the
   current conversation to locate active work, then verify its project-state
   claims by the same triangulation rule. Default to the current repository and
   worktree; include other worktrees or repositories only when the user
   requests an aggregate view.
2. Capture the local baseline before relying on status documents:
   - current branch, upstream, and ahead/behind state;
   - staged, unstaged, and untracked work, summarized by purpose when the diff
     makes that purpose evident;
   - a small recent history window, widened only when it cannot explain the
     apparent current work.
3. Discover the repository's compact routing and status sources. Prefer
   user-named files, then applicable README or architecture routing, roadmaps,
   plans, task lists, progress files, and handoffs. Follow only links that could
   determine current work, recent completion, blockers, or the next action.

This phase is complete when the reported current and next work, recent
movement, local in-progress work, and relevant scope are known as claims to
verify. Absence of a planning system is a finding, not a blocker.

## 2. Triangulate Plan and Implementation

For every claim that could change the headline status or next action:

1. Test it against at least one independent source: the working-tree diff,
   commit history, actual implementation and wiring, relevant tests, generated
   artifacts, or recorded validation and CI results.
2. Open the implementation entry points and tests needed to determine whether
   described work exists, is connected, and has validation coverage. Limit
   this inspection to establishing delivery state.
3. Run a targeted check only when it is an established, reasonably fast
   diagnostic whose scope and side effects are understood. Prefer existing CI
   evidence when a local check would be slow or environment-dependent. Name
   important checks that remain unrun.
4. Classify the claim:
   - **Verified** — independent evidence supports it.
   - **Drift** — present repository evidence materially contradicts it.
   - **Unverified** — affordable evidence is insufficient; lack of evidence
     alone is not contradiction.

When sources conflict, describe the document's intent separately from the
observed implementation state. Current repository evidence governs the status
briefing; documents may still govern desired direction.

This phase is complete when every headline claim is verified, identified as
drift, or explicitly left unverified, and neither current work nor the next
action rests only on a planning document.

## 3. Add Relevant GitHub State

When the repository has a GitHub remote and access is available:

1. Use `gh` to list open pull requests with their draft state, branches,
   review/check state, and recent activity. Inspect the current-branch or
   current-work pull request more deeply when one exists.
2. Inspect issues only when relevance is established by the current plans,
   commits, pull requests, an explicitly active milestone, or the user's named
   scope. Keep the general issue backlog outside the briefing.
3. Connect remote work to local and documented state instead of reporting it as
   a separate inventory.

If the remote, authentication, or network is unavailable, finish the local
briefing and state the resulting blind spot.

This phase is complete when relevant open pull-request state is included or its
absence is established, and any included issue has a concrete connection to
the work being reported.

## 4. Choose the Next Action

Recommend the smallest concrete action that best advances the observed project
state. Weigh, in order:

1. explicit user direction;
2. coherent in-progress local or pull-request work;
3. blockers, failing checks, or review gates preventing that work from landing;
4. verified current planning intent;
5. the best-supported queued plan or relevant issue.

Treat dirty or remote work as evidence of activity, not proof of ownership or
priority. When two materially different actions remain equally plausible,
state the decision needed instead of inventing certainty.

The recommendation is complete when it names the target, explains why it comes
next, and gives a concrete first step supported by the briefing's evidence.

## 5. Deliver the Briefing

Lead with the bottom line: what state the project is in and what to do next.
Then include only useful sections from:

- **Current state** — active work, branch/worktree condition, and project
  health signals.
- **Recent movement** — meaningful local commits, uncommitted work, and merged
  or active remote work.
- **Evidence and drift** — material verified, contradicted, or unverified
  planning claims.
- **Pull requests and issues** — only work relevant to the current state.
- **Blockers and blind spots** — failures, missing access, ambiguity, and
  validation not run.
- **Recommended next action** — one action and its evidence-backed rationale.

Distinguish observation, document claim, and inference. Attach concise file,
commit, check, PR, or issue references to the claims they support. Avoid raw
command output, exhaustive file inventories, a full issue backlog, or a
general code review.

Before responding, compare repository status with the baseline and confirm that
the run introduced no repository or remote changes. Remove only incidental
artifacts created by diagnostics run during this invocation, and report any
change that cannot be safely reversed.
