# Local changes

This skill vendors the official Impeccable skill and keeps local workflow guardrails in the same skill so there is one Impeccable entrypoint to maintain.

## Added guardrails

- `SKILL.md` includes a local operating contract for live-mode setup, helper/app server separation, runtime state hygiene, and cleanup expectations.
- User-facing command references point at `/impeccable-june`, while shell/package commands still use the upstream `impeccable` CLI where appropriate.
- `reference/live.md` includes a local live preflight that prioritizes repo-specific strict dev commands, rejects blind Vite default-port inference, documents common failure signatures, and defines source-control conventions for `.impeccable/live`.

## Upstream update rule

When updating from upstream, copy upstream changes first, then reapply the local sections marked as local workflow guardrails. Keep local edits small and focused on setup, lifecycle, and repo hygiene rather than design taste.
