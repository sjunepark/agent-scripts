# Skill registry

`skill-registry.json` is the authoritative classification and installation
policy for this repository's published skills and deliberately recommended
external skills.

## Contract

Each skill has one record that separates three concerns:

- `source` identifies where the skill definition comes from.
- `recommendation` identifies its intended scope, target agents, and, for a
  project recommendation, when it applies.
- `installation` identifies who manages the install and any required mode.

Recommendation scopes are:

- `global`: part of a named machine audience and desired for the compatible
  agents when the selected profile includes that audience.
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

The version 2 global contract defines two profiles:

- `dev` resolves `common` and `dev` audiences.
- `kicpa` resolves `common` and `kicpa` audiences.

Every global recommendation declares exactly one `audience`: `common`, `dev`,
or `kicpa`. Project and catalog recommendations cannot declare an audience.
Profile audience arrays are nonempty, sorted, deduplicated, and fixed to the
compositions above. Callers must select `dev` or `kicpa` explicitly; there is
no inferred or default profile.

`global.allowUnlistedSkills` is retained as an explicit strictness declaration
and must be `false`; installed skills outside the selected profile are drift.

`recommendation.agents` records the agents targeted by installation commands.
For exact global reconciliation, Codex- or Pi-compatible entries map to the
shared `~/.agents/skills` root through an explicit Codex target; Claude Code
compatibility maps to `~/.claude/skills` through an explicit Claude target.
The reconciler never synthesizes a Pi target or uses `--all`.

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
the supported scope and installation combinations.

Run `scripts/audit-global-skills --profile dev|kicpa` to compare the selected
profile with exact entries in the managed and explicitly known legacy roots.
The command materializes Skills CLI-managed remote sources in a temporary home
to establish expected content; it never uses a local path as an apply source.
The default mode is read-only and exits nonzero for missing, outdated,
modified, misplaced, unexpected, unclassified, or verified legacy duplicate
state. Runtime-owned roots are reported as protected and not enumerated.

`--apply` installs or updates only unambiguous Skills CLI-managed placements.
It also adopts already-exact placements into the reconciler-owned verified
state file; a later update is authorized only while the installed tree still
matches that record. Skills CLI v3 `skillFolderHash` values are source-tree
metadata, not verified local-content hashes. Apply does not prune.
When a first-run copy differs from remote content and has no trustworthy local
hash, the audit proposes `--replace-unverified --yes` instead of treating it as
an update. That separate boundary quarantines the existing copy with a restore
manifest before installing and verifying remote content. For rollback after a
partial or complete replacement, first inspect all modeled destinations and
move every active replacement aside; restore never overwrites a destination.
`--prune --yes` is a separate boundary that moves only
verified legacy duplicates into a timestamped quarantine. The output prints a
manifest-specific `--restore ... --yes` command, and restoration refuses to
overwrite an existing path. Manual entries remain with their recorded manager:
their exact placement is audited, their content is reported as externally
managed when no remote expected content exists, and no install command is
synthesized.

Project, workflow, and catalog-only records are excluded from profile
resolution. The public `kicpa` profile currently contains only public common
entries; private KICPA source overlays are not part of this registry contract.
