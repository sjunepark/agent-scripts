---
name: interview
description: "Interview the user to resolve consequential open items from the preceding conversation or the prompt that invokes $interview, then act on the confirmed decisions. Use only when the user explicitly invokes $interview or asks to be interviewed about findings or questions; do not self-trigger."
---

# Interview

Treat unresolved findings and questions from the relevant preceding conversation, plus any items named in the same prompt as `$interview`, as a decision agenda. Clear it at reviewer altitude, while leaving implementation details and strongly preferred answers to the implementer.

## Build and Triage the Agenda

1. Gather every unresolved finding or open question from the preceding conversation that the invocation refers to and from the invoking prompt itself. Honor any scope the current prompt sets, then group tightly related items.
2. Inspect the relevant conversation, workspace, and other available evidence. Finding facts is the agent's job; ask the user only for judgment they need to own.
3. Classify every agenda item:
   - **Implementer-owned:** an implementation detail or an item with a strongly preferred answer. Choose the answer and state the intended handling briefly.
   - **User-owned:** multiple reasonable answers remain and the choice materially affects direction, scope, user-visible behavior, an external contract, risk, or review policy.
   - **Deferred:** the user intentionally leaves the item open. Record what remains unresolved and, when useful, what would unblock it.

Only user-owned items enter the interview. A recommendation strong enough that a responsible implementer should simply take it is not a user decision.

## Discuss at Reviewer Altitude

Work through a small group of tightly related user-owned decisions at a time:

1. Give the concise background needed to decide: relevant evidence, dependencies, stakes, and tradeoffs.
2. Recommend an answer when the evidence supports a meaningful preference. If the recommendation is clearly dominant, reclassify the item as implementer-owned instead of asking.
3. Ask focused questions whose answers resolve the group. Use a structured question tool for genuinely enumerable choices when available; otherwise use numbered plain text.
4. Follow an answer only when it conflicts with another decision, remains materially ambiguous, or exposes a consequential prerequisite. Resolve prerequisites before downstream decisions, and stop when the remaining choices belong to the implementer.

## Check Consequential Blind Spots

After clearing the agenda, make one bounded pass for consequential issues neither side raised. Admit an issue only if it could materially change direction, scope, user-visible behavior, an external contract, or risk exposure. Leave minor issues and ordinary implementation concerns with the implementer. If no issue meets the bar, move directly to the readback.

## Confirm, Then Act

Give one compact decision readback that accounts for every original item and any admitted blind spot as:

- a confirmed decision,
- an implementer-owned action, or
- an explicit deferral.

Include rationale only where it preserves an important tradeoff, and label remaining assumptions. Ask for one confirmation of the complete decision set. Corrections update the affected items before confirming the revised set.

Use the confirmed readback as the gate for action. Then carry the decisions into the original task. When the user requested only an interview, the confirmed readback is the deliverable.

The interview is complete when every agenda item is accounted for and no consequential user-owned decision remains unresolved; it does not need to visit every possible branch.
