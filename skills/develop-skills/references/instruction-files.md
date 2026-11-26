# Instruction Files

Use when designing, auditing, or migrating `AGENTS.md`, `CLAUDE.md`, or other
rules-file hierarchies, or when an instruction is ignored or not loaded.

These facts were checked on 2026-11-27 against the
[Codex AGENTS.md guide](https://learn.chatgpt.com/docs/agent-configuration/agents-md)
and the [Claude Code memory guide](https://code.claude.com/docs/en/memory).
Both tools change their discovery rules often. When a decision depends on a filename,
precedence rule, limit, or version, re-check the guide first. Other clients
have their own rules; look them up rather than assuming these.

## Codex

- **Global:** in `$CODEX_HOME` (default `~/.codex`), Codex reads the first
  non-empty file of `AGENTS.override.md` and `AGENTS.md`.
- **Project chain:** Codex walks from the project root (usually the Git root;
  if there is none, only the current directory) down to the launch directory.
  In each directory it checks `AGENTS.override.md`, then `AGENTS.md`, then the
  names in `project_doc_fallback_filenames`, and it takes at most one file. So
  an override replaces that directory's `AGENTS.md` instead of adding to it.
  Empty files are skipped.
- **Nothing below the launch directory loads.** A nested `AGENTS.md` reaches
  only sessions launched at or below its directory.
- **Order:** files are joined from the root down, so closer files come later
  and win conflicts.
- **Size cap:** `project_doc_max_bytes` limits the combined size (default
  32 KiB). Once the cap is reached, Codex stops adding files, so the deepest,
  most specific files are the ones that get dropped.
- **Timing:** the chain is built once per run (in the TUI, once per launched
  session). Edits take effect only in a new session.

## Claude Code

- **Locations, in load order:** managed policy
  (`/Library/Application Support/ClaudeCode/CLAUDE.md`,
  `/etc/claude-code/CLAUDE.md`, `C:\Program Files\ClaudeCode\CLAUDE.md`), then
  `~/.claude/CLAUDE.md`, then the project's `CLAUDE.md` or `.claude/CLAUDE.md`,
  then `CLAUDE.local.md`. All of them are concatenated and none overrides
  another. Within a directory, `CLAUDE.local.md` comes after `CLAUDE.md`, and
  across directories files run from the root down.
- **Timing:** files in the launch directory and its ancestors load at launch.
  A subdirectory's file loads only when Claude reads, writes, or edits a file in
  that subdirectory. After `/compact`, Claude re-reads the project-root file
  from disk; nested files and path-scoped rules load again on demand.
- **Rules:** `.claude/rules/*.md` (project) and `~/.claude/rules/` (user) load
  unconditionally unless they have `paths:` frontmatter globs. `paths` is the
  only field Claude reads, and if the YAML doesn't parse the rule loads
  unconditionally.
- **Imports:** `@path` resolves relative to the importing file and nests at
  most four hops. It is skipped inside code spans and fences, and spaces in a
  path need a backslash. An import keeps the file organized but still loads
  everything into context. Imports from outside the working directory need
  approval.
- **Excluding files:** block-level HTML comments are stripped before injection.
  `claudeMdExcludes` (absolute-path globs, merged across settings layers) skips
  files but cannot exclude managed policy. CLAUDE.md files in an `--add-dir`
  directory load only if `CLAUDE_CODE_ADDITIONAL_DIRECTORIES_CLAUDE_MD=1` is set.
- **Size:** keep each file under about 200 lines; startup and `/status` warn
  when one file or the combined total runs long. Claude skips any file over
  4 MiB. Files that load on demand (subdirectory `CLAUDE.md` and their imports,
  subdirectory rules, and path-scoped rules) are cut off past a per-file
  character limit (40,000 characters with a 200K context window), and Claude
  sees only the beginning.
- **Built-in git guidance** can compete with commit or PR rules. Turn it off
  with `includeGitInstructions` and set the attribution text with
  `attribution`.

### AGENTS.md in Claude Code

- The **Project instructions** setting defaults to `claude-md-or-agents-md`
  and needs v2.1.277 or later. With that default, Claude reads every `AGENTS.md`
  and `.claude/AGENTS.md` at and above the working directory only when there is
  no `CLAUDE.md`, `.claude/CLAUDE.md`, or `CLAUDE.local.md` at or above it. A
  user or managed `CLAUDE.md` and `.claude/rules/` don't count toward that
  check. A subdirectory's `AGENTS.md` loads on demand when that subdirectory has
  none of the three CLAUDE files.
- Adding a personal `CLAUDE.local.md` quietly stops Claude from reading
  `AGENTS.md`.
- Claude never reads `AGENTS.override.md`, `AGENTS.local.md`, or anything under
  `.agents/`.
- The other values are `claude-md-and-agents-md`, `claude-md`, and
  `managed-only`. Set the value in `/config`, or in user, `--settings`, or
  managed settings at
  `pluginConfigs."cc-plugin-agents-md@builtin".options.instructionFiles`.
  Project and local settings files are ignored for this entry.
- To share one file, put a `CLAUDE.md` containing `@AGENTS.md` next to it, with
  Claude-only lines below the import. The import never loads the file twice.
  Avoid a `CLAUDE.md` symlink if anyone works on Windows: without
  `core.symlinks`, Git checks the link out as a one-line text file, and Edit and
  Write refuse to write through symlinks. A `CLAUDE.md` that only tells Claude
  in words to read `AGENTS.md` doesn't load it.
- An `AGENTS.md` that Claude reads through the setting doesn't fire
  `InstructionsLoaded` hooks and doesn't load from `--add-dir` directories.

## Placement across tools

- A shared file serves every tool only if each tool reads it in the actual
  layout. For example, a bare `AGENTS.md` reaches Claude Code only while no
  CLAUDE file sits on the path. `AGENTS.override.md` is Codex-only. A nested
  file reaches Codex only for sessions launched at that depth, while Claude
  loads it once it touches a file there.
- Place each instruction by when it must load:
  - Facts every session needs go in the root file.
  - Subtree rules go in the closest file that loads for the sessions that need
    them, or in a Claude path-scoped rule.
  - Procedures and occasionally needed knowledge go in skills.
  - Anything that must happen every time goes in a hook, linter, or CI.
- Claude's `/doctor` trims checked-in `CLAUDE.md` content that can be derived
  from the code, such as layouts, dependency lists, and architecture overviews,
  and keeps pitfalls, rationale, and non-default conventions. Apply the same
  test to `AGENTS.md`.

## Verify the effective chain

Re-read each affected chain in its tool's load order. Look for contradictions,
including with skills installed at that scope, stale commands, and size against
the limits above. When discovery or hierarchy changes, check from the root and
from a representative nested directory in a new session:

- **Codex:** run
  `codex --cd <dir> --ask-for-approval never "Summarize the current instructions."`.
  For a plaintext TUI log, add `-c log_dir=./.codex-log`.
- **Claude Code:** `/context all` lists the launch-time files under Memory
  files, and `/memory` lists the files it loaded, including an `AGENTS.md` read
  directly (v2.1.280 or later). A file that loads on demand prints a `Loaded`
  line when it loads, and an `InstructionsLoaded` hook can log every load.
  `/doctor` flags oversized files. `/doctor prompt-audit [path]` reports stale
  or conflicting instructions without changing anything.

Report any client behavior you could not confirm.
