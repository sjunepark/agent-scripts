# Read

Restore only the memo routed to this session.

1. Run the helper's `path` operation and consume its TOON result.
2. If `exists` is false, state that this session has no memo and stop. Do not search other files or repositories.
3. Read the entire regular file at the returned path.
4. Reconcile it with the latest user instruction and current repository state. Treat the memo as a possibly stale checkpoint, not as authority; current user intent wins.
5. Give a short restoration acknowledgement that states the recovered goal or next action without echoing the marker or opaque ID.
6. Continue working only when the current user request explicitly asks to continue. A bare `$memo read` restores context and acknowledges it without automatically executing the recorded next action.

Do not parse native transcripts or expose secrets that an old memo should not have contained.
