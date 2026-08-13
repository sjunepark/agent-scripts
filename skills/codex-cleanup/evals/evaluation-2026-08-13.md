# Codex Cleanup trigger revision — 2026-08-13

## Intended behavior

`codex-cleanup` is manual-only. Codex may load it when the user explicitly
invokes `$codex-cleanup` or names the `codex-cleanup` skill, but must not select
it merely because a request concerns Codex storage, memory, tasks, releases, or
runtime processes.

## Candidate

- The portable description says `Explicit invocation only` while retaining the
  cleanup scope and near-miss boundaries.
- `agents/openai.yaml` sets `policy.allow_implicit_invocation` to `false`.
- All four positive trigger fixtures explicitly invoke or name the skill.
- Two in-scope but uninvoked cleanup requests are now near misses, alongside the
  four existing out-of-scope requests.

## Validation

- `scripts/validate-skills`: passed for 33 skills and the registry.
- `bunx skills add ./skills/codex-cleanup --list`: discovered exactly
  `codex-cleanup` and displayed the manual-only description.
- JSON parsing and `git diff --check`: passed.

The Codex adapter policy provides the deterministic manual-only gate. The
revised trigger fixture records 4 explicit positives and 6 near misses; a fresh
probabilistic metadata-classifier run was not needed to enforce the adapter
policy and was not performed.
