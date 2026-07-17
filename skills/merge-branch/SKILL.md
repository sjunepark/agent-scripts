---
name: merge-branch
description: "Manually integrate Git branch work without blind mechanical merges. Use when merging, dry-planning a merge, transplanting branch work, resolving conflicts, or auditing a completed merge."
---

# Merge Branch

Integrate branch intent, not just patches. Prefer Git's merge machinery as a draft, then edit, refactor, audit, validate, and report the deliberate result.

Use clear branch terms throughout the task: the current `HEAD` branch is the destination branch, and `<source>` is the branch being integrated into it. If the user says "target branch" ambiguously, confirm whether they mean the source branch to merge or the destination branch to receive the work.

## Choose One Mode

- **Dry plan:** When the user asks to analyze or plan without changing repository state, read and follow [modes/dry-plan.md](modes/dry-plan.md). Stop after presenting the plan and asking how to continue.
- **Audit:** When the user asks to review a completed or in-progress merge result, read and follow [modes/audit.md](modes/audit.md). Do not start or redo a merge unless separately authorized.
- **Integrate:** When the user asks to perform a merge, transplant branch work, or resolve conflicts, read and follow [modes/integrate.md](modes/integrate.md).

If the request combines modes, complete the read-only mode first and get any decision or authorization it requires before entering integration mode.

## Shared Guardrail

Do not move, force-update, or delete the source branch automatically. Convergence means the source tip is reachable from destination `HEAD`; aligning branch refs requires explicit approval.
