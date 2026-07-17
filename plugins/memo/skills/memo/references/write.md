# Write

Create or coherently rewrite the memo for this session.

1. Run the helper's `path` operation and consume its TOON result.
2. If `exists` is true, read the entire memo before drafting its replacement. If false, do not look for another memo.
3. Reconcile the latest user instruction, the current conversation, relevant repository state, and any existing memo. Current evidence and explicit user instructions override stale memo content.
4. Draft a compact Markdown checkpoint around current truth. Use only headings that help this session, drawn from:
   - goal and user intent;
   - constraints and exact wording whose nuance matters;
   - decisions and rationale;
   - rejected alternatives that still prevent repeated mistakes;
   - completed work and concise evidence;
   - unresolved questions or blockers;
   - the next concrete action.
5. Omit the session marker and opaque ID, secrets, raw transcripts, bulky logs, resolved history, duplicated material, and a chronological activity log.
6. Run `prepare` immediately before mutation. In Git, stop if the target is not already ignored; do not edit `.gitignore` or `.git/info/exclude`.
7. Replace the whole file at the returned path with the new checkpoint. Never append. Use the available file-editing tool and do not write through a symlink.
8. Verify the file contains the coherent replacement, then briefly report that the current-session memo was saved.
