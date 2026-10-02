# Portability Contract

Use when creating a package or changing structure, metadata, dependencies, or
cross-client assumptions. Keep shared guidance usable in environments meeting
its declared prerequisites.

## Package

- Begin `SKILL.md` with portable YAML frontmatter containing `name` and
  `description`. Match the lowercase hyphenated name to its directory.
- Describe the capability, when it applies, and its selection boundary
  concisely. Preserve invocation
  intent under the entry point's policy, with client enforcement and UI settings
  in separate adapter metadata. The exception is a client that reads invocation
  policy only from frontmatter; follow its invocation guide for that one field.
- Keep runtime resources inside the package, without external symlinks,
  machine-specific paths, or parent-directory traversal.
- Name each runtime resource directly from `SKILL.md` with its use condition.
  Use exact, case-consistent relative paths; readers should not need recursive
  discovery or chains of references to find required guidance.
- Include scripts, assets, and references only when they serve the workflow.

## Runtime requirements

Describe outcomes, decisions, and constraints without assuming incidental host
features. A skill for a named tool, language, or platform may require its commands;
declare those prerequisites. Optional client metadata must not be required to
understand the procedure.

Use capability checks for optional facilities and provide a simpler fallback when
one preserves the result. If a missing prerequisite prevents correct completion,
hold the affected action and report its consequence; continue independent work.
Do not claim completion after silently skipping required work.

## Host-bound capabilities

Check the skill against every client it installs to, not only the one it was
written in. Invocation gating controls selection; it does not make required
tools, model settings, or commands exist in that client.

- When the whole capability depends on one client, narrow its install targets.
- When only a mode depends on one client, name that host in the mode and the
  description, check the capability before mutation, and state the fallback.
- When a named setting has an equivalent elsewhere, map it per client and
  report settings that could not be confirmed.

Refer to other skills by name in shared runtime text. Invocation syntax differs
between clients (`$name` in Codex, `/name` in Claude Code), and Claude Code's
model cannot load a skill that sets `disable-model-invocation`. Use host syntax
only in text generated for a known host, such as a prompt the user will paste.

## Third-party material

When merging or reusing text, code, templates, examples, or assets, inspect
provenance and applicable licensing. Retain required notices and attribution.
Do not infer copying rights from local availability or assume that restructuring
copied material removes its obligations.
