# Plan Mode

Bootstrap or replan the campaign ledger.

1. Inventory tracked-file counts per directory, languages and frameworks (for example, Effect usage via import grep), test locations, colocated docs, existing review docs, and TODO conventions.
2. Write or refresh the repo profile. At first bootstrap, use the campaign branch,
   merge policy, and export convention established by the user's direction or
   applicable repository instructions; ask only for unresolved choices. Derive
   verification commands from repository scripts and agent instructions.
3. Propose or update the area table: paths, id prefix, pass applicability, priority, and `small` flag. Small areas batch all their passes into one session.
4. For every `✓ <sha>` cell, run `git diff --stat <sha>..HEAD -- <paths>` and flip changed cells to `~`, preserving the old sha.
5. Seed missing `reviews/` files. Apply row changes already authorized by the user
   and splits recorded by structure passes; ask before other changes to existing rows.

Finish with the resulting profile, area/matrix changes, stale cells, decisions still needed, and next-up cell. Do not begin a review pass in the same invocation unless the user explicitly requested both modes.
