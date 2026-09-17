# Whole-Application Orientation

Use for an exact target of `project`, an entire-application orientation, or a
follow-up from its learning map. The first response should let the learner
choose what to study next without compressing every subsystem into one lesson.

## Establish the application map

Identify the application's actors, core capabilities, and major runtime
responsibilities. Verify high-level documentation against entry points and
runtime wiring, then trace one representative request, command, event, or data
lifecycle to confirm how the parts connect. Group by behavior and ownership
rather than directory layout.

Inspect enough to explain capabilities, owners, and important boundaries
without guessing. Defer internal mechanisms and omit tests, build tooling,
deployment, and documentation unless they materially change the runtime model.

## First response

Give a capabilities-first overview with these useful elements; combine them
when the application is small:

- What the application enables and who uses it.
- A system map connecting capabilities to the source regions that own them,
  with a diagram when relationships need one.
- One representative flow at lifecycle depth to make that map concrete.
- A small numbered learning map with stable, descriptive topic labels and what
  the learner would understand by selecting each.

Invite the learner to choose a number or name a topic. Recommend a starting
point when understanding it clearly unlocks the others.

## Follow-up turns

Treat the selected area as a focused teaching target. Briefly reconnect it to
the system map, then explain its coherent model without repeating the full
orientation. Preserve established labels; if a referenced number or label is
unavailable, ask what it represented rather than guessing.

For a broad selected area, offer a smaller learning map after teaching its
high-level model. Introduce adjacent areas only when necessary to understand
the selection.
