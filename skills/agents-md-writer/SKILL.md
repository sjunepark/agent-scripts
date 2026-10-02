---
name: agents-md-writer
description: "Design or audit AGENTS.md, CLAUDE.md, and other agent rules files and their hierarchies. Use for scope, conflicts, migration, bloat, and instructions that are ignored or not loaded."
---

# AGENTS.md Writer

Produce durable instructions at the scope that needs them. Ground additions in
non-obvious repository facts, recurring failures, hard constraints, or explicit
team decisions; prefer executable enforcement for mechanical rules.

## Establish the effective hierarchy

Identify the requested operation and target tools. Inspect the relevant global,
root, nested, tool-specific, and legacy sources and how they combine. A local
wording fix needs its affected instruction chain; a hierarchy migration needs
coverage of every candidate source and intended tool.

When a decision depends on filenames, discovery, precedence, overrides, size
limits, or reload timing, verify current official tool documentation. Start with
the [Codex AGENTS.md guide](https://learn.chatgpt.com/docs/agent-configuration/agents-md)
and the [Claude Code memory guide](https://code.claude.com/docs/en/memory). Do
not assume other clients use the same rules. Confirm these divergences against
those guides before relying on them:

- Codex builds its chain once at launch, from the project root down to the
  launch directory, with at most one file per directory. Files below the launch
  directory do not load, `AGENTS.override.md` replaces the same directory's
  `AGENTS.md`, and Codex stops adding files at a combined size cap.
- Claude Code concatenates ancestor `CLAUDE.md` files at launch and loads a
  subtree's file only when it reads a file there. By default it reads
  `AGENTS.md` only when no `CLAUDE.md`, `.claude/CLAUDE.md`, or
  `CLAUDE.local.md` exists at or above the working directory, and it never
  reads `AGENTS.override.md`. A `CLAUDE.md` that imports `@AGENTS.md` shares
  one file with both tools.

## Place and revise guidance

- Keep shared guidance at the broadest intended scope inside the authorized
  target. User-global files are for explicitly personal or global guidance.
- Put subtree rules in the closest suitable existing file when nesting is
  supported and that file loads for the sessions that need it. Add a file only
  for a distinct scope or intentional precedence change.
- Use a common cross-tool file only if every intended tool reads it in the
  target layout.
- Keep only what every session at that scope needs. Move multi-step procedures
  and occasionally needed knowledge to skills, path-specific rules to a nested
  or path-scoped file, and anything that must happen every time to hooks,
  linters, or CI.
- Give duplicate or conflicting rules one authoritative owner where possible.
  Link to explanatory documentation; retain commands, invariants, and context
  needed to act. Consult scripts, CI, configuration, and docs for the claims
  being changed rather than requiring a full repository survey for every edit.
- State desired behavior directly and concretely enough to verify. Preserve
  real authority boundaries and exact fragile steps. Remove a line when the
  agent would act correctly without it: facts derivable from the code, standard
  conventions, generic advice, redundant approvals, workarounds for older
  models, and unconditional reading that does not affect the task. Reserve
  emphasis for a line the agent has been observed skipping.

## Verify and finish

Re-read the affected chains in documented load order for contradictions, stale
commands, and size against each tool's documented limit or recommended length.
When hierarchy or discovery changes, inspect effective sources from the root and
a representative nested directory: in Codex, run
`codex --cd <dir> --ask-for-approval never "Show which instruction files are active."`;
in Claude Code, check Memory files in `/context`. For other clients, use their
documented introspection or a dry run. Report changed scopes, validation, and
any unverified client behavior. An audit remains read-only; an authorized
revision finishes through fixes and verification.
