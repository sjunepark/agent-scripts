# Whole-Application Orientation

Read this guide when the teaching target is exactly `project`, when the learner
clearly asks to understand an entire codebase or application before choosing a
narrower area, or when the learner continues from the learning map produced by
that first response.

## Goal

Give the learner a capabilities-first map of the running application and the
source-code boundaries that make it work. The first response is an orientation,
not a compressed explanation of every subsystem. It should leave the learner
able to choose what to study next.

## Inspect breadth-first

1. Establish what the application does, who or what uses it, and its core
   capabilities. Use high-level documentation for orientation when useful, but
   verify the model against source entry points and runtime wiring.
2. Identify the major runtime regions and their responsibilities, such as user
   interfaces, APIs, domain logic, persistence, background processing, and
   external integrations. Group by behavior and ownership rather than mirroring
   the directory tree.
3. Map capabilities to those source-code regions and name the important
   boundaries between them.
4. Trace one representative path end to end to confirm the map. Keep it at
   lifecycle depth; defer internal mechanisms.
5. Omit tests, build configuration, deployment, developer tooling, and
   documentation unless they materially change the application's runtime model.

Stop inspecting when you can explain the application's purpose, map its major
capabilities to runtime responsibilities without guessing, trace one confirming
flow, and offer a useful set of follow-up areas.

## First response

### Application at a Glance

Explain what the application enables, its primary actors, and its main
capabilities. Begin with user-visible behavior before naming code structure.

### System Map

Relate the major capabilities to the source-code regions that own them and show
the important runtime boundaries. Use a compact relationship diagram when it
materially improves the map; do not give a file-tree tour.

### One Representative Flow

Walk through one central request, command, event, or data lifecycle just deeply
enough to make the system map concrete.

### Learning Map

Offer a small numbered set of stable, descriptive topics that together cover
the application's important source-code areas. For each topic, state what the
learner would understand by selecting it. Include cross-cutting concerns only
when they materially shape application behavior.

### Where to Go Next

Invite the learner to choose a number or name a topic. Recommend one starting
point only when understanding it clearly unlocks the others.

## Follow-up turns

- Treat the selected area as the new focused teaching target. Briefly reconnect
  it to the system map, then explain that area without repeating the full
  orientation.
- Preserve established topic labels so the learner does not have to rebuild
  context between turns.
- If the learner refers only to a topic number or label that is unavailable in
  the current context, ask what it represented instead of guessing.
- If the selected area is still broad, teach its coherent top-level model and
  end with a smaller learning map for optional deeper exploration.
- Introduce adjacent areas only when they are required to understand the
  selected topic; otherwise leave them for later selection.
