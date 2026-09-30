# Roadmap

## Current

1. [Publish and activate the check-only sjskills hook](plans/sjskills-startup-check.md)
   is published to main and installed on the current Windows machine; configured
   skills are synchronized. User review through `/hooks` and fresh trusted-session
   acceptance remain pending.

## Plans

1. [Read the skill registry from the published source](plans/sjskills-remote-registry.md)
   is proposed so registry changes stop requiring an sjskills release.
   Registry resolution and caching are implemented; switching production
   loading, commit pinning, and evidence are next.

## Tasks

_None._

## Completed

- [Roll out the fixed global skill baseline](plans/sjskills-global-rollout.md):
  evidence-bound configured sync is delivered; per-machine runs keep their own
  execution evidence.
- [Install private GitHub skills through authenticated profiles](plans/sjskills-private-github-sources.md)
  is recorded as shipped in v1.3.0; live private-skill installation remains separate.
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
