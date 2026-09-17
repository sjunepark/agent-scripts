---
name: delegate
description: "Orchestrate code and file implementation through narrow GPT-5.6 Luna subagent assignments, then independently review and integrate the result. Explicit invocation only."
---

# Delegate

The parent owns the complete requested outcome, decomposition, decisions,
review, validation, and final response. GPT-5.6 Luna workers perform substantive
implementation in narrow slices. The parent may directly make a clearly correct,
localized, low-risk edit when delegation and review would cost more than making
and validating it.

## Define the next slice

Inspect the relevant repository state and instructions, and resolve consequential
product or authority decisions that the worker is not authorized to make.
Decompose broad features, refactors, migrations, and multi-surface requirements;
never transfer the end-to-end requirement to one worker.

Each slice needs one concrete outcome, bounded behavior and file ownership,
constraints, dependencies, acceptance criteria, and independently reviewable
validation. Size by behavior, not line count. A source change and directly
coupled regression test may share a slice when separating them would make
validation meaningless; this does not justify a broad assignment.

## Dispatch with complete context

Use one primary implementer per slice with these exact settings:

- `model: "gpt-5.6-luna"`
- `reasoning_effort: "max"`
- `fork_turns: "none"`

The self-contained prompt must reproduce the original user request verbatim and
include, without paraphrasing, compressing, or omitting, every known requirement,
constraint, acceptance criterion, decision, failure or error output, repository
fact, and integration dependency that can affect this slice. Include exact
files to read in full when needed. Omit unrelated background, but do not replace
result-affecting context with a vague summary.

Include the working directory, slice contract, validation, and how this narrow
assignment contributes to the larger outcome without owning it. Tell the worker
to inspect before editing, preserve unrelated work, avoid commits or publication
unless authorized, and return changed files, validation, and unresolved issues.

Use one writer for overlapping files or behavior. Parallelize only independent
slices whose ownership cannot overlap; never edit worker-owned files concurrently.
If the required model or reasoning setting is unavailable, report the limitation
and preserve the current state instead of silently substituting.

## Verify and finish

Inspect each worker's diff and validation against its contract and repository
instructions. Independently review and validate a slice before dispatching
anything dependent on it; a success report alone is insufficient.

Send material defects or missing coverage back to the same worker as a precise,
bounded correction with the evidence and required validation. The parent may
apply an obvious localized correction directly and validate it; return broader
or judgment-heavy rework to the worker. Continue through the authorized slices
until the complete requested behavior and required checks pass, or a real
blocker prevents further progress.

Report the verified outcome, important decisions, changed files, validation,
and remaining limitations across all slices.
