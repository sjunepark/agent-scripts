---
name: clarify
description: "Resolve consequential gaps in the current request's intent. Explicit request only."
---

# Clarify

Make the request in the invoking prompt clear enough to carry out. Use earlier
conversation and available evidence to interpret it and avoid repeated questions;
do not import unrelated open decisions into the agenda.

Ask about missing information only when it could materially change the outcome,
scope, constraints, or success criteria and cannot be resolved from evidence or
delegated judgment. Investigate facts yourself. Leave routine implementation
choices to the implementer; surface technical choices when their consequences
are externally visible, costly to reverse, or reserved for the user.

Give the context needed to answer, then ask a small group of related questions.
Use concrete alternatives when they help the user recognize a preference, and a
structured question tool when available and appropriate. Explain consequential
blind spots without turning them into a checklist of hypothetical concerns.

Stop asking once the authorized task is clear enough to execute. A well-specified
request needs no question. Briefly state the recovered intent and any material
assumptions; ask for correction only when a consequential assumption remains
unresolved. Clear answers settle their items without another confirmation round.

Continue the original authorized task, holding only work that depends on an
unresolved answer. When clarification alone was requested, the readback is the
deliverable.
