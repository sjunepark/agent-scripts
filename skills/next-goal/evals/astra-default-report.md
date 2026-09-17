# Astra default evaluation — 2026-09-17

Scope: change unqualified spawn to Astra/medium while preserving explicit
presets and reasoning overrides. Baseline: commit
`9c7e6b06d910cb3bdc475f666a836b712b12d655`. Candidate: the accompanying
entry-point and spawn-workflow changes.

Before trials, the acceptance criteria were fixed: bare spawn must resolve to
`gpt-6-astra`/`medium`; Luna must retain `gpt-5.6-luna`/`max`; Astra high must
retain `gpt-6-astra`/`high`. Two separate fresh workers read baseline and
candidate instructions and simulated argument resolution without creating tasks.

| Input | Baseline output: model / thinking | Candidate output: model / thinking |
| --- | --- | --- |
| `$next-goal spawn` | omitted / omitted | `gpt-6-astra` / `medium` |
| `$next-goal spawn luna` | `gpt-5.6-luna` / `max` | `gpt-5.6-luna` / `max` |
| `$next-goal spawn astra high` | `gpt-6-astra` / `high` | `gpt-6-astra` / `high` |

All three candidate assertions passed. These are single synthetic resolution
trials per condition, not live task launches or general reliability evidence.
No timing or token measurements were collected. The final wording narrows the
existing omit-thinking exception to explicit exact model identifiers; the
candidate reviewer confirmed the three resolutions still hold.

Independent bounded code review found a stale expectation in reusable case 24.
It now explicitly supersedes the historical omitted-model assertion. Frozen
2026-09-09 inputs and results remain unchanged. The README now agrees with the
runtime default. No other actionable findings were reported.

Static validation: `node scripts/validate-skills` and scoped `git diff --check`
passed. Authoring and portability checks pass for the changed scope: one
default, explicit overrides, existing Codex capability boundary, and no new
runtime resources or dependencies. Discovery is unchanged: the description and
adapter retain explicit-only invocation, so no new trigger trials apply.
