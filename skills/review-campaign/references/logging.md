# Pass: logging

North star: in production, logs are the primary debugging tool. Pick the area's three most plausible incidents and ask: could you localize the cause from logs alone — no debugger, no repro? Review against those incidents, not against a checklist of log lines.

## Check

- **Decision points, not just failures.** Branches that choose behavior (fallbacks taken, retries scheduled, feature gates, cache hits/misses, degraded modes) log which way and why, with structured context. A failure log without the preceding decisions is a dead end.
- **Actionable failure sites.** Errors carry the ids and state needed to act: which workspace/workflow/request/provider, what input shape, what the code did next. "Something failed" + stack is not enough; neither is logging the error object without the operation.
- **Correlation.** Ids propagate across boundaries — UI process → IPC → main process, client → service request id, job/run id through workers. One incident must be traceable as one thread.
- **Level discipline.** `error` = somebody should act; `warn` = degraded but coping; `info` = lifecycle and decisions; `debug` = detail. Flag errors used for normal paths and info spam inside loops.
- **Noise audit.** Lines nobody could act on, double-logging the same failure at every layer (log where handled, propagate elsewhere), and per-iteration spam that buries signal.
- **Greppability.** Stable event names/keys over interpolated prose; consistent key names for the same concept across the area (and across service seams).
- **Sinks and survival.** Multi-process apps: do logs from every process reach a durable sink, with rotation? Crash-time logs flushed? Services: a structured logger with consistent keys, not raw prints.
- **Secrets.** Tokens, API keys, and user content (per the product's policy) must not be logged — record the finding under the security pass's area file if one exists; severity floor major.

## Do not flag

- Missing logs in pure functions and trivial glue.
- Verbose `debug`-level detail — that is what the level is for.

## Severity and tier

Undiagnosable critical incident path: major. Secret in logs: blocker. Auto tier only for adding an id already in scope to an existing log call, or fixing a message/key typo — level changes and new log points are triage.
