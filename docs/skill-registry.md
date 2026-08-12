# Skill registry

`skill-registry.json` is the authoritative classification and installation
policy for this repository's published skills and deliberately recommended
external skills.

## Contract

Each skill has one record that separates three concerns:

- `source` identifies where the skill definition comes from.
- `recommendation` identifies its intended scope, installation targets, and, for a
  project recommendation, when it applies.
- `installation` identifies who manages the install and any required mode.

Recommendation scopes are:

- `global`: part of a named machine audience and desired at its declared
  installation targets when the selected profile includes that audience.
- `project`: install only when the record's `when` condition matches a
  repository.
- `catalog`: published or tracked for discovery, but not recommended for
  installation by default.

Installation managers are:

- `skills-cli`: install from the recorded source with `bunx skills`.
- `manual`: audit the desired presence, but do not synthesize an install
  command because the source or procedure is maintained elsewhere.
- `workflow`: let the named workflow provision the skill in context.
- `none`: no default installation; required for catalog-only entries.

The manager and the source location are coupled. `skills-cli` clones git
shorthand (`owner/repo[/path]`) or credential-free `https:` remotes; it cannot
install from an npm specifier, another URL scheme, or a local path. A record
whose `source` location falls outside that set must therefore use `manual`,
`workflow`, or `none`, and validation rejects the `skills-cli` combination at
any scope.

For those records, `location` names the provisioning entry point rather than a
tree the Skills CLI could clone. A skill whose own installer compiles or
selects per-agent content is the usual reason: name-based skill discovery can
pick the wrong agent's variant or the uncompiled source, and the resulting
install looks successful while pointing at paths and command prefixes that do
not match the target agent. `impeccable` is the current example; the
`delegate-ui-to-claude` workflow provisions it with its own CLI, scoped to one
project and one agent.

The version 3 global contract defines two profiles:

- `dev` resolves `common` and `dev` audiences.
- `kicpa` resolves `common` and `kicpa` audiences.

Every global recommendation declares exactly one `audience`: `common`, `dev`,
or `kicpa`. Project and catalog recommendations cannot declare an audience.
Profile audience arrays are nonempty, sorted, deduplicated, and fixed to the
compositions above. Callers must select `dev` or `kicpa` explicitly; there is
no inferred or default profile.

`global.allowUnlistedSkills` is retained as an explicit strictness declaration
and must be `false`; installed skills outside the selected profile are drift.

`recommendation.targets` records installation destinations directly. The only
supported values are `.agents`, which maps to `~/.agents/skills`, and
`.claude`, which maps to `~/.claude/skills`. Global and project recommendations
must declare at least one sorted target; catalog recommendations declare none.
The Skills CLI adapter uses an
explicit Codex command target to populate `.agents` and an explicit Claude Code
command target to populate `.claude`; those client names are implementation
details rather than registry classification. The reconciler never synthesizes
a Pi target or uses `--all`.

## Ownership

The registry owns desired classification. Each `SKILL.md` owns skill behavior.
Installation and audit scripts are consumers of the registry; documentation
should explain policy without copying skill lists from it.

The registry is not an inventory of every system skill or every skill exposed
by an installed plugin. Include an external skill only when this repository
deliberately recommends or manages it. A command shown as an example does not
make its referenced skill a recommendation.

## Validation and consumers

Run `scripts/validate-skills` after changing the registry or published catalog.
It requires every `skills/*/SKILL.md` to have exactly one repository-source
record, rejects missing sources and incomplete classifications, and validates
the supported scope and installation combinations. It also rejects a
`skills-cli` record whose source location the Skills CLI cannot clone. That
check is total: the reconciler's runtime guard sees only the selected profile,
so validation is the only protection for project-scoped records.

Run `scripts/audit-global-skills --profile <dev|kicpa>` to compare the selected
profile with exact entries in the managed and explicitly known legacy roots.
The command materializes each Skills CLI-managed remote skill once in a
temporary home to establish expected content; apply reuses that exact verified
snapshot for every modeled placement in the run and never uses a local source
path.
The default mode is read-only and exits nonzero for missing, outdated,
modified, misplaced, unexpected, unclassified, or verified legacy duplicate
state. Runtime-owned roots are reported as protected and not enumerated.

`--apply` installs or updates only unambiguous Skills CLI-managed placements.
It also adopts already-exact placements into the reconciler-owned verified
state file; a later update is authorized only while the installed tree still
matches that record. Before a verified update, apply moves the old tree into a
manifest-backed quarantine and installs the staged snapshot at the now-absent
target. A failed or interrupted update therefore leaves the prior tree and a
restore manifest available; move any active replacement aside before restore.
Skills CLI v3 `skillFolderHash` values are source-tree
metadata, not verified local-content hashes. Reconciler tree hashes cover file
content, paths, symlink targets, and executable bits. Apply does not prune.
When a first-run copy differs from remote content and has no trustworthy local
hash, the audit prints an exact `--replace-unverified <sha256:digest> --yes`
command instead of treating it as an update. That separate boundary
quarantines the existing copy with a restore manifest before installing and
verifying remote content. For rollback after a partial or complete
replacement, first inspect all modeled destinations and move every active
replacement aside; restore never overwrites a destination. The printed
`--prune <sha256:digest> --yes` command is a separate boundary that moves only
the exact reviewed legacy duplicates into a timestamped quarantine. A changed
candidate set invalidates either digest, and a replacement digest also changes
when its verified remote snapshot changes. The output prints a
manifest-specific `--restore ... --yes` command, and restoration refuses to
overwrite an existing path. Manual entries remain with their recorded manager:
their exact placement is audited, their content is reported as externally
managed when no remote expected content exists, and no install command is
synthesized.

Project, workflow, and catalog-only records are excluded from profile
resolution. The public `kicpa` profile currently contains only public common
entries; private KICPA source overlays are not part of this registry contract.
