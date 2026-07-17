# Diagrams

Add a diagram only when it materially improves understanding of control flow, data flow, component relationships, state transitions, layered architecture, or a request lifecycle.

## Rules

- Use compact ASCII that renders clearly in plain Markdown; do not use Mermaid.
- Make each diagram answer one question.
- Prefer a request or event flow, module relationship map, state-transition sketch, or data-transformation pipeline.
- Put the diagram next to the explanation it supports and explain it briefly.

Example:

```text
Route -> Service -> Repo -> DB
```

Use this when the teaching problem is ownership or call flow rather than syntax.
