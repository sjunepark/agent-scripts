---
name: source-investigator
description: "Investigate external libraries, frameworks, and unfamiliar repositories from exact source in a project-local temp clone. Use when docs are incomplete, version-specific behavior matters, codebase learning is needed, or large source exploration should be delegated to focused subagents instead of polluting the main context."
---

# Source Investigator

Inspect real source when that is the most reliable way to answer the question. Keep the main thread focused on synthesis and decisions, not on carrying around large amounts of raw source code.

## Workflow

1. Decide whether source inspection is warranted.
- Prefer source over docs when the question depends on real control flow, defaults, edge cases, or version-specific behavior.
- Prefer docs first for public API overviews or setup instructions. If docs are still useful, have a subagent use `find-docs` rather than loading long documentation into the main thread.
- If the user named a specific repo, library, or framework version, treat that as a strong signal to inspect the exact source.

2. Create a project-local temp workspace before cloning.
- Work from the current project root.
- Create `.tmp/` if it does not exist.
- Create or update `.gitignore` so `.tmp/` is ignored. Do not assume a `.gitignore` file already exists.
- Keep disposable investigation clones under `.tmp/repos/`.
- Keep any generated notes or scratch outputs under `.tmp/work/`.

3. Clone the exact repository and ref you need.
- Clone into a clearly named directory such as `.tmp/repos/react` or `.tmp/repos/context7`.
- Prefer `git clone --depth 1` when current source is enough.
- If the question depends on a branch, tag, or commit, check out that exact ref and report it in the answer.
- Keep the clone disposable. Do not wire it into the current project or treat it as a dependency.

4. Delegate exploration to subagents.
- Do not inspect a large repository inline in the main thread unless a very small local excerpt is all you need.
- Spawn focused subagents with narrow questions such as:
  - where the entry point is
  - how configuration is loaded
  - which module implements a specific behavior
  - what changed between two refs
  - why a specific error or default occurs
- Tell subagents to return concise findings, supporting file paths, and only the minimum quoted code needed to justify the answer.
- Use subagents because their exploration stays out of the main thread's context window. The main thread should receive distilled findings, not raw repository contents.
- When docs matter, give a subagent the `find-docs` skill or have it read primary documentation while the main thread stays focused on the decision.

5. Synthesize the answer in the main thread.
- Combine the subagent findings into a direct answer.
- Cite concrete repo paths, symbols, and refs so the user can verify the conclusion.
- Separate direct observations from inference.
- Keep the answer focused on what the user asked, not on everything the exploration uncovered.

## Guardrails

- Do not clone into tracked source directories.
- Do not leave temp directories unignored.
- Do not paste large source files into the main thread just because they seem relevant.
- Do not use a single broad subagent task when several smaller, well-scoped questions would keep findings cleaner.
- Do not let subagents make the final user-facing conclusion when the main thread should synthesize the result.

## Example Triggers

- "Figure out how this library resolves config defaults. Clone the repo and inspect the source."
- "The docs are vague. Check the actual implementation for this behavior."
- "Investigate how this framework handles streaming responses in the current version."
- "Use a temp clone and subagents to learn how this repo is structured without polluting the main thread."
