---
name: ai-reviewer-setup
description: "Set up or update Greptile and CodeRabbit AI code review configuration in a project. Use when adding root greptile.json or .coderabbit.yaml files, tuning PR auto-review behavior, restricting Greptile by PR target branch, excluding generated files or bots, or documenting installation steps for these review bots."
---

# AI Reviewer Setup

## Workflow

1. Inspect the project before editing.
   - Check `git status --short` for pre-existing changes.
   - Look for existing `greptile.json`, `.coderabbit.yaml`, `.coderabbit.yml`, `.github/`, agent docs, style guides, generated-output directories, bot usage, and current branch/default-branch evidence.
   - Merge into existing config files rather than replacing working settings wholesale.

2. Load only the provider guidance needed for the requested change.
   - When adding or updating `greptile.json`, changing Greptile review cadence or
     target filters, or documenting Greptile installation, read
     [providers/greptile.md](providers/greptile.md) before editing.
   - When adding or updating `.coderabbit.yaml` or `.coderabbit.yml`, changing
     CodeRabbit auto-review behavior, or documenting CodeRabbit installation,
     read [providers/coderabbit.md](providers/coderabbit.md) before editing.

3. Confirm service installation boundaries.
   - Verify that each requested service's GitHub or GitLab integration is
     enabled, authorized, and includes the repository when local evidence can
     establish that state.
   - If local evidence cannot prove installation, complete the configuration
     change and leave a concise manual follow-up.

4. Choose a deliberate review cadence.
   - Prefer an initial automatic review plus manual follow-ups unless the user
     asks for continuous review.
   - Avoid configuring both services to review every push on the same PR; use
     the selected provider guides to express the cadence correctly.

5. Add only project-earned configuration.
   - Base exclusions on generated, vendored, build, coverage, snapshot, or
     lockfile paths that exist or are clearly produced by the repository.
   - Add project instructions or context only when existing style,
     architecture, or agent docs provide concrete review guidance.
   - Keep status checks, merge blocking, and request-changes behavior advisory
     unless the user explicitly asks for enforcement.

6. Validate before finishing.
   - Run each changed provider guide's syntax validation.
   - Run repository validation that covers the changed configuration files.
   - When the services are installed and a test PR is practical, verify target
     branch, draft, author, path, and update behavior affected by the change.

## Final Response

Report the files changed, each provider's effective target and update-review
behavior, validation results, and any remaining app-install or dashboard work.
Mention skipped service-level verification when no installed app or practical
test PR was available.
