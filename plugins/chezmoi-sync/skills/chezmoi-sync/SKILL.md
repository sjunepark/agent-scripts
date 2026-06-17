---
name: chezmoi-sync
description: Review and synchronize chezmoi-managed local configuration from Codex. Use when the user asks to check chezmoi drift, inspect chezmoi status or diff, apply source changes, add target changes back to chezmoi, run chezmoi update, or commit/push reviewed chezmoi source changes. The bundled startup hook is read-only and only reports pending sync work.
---

# Chezmoi Sync

Use this skill to turn a chezmoi sync warning into reviewed, explicit action.

## Core rules

- Treat lifecycle hook output as informational only.
- Do not run mutating commands until the user has reviewed the relevant status or diff and confirmed the intended direction.
- Prefer target-specific commands when the user names files.
- Re-run the review after any mutating command.
- Keep `chezmoi update`, `chezmoi apply`, `chezmoi add`, `git commit`, `git pull --rebase`, and `git push` explicit.

## Scripts

Resolve script paths relative to this skill directory:

- `../../scripts/chezmoi-check.sh`: read-only summary used by the startup hook.
- `../../scripts/chezmoi-review.sh`: read-only review helper. Use `--diff` to include `chezmoi diff`; use `--fetch` only when the user wants fresh remote state.

Run the normal review first:

```bash
bash ../../scripts/chezmoi-review.sh --diff
```

## Workflow

1. Review state with `chezmoi-review.sh --diff`.
2. Explain the direction choices:
   - Keep local target changes: run `chezmoi add <target...>`.
   - Restore source state to this machine: run `chezmoi apply <target...>`.
   - Pull and apply source repository updates: run `chezmoi update`.
3. Ask for confirmation before running the chosen mutating command.
4. If committing source changes, inspect the chezmoi source path with `chezmoi source-path`, show `git -C <source> status --short` and the relevant diff, then ask for a commit message or propose one.
5. Before pushing, run `git -C <source> pull --rebase` when an upstream is configured, then `git -C <source> push`.
6. Re-run `chezmoi-review.sh --diff` and report the final state.

## Target aliases

Use these only as conveniences after resolving them for the command:

- `agents` or `AGENTS.md` means the relevant agent guidance file under the user's home configuration when chezmoi manages one.
- `~` and `~/...` paths should be expanded before passing them to shell commands.
