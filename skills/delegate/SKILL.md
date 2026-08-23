---
name: delegate
description: "Orchestrate substantive code and file implementation through a GPT-5.6 Luna subagent, then review and iterate on the result. Explicit invocation only."
---

# Delegate

Treat the request supplied with `$delegate` as the desired implementation.
Keep the parent agent accountable for scope, decisions, review, validation, and
the final response while a GPT-5.6 Luna worker performs the detailed edits.

## 1. Frame the Work

1. Inspect the request, applicable instructions, repository state, and relevant
   implementation surfaces far enough to define the work accurately. Resolve
   consequential product or authority questions before delegation; do not make
   the worker guess them.
2. Define a concrete implementation contract containing the desired outcome,
   in-scope behavior or files, constraints, acceptance criteria, validation,
   and any known unrelated changes that must be preserved.
3. Delegate substantive implementation. The parent may instead perform a
   clearly correct, localized, low-risk change when preparing and reviewing a
   delegation would cost more than making and validating the change directly.

Framing is complete when a fresh worker can act without needing the parent
conversation or making a consequential product decision.

## 2. Dispatch the Implementer

Spawn one primary implementation subagent with:

- `model: "gpt-5.6-luna"`;
- `reasoning_effort: "max"`; and
- `fork_turns: "none"`.

Give it a self-contained prompt containing the implementation contract,
working directory, relevant evidence and paths, applicable constraints, and
the validation it must run. Tell it to inspect before editing, preserve
unrelated work, avoid commits or publication unless authorized, and return a
concise account of changed files, validation, and unresolved issues.

Use one writer for overlapping files or behavior. Parallel workers are
appropriate only for scopes that are independent and cannot edit the same
surfaces. Give every implementation worker the same GPT-5.6 Luna, maximum
reasoning, and no-history configuration and a complete scope-specific contract.
Do not silently substitute another model or reasoning level when the requested
worker configuration is unavailable; report that limitation and preserve the
current state.

Once implementation is dispatched, coordinate it rather than duplicating its
detailed work. Do not edit worker-owned files concurrently.

## 3. Review and Iterate

1. Wait for every implementer, then inspect the individual and combined diffs
   and validation evidence against the contracts and repository instructions.
   The parent's review is an independent quality gate, not a restatement of the
   workers' reports.
2. Send material defects, missing coverage, or unclear decisions back to the
   same implementer as a precise follow-up. State what failed, the required
   outcome, and the validation to rerun. Repeat until the contract passes or a
   genuine blocker requires the user.
3. The parent may directly apply an obvious, localized correction found during
   review when another worker turn would be less efficient. Validate every
   direct correction and return broader or judgment-heavy rework to the
   implementer.
4. Run or independently confirm the proportionate final checks. Report the
   completed outcome, important decisions, changed files, validation results,
   and any residual risk.

The delegation is complete only when the parent has verified the requested
behavior and required checks, not merely when the implementer reports success.
