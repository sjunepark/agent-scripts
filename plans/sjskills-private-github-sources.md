# Install private GitHub skills through authenticated profiles

## Outcome

Projects install and update selected private GitHub.com skills through ordinary
`sjskills plan` and `sjskills apply`, using the machine's existing `gh` login.
Public skills retain their current fetching convention. Public and authenticated
profiles are separate, explicitly describe their access requirement, and compose
in the same project. Both use the existing verified installation, ownership,
quarantine, and recovery machinery.

## Current state

Design settled with the user on 2026-09-09; implementation has not started.
Repository evidence was inspected at `076a487`:

- [The registry](../skill-registry.json) defines central profiles; its copy in
  [the embedded registry](../internal/sjskills/data/registry-v4.json) ships in
  the CLI. [Go validation](../internal/sjskills/validate.go) and
  [JavaScript validation](../scripts/lib/skill-registry.js) restrict profiles
  to exactly `dev`, `go`, `kicpa`, and `rust`. Neither profiles nor direct
  declarations have access metadata.
- [Parsing](../internal/sjskills/parse.go) rejects unknown fields.
  [Resolution](../internal/sjskills/resolve.go) combines profiles and direct
  skills, rejects duplicate selections, and independently resolves the fixed
  global baseline.
- [Materialization](../internal/sjskills/materialize.go) uses pinned Skills CLI
  `1.5.23` in a temporary home. It accepts shorthand and credential-free HTTPS,
  retains ordinary environment variables, and redirects home/config paths.
  This accepts a private repository's syntax but does not establish reliable
  access through the user's `gh` configuration.
- [Status](../internal/sjskills/status.go) shares upstream evidence by fetch
  inputs. [Provenance](../internal/sjskills/provenance.go) records source and
  content ownership; temporary staging paths must never become source identity.
- Cleanup on 2026-09-10 recovered seven registrations from the private
  `sjunepark/kicpa` catalog on `codex/integrate-private-kicpa-skills`. That draft
  currently extends `kicpa` and is not ready to merge: move the registrations
  into the selected `kicpa-private` design after authenticated fetching is
  implemented and verified. Recovery did not validate private materialization.

## Next action

Start with the authenticated-fetch feasibility checkpoint below: inspect the
pinned Skills CLI's actual Git path and prove a narrowly scoped `gh` integration
under the isolated materializer environment. Then implement the connected
configuration, fetching, reconciliation, and acceptance results in this plan.

## Selected contract

### Selection and access

- Central profile metadata uses `access: public` or
  `access: github-authenticated`. Omission means `public` for existing input.
  Write explicit access metadata on shipped profiles once supported.
- Access is a fetching requirement, not an assertion about repository
  visibility. Do not query GitHub merely to prove that every repository in an
  authenticated profile is private. An authenticated profile applies its policy
  to all its managed skill sources; there are no per-member overrides.
- Keep the existing `kicpa` name and members public. Support an additional
  `kicpa-private` profile once actual sources are enrolled. A project can select
  `profiles = ["kicpa", "kicpa-private"]`; selecting only `kicpa` keeps the
  current behavior and does not introduce a `gh` dependency.
- Keep profiles centrally defined. Accept additional valid, nonempty profile
  names without changing a hard-coded list each time. Preserve the existing
  required public profiles, sorted selections, reference checks, and disjoint
  skill membership rules.
- Direct project declarations receive the same optional `access` field,
  defaulting to public. This gives private sources the existing direct-install
  route without requiring central enrollment or remotely hosted profile files.
- The fixed global baseline stays public and unchanged. Adding global private
  selection or machine-specific baseline policy is not required for this feature.
- Private-source names and URLs may appear in this public repository and the
  distributed binary. Credentials and private skill contents do not belong in
  the registry. External private repositories use the existing `external`
  source kind; `repository` currently means this checkout's published catalog.

### Fetching and installation

- Only authenticated GitHub.com selections use the new authentication path.
  Require available `gh` tooling and an existing login with repository access;
  do not start interactive login, change accounts, or edit persistent Git/gh
  configuration. Public selections retain the existing materialization path.
- Accept the supported GitHub shorthand and credential-free GitHub.com HTTPS
  forms. Preserve their existing subpath/ref semantics. Reject unsupported
  authenticated hosts and credential-bearing URLs before invoking a helper.
  Existing public-source URL support is not narrowed by this feature.
- Fetch into temporary staging, discover only the requested skills, and reuse
  current tree verification and final placement. Never clone directly into
  `.agents/skills` or `.claude/skills`, and never create a second installer or
  give the fetching process ownership of real installation directories.
