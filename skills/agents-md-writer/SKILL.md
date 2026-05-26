---
name: agents-md-writer
description: "Create, edit, review, or improve AGENTS.md files for agentic coding tools. Use when changing AGENTS.md guidance, drafting a new file, migrating legacy instruction files, defining nested overrides in monorepos, or debugging tool discovery and precedence behavior."
---

# AGENTS.md Writer

Edit AGENTS.md guidance to be short, specific, and executable. Prefer repository facts over generic advice.
Default to editing existing instruction files; create new files only when no suitable file exists for the scope.

## Workflow

1. Identify target tools and audit instruction sources.
- Detect which agentic tools must read the file (for example Codex, OpenCode, Claude Code, Cursor).
- Inspect repo-root and nested instruction files already in use (`AGENTS.md` and tool-specific alternatives).
- Read the matching tool docs for discovery/precedence behavior before editing.
- Read [references/openai-codex-agents-md.md](references/openai-codex-agents-md.md) only when Codex behavior matters.

2. Prefer edit-in-place over scaffolding.
- Start from the closest existing instruction file and keep diffs minimal.
- Create a new file only when guidance is missing for a scope or precedence must change.

3. Collect project facts before writing.
- Read package manager scripts, CI workflows, linters/formatters, and test commands.
- Extract only commands and constraints that are true in the current repository.
- Add instructions only for recurring failures agents cannot resolve reliably from code discovery.

4. Draft the smallest useful change.
- Include only instructions an agent cannot infer quickly from code.
- Use imperative bullets with concrete commands and paths.
- Cover setup/build/test, code conventions, testing expectations, safety constraints, and change-management rules.
- Prefer positive direction over broad prohibitions when possible.
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
- Remove architecture summaries and generated context that tools can infer from source in seconds.

## Authoring Rules

- Prefer exact commands over abstract advice.
- Encode hard constraints: security boundaries, data policies, migration safeguards, approval gates.
- Keep architecture prose brief; link docs by path instead.
- Keep instructions concise to avoid truncation by byte limits.
- Resolve conflicts explicitly so nearest-file precedence is unambiguous.
- Keep guidance tool-agnostic unless a rule is truly tool-specific.
- Treat instruction text as a patch for repeated failures, not a full project explainer.
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

## Resources

- [references/openai-codex-agents-md.md](references/openai-codex-agents-md.md): Codex-specific discovery order, override behavior, fallback filename config, limits, and verification commands.
