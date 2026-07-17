# Setup Workflow

Use this workflow only when the user explicitly asks to add Release Please.

1. Inspect the package or language ecosystem, existing release docs, CI and publishing flow, required secrets, branch protection, package manager, tag conventions, and current changelog/version files.
2. Identify where the repository expects release automation, configuration, and ownership documentation to live.
3. Propose or implement the smallest coherent Release Please configuration that fits those conventions. Do not assume publishing permissions exist.
4. Update release docs and agent instructions when they define release ownership or manual fallback steps.
5. Validate configuration and workflows with repository-provided checks or documented read-only tooling. Do not trigger release side effects as part of setup unless the user separately authorizes them under the hard gates in `SKILL.md`.
