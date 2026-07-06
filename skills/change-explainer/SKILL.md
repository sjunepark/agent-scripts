---
name: change-explainer
description: "Explain a code or document change set as a self-contained interactive HTML doc for reviewers and maintainers. Use to understand a git diff (pasted, staged, or unstaged), a commit or commit range, a PR patch, or two versions of a file or document."
---

# Change Explainer

Teach the change, not the patch mechanics — a concept-first, interactive HTML doc the reader can open in a browser, navigate, and expand at their own pace.

This is an explanatory skill, not an implementer tutorial and not a formal code review. Do not explain syntax or line-by-line execution unless asked.

## Explanation principles

Carry these from prose into the doc; they decide what each section says.

1. **Start from the reader's question.** Identify what they want to understand, not just which files moved.
2. **Build a mental model before hunk details.** Lead with purpose and visible behavior, then the main flow, then supporting edits.
3. **Teach the changed system, not just the delta.** Reconstruct the smallest useful model of how the code works *after* the patch; use only the minimum "before" needed to make the change legible. Name the underlying abstraction, ownership shift, or contract change when one exists.
4. **Group scattered hunks into a few conceptual buckets.** Name each bucket by its technical decision, not the file it touched. Split one large diff when it mixes concerns.
5. **Prefer `before → after → why`** when the contrast helps; otherwise explain the post-change design and then name the delta.
6. **Use evidence deliberately.** A before/after panel earns its place when the exact code shape carries the decision; a diagram earns its place when flow or ownership is easier seen than described. Do not paste whole functions or add evidence that only proves a file changed.
7. **Separate explanation from judgment.** Stay descriptive. Surface a concern only for correctness, regression, broken-contract, security/data, or a design tradeoff that materially affects future work. If there is none, say so and omit the concerns section.

## Doc structure

Fill the bundled template's sections (each maps to a region in [assets/template.html](assets/template.html)):

- **Big Picture** — one or two sentences on intent and where it matters, plus a TL;DR takeaway.
- **How the Change Works** — one collapsible block per conceptual change, numbered in learning order. Each has prose, an optional before/after panel, and a "why it matters" note. Add a diagram only when it clarifies flow, ownership, or state.
- **Key Decisions** — the few abstractions, invariants, or contracts that make the change cohere.
- **Reviewer & Maintenance Focus** — consequences for usage, compatibility, testing, and future edits; important tradeoffs and risks.
- **What to Remember** — the new mental model in one short paragraph or a couple of points.
- **Concerns** — include only if there is a material issue; otherwise delete the section and its TOC link.

## Workflow

1. **Identify the comparison unit.** Confirm whether you are explaining local changes, a commit, a range, a PR patch, or a file pair. When no target is given, do not ask first — default to uncommitted changes: run `git status --short`, then inspect staged and unstaged diffs and relevant untracked files; if the tree is clean, say so and ask for a commit, range, patch, or file versions. Read the diff first, then pull in neighboring code, tests, and docs only where the diff alone does not explain behavior, ownership, or invariants.
2. **Build the conceptual buckets** using the principles above. Done when every hunk in the diff is accounted for by a bucket; mechanical or trivial edits may share one briefly summarized bucket.
3. **Prepare the output directory.** Create `.change-explainer/` in the working directory. Make it self-ignoring without touching the repo's tracked files: ensure `.change-explainer/.gitignore` exists containing a single line `*`. (This ignores the whole directory in any git repo and leaves no tracked diff; create the dir even when not in a git repo.)
4. **Author the HTML from the template.** Read [assets/template.html](assets/template.html), keep its `<style>` and `<script>` intact, and replace the regions marked `<!-- FILL: ... -->`. Duplicate the conceptual-change `<details>` block once per bucket and number them. Remove any optional block (before/after panel, diagram, concerns section) you do not use, and keep the TOC links in sync with the sections you keep. Choose `language-*` classes on `<code>` to match the source language.
5. **Write the file** to `.change-explainer/<slug>.html`, where `<slug>` is a short kebab-case name for the change (e.g. `auth-session-refactor`). Reuse the same slug when regenerating the same change so it overwrites rather than accumulates.
6. **Open it (lavish-aware).** Open the file for the reader:
   - Default: `open "<path>"` on macOS (`xdg-open` on Linux).
   - If the user wants an annotate / feedback loop and the `lavish` skill or `lavish-axi` CLI is available, offer to open it with `npx -y lavish-axi "<path>"` instead, which serves the same file in a review session.
7. **Give a brief in-chat orientation.** State the big picture in two or three sentences and print the absolute path. Keep the depth in the doc; do not duplicate the full walkthrough in chat unless asked.

## HTML authoring rules

- Keep the doc self-contained: inline CSS and JS only; the sole external dependency is the optional highlight.js CDN, which the template already degrades from gracefully when offline.
- Prevent horizontal overflow at every level. Long unbreakable tokens belong in `<pre><code>` blocks (which scroll), not in prose.
- Keep before/after panels tightly scoped: the changed lines plus just enough context to read them. Excerpt a single branch, signature, or data shape rather than a whole function. For unified diffs, wrap added/removed lines in `<span class="add">` / `<span class="del">` inside a `<pre class="diff">` block.
- Diagrams are ASCII inside `figure.diagram > pre` by default. Use inline SVG only when an annotated picture genuinely helps. Do not pull in Mermaid or other diagram runtimes unless the user asks.
- Do not invent design systems or heavy chrome. The template's typography-led, low-border style is intentional; match it.

## Communication rules

- Optimize for reader understanding over exhaustiveness; do not collapse the doc into "file A changed X, file B changed Y" without naming the concept those edits serve.
- Do not use review language like "finding" or "severity" unless the user wants review mode.
- When inferring intent or tradeoffs, label it as an inference.
- If the change is tiny or the user wants a quick look, keep the doc short — fewer sections, fewer panels. If they want deep teaching, go further into mechanisms and tradeoffs.
