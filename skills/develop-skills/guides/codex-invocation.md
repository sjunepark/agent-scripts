# Codex invocation policy

Read this guide when a target skill includes `agents/openai.yaml` for Codex.

## Default

Preserve an existing skill's invocation policy unless changing it is requested.
For a new skill, use this repository's manual-only default unless the entry
point's full implicit-discovery gate passes:

```yaml
policy:
  allow_implicit_invocation: false
```

Also state `Explicit invocation only` in the portable description so the intent
remains visible to hosts that do not enforce the adapter setting.

In current Codex behavior, `false` keeps the skill out of the model-visible
catalog, so its name, description, and path do not occupy the initial skills
context. Explicit `$skill-name` invocation still loads the skill instructions
into that turn's context.

## Implicit discovery opt-in

Set `allow_implicit_invocation: true` only after evidence supports broad
recurrence within the installation scope, reliable prompt matching, safe,
useful activation without explicit intent, and value worth the catalog-context
cost.
Keep this decision independent from whether the skill is installed globally or
only for one repository.

Test explicit invocation in either mode. For manual-only skills, label in-scope
but uninvoked prompts as negative cases. For implicitly discoverable skills,
test those prompts as positives alongside ambiguous and near-miss negatives.
