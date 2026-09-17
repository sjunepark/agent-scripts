# Authoring Review

Review the candidate against the task and surrounding instructions. Apply the
criteria relevant to the change; for a whole-skill audit, account for all runtime
resources. Report consequential findings and evidence rather than a mandatory
row for every criterion.

## Questions that affect acceptance

- **Purpose:** Does the skill provide reusable, non-obvious help for one coherent
  capability? Can ordinary reasoning or an existing instruction already do it?
- **Selection:** Does the concise description distinguish intended tasks from
  likely near misses? Do invocation intent, adapter metadata, and cases agree?
- **Context:** Does the entry point expose the necessary constraints and route
  only to relevant resources? Can a short skill remain self-contained?
- **Specificity:** Are exact steps reserved for real invariants or fragile
  operations? Can the agent adapt judgment-heavy work to the evidence?
- **Authority:** Does the skill preserve user scope and delegated decisions,
  ask only about unresolved consequential boundaries, and continue independent
  work while an answer is pending?
- **Completion:** Is the required outcome observable? Does the workflow carry
  authorized work through verification and fixes without an unnecessary stop
  for approval or repeated passing checks?
- **Package:** Are resources reachable, necessary, and portable within declared
  prerequisites? Are licensing obligations preserved?
- **Evidence:** Do checks cover changed decisions and relevant failure boundaries?
  Are structural checks, simulated behavior, live execution, and untested claims
  distinguished?

## Disposition

Fix failures that prevent correct selection, execution, safety, or completion
before declaring the candidate ready. Keep minor limitations explicit when they
matter. Treat unsupported stylistic preferences as hypotheses, not required
changes. If a choice is already sound, retain it without cosmetic churn.

Use the evaluation workflow for experiments needed to resolve uncertain behavior.
Do not turn each review question into another mandatory trial or output section.
