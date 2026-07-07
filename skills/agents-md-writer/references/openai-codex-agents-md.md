# OpenAI Codex AGENTS.md Reference

Source documents:
- https://developers.openai.com/codex/guides/agents-md/
- https://agents.md/

Last checked: 2026-03-06

## Discovery and Precedence

- Build one instruction chain per run at startup.
- Load global instructions from `CODEX_HOME` (default `~/.codex`):
1. Use `AGENTS.override.md` if it exists and is non-empty.
2. Otherwise use `AGENTS.md` if it exists and is non-empty.
- Load project instructions from repo root down to current working directory:
1. `AGENTS.override.md`
2. `AGENTS.md`
3. Fallback names from config (`project_doc_fallback_filenames`)
- Load at most one non-empty file per directory.
- Merge files in root-to-leaf order; nearest directory instructions apply last.

## Limits and Filtering

- Ignore empty files.
- Stop loading project files when aggregate bytes reach `project_doc_max_bytes` (default `32768` bytes).
- Use Markdown only; no required schema.

## Config Knobs

Configure in `~/.codex/config.toml`:

```toml
project_doc_fallback_filenames = ["TEAM_GUIDE.md", ".agents.md"]
project_doc_max_bytes = 65536
```

- Use fallback filenames only when teams cannot standardize on `AGENTS.md`.
- Start a new Codex run after config changes.

## Verification Commands

Use these checks when discovery looks wrong:

```bash
codex --ask-for-approval never "Summarize the current instructions."
codex --cd /path/to/subdir --ask-for-approval never "Show which instruction files are active and why."
```

Then inspect logs:
- `~/.codex/log/codex-tui.log`
- `~/.codex/sessions/<date>/session-*.jsonl`
