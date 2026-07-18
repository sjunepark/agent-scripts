---
name: record
description: "Maintain a distilled, temporary conversation record across a long-running session. Use only when the user explicitly invokes $record."
---

# Record

Run this skill only after the user explicitly invokes `$record`. Activate a
write-through continuity cache for the current conversation. Keep the latest
truth, not a transcript.

## Start Recording

1. Resolve one dedicated record file.
   - Honor a dedicated path named by the user. Otherwise prefer `RECORD.md` at the repository root, or the current working directory outside a repository.
   - Reuse an existing file only when it is clearly managed by `$record`. If the preferred path is tracked, occupied by unrelated content, or otherwise unsuitable, choose a dedicated untracked path such as `.record/RECORD.md`.
   - Keep the record separate from project documentation, plans, progress files, source files, and instruction files.
   - In a Git repository, keep the chosen path untracked and ignored. Add its repository-relative path to the local exclude file returned by `git rev-parse --git-path info/exclude` when needed, then confirm it is ignored with `git check-ignore -q -- <record-path>`. Use another path if `git ls-files --error-unmatch -- <record-path>` succeeds.

   This step is complete when the record has a dedicated path that cannot enter an ordinary commit.

2. Capture the conversation so far.
   - Read any existing managed record and all available conversation context.
   - Write a compact living record with this default shape, omitting empty sections:

```markdown
# Conversation Record

> Temporary local continuity cache managed by `$record`. Keep untracked and uncommitted; while active, reconcile it after substantive turns.

Status: active

## Current context

## Decisions and rationale

## Constraints and preferences

## Durable details

## Open threads
```

   Capture goals, current state, decisions, rationale that still matters, constraints, preferences, discoveries, unresolved questions, and likely next actions. Preserve exact wording only when its nuance matters. Represent sensitive values by purpose rather than literal contents.

   This step is complete when a future agent could resume the conversation accurately from the record without needing a transcript.

3. Activate recording.
   - Tell the user once that recording is active and give the exact record path.

   This step is complete when the active status and recovery path are visible in the conversation.

## While Recording

- On every later substantive turn, read the record before relying on earlier context and reconcile it before responding whenever the durable state changed.
- Treat the latest explicit user direction and observed current state as authoritative when they conflict with the record.
- Replace superseded state in place. Retain an earlier choice only when its rationale explains the current choice; label it as superseded context rather than presenting it as current.
- Remove resolved questions, obsolete next actions, duplicates, and detail that no longer helps recovery. Keep the file compact and topic-agnostic as the conversation moves between tasks.
- Maintain the record quietly. Mention it after activation only when its integrity is at risk or when the user asks about it.

Recording remains active for the current conversation until the user explicitly stops it; later turns do not require another invocation.

## Stop Recording

When the user explicitly stops recording, reconcile the record one final time and change its status to `inactive`. Leave the ignored file in place for recovery unless the user asks to remove it. A separate conversation must invoke `$record` to activate recording there.
