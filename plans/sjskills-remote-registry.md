# Read the skill registry from the published source

## Outcome

`sjskills` reads `skill-registry.json` from the published agent-scripts source
at the same commit as the skill contents it installs. Changing profiles, the
global baseline, targets, or skill entries takes effect once it reaches `main`,
without a new `sjskills` release. A release is needed only when the registry
schema or the CLI itself changes.

## Current state

Proposed on 2026-09-30. The step 1 spike is complete; implementation has not started.

The executable embeds `internal/sjskills/data/registry-v4.json`
(`internal/sjskills/registry.go`) and reads no other registry, while skill
contents come from `.../tree/main/skills` at run time. The two can drift
unnoticed: the CLI release notice compares version numbers only, and nothing
records which commit supplied the contents. `PROGRESS.md` records that syncing
a registry change on 2026-09-10 needed a temporary binary until v1.2.0 shipped it.

## Decisions

- **Source of truth.** The binary embeds the registry's location (repository,
  branch `main`, path `skill-registry.json`), not its contents. Remove the
  embedded registry data and the Node test that keeps it equal to the root
  file. Tests use fixtures through the existing `loadRegistry` seam.
- **One commit per invocation.** Resolve `main` to a commit once, fetch the
  registry at that commit, and pin the agent-scripts source for every skill to
  `.../tree/<commit>/skills`. Load the registry once per command and share it
  across `prepare`, status, and notices; today `plan`/`apply` load it three times.
- **Evidence.** Replace the `registry: embedded version 4` evidence with the
  registry commit and SHA-256. It then flows into the reviewed-plan semantic
  comparison unchanged. Global apply uses the commit recorded in the reviewed
  plan instead of re-resolving `main`, so a push between plan and apply cannot
  change the reviewed desired state. Project apply resolves fresh, as it has no
  artifact.
- **Caching and offline use.** Cache verified registries by commit in a new
  `registry/` namespace, reusing the status cache's bounded read, lock, and
  atomic-write helpers. `status`, `profiles`, and `init` may use the last
  verified registry when resolution fails and must label it stale with its
  commit and age. Status refreshes follow the existing 24-hour interval and
  15-minute cooldown. `plan` and `apply` always require a fresh resolution;
  they already need the network. With no cache and no network, status reports
  the registry unavailable, which is an existing supported state.
- **Schema compatibility.** Keep exact `RegistryVersion` equality. A fetched
  registry with a different version fails with a message that names the
  required `sjskills` update; never partially interpret it.
- **Resolution method.** The implementer chooses, subject to no required
  credentials, bounded time and size, and tolerance of unauthenticated GitHub
  API rate limits, since the startup hook runs status often. The hook's
  resolve-then-fetch in `plugins/sjskills-maintenance/scripts/observation.cjs`
  is a precedent.
- **No local override.** Reconciliation consumes only published content
  (`AGENTS.md`), so the CLI gets no flag to read a working-tree registry.
  Development validates the root file with `scripts/validate-skills` and the
  Node registry tests.
- **Unchanged.** Other external and authenticated sources keep their own refs.
  Provenance identity stays `github:owner/repo`, so pinning the ref does not
  change ownership of installed copies.

## Sequence

1. ~~Spike: confirm Skills CLI 1.5.23 installs a commit-pinned public source.~~
   Done 2026-09-30; see [Spike result](#spike-result). Materialize the
   agent-scripts source by fetching the commit with Git and passing the local
   `skills/` path to Skills CLI, generalizing the authenticated path's
   commit fetch to anonymous public sources.
2. Add registry resolution, fetch, validation, and cache with the fixed
   location; load once per invocation.
3. Pin agent-scripts sources to the resolved commit and add registry
   commit/digest evidence; make global apply reuse the reviewed commit.
4. Remove the embedded registry data and equality test; move Go tests to
   fixtures.
5. Update `docs/skill-registry.md`, `docs/sjskills-releases.md`,
   `skills/sjskills/` (including its global-rollout reference), and comments
   that describe the embedded registry.
6. Release a new minor version and install it on each machine under the
   existing release process.

## Spike result

Run on Windows on 2026-09-30 with the materializer's isolated environment and
Skills CLI 1.5.23, installing `next-goal` from `bbae6d1`, whose tree differs
from `main`:

- `.../tree/<40-hex commit>/skills` fails: Skills CLI clones with
  `--branch <ref>`, and Git reports the remote branch not found. A nonexistent
  commit fails the same way. `.../tree/main/skills` installs correctly.
- Anonymous `git init`, `fetch --depth=1 --no-tags <url> <commit>`, and
  `checkout --detach FETCH_HEAD`, then `skills add <checkout>/skills`, installs
  a tree identical to `git archive bbae6d1 skills/next-goal`.
- Anonymous `git ls-remote <url> refs/heads/main` resolves `main` without API
  rate limits, and `raw.githubusercontent.com/<owner>/<repo>/<commit>/skill-registry.json`
  returns the registry at that commit, matching `git show`.

Consequence: public agent-scripts materialization gains a Git prerequisite that
today applies only to authenticated sources. Resolution uses `git ls-remote`;
the registry fetch may use the raw URL or the same shallow checkout.

## Acceptance and validation

- Changing a profile on `main` changes the next plan's desired state with the
  same binary.
- A reviewed global plan applies against its recorded commit even after `main`
  moves, and a tampered commit or digest fails the reviewed-plan check.
- Offline `status`, `profiles`, and `init` use a cached registry with a stale
  label; with no cache they report the registry unavailable without failing.
- A registry with an unsupported version fails closed with an update message.
- Contents and registry in one plan come from the same commit.
- Full Go suite, Node tests, `scripts/validate-skills`, and native release
  checks on every release target pass.

## Out of scope

Registry schema changes beyond evidence, CLI self-update, changes to external
or private sources, and rollout of the new binary beyond the machines the
release step authorizes.

## Next action

Implement step 2: registry resolution, fetch, validation, and cache.
