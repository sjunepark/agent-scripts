---
name: brainstorming
description: "Brainstorm from diverse perspectives through fresh subagents, then synthesize their insights. Explicit request only."
---

# Brainstorming

Explore the invoking request through fresh, independent perspectives, then
synthesize a coherent landscape of ideas. Ground every worker in the topic,
goal, known facts, constraints, and desired outcome.

## Design and dispatch the perspectives

Choose at least three pairwise-distinct lenses suited to the topic. Examples
include stakeholder needs, feasibility, failure, assumption inversion, analogy,
second-order effects, and time horizons. Add a lens when it covers consequential
territory the others miss; avoid duplicating perspectives to increase the count.

Write each prompt independently so its questions and thinking operation embody
its lens. Include enough shared facts to work without conversation history,
limit the assignment to analysis and response, and request concrete insights
with their reasoning.

Spawn one fresh subagent per prompt with `fork_turns: "none"`. Omit `model` and
`reasoning_effort` so both inherit from the caller. Run independent prompts in
parallel when capacity allows; otherwise dispatch as slots become available.
Do not claim the requested fan-out occurred when fresh subagents are unavailable.

## Synthesize

Consider every worker's contribution. Explain the strongest ideas, unique
angles, recurring themes, tensions, and useful combinations. Preserve meaningful
disagreement and use parent judgment instead of treating consensus as proof.
Include open questions when another round would resolve something consequential.
The result should make the useful possibilities and tradeoffs clear without
repeating the workers' reports in full.