- Preserve original remote source identity through staging, apply, provenance,
  quarantine, and restore. Changing only access policy does not make a trusted
  installation a different repository or authorize adopting an unmanaged copy.
- Missing tooling, missing login, denied access, and fetch failure produce
  actionable, sanitized errors. Report distinctions only when supported by
  evidence: a repository-not-found response may also mean denied access.
  A failing selected source prevents successful reconciliation of that scope;
  do not silently skip it, retry anonymously, or apply a partial desired set.
- Status retains its current freshness, failure cooldown, and independent-scope
  behavior. Cached hashes describe previously observed content, not current
  access permission. Explicit plan/apply still obtains fresh required evidence;
  ancillary status failure does not change a successful primary command's result.

## Implementation results

### Authenticated fetch integration

1. Inspect the pinned Skills CLI's Git implementation, source parsing, process
   environment, and local staging capabilities. Retrieve current official Git
   and `gh` documentation through the repository's documentation workflow before
   depending on exact helper syntax. Record the chosen mechanism and evidence.
2. Prefer a process-scoped Git credential-helper bridge backed by the existing
   `gh` login if it works through the pinned CLI. This keeps existing skill
   discovery and remote parsing intact. If the pinned CLI prevents that, use
   a bounded `gh` fetch/clone into private staging and an internal handoff to
   existing discovery. Select one implementation path after the experiment;
   do not ship both as speculative fallbacks or expose arbitrary local sources.
3. Preserve access to the selected machine `gh` configuration for the narrow
   authentication operation while retaining the materializer's isolated home
   and write boundaries. Merely inheriting an environment that redirects
   `HOME`/`XDG_CONFIG_HOME` is insufficient proof that existing login works.
4. Scope credential requests to the intended GitHub.com source. Do not forward
   credentials to other hosts, credential-bearing redirects, or fetch unrelated
   submodules. Do not copy the user's entire Git configuration into staging.
   Avoid shell interpolation and keep tokens out of argv, URLs, diagnostics,
   registry/manifest data, temporary credential files, and durable records.
   A credential helper may transmit credentials only over its private protocol
   pipe to Git; that stream must not become captured diagnostic output.
5. Integrate helper/clone subprocesses with existing output limits, deadlines,
   cancellation, Unix process groups, Windows jobs, and cleanup verification.
   Credential failure must not leave descendants or a reported-success snapshot.

Completion: the selected mechanism demonstrably reaches the authenticated Git
boundary with isolated configuration, returns verified requested trees, and
passes the failure/cleanup tests below. Public-only materialization invokes no
new `gh` operation. A machine's login is reused without changing it.

### Access model and profile discovery

1. Introduce a small validated access type in
   [the shared model](../internal/sjskills/types.go); normalize missing access
   to public at input boundaries and reject unknown or malformed values.
2. Update Go and JavaScript registry validators together, the strict parser,
   profile/direct resolution, and relevant fixtures. Relax the exact-profile
   restriction into validated central additions while retaining current public
   names and all selection invariants.
3. Carry effective access in resolved desired skills, never infer it from
   `-private`, and validate authenticated host/source combinations before
   subprocess work. Keep global resolution explicitly public.
4. Update `profiles` human and JSON output so users can see the access
   requirement before selecting a profile. Preserve existing direct manifest
   initialization/serialization behavior and document the new direct field.
5. Keep root and embedded registries synchronized. Mark current profiles public;
   use test fixtures to demonstrate `kicpa-private`. Do not add an empty profile,
   placeholder remote, or fabricated private skill to the production registry.

Completion: old public manifests and profiles behave as before; authenticated
central profiles and direct selections resolve through one access model;
mixed public/authenticated projects work without mixed policy inside a profile.

### Reconciliation, evidence, and compatibility

1. Include effective access in materialization batch and deduplication keys,
   desired-skill equality, status upstream cache keys, and reviewed-plan
   comparisons. Different access policies must not share a fetch or review
   approval accidentally. Do not include secrets or login identity in artifacts.
2. Audit serialization and comparison consumers of `DesiredSkill` and profile
   metadata. Old missing access means public; explicit public and omitted public
   have equivalent meaning. Document older executables' inability to read the
   new fields rather than silently dropping authenticated intent for them.
3. Preserve remote-only installed source identity and the existing provenance
   and quarantine contracts. Do not migrate installed ownership records merely
   because the fetch policy is new. Invalidate or version derived cache formats
   as necessary; do not reinterpret an old public cache as authenticated evidence.
4. Verify warm/cold status, unavailable/stale evidence, command-produced snapshot
   reuse, independent global status, and failure cooldown. Authentication failure
   must not publish a partial successful cache entry or authorize apply.
