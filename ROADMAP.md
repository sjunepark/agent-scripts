# Roadmap

## Current

_None._

## Plans

1. [Install private GitHub skills through authenticated profiles](plans/sjskills-private-github-sources.md)
   has a settled design; implementation and verification have not started.
2. [Roll out the fixed global skill baseline](plans/sjskills-global-rollout.md)
   is proposed and requires separate evidence-bound authorization for each
   machine. Exact-content approval binding is delivered.

## Tasks

_None._

## Completed

- [Prepare and spawn the next goal](tasks/next-goal-preparation-spawn.md)
  has implemented and evaluated source.
- [Include the sjskills CLI version in automatic status checks](tasks/sjskills-cli-version-status.md)
  shipped in `sjskills-v1.0.0` after validation on all supported native targets;
  installed-binary upgrades remain separate.
- [Reduce repeated sjskills status work](tasks/sjskills-status-performance.md)
  shares upstream evidence across matching selections and prefers the native CLI
  locally; implemented, reviewed, and verified with the updated local binary.
- [Show useful status when sjskills runs without a subcommand](tasks/sjskills-default-status.md)
  shipped in `sjskills-v1.0.0`.
- [Deliver `sjskills` v1](goals/sjskills-v1.md)
- [Build the profile-aware global reconciler](goals/profile-aware-global-skill-reconciler.md)
- [Bind global apply to reviewed expected-content evidence](goals/sjskills-global-rollout-approval-binding.md)
- [Deliver automatic project and global skill-status notices](goals/sjskills-status-notices.md)
