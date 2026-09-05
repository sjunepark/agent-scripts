# sjskills strict sync

Scope: enforce exact selections in global and project `.agents/skills` and
`.claude/skills`. Built-in skills, plugins, and legacy Pi roots stay outside
this boundary. Remove extras into recoverable quarantine; never claim ownership
of restored unknown or locally modified copies. Desired-path conflicts remain
blocking rather than silently overwriting local work.

- Reproduced warning-only preservation through isolated project/global CLI tests.
- Implemented removal classification and optional removal provenance through
  quarantine, restore, and crash recovery.
- Completed bounded code review and documentation harmonization. Fixed unsafe
  unselected-root handling and recovery-record collisions on restore retries.
- Validation passed: full Go suite on macOS, 12 registry/audit Node tests,
  skill/link validation, skill-creator validation, and diff whitespace checks.
  Isolated CLI tests cover project/global removal and restore; recovery tests
  cover interrupted apply/restore for unknown, modified, and owned copies.
- Windows test binaries compile successfully; native Windows execution remains
  unverified because no Windows test runner was available for this task.
- Implementation is complete on `codex/sjskills-strict-sync`. The user has
  requested committing and pushing all current changes, followed by `sjskills`
  sync. Full Go, Node, skill, and whitespace validation passed again before
  publication. Native Windows execution remains unverified.
- Registry-backed rollout requires the reviewed skill trees at the configured
  `main` source. Global mutation still requires the separate exact-evidence
  approval described in the rollout plan.
