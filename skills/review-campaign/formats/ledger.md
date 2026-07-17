# Campaign Ledger Format

Read this file when creating or mutating campaign ledger state.

## Ledger Contract

```text
reviews/
  REVIEW.md       dashboard: repo profile, phases, next-up pointer, area config table, matrix, sync log
  areas/<id>.md   per-area: Map / Findings / Pass log
  decisions.md    append-only triage log (accept/reject + reason)
```

Matrix cell states: `·` pending | `✓ <sha7>` done at that commit | `~ <sha7>` stale (area changed since) | `—` not applicable.

`REVIEW.md` opens with the **repo profile**: the per-repo facts every session reads instead of assuming. Record the campaign branch and merge policy, export convention for accepted findings (TODO files, issue tracker, or another target), verification commands (typecheck/lint/test, targeted file lint/format, markdown checks), stack notes per area, agent-instruction rules to honor, and anchors (existing review docs to fold in, recorded keep-as-is verdicts, growth domains).

## Area File Template

```markdown
# <id> — review notes

paths: <from REVIEW.md config>

## Map
<structure pass writes this: responsibilities, dependency direction, public surface,
keep-as-is verdicts. Detail passes read it and do not re-litigate verdicts.>

## Findings

### <PREFIX>-001 · major · triage · open — <title>
- pass: structure @ <sha7>
- where: path:line[, path:line]
- problem: <what is wrong and why it matters>
- action: <concrete change>

## Pass log
- <date> · structure @ <sha7> · 2 findings (1 major, 1 nit)
- <date> · tests @ <sha7> · clean
```

## Finding Schema and State

- **id**: `<PREFIX>-<NNN>` from the area's prefix in `REVIEW.md`, sequential and never reused.
- **severity**: `blocker` (correctness, security, data loss) | `major` (will cost real time or quality) | `minor` (worth fixing opportunistically) | `nit` (mention once, never argue).
- **tier**: `triage` by default; `auto` only when every eligibility rule below holds, with a one-line `why auto:` justification.
- **status**: `open → accepted | rejected | fixed | obsolete`. Auto tier may go `open → fixed` directly.

Accepting a finding exports it with its id according to the repo profile. The exported entry becomes the work item; the campaign is done with it.

## Auto-Tier Eligibility

Default to `triage`. Mark `auto` only when **all** hold:

1. The blast radius is one file or one tight cluster, with no exported API, contract, or schema change and no behavior change beyond restoring obviously intended behavior.
2. Existing checks or tests prove the fix, or the defect is self-evident, such as a dead import, broken doc link, or log-field typo.
3. No design opinion is involved. If a reasonable maintainer could argue, use `triage`.
4. The fix does not remove intentional-looking code. Only trivially dead artifacts qualify, such as an unused import, unreferenced local, or byte-identical duplicate test.
5. The finding is not from the security pass; security findings are never auto.
