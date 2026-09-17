# Codex Invocation Policy

Use when creating or changing `agents/openai.yaml`.

Preserve an existing policy unless changing it is requested. For new skills,
apply the explicit-invocation preference and discovery criteria in the entry
point. Encode the chosen value explicitly:

```yaml
policy:
  allow_implicit_invocation: false
```

Use `true` for an intentional implicit-discovery opt-in. With `false`, Codex does
not inject the skill into context by default; explicit invocation can still load
it. State `Explicit invocation only` in a manual-only skill's description so its
intent is visible to other clients.

Keep interface metadata optional and consistent with the skill. Test explicit
invocation in either mode and uninvoked requests against the chosen policy.
Source metadata, installed metadata, and a running session are separate evidence;
use client discovery to verify installation behavior.
