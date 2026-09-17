---
name: agents-md-writer
description: "Design or audit AGENTS.md and tool-specific instruction hierarchies, including scope, conflicts, migration, and discovery failures."
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
limits, or reload timing, verify current official tool documentation. For Codex,
start with the [AGENTS.md guide](https://developers.openai.com/codex/guides/agents-md/).
Do not assume other clients use the same rules.

## Place and revise guidance

- Keep shared guidance at the broadest intended scope inside the authorized
  target. User-global files are for explicitly personal or global guidance.
- Put subtree rules in the closest suitable existing file when nesting is
  supported. Add a file only for a distinct scope or intentional precedence change.
- Use a common cross-tool file only if every intended tool reads it. Codex's
  `AGENTS.override.md` intentionally replaces the same directory's `AGENTS.md`;
  verify other tools independently.
- Give duplicate or conflicting rules one authoritative owner where possible.
  Link to explanatory documentation; retain commands, invariants, and context
  needed to act. Consult scripts, CI, configuration, and docs for the claims
  being changed rather than requiring a full repository survey for every edit.
- State desired behavior directly. Preserve real authority boundaries and exact
  fragile steps; remove generic advice, redundant approvals, and unconditional
  reading that does not affect the task.

## Verify and finish

Re-read the affected chains in documented load order for contradictions and stale
commands. When hierarchy or discovery changes, inspect effective sources from
the root and a representative nested directory through supported introspection
or dry runs. Report changed scopes, validation, and any unverified client behavior.
An audit remains read-only; an authorized revision finishes through fixes and
verification.
