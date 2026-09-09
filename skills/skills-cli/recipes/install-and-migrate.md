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

For selected profiles, initialize once, commit `sjskills.toml`, and review
before applying:

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

For direct-only selection, create `sjskills.toml` with `version = 1` and the
requested `[[direct]]` entries, omitting `profiles`; `init` requires a profile.
For an existing manifest, edit the requested declarations in place. Use actual
source skill names, sort direct entries by name, and avoid collisions with
selected profiles or the global baseline. Review `plan` before an authorized
`apply`; merely editing the manifest does not authorize installation.

Treat `.sjskills/`, `.agents/skills/`, and `.claude/skills/` as generated
machine-local state only after reviewing any content already committed at
those paths. Verifiable undeclared directories move into recoverable quarantine,
including unknown or modified copies. Unmanaged or locally modified desired
paths, unsafe filesystem boundaries, and untrusted provenance block apply.

An update or removal moves the prior tree into manifest-backed quarantine with
its applicable provenance and prints an opaque identifier. Retain it through a normal work
cycle. Restore refuses to overwrite:

```bash
sjskills restore <quarantine-id>
```

## Inspect the fixed global baseline

There is one machine-independent baseline and no profile argument:

```bash
sjskills --json plan --global > plan.json
plan_sha256=$(shasum -a 256 plan.json | awk '{print $1}')
```

`scripts/audit-global-skills` is only a read-only transition wrapper for this
plan. Its former profile, apply, prune, replacement, and path-based restore
interfaces are retired.

Real-home mutation is a separate operational rollout. Do not run these as
repository validation or infer authorization from a request to inspect:

```bash
sjskills apply --global \
  --approved-plan plan.json \
  --approved-plan-sha256 "$plan_sha256"
sjskills restore --global <quarantine-id>
```

For a requested configured sync, the agent prepares and reviews the evidence;
no separate user approval of the plan hash is required. Use one verified
executable throughout, and use `--yes` when the request already authorizes apply.
Global apply rejects missing or changed evidence and recomputes the complete
plan before mutation. `sjskills` never adopts a
preexisting desired tree from byte equality alone and has no force-adopt or
force-replace flag. Former profile placements inside managed roots are
undeclared extras; legacy Pi copies remain outside reconciliation.

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