5. Keep strict sync, protected paths, locally modified trees, reviewed global
   approval binding, transactional rollback, and restore behavior intact.

Completion: access changes invalidate the relevant fetch/review evidence while
same-repository ownership remains valid. Legacy public inputs and installed
states remain usable; failures preserve managed contents and trusted state.

### Operator documentation and acceptance

Update [the registry contract](../docs/skill-registry.md) as the canonical owner
of access semantics, compatibility, source formats, and authentication behavior.
Update [README usage](../README.md) and
[the sjskills operator skill](../skills/sjskills/SKILL.md) with the public/private
profile distinction, direct declarations, dependencies, login/access error
remediation, and the fact that declared pointers are public. Change supporting
skill files only where their behavior is affected, preserving explicit invocation
and existing sync authorization boundaries. Explain that `kicpa-private` is an
example until actual catalog enrollment; do not claim it ships.

## Validation and completion

All automated execution uses temporary homes, projects, caches, and controlled
authentication fixtures. CI must not require a private-repository credential or
fall through from a fake helper to the user's real `gh`, Git, or Skills CLI.

| Obligation | Required evidence |
| --- | --- |
| Public behavior unchanged | Existing public registry/manifest fixtures, no new `gh` invocation for public fetching, same selected skills and installation targets |
| Explicit, composable access | Go/JS parser and validator parity; additional profile, omitted public default, unknown access, name/reference/collision errors, mixed-profile and authenticated-direct CLI fixtures |
| GitHub-only authenticated boundary | Shorthand/HTTPS/ref/subpath cases; non-GitHub host and malicious URL rejection; controlled helper proves credentials are not supplied to unrelated hosts |
| Real integration seam | Pinned Skills CLI experiment plus process-level integration of the chosen bridge/fetch path, including isolated `gh` configuration; a fake runner returning success alone is insufficient |
| Ordinary sync works | CLI E2E fixture: public plus authenticated selections, plan, install, upstream change, update, removal/quarantine, restore, and repeated no-op apply across both project targets |
| Failed access is safe and useful | Missing `gh`, missing login, denied/missing repository, network error, unavailable skill, timeout, cancellation, oversized output, and surviving-descendant cases; no partial placement or successful cache publication |
| Source/evidence correctness | Remote source remains in provenance; access separates batches/caches and invalidates reviewed plans; access-only change preserves same-repository ownership; old public state still works |
| Credentials remain transient | Sentinel-secret absence in user-visible/captured diagnostics, non-auth subprocess output, argv, temporary files, plans, caches, provenance, and quarantine; only the private helper-to-Git credential protocol pipe may transmit the secret; no persistent user config writes |
| Status and platform contracts survive | Fresh/stale/unavailable and cooldown fixtures, primary/advisory independence, global-public regressions, and native tests on every supported target |

Run focused tests during implementation, then the existing required checks:

```sh
go test ./internal/sjskills ./cmd/sjskills
go test -race ./internal/sjskills ./cmd/sjskills
go vet ./...
node --test scripts/lib/skill-registry.test.js scripts/audit-global-skills.test.js
scripts/validate-skills
python3 scripts/release_test.py
```

[Checks](../.github/workflows/checks.yml) and
[native artifact verification](../.github/workflows/release-artifacts.yml) own
integration validation. Supported runtime targets are defined by
[the release contract](../docs/sjskills-releases.md); Linux source checks alone
do not establish macOS/Windows authentication or process-cleanup correctness.
Run applicable native acceptance through the existing CI delivery path.

Complete one bounded `$code-review` for each reviewable implementation slice,
using subagents for shared/authentication contracts, and `$harmonize-docs changes`
after review. Resolve in-scope findings and record command/native evidence here.
Completion requires all four implementation results above, meaningful process
and CLI coverage, passing required checks, and accurate documentation. Report
controlled-fixture evidence separately from a live GitHub smoke test. A live
test is supplementary and only uses an explicitly identified, authorized source;
do not create a remote test repository or require undisclosed private inputs to
finish the implementation.

## Delivery boundary

Deliver the coherent feature through a reviewed PR with initial CodeRabbit
review and the existing required CI, preserving individual commits. The goal
contract owns the precise delivery lifecycle and completion checkpoint.

Stop before operational onboarding: enrolling the user's actual private skills,
publishing a new release, and installing or syncing it on real machines require
their concrete inputs and separate authority. GitHub Enterprise/other Git hosts,
privately hosted profile definitions, interactive login/account management,
visibility enforcement, optional-skill skipping, and a new global-private policy
are not part of this feature. Exact helper mechanics and minimal internal schema
handling are implementer-owned; no product decision remains open.
