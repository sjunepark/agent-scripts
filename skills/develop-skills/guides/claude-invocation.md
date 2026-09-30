# Claude Invocation Policy

Use when a manual-only skill installs to Claude Code, or when adding or removing
`disable-model-invocation`.

Claude Code has no adapter file; it reads invocation policy from frontmatter.
Without the field, Claude lists every installed skill for automatic selection,
so `Explicit invocation only` in the description is advice, not enforcement.
Enforce a manual-only policy with:

```yaml
disable-model-invocation: true
```

The skill then leaves the model's catalog and loads only when the user invokes
`/skill-name`. Clients without this control ignore the field. Keep it consistent
with other adapters: set it only when `agents/openai.yaml` also disables implicit
invocation, and never use it to make Claude stricter than the stated policy.

Omit the field when a shared instruction or another skill tells the agent to load
this skill by name, such as a global review follow-up or a goal-recovery route.
With the field set, Claude cannot load the skill for that route. Keep such a
skill model-invocable and let its description carry the manual-only intent.

Verify with the client's discovery path after installation: explicit
`/skill-name` loads the skill, and an uninvoked in-scope request does not.
