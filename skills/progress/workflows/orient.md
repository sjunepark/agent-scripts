# Orient to Work

Use this read-only branch to regain context or audit stale planning claims.

1. Resolve the selected scope's index and read all linked item files. If no
   index exists, report that fact and recommend initialization without
   creating files. During parallel work, keep other worktree namespaces out of
   scope unless the user requests a cross-worktree status report.
2. For an ordinary orientation, verify the current item and first queued plan
   against relevant code, tests, git status, and recent history. Scan the
   remaining links for obvious contradictions or missing targets.
3. For an explicit audit, verify every in-scope item and classify it as:
   `current`, `queued`, `available`, `blocked`, `complete`, or `stale`.
4. Keep document claims distinct from verified repository evidence. A stale
   item remains a claim until the user asks to repair it.

Return, in order:

- current work and its concrete next action;
- next scheduled plan, or that none exists;
- for an explicit audit, every remaining queued plan in order with its
  classification;
- unordered tasks, explicitly noting that their list order has no meaning;
- blockers, missing files, and stale claims;
- the smallest useful documentation repair, if any.

Complete the branch only when the reported current and next work are supported
by evidence. Keep the repository, index, git state, and external systems
unchanged.
