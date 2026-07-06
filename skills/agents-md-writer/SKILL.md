---
name: agents-md-writer
description: "Create, edit, or review AGENTS.md files for agentic coding tools. Use when migrating legacy instruction files, defining nested overrides in monorepos, or debugging tool discovery and precedence behavior."
---

# AGENTS.md Writer

Edit AGENTS.md guidance to be short, specific, and executable. Prefer repository facts over generic advice.

## Workflow

1. Identify target tools and audit instruction sources.
- Detect which agentic tools must read the file (for example Codex, OpenCode, Claude Code, Cursor).
- List every instruction file in scope — repo root and nested, including tool-specific alternatives — before editing.
- Read the matching tool docs for discovery/precedence behavior before editing.
- Read [references/openai-codex-agents-md.md](references/openai-codex-agents-md.md) only when Codex behavior matters: discovery order, overrides, fallback filenames, byte limits, verification commands.

2. Prefer edit-in-place over scaffolding.
- Start from the closest existing instruction file and keep diffs minimal.
- Create a new file only when guidance is missing for a scope or precedence must change.

3. Collect project facts before writing.
- Read package manager scripts, CI workflows, linters/formatters, and test commands.
- Extract only commands and constraints that are true in the current repository.
- Treat instruction text as a patch for recurring failures agents cannot resolve from code discovery; add nothing else.

4. Draft the smallest useful change.
- Use imperative bullets with concrete commands and paths.
- Cover the sections in the Recommended Skeleton below.
- If a behavior can be fixed in codebase structure/tooling, prefer that over adding instruction text.

5. Place overrides intentionally.
- Put shared defaults at repo root.
- Put subproject-specific guidance near the subtree it governs.
- Use `AGENTS.override.md` only when replacing same-directory `AGENTS.md` is intentional.
- Keep cross-tool guidance in common files; isolate tool-specific quirks to clearly labeled sections.

6. Verify and tighten.
- Remove duplicates from README or docs unless needed for execution.
- Remove vague guidance.
- Ensure each rule is observable, testable, or clearly enforceable.
- Validate behavior from repo root and a nested subdirectory with each target tool when possible.

## Authoring Rules

- Encode hard constraints: security boundaries, data policies, migration safeguards, approval gates.
- Keep architecture prose brief; link docs by path; cut summaries tools can infer from source in seconds.
- Resolve conflicts explicitly so nearest-file precedence is unambiguous.
- Avoid "pink elephant" phrasing (for example broad "do not use X") unless required for safety/compliance.

## Recommended Skeleton

```md
# AGENTS.md

## Setup commands
- Install deps: `...`
- Start dev env: `...`

## Build and test
- Full checks: `...`
- Fast checks for touched package: `...`

## Code conventions
- Formatter/linter commands: `...`
- Framework or language constraints: `...`

## Change expectations
- Add or update tests for changed behavior.
- Update docs in `...` when public APIs change.

## Safety and approvals
- Do not ...
- Ask before ...

## Tool-specific notes (optional)
- Codex: ...
- OpenCode: ...
```

## Debug Discovery

- Use each target tool's introspection or dry-run command to report active instruction files.
- Compare behavior from repo root and a nested subdirectory to confirm precedence.
- If source attribution is unclear, inspect tool logs/session traces when available.
- If instructions are missing, verify filename support, precedence rules, non-empty files, size limits, and restart requirements.
