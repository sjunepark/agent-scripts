---
name: teach
description: "Explain the design and behavior of source code, whole applications, and code or document changes. Use for understanding rather than a requested code review."
---

# Teach

Build a source-grounded mental model of purpose, responsibilities, contracts,
control or data flow, and the decisions that make the design understandable.
This is explanatory work: teach syntax or implementation mechanics when asked,
and include external context only when it changes the local model.

## Select the target and guidance

- For a diff, commit, range, patch, or code/document comparison, read
  [Teaching changes](guides/changes.md). With no named target in a Git
  repository, default to staged, unstaged, and relevant untracked changes.
  List changed paths before reading content and exclude secret-bearing
  environment or credential files and private documents from that default
  selection. Ask for a target if none remain. Redact secrets and private
  personal data from any selected content before quoting or summarizing it.
- For an exact target of `project`, an entire-application orientation, or a
  follow-up from its learning map, read
  [Whole-application orientation](guides/whole-application.md).
- Otherwise choose the smallest coherent subsystem, feature, boundary, or flow
  that answers the request.
- Read [Snippets](guides/snippets.md) when exact code is needed to explain a
  contract, condition, data shape, or transition.
- Read [Diagrams](guides/diagrams.md) when a visual relationship would materially
  improve understanding.

## Ground and explain the model

Start at the relevant entry point or public boundary. Read supporting code,
tests, and nearby instructions or architecture notes as needed to trace the
main flow and explain its contracts without guessing. Label inferred intent.

Teach in learning order: purpose and context, then the important flow,
ownership, and decisions. Match depth and format to the learner and target;
a small change may need only a paragraph. Use evidence selectively and attach
source locations to the claims they support rather than making paths the
reader's primary map.

Include invariants, edge cases, tradeoffs, and maintenance consequences when
they affect understanding. Call out misleading names or blurred boundaries as
teaching notes; do not turn the explanation into an unsolicited full review.

Finish with a coherent model the learner can use to reason about this area or
change. Highlight the most useful recall point when it adds value; avoid
repeating a short lesson merely to satisfy a closing section.
