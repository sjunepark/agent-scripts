# Sequential issue tasks

Use the current task as coordinator. Resolve each selected issue in its own
visible Codex app task with `model: "gpt-6-astra"` and `thinking: "medium"`.
Keep at most one issue task running, including resumed tasks. Do not pre-create
the rest of the queue or substitute ordinary subagents for these app tasks.

## Confirm the queue

1. Discover the currently callable app tools for `list_projects`, `create_thread`,
   `list_threads`, `wait_threads`, `read_thread`, and `send_message_to_thread`.
   Verify the selected host supports the exact model and reasoning setting.
   If a required capability is absent, explain the limitation and discuss an
   alternative before dispatch; do not silently switch model, host, or mode.
2. List issues through `gh issue list` and read candidate bodies/comments before
   proposing a selection. Retrieve additional pages when the candidate scope
   exceeds the returned limit; identify any list that is intentionally partial.
   Check dependencies, existing PRs, and required environments.
3. Present a compact ordered proposal with issue numbers, titles, URLs, scope,
   and dependency/blocker notes. State that each issue goes through validated PR,
   review, merge, and resolution verification in its own Astra/medium task.
   Ask the user to confirm both membership and order before creating any task.
   An exact queue already confirmed in this conversation needs no repeat.
4. Keep a coordinator-owned queue record in the repository's existing progress
   convention; otherwise use a scoped Markdown file under `plans/`. Record
   confirmed order, dependencies, integration branch, project/host, issue
   disposition, task IDs, branch/PR/commit evidence, blockers, and next action.
   Do not mix this coordinator state into issue PRs. Reconcile this record with
   current task and GitHub state on resume before dispatching anything.

## Dispatch one issue

1. Choose the earliest runnable issue in the confirmed relative order. Recheck
   whether it was already resolved and whether its prerequisites are satisfied.
   Select a matching saved project using `list_projects`; do not guess project
   IDs or assume the coordinator's repository is the implementation repository.
2. For a Git project, default to a worktree. Honor an explicit request to use the
   saved project directly. Follow the tool's current starting-state contract:
   omit `startingState` unless the user explicitly requested a particular Git
   state, and only pass an observed/requested branch name. Instruct the child to
   fetch and start from the intended current integration branch before editing,
   including changes landed by earlier issues. Never inherit another issue's
   unfinished diff. If the project or base cannot be resolved, hold dispatch.
3. Compose a self-contained prompt containing:
   - the exact issue URL and agreed outcome, relevant issue evidence and user
     constraints, acceptance criteria, integration branch, and dependencies;
   - the single-issue workflow and authorized delivery boundary from `SKILL.md`,
     either via a verified accessible skill in the child environment or by
     including that contract directly; do not assume a parent-local path exists
     on another host;
   - the requirement to handle only this issue, avoid spawning further issue
     tasks, and return the completion/blocker evidence required by the skill;
   - instructions to end the turn with a blocked result when progress requires
     user input, and to leave resumption to the coordinator so the queue remains
     sequential. Do not schedule a background continuation.
4. Call `create_thread` once with the resolved project/environment and the exact
   model and thinking arguments above. Record the returned identity immediately.
   If setup returns only `clientThreadId`, retain it as pending; resolve the ready
   task from subsequent app state using correlated identity evidence. Never use
   a pending ID in a `threadId` field or select a match by title alone.
   If creation times out or its outcome is ambiguous, inspect task state before
   retrying; do not duplicate the issue task. Pending or uncertain creation
   blocks further dispatch.

## Collect the result and advance

1. Once a ready `threadId` is known, wait on that one task through `wait_threads`,
   retaining `hostId` and the returned cursor as `afterCursor`. Prefer bounded
   waits up to 60 seconds so the coordinator remains responsive; use a zero-time
   snapshot for an immediate status check. Avoid narrating unchanged snapshots.
2. Use `read_thread` when the compact result lacks necessary evidence, and
   `send_message_to_thread` for a scoped follow-up to the same issue task.
   Omit model/thinking overrides for follow-ups to retain its configuration.
   The coordinator reads the child's result; no unsolicited child-to-parent
   callback is required. Do not assume the original task wakes automatically
   after its turn has ended. Keep coordination active while the queue runs, or
   explicitly report the saved checkpoint when external input is required.
3. Treat a finished turn as a result to assess. Verify PR/issue state through
   `gh`, inspect the returned validation and integrated commit evidence, and
   distinguish resolved, skipped, blocked, and failed. Send incomplete work back
   to the same child within scope before advancing; do not equate a successful
   wait response or a child's assertion with the issue completion gate.
4. For a blocker, record and surface it, defer dependent issues, and continue
   with the next independent issue only after the child has ended its turn and
   has no continuing work. An outstanding approval or uncertain/running task
   is not a safe handoff. If the user resumes an earlier child independently,
   reconcile active task states and restore sequential execution before any
   further dispatch.
5. Update the queue record after every result. Reconfirm only material changes
   to membership or relative order; dependency deferrals follow the agreed
   blocker policy. Reuse the same child for later follow-ups and never recreate
   a task merely because the coordinator was interrupted.

Finish with one concise queue summary linking each issue, PR, and task, with
verified outcomes and remaining blockers/dependents. Include the app-required
`created-thread` directive for each task actually created, using its returned
ready or pending identifier. An exhausted runnable queue with blocked entries
is partial completion, not all issues resolved.
