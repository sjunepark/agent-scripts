# Pass: security

Trust-boundary areas only (set in REVIEW.md applicability). Every finding needs an attack story — who reaches this surface with what, and what they gain. Checklist items without a path are noise, not findings.

## Check by surface

- **Desktop shell** (Electron and kin, when present): window webPreferences (`contextIsolation` on, `sandbox` on, `nodeIntegration` off); preload exposes a minimal, typed API — never raw `ipcRenderer`/`shell`/fs handles; navigation and `window.open` constrained (`setWindowOpenHandler`, will-navigate); remote/untrusted content never rendered with privileges.
- **IPC** (privileged ↔ unprivileged processes): every handler validates payloads with its guards (verify guards are actually invoked, not just defined); sender/frame checks where it matters; the UI process is untrusted once it renders third-party or model-generated content; no handler proxies arbitrary paths/URLs/commands from UI input.
- **Process execution** (anything spawning runtimes or subprocesses): argument construction from user input, PATH/env hygiene, download verification for fetched runtimes, no shell-string interpolation.
- **Network services**: authn on every route (middleware coverage, not per-handler hope); tenant/owner scoping in every query touching tenant data; upstream API keys — storage, never logged, never echoed in errors; SSRF posture on configurable upstream URLs; request size/rate limits; webhook signature verification.
- **Web control plane**: authz enforced server-side (server routes, actions, API handlers), not in UI; auth-library session and CSRF configuration; payment/billing mutations idempotent and operator-gated; SQL parameterized throughout.
- **Secrets everywhere**: nothing committed (scan config, scripts, and fixtures for key-shaped strings); env handling in scripts and config; secrets absent from logs (pull anything the logging pass noted); `.env` files gitignored and documented.
- **db**: roles/grants least-privilege in migration scripts; schema verification asserts grants, not just object existence.

## Do not flag

- Attacks requiring an already-compromised local machine on the desktop app.
- Hardening that contradicts the product (a local tool that executes user-authored code does so by design — the boundary is *whose* code and *which* privileges).
- Dependency CVEs without a reachable path — note them in one line, don't inflate.

## Severity and tier

Severity floor is major; reachable secret exposure, missing authn/tenant scoping, or renderer→main escalation is blocker. **Blockers are surfaced to the user at session end, immediately — they do not wait for triage cadence.** Nothing in this pass is ever auto; fixes here deserve eyes.
