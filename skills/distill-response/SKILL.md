---
name: distill-response
description: "Distill a prior AI response into a standalone explanation of its decision spine."
---

# Distill Response

Produce an executive compression: the shortest standalone restatement that
preserves the source's decision spine.

## Select the Source

1. Use the assistant response or responses identified in the invocation.
2. Otherwise, use the immediately preceding assistant response.
3. Apply any requested lens or length as an additional constraint. Let
   understanding determine the length when the user gives no limit.

Finish source selection when the exact material to distill is identifiable from
the conversation.

## Recover the Decision Spine

Read the complete source and recover the content that governs understanding:

- the central conclusion or outcome;
- the mental model: the important relationships, causes, or flow;
- consequential insights and technical decisions, including their rationale;
- material tradeoffs, consequences, constraints, and uncertainties;
- critical evidence or caveats that qualify the conclusion; and
- current state and next action when the source includes them.

Use only applicable categories. Preserve the source's level of certainty and
distinctions that would change a reader's interpretation or decision. Build the
decision spine only from source-supported claims.

Finish recovery when every consequential claim within the selected scope is
represented in the decision spine.

## Distill

Rewrite around the decision spine rather than the source's order. Collapse
examples into the principle they establish and repeated explanations into one
clear statement. Retain an example, implementation detail, command, path,
citation, or exact value only when it is needed to understand, verify, decide,
or act.

Prefer a compact paragraph for a simple response and a short, idea-grouped list
or sections when the relationships would otherwise become unclear. Lead with
the bottom line. Supply enough context for the result to stand alone, and emit
the distilled response directly.

Finish when a reader who has not reread the source can explain what matters,
why it matters, the consequential choices and tradeoffs, and what follows. Every
remaining detail must support that explanation; removing any further content
would weaken it.
