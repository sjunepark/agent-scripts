# Plan Mode

Bootstrap or replan the campaign ledger.

1. Inventory tracked-file counts per directory, languages and frameworks (for example, Effect usage via import grep), test locations, colocated docs, existing review docs, and TODO conventions.
2. Write or refresh the repo profile. At first bootstrap, ask the user to choose the campaign branch and merge policy and the export convention; derive verification commands from repository scripts and agent instructions.
3. Propose or update the area table: paths, id prefix, pass applicability, priority, and `small` flag. Small areas batch all their passes into one session.
4. For every `✓ <sha>` cell, run `git diff --stat <sha>..HEAD -- <paths>` and flip changed cells to `~`, preserving the old sha.
5. Seed missing `reviews/` files. Confirm row changes with the user when rows already exist; otherwise apply only splits already recorded by structure passes.

Finish with the resulting profile, area/matrix changes, stale cells, decisions still needed, and next-up cell. Do not begin a review pass in the same invocation unless the user explicitly requested both modes.
