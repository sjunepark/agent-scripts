# Install and reconcile skills

Use the registry-backed reconciler for this repository's global profiles. Use
raw `bunx skills add` only for project-scoped work, local validation, or a
repository without `skill-registry.json`.

## Inspect this repo as a remote skill source

```bash
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --list
```

`./skills` is valid only for local source validation:

```bash
bunx skills add ./skills --list
```

## Reconcile one machine profile

Choose the profile explicitly. `dev` resolves `common + dev`; the current
public `kicpa` profile resolves the public `common + kicpa` entries in the
registry. Private KICPA source overlays are not supported by this command.

```bash
PROFILE="dev" # or kicpa

# Read-only exact-state audit. It may exit nonzero while printing the safe plan.
scripts/audit-global-skills --profile "$PROFILE"

# Install or update only unambiguous remote Skills CLI entries.
scripts/audit-global-skills --profile "$PROFILE" --apply

# Review remaining modified, unknown, manual, and legacy state.
scripts/audit-global-skills --profile "$PROFILE"
```

The command materializes desired remote content in a temporary home before it
reads the inspected roots. Apply uses copy mode and exact skill selection. It
passes an explicit Codex target for the shared `~/.agents/skills` copy and an
explicit Claude Code target for `~/.claude/skills`; it never passes a Pi target
or `--all`. The current CLI already places a raw explicit Codex global copy in
the shared root; the reconciler pins `HOME` and `CODEX_HOME` defensively so the
target cannot vary with caller configuration.

On apply, already-exact Skills CLI copies are adopted into
`~/.agents/.global-skill-state.json`. A future upstream change is eligible for
automatic update only while the installed tree still matches its recorded
hash. Before that verified update, apply moves the old tree into a
manifest-backed quarantine and prints its restore path. Local modifications
remain blocked for operator review.

If a preexisting copy differs from current remote content before it has a
verified record, ordinary apply leaves it unclassified. Review the audit's
recoverable replacement plan, then copy its exact approval command:

```bash
scripts/audit-global-skills --profile "$PROFILE" \
  --replace-unverified 'sha256:PRINTED_DIGEST' --yes
scripts/audit-global-skills --profile "$PROFILE"
```

The command writes the manifest and moves every old copy before installing the
same verified staged remote snapshots. A changed candidate set or verified
snapshot invalidates the digest. Whether installation fails partway through or
succeeds, first inspect every modeled
destination and move each active replacement aside. Then use the printed
restore command; restore never overwrites a destination.

## Publish before applying

For a changed skill from this repository, the registry source points at the
public GitHub `main` ref. Commit, push, merge, and pull the intended version
before apply. A local working tree or pushed feature branch is not enough.

```bash
git status --short
SKILL_NAME="skills-cli"
INTENDED_COMMIT=$(git rev-parse HEAD)
git push
# After merge to the registry source's pinned main ref:
git fetch origin main
git diff --quiet "$INTENDED_COMMIT:skills/$SKILL_NAME" \
  "origin/main:skills/$SKILL_NAME"
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --list
scripts/audit-global-skills --profile "$PROFILE" --apply
```

Compare every changed skill path. Tree comparison verifies publication after
merge, squash, or rebase without requiring the original commit object to be an
ancestor of `origin/main`.

## Quarantine verified legacy duplicates

Prune is intentionally separate from apply. First run the default audit and
review every printed source candidate plus its proposed restore path. Copy the
exact digest-bound command from that report: the digest covers the candidate
entries and source state, not the timestamped destination. Prune creates a
fresh run directory and revalidates every candidate immediately before moving:

```bash
scripts/audit-global-skills --profile "$PROFILE" \
  --prune 'sha256:PRINTED_DIGEST' --yes
scripts/audit-global-skills --profile "$PROFILE"
```

A changed candidate set invalidates the digest. Prune moves verified duplicate
children one by one into a new timestamped directory under
`~/.skill-quarantine`; it does not delete them. Modified, unknown, ambiguous,
unexpected managed, and externally managed entries are never prune operations.
Retain the manifest through a normal machine work cycle.

## Restore a quarantine

Use the manifest-specific command printed by prune:

```bash
scripts/audit-global-skills --restore \
  /absolute/path/to/.skill-quarantine/global-skills-TIMESTAMP/manifest.json \
  --home /absolute/path/to/home \
  --yes
```

Restore validates that the manifest and both paths remain inside the recorded
home and quarantine roots. It updates the manifest after every move and refuses
to overwrite an active destination.

## Project-scoped install

Install a project recommendation only when its registry `when` condition
matches. Omit `--global` and select the intended agent explicitly:

```bash
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills \
  --skill modern-rust --copy --agent codex --yes
```
