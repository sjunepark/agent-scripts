# Merge Audit Mode

Review whether a completed or in-progress merge preserves the intended behavior
and history. Keep the audit read-only: recommend fixes without editing, staging,
committing, or restarting a merge unless integration work is also authorized.

Establish destination, source or merge parent, pre-merge state, and whether the
result is staged, uncommitted, or committed. Use the request, `MERGE_HEAD`,
`ORIG_HEAD`, merge commits, reflog, and history as evidence. Resolve uncertainty
that prevents judging intent before issuing a verdict.

Compare both sides with the result, including conflict resolutions and manual
edits. Focus on:

- Expected parents and ancestry; a transplant or squash does not prove a
  requested whole-branch merge.
- Omitted or weakened behavior, intentional deletions, renamed paths, tests,
  documentation, configuration, migrations, and assets.
- Conflicting product or compatibility assumptions, duplicate owners or paths,
  and interactions across affected public contracts.

Run required checks and existing validation relevant to the merged area when
compatible with the read-only scope. Explain unavailable checks and distinguish
confirmed defects from risks.

Report a verdict with source/destination evidence, consequential findings,
validation, and any next action. For a sound merge, explain the evidence for
preserved intent; use additional sections only when the findings need them.
