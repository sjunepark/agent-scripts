# Spawn a Codex Goal

Use only for an explicit `spawn` operation or equivalent request to prepare and
create a new task. Requires Codex app task tools: `list_projects`,
`create_thread`, `list_threads`, `read_thread`, and `wait_threads`. The child
requires native `create_goal`, `get_goal`, and `update_goal`, plus `$progress`.
Use the available tools' schemas as authority for supported arguments and
destination capabilities; do not substitute an ordinary subagent or simulate a
native goal with Markdown. General goal behavior is described in
[Follow a goal](https://learn.chatgpt.com/use-cases/follow-goals).

## Resolve Arguments and Target

Read this section before preparation. Treat these aliases as explicit user
model choices and pass settings to `create_thread`, not just to the prompt:

| Model argument | `model` | Default `thinking` |
| --- | --- | --- |
| `astra` | `gpt-6-astra` | `medium` |
| `luna` | `gpt-5.6-luna` | `max` |
| omitted | omit | omit |

An explicit reasoning argument overrides the preset, e.g. `spawn astra high`.
Accept an exact model identifier only when the user names it and the destination
supports it; with no preset or explicit reasoning, omit `thinking`. Reject an
unknown alias or unsupported combination before mutation; never silently
substitute a model or lower reasoning. If destination support can only be
validated at creation, report a rejection and retain preparation. With neither
argument, omit both overrides to use the configured defaults.

Call `list_projects` and resolve the user-named target, otherwise the current
repository's saved project, accounting for the current worktree's repository.
Prepare in that target repository. Never use a similarly named project or
silently transfer plans between repositories or hosts. Ask only when the target
cannot be resolved from evidence. If task tools or a suitable saved project are
unavailable, complete independently authorized preparation where its repository
is known and return the ready contract with the launch limitation; do not claim
that a task or goal started.

For Git projects, default to a new worktree. `spawn` requests its prepared
current state: pass `environment.startingState` with `type: branch` and the
verified prepared branch/ref as `branchName`. Do not omit it and accidentally
start from the project's default branch. Record the prepared commit for child
verification. Honor an explicit request to use the saved checkout directly with
`type: local`; for non-Git projects, also use local, but report that the required
Git preparation commit is unavailable and resolve that requirement before
launch. A remote destination must actually have the prepared ref and sources;
preparation alone does not authorize publishing or pushing them.

## Create One Task

After preparation and readiness pass, build exactly one contract using the
entry point's delivery rules. Check that the prepared branch/ref still resolves
to the recorded commit, and that every referenced source is tracked and present
there. If a dependency exists only in unrelated dirty work, resolve its commit
authority rather than starting an incomplete checkout.

Call `create_thread` once with the resolved project, environment, explicit
model/`thinking` settings when supplied, a concise outcome title, and a
self-contained prompt containing:

1. The user's request to create and execute a real native goal in this new task.
2. The complete closed goal contract, unchanged, plus the prepared commit and
   source paths as startup evidence rather than new scope fields.
3. Instructions to read applicable repository instructions and verify the
   prepared commit and cited sources in this checkout before implementation.
   Verify the prepared commit remains an ancestor and its required content is
   retained after any delivery-base selection; never switch to a default branch
   that drops the plan or prerequisites. If a required skill, source, or native
   tool is missing, report the failure and stop before implementation.
4. Instructions to call `get_goal` first; create a native goal with
   `create_goal` only if no unfinished goal exists. Use the contract's outcome
   and completion boundary as the objective. Omit `token_budget` unless the user
   explicitly supplied one, in which case pass that exact budget. Reuse a
   matching existing native goal on retry; never replace a different unfinished
   goal. Initialize/recover the supplied contract through `$progress` before
   implementation, execute only its scope, and call `update_goal` complete only
   after its completion and delivery conditions hold.
5. Instructions to report startup evidence promptly: checkout verification,
   durable contract initialization, and the actual native goal creation/recovery
   result. Report later completion truthfully through native goal status.

The parent remains the preparer and launcher; it does not set its own goal or
begin implementing the child's objective. The child's contract grants only the
chosen scope and delivery authority, regardless of its model or reasoning level.

## Verify Startup and Return

Creation is asynchronous. Retain the returned identifiers immediately. A
`clientThreadId` is a pending setup identifier, not a `threadId`; resolve the
ready task through `list_threads` using the returned setup identity and project
context, never title alone. Do not pass it to tools requiring a ready task ID.

Once a ready `threadId` is known, use `wait_threads` with its host and cursor in
bounded waits, and read only the startup evidence needed with `read_thread`.
Task creation or an `active` task status alone does not prove native goal
creation. Confirm checkout/source verification, `$progress` initialization, and
successful native goal creation or matching recovery before saying it started.
Return once startup is verified; do not wait for the entire implementation.

If setup is still pending after a bounded observation, report pending setup or
unverified goal startup accurately and return the existing task reference. On
retry, recover that task and inspect it before any new creation. A timeout,
ambiguous response, missing native tools, or child failure is never a reason to
blindly create another task or to report a running goal. Report the concrete
failure and completed preparation; offer the contract for manual use when
useful, but do not silently downgrade successful launch to prompt generation.

For a created task, finish with its actual startup state, project, preparation
commit, and requested model/reasoning (or configured defaults). Distinguish
requested settings from settings confirmed by the tool response. Emit the app's
`created-thread` directive using the returned `threadId`, or `clientThreadId`
while setup is pending, following the current tool instructions. Prompt-only
formatting from the delivery resource does not apply to this launch report.
