# Install private GitHub skills through authenticated profiles

## Outcome and delivery

PR #23 completes the private-source functionality recovered in PR #22 while
preserving local skill-authoring work and upstream history. The source now
supports private GitHub.com selections through ordinary project plan/apply.
Public selections and the fixed global baseline retain their fetching policy.
The [registry contract](../docs/skill-registry.md) owns access semantics,
compatibility, operator setup, and error handling.

Implementation, bounded independent review, local acceptance, and supported
native CI are complete for `84ba2e7`. [PR #23](https://github.com/sjunepark/agent-scripts/pull/23)
owns the final merge status. No release, actual private-skill installation,
real-machine reconciliation, or authentication setup is included.
The seven already-authorized recovered registrations are in `kicpa-private`;
`kicpa` retains its public members. Additional enrollment is outside this change.

## Implemented design

- [Access model](../internal/sjskills/access.go), strict parsers, Go/JavaScript
  validators, resolution, and profile discovery carry explicit fetching policy.
  Omitted access remains public; malformed access and unsupported authenticated
  hosts are rejected. Required public profiles stay public, while additional
  nonempty profiles obey the same collision/reference rules.
- [Authenticated staging](../internal/sjskills/github_auth.go) uses a controlled
  HTTPS Git clone and an exact-repository gh credential bridge. The existing
  pinned Skills CLI discovers from that internal local checkout. Public fetches
  do not enter this path. The original remote identity survives staging,
  ownership, quarantine, and restore.
- [Materialization](../internal/sjskills/materialize.go) batches and desired
  equality include effective access. [Status](../internal/sjskills/status.go)
  separates upstream evidence by access; [reviewed plans](../internal/sjskills/reviewed_plan.go)
  normalize legacy omitted-public fields without ignoring authenticated changes.
  Credential failure cannot publish partial expected content or bypass apply.
- Documentation and the bundled maintenance procedure describe profile/direct
  selection, prerequisites, migration refusal, and failure recovery. Older
  executables may reject the new fields; use one compatible binary for plan/apply.

### Authenticated-fetch checkpoint (2026-09-10)

- Inspected Skills CLI `1.5.23`: remote clone broadens allowed Git protocols
  and retries authentication through gh and SSH. Its local-source path performs
  discovery without invoking either tool. Controlled real-CLI experiments
  verified both boundaries with a local fixture and a sentinel credential.
- Selected one bounded HTTPS Git clone with an exact-repository credential
  helper, followed by the existing CLI's internal local-path handoff. Disable
  redirects, other protocols, submodules, inherited Git helpers/configuration,
  and tracing. Preserve the remote identity for provenance and evidence.
- `gh 2.100.0` migrates legacy configuration even during credential lookup.
  A controlled legacy fixture reproduced writes; a version-1 fixture remained
  unchanged. Check configuration read-only before invoking gh, refusing legacy
  or unsupported schemas. Environment-token login uses only a temporary schema
  file, with its token remaining in process environment/private credential pipe.
- The helper admits only HTTPS `github.com` requests for the selected repository;
  gh output goes to Git's private credential pipe with a fixed bound. Subprocess
  diagnostics are not included in authenticated-fetch failure reports.
- Sources: [Git credentials](https://github.com/git/git/blob/master/Documentation/gitcredentials.adoc),
  [Git process configuration](https://github.com/git/git/blob/master/Documentation/git-config.adoc),
  [gh configuration paths](https://cli.github.com/manual/gh_help_environment),
  [gh credential helper](https://github.com/cli/cli/blob/v2.100.0/pkg/cmd/auth/gitcredential/helper.go),
  [gh startup migrations](https://github.com/cli/cli/blob/v2.100.0/internal/ghcmd/cmd.go),
  and [configuration migration guard](https://github.com/cli/cli/blob/v2.100.0/internal/config/config.go).

## Acceptance evidence

All automated execution uses temporary homes/projects/caches and controlled
credentials. No private repository or real user login is needed by CI.

| Obligation | Evidence |
| --- | --- |
| Public behavior unchanged | Contract tests retain the five public KICPA members; public process test invokes neither Git nor gh authentication fixtures; baseline stays public |
| Explicit, composable access | Go/JS validation tests cover additional profiles, defaults, malformed access, required-public policy, GitHub-only sources, mixed selection, direct serialization and profile output |
| Credential boundary | Native fixture uses real Git credential machinery to deny other hosts/repositories before gh; supported lookup returns a sentinel only through the private pipe |
| Real integration seam | Pinned real-CLI experiments verify protocol/fallback behavior and local handoff without Git/gh; native acceptance uses real local Git clone and repository content through production helper processes |
| Ordinary sync works | Mixed CLI plan/install, upstream commit/update, repeated no-op apply, access-only ownership continuity, quarantine/removal and restore across both project targets |
| Failed access is safe | Missing tools, legacy configuration, denied login, oversized output, unavailable skill, cancellation and timeout fixtures; failed apply preserves installed contents |
| Descendant cleanup | Blocking native gh fixture is stopped by cancellation and timeout through Git → helper → gh before staging cleanup; generic process-tree tests retain staging when termination is unverified |
| Source and evidence | Original remote source in plans/provenance; access separates batch/cache/review identity; warm/cold/cooldown tests preserve public evidence when private refresh fails |
| Credentials remain transient | Sentinel absent from CLI diagnostics and project files; token login uses version-only temporary configuration and leaves original legacy config unchanged; helper output is bounded and never included in diagnostics |
| Platform compatibility | Full Go/race suites and supporting checks passed locally; hosted macOS Intel/ARM and Windows acceptance passed in run 34482396368 |

The independent review found no concrete implementation defect or exploitable
credential leak. Its requested full-chain cancellation/timeout coverage and
token-login coverage were added and passed. Fixtures never fall through to
real remote GitHub, gh credentials, or Skills CLI during automated tests.
A live private-GitHub smoke test is supplementary and was not performed.

## Delivery record

- Implementation: `84ba2e7`, retaining local authoring work and the original PR
  #22 recovery tip through merge ancestry.
- [Hosted acceptance](https://github.com/sjunepark/agent-scripts/actions/runs/34482396368)
  passed source checks, artifact builds, all supported native consumers, and the
  required aggregate gate. The final PR checks also cover documentation updates.
- CodeRabbit's requested initial review was skipped because automatic reviews
  are disabled. The bounded independent review and its acceptance follow-ups
  are complete; no external review findings remain.
- [PR #23](https://github.com/sjunepark/agent-scripts/pull/23) is the authoritative
  publication/merge record. No implementation or acceptance work remains.
- Release publication and machine onboarding require a separate request.

GitHub Enterprise/other hosts, privately hosted profile definitions,
interactive account management, visibility enforcement, optional-skill skipping,
and global-private selection remain outside the selected design.
