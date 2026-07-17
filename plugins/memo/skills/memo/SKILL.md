---
name: memo
description: Save, restore, or clean a project-local memo for the current Codex session. Use only when the user explicitly invokes $memo, $memo write, $memo read, or $memo clean.
---

# Memo

Treat native Codex compaction and transcripts as primary. Use this skill only as a selective supplemental checkpoint; never save or restore a memo automatically.

## Route the operation

- No operation or `write`: read [references/write.md](references/write.md) completely and follow it.
- `read`: read [references/read.md](references/read.md) completely and follow it.
- `clean`: read [references/clean.md](references/clean.md) completely and follow it.
- Any other operation: report the supported operations without guessing.

Do not combine operations unless the user explicitly requests the combination.

## Resolve identity and helper

1. Resolve the helper only from this installed skill: from the directory containing this `SKILL.md`, the plugin root is `../..`, and the helper is `scripts/memo.py` under that root. Do not search for another copy.
2. Inspect developer context for `codex.runtime.session_id.v1=<opaque-id>`.
   - If one distinct non-empty marker is present, pass the entire marker to the helper in the `CODEX_MEMO_SESSION_MARKER` process environment. Repeated identical markers are equivalent.
   - If the marker is absent, leave that environment variable unset so the helper falls back to `CODEX_THREAD_ID`.
   - If marker values conflict or a present marker is malformed, stop with a routing error. Do not fall back, ask the user for an ID, or inspect existing memo names.
3. Invoke `python3 <plugin-root>/scripts/memo.py <operation>` on macOS or Linux, or `py -3 <plugin-root>\scripts\memo.py <operation>` on Windows. For the no-argument status view, omit `<operation>`.
   - Set `CODEX_MEMO_SESSION_MARKER` only for that process and pass the marker as inert environment data, never as an executable fragment or persistent export.
4. Consume the helper's TOON output. Use only its returned project root and path; never construct a memo path, expose the opaque ID, select by modification time, or search another repository.
5. On a helper error, preserve its boundary: do not modify ignore configuration, follow symlinks, or choose another memo.

Never print the marker, place the opaque ID in a memo, parse a transcript, or add compaction hooks or prompts.
