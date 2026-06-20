---
name: handoff
description: "Manual workflow for moving current Codex work into a fresh thread through a durable Markdown handoff plan and compact continuation prompt. Use only when the user explicitly invokes `$handoff` or explicitly asks to run/use the handoff skill; do not trigger automatically for ordinary handoff planning, context-pressure warnings, or casual mentions of handoff."
---

# Handoff

Move active work to a fresh Codex thread without carrying the old transcript forward. Preserve the detailed state in a durable Markdown handoff plan first, then create the thread when tooling is available.

## Workflow

1. Confirm the transfer target.
   - Treat the user invocation as the latest instruction.
   - If the user asks for prompt-only output, do not create a thread; still prefer a durable Markdown file unless the user also says not to edit files.
   - If thread creation tools are not already visible, search for `create_thread` before falling back to a manual prompt.
   - Prefer a fresh `create_thread` flow over a fork unless the user explicitly asks for copied thread history.

2. Capture the current execution context.
   - Record the current working directory and repository root.
   - Record the current branch with `git rev-parse --abbrev-ref HEAD` when inside git.
   - Check `git status --short` and include only meaningful changed-file context.
   - Capture the source thread title when the active environment exposes it; do not invent it.
   - Include model and reasoning effort only when the active environment exposes them; do not invent them.
   - Preserve user constraints, confirmed decisions, blockers, validation results, and open questions that affect continuation.

3. Create or update a durable Markdown handoff plan.
   - For repo-scoped or workspace-scoped work, create or update a compact Markdown handoff plan before creating the new thread.
   - Treat the file as the durable source of truth; the fresh-thread prompt should point to it, not replace it.
   - Skip the file only when the user explicitly forbids file edits or no writable repo/workspace exists; state that fallback clearly.
   - Inspect repository conventions first: a relevant existing plan or handoff file, `plans/`, `docs/plans/`, `.pi/plans/`, `todo/`, `todos/`, `docs/todos/`, existing `PLAN*.md` or `TODO*.md`, or similar handoff files.
   - If no convention exists, use `PLAN-<SPECIFIC-SLUG>.md` at the repository root.
   - If a relevant handoff already exists, update it rather than creating a duplicate.
   - Keep the file implementation-ready, not transcript-like.

Use this structure for new handoff plan files:

```markdown
# <Task title>

## Purpose

Short paragraph explaining what is being done and why.

## Current state

- Confirmed decisions, constraints, completed work, blockers, assumptions, and open questions.
- Changed files or repository evidence only when they matter for continuation.

## Next implementation slice

- [ ] Small, reviewable next step.
- [ ] Small, reviewable next step.
- [ ] Small, reviewable next step.

## Files to inspect first

- `<path>`: reason.

## Validation

- Commands run, results, blockers, or manual QA notes.

## Fresh-thread prompt

Paste the exact continuation prompt here.
```

Omit optional sections that would be empty or obvious. Keep ordinary handoff plan files compact, around 80-140 lines.

4. Build the fresh-thread prompt.
   - Make the prompt self-contained enough to work even if the next thread has not seen this conversation.
   - Reference the durable handoff plan path and tell the new thread to read it first.
   - If no durable handoff plan exists, explain why in the prompt and include the missing details inline as a fallback.
   - Tell the new thread to check git status before editing.
   - Tell the new thread to continue from the listed next implementation slice.
   - Tell the new thread to preserve confirmed decisions unless repository files contradict them.
   - Tell the new thread not to stage, commit, push, or open a PR unless the user asked for that.
   - Include exact open questions or blockers that should be resolved before implementation.
   - Include the suggested new thread title when a thread will be created.

Fresh-thread prompt template:

```text
Continue this Codex task in a fresh thread.

Context:
- cwd: <absolute working directory>
- repository: <repo root or "none">
- branch: <branch or "unknown">
- model: <current model, "current/default", or "unknown">
- reasoning effort: <current reasoning effort, "current/default", or "unknown">
- handoff plan file: <absolute path, or "none: <reason>">
- suggested new thread title: <title or "none">
- source thread state: <one-sentence summary>

Start by:
1. Read the handoff plan file first; if it is listed as `none`, use the inline fallback context.
2. Run `git status --short`.
3. Inspect only the files listed in the handoff plan unless repository evidence requires widening scope.
4. Continue from the `Next implementation slice`.

Carry forward:
- Confirmed decisions: <bullets>
- Constraints: <bullets>
- Validation state: <bullets>
- Open questions/blockers: <bullets>

Do not stage, commit, push, create a PR, or archive anything unless explicitly asked.
```

5. Create the thread when tooling is available.
   - Use the Codex app thread tool, not shell commands, for thread creation.
   - If title-management tools are not already visible, search for `set_thread_title` before creating the thread.
   - Choose the title before creating the thread. Prefer `Handoff: <source title>` when a useful source title exists; otherwise use `Handoff: <repo or task>`.
   - Use a prefix, not a suffix, because the start of titles survives sidebar truncation better. Keep the name under about 70 characters, trim at a word boundary when practical, and avoid duplicate `Handoff:` prefixes.
   - Call `list_projects` first and choose the project that matches the current repository or working directory.
   - Use `create_thread` with a project target and local environment for same-checkout continuation when possible.
   - After `create_thread` succeeds, immediately call `set_thread_title` with the created thread id and the chosen title. Do not assume `create_thread` can set the title directly.
   - Omit the `model` argument unless the user explicitly requested a model override.
   - Omit reasoning overrides unless the user explicitly requested them or the current environment exposes an exact value that should be preserved.
   - If no matching project is available, provide the fresh-thread prompt for manual use instead of creating an unrelated projectless thread.

6. Final response.
   - Report the durable handoff plan path, if created or updated.
   - Report the created thread, if creation succeeded, following any active app directive for created-thread output.
   - If thread creation was unavailable or not requested, provide the exact fresh-thread prompt.
   - Keep the response short; do not include a transcript recap.
