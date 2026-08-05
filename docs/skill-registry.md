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
The global audit requires those targets to be discoverable, but it does not
treat incidental visibility to other harnesses from a shared skill directory
as drift.

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
profile's `global` recommendations with `bunx skills list -g --json`. The
registry-v2 audit is read-only until exact filesystem reconciliation is
implemented. Project, workflow, and catalog-only records are excluded from
profile resolution. Global manual records remain desired and audited, but no
install command is synthesized for them.
