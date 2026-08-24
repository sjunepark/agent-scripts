# Install and reconcile skills

Use `sjskills` for registry-backed project and global reconciliation in this
repository. Use raw `bunx skills add` for local source validation, explicitly
requested unmanaged installs, or repositories without an `sjskills` contract.

## Inspect this repository as a source

```bash
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --list
```

Use `./skills` only for local validation or unpublished work:

```bash
bunx skills add ./skills --list
```

## Reconcile committed project intent

Initialize once, commit `sjskills.toml`, and review before applying:

```bash
sjskills init dev go
sjskills plan
sjskills apply
```

Profiles are composable project selections. A direct third-party declaration
can be added without changing the central registry:

```toml
version = 1
profiles = ["dev", "go"]

[[direct]]
name = "third-party-review"
source = "example/third-party-skills/review"
full_depth = true
```

Treat `.sjskills/`, `.agents/skills/`, and `.claude/skills/` as generated
machine-local state only after reviewing any content already committed at
those paths. Unknown entries are preserved. Unmanaged desired paths, locally
modified managed trees, unsafe filesystem boundaries, and untrusted provenance
block the affected change.

An update or removal moves the trusted prior tree into manifest-backed
quarantine and prints an opaque identifier. Retain it through a normal work
cycle. Restore refuses to overwrite:

```bash
sjskills restore <quarantine-id>
```

## Inspect the fixed global baseline

There is one machine-independent baseline and no profile argument:

```bash
sjskills plan --global
```

`scripts/audit-global-skills` is only a read-only transition wrapper for this
plan. Its former profile, apply, prune, replacement, and path-based restore
interfaces are retired.

Real-home mutation is a separate operational rollout. Do not run these as
repository validation or infer authorization from a request to inspect:

```bash
sjskills apply --global
sjskills restore --global <quarantine-id>
```

Before a real-machine apply, require a separately reviewed evidence-bound plan
and explicit authorization. `sjskills` never adopts a preexisting desired
tree from byte equality alone and has no force-adopt or force-replace flag.
Former profile placements and legacy Pi copies remain report-only.

## Publish before reconciliation

The repository source points at public GitHub `main`. Commit, push, merge, and
pull the intended version before applying it. Verify every changed skill tree,
including after squash or rebase:

```bash
SKILL_NAME="skills-cli"
INTENDED_COMMIT=$(git rev-parse HEAD)
git fetch origin main
git diff --quiet "$INTENDED_COMMIT:skills/$SKILL_NAME" \
  "origin/main:skills/$SKILL_NAME"
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills --list
```

## Raw project-scoped install

When no `sjskills.toml` owns the project selection, omit `--global` and
select the intended agent explicitly:

```bash
bunx skills add https://github.com/sjunepark/agent-scripts/tree/main/skills \
  --skill modern-rust --copy --agent codex --yes
```
