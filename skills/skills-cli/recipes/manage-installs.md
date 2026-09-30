# Manage installed skills

Use the same scope and agent filters used for installation so inspection, updates, and cleanup act on the intended targets.

## List current installs

```bash
bunx skills list
bunx skills list -g
bunx skills list -g -a claude-code
```

## Remove a global skill from Claude Code

```bash
bunx skills remove find-skills -g -a claude-code -y
```

## Update installed skills

Use this for generic CLI-managed updates. Reinstall this repository's published skills from the remote source after their changes are committed and pushed.

```bash
bunx skills update
```

## Restore from the project lock file

```bash
bunx skills experimental_install
```

## Sync package-provided skills

```bash
bunx skills experimental_sync
```
