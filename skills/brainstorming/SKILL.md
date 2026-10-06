---
name: brainstorming
description: "Brainstorm from diverse perspectives through fresh subagents, then synthesize their insights. Explicit request only."
disable-model-invocation: true
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

Spawn one fresh subagent per prompt that does not inherit this conversation:
in Codex pass `fork_turns: "none"`; in Claude Code start a new general-purpose
agent rather than a fork. Workers use the caller's model and reasoning effort:
in Codex leave `model` and `reasoning_effort` unset; in Claude Code set `model`
to the caller's model, because an unset model can resolve to a configured
subagent default. Run independent prompts in
parallel when capacity allows; otherwise dispatch as slots become available.
Do not claim the requested fan-out occurred when fresh subagents are unavailable.

## Synthesize

Consider every worker's contribution. Explain the strongest ideas, unique
angles, recurring themes, tensions, and useful combinations. Preserve meaningful
disagreement and use parent judgment instead of treating consensus as proof.
Include open questions when another round would resolve something consequential.
The result should make the useful possibilities and tradeoffs clear without
repeating the workers' reports in full.
