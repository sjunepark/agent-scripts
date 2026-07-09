---
name: change-explainer
description: "Explain a code or document change set as a self-contained interactive HTML doc for reviewers and maintainers. Use to understand a git diff (pasted, staged, or unstaged), a commit or commit range, a PR patch, or two versions of a file or document."
---

# Change Explainer

Teach the change, not the patch mechanics — a visual-first, skimmable, interactive HTML doc the reader can open in a browser, navigate, and expand at their own pace.

This is an explanatory skill, not an implementer tutorial and not a formal code review. Do not explain syntax or line-by-line execution unless asked.

## Explanation principles

Carry these from prose into the doc; they decide what each section says.

1. **Start from the reader's question.** Identify what they want to understand, not just which files moved.
2. **Build a mental model before hunk details.** Lead with purpose and visible behavior, then the main flow, then supporting edits.
3. **Teach the changed system, not just the delta.** Reconstruct the smallest useful model of how the code works *after* the patch; use only the minimum "before" needed to make the change legible. Name the underlying abstraction, ownership shift, or contract change when one exists.
4. **Group scattered hunks into a few conceptual buckets.** Name each bucket by its technical decision, not the file it touched. Split one large diff when it mixes concerns.
5. **Two reading layers.** The collapsed doc must be a complete 30-second read: title, TL;DR, overview diagram, one skim sentence per change, decision list. Expanded bodies are the dive; nothing essential may live only in prose behind an expander.
6. **Show, then tell.** When the content is a flow, structure, sequence, ownership shift, or comparison, render it as a diagram, walkthrough, or table and let prose carry only what the visual cannot. A paragraph describing a shape the reader could be shown is a defect.
7. **Code is a last resort.** Default to no code; reference locations as `path:line`. Include a tightly-excerpted snippet or before/after code panel only when the exact code shape itself carries the decision and no visual or sentence can.
8. **Separate explanation from judgment.** Stay descriptive. Surface a concern only for correctness, regression, broken-contract, security/data, or a design tradeoff that materially affects future work. If there is none, say so and omit the concerns section.

## Doc structure

Fill the bundled template's sections (each maps to a region in [assets/template.html](assets/template.html)):

- **Big Picture** — one or two sentences on intent, a TL;DR takeaway, and an **overview diagram** (standard, not optional): the change's pipeline, flow, or ownership shift as SVG.
- **How the Change Works** — one collapsible block per conceptual change, numbered in learning order, collapsed by default. Each block: a **skim sentence** in the summary (visible without expanding), then a body that leads with its strongest visual, at most two short paragraphs of prose, and a "why it matters" note.
- **Key Decisions** — one line each: bolded decision, one clause of consequence.
- **Reviewer & Maintenance Focus** — consequences for usage, compatibility, testing, and future edits; important tradeoffs and risks. Tight bullets, or a table when comparing.
- **What to Remember** — the new mental model in one short paragraph or a couple of points.
- **Concerns** — include only if there is a material issue; otherwise delete the section and its TOC link.

## Choosing the visual

The template ships a component library; pick per change block when a visual is useful, and don't stack components:

| Content of the change | Component |
| --- | --- |
| Multi-step flow or pipeline | Guided walkthrough (step pills + SVG spotlight) |
| Structural / ownership / architecture shift | Before/after toggle holding two diagrams |
| Options, conventions, old-vs-new contracts | Comparison table |
| Single flow or data shape, no steps | Plain SVG diagram |
| Jargon the reader may not know | Term tooltip at first use, tip ≤ 2 sentences |
| Exact code shape carries the decision (rare) | Before/after code panel, tight excerpt |

A change block whose idea fits in its skim sentence plus one paragraph needs no visual at all — don't decorate.

## Workflow

1. **Identify the comparison unit.** Confirm whether you are explaining local changes, a commit, a range, a PR patch, or a file pair. When no target is given, do not ask first — default to uncommitted changes: run `git status --short`, then inspect staged and unstaged diffs and relevant untracked files; if the tree is clean, say so and ask for a commit, range, patch, or file versions. Read the diff first, then pull in neighboring code, tests, and docs only where the diff alone does not explain behavior, ownership, or invariants.
2. **Build the conceptual buckets** using the principles above. Done when every hunk in the diff is accounted for by a bucket; mechanical or trivial edits may share one briefly summarized bucket. For each bucket, pick its visual from the table above before writing prose.
3. **Prepare the output directory.** Create `.change-explainer/` in the working directory. Make it self-ignoring without touching the repo's tracked files: ensure `.change-explainer/.gitignore` exists containing a single line `*`. (This ignores the whole directory in any git repo and leaves no tracked diff; create the dir even when not in a git repo.)
4. **Author the HTML from the template.** Read [assets/template.html](assets/template.html), keep its `<style>` and `<script>` intact, and replace the regions marked `FILL:` in comments or placeholder attributes. Duplicate the conceptual-change `<details>` block once per bucket and number them. Copy components you need out of the template's COMPONENT LIBRARY into the change blocks, then delete the library and any optional block you did not use. Keep the TOC links in sync with the sections you keep.
5. **Write the file** to `.change-explainer/<slug>.html`, where `<slug>` is a short kebab-case name for the change (e.g. `auth-session-refactor`). Reuse the same slug when regenerating the same change so it overwrites rather than accumulates.
6. **Open it (lavish-aware).** Open the file for the reader:
   - Default: `open "<path>"` on macOS (`xdg-open` on Linux).
   - If the user wants an annotate / feedback loop and the `lavish` skill or `lavish-axi` CLI is available, offer to open it with `npx -y lavish-axi "<path>"` instead, which serves the same file in a review session.
7. **Give a brief in-chat orientation.** State the big picture in two or three sentences and print the absolute path. Keep the depth in the doc; do not duplicate the full walkthrough in chat unless asked.

## HTML authoring rules

- Keep the doc self-contained: inline CSS and JS only; the sole external dependency is the optional highlight.js CDN, which the template already degrades from gracefully when offline.
- Prevent horizontal overflow at every level. Long unbreakable tokens belong in `<pre><code>` blocks (which scroll), not in prose.

SVG diagrams:

- Style only through the template's classes — `.sn` (node), `.se` (edge), `.st`/`.st.sub` (text), `.sd` (arrowhead), with `.hi`/`.add`/`.del` variants — so diagrams follow the theme toggle. No hard-coded colors.
- Use a `viewBox` about 760 wide (height to fit); flow left→right; keep it to roughly 12 nodes. If a diagram wants more, split it across change blocks or make it a walkthrough and let the spotlight carry the complexity.
- Text must not overlap or escape its node: keep labels short, size nodes to their text, use a `.sub` second line for detail. Give each SVG's `<marker>` a unique id (ids are document-global). Set `role="img"` and a one-line `aria-label`.

Walkthroughs:

- 3–7 steps. Tag every SVG element belonging to a step with `data-step="n"` (group node + label in a `<g>`); assign edges to the step they lead into. Keep each walkthrough's marker and panel ids unique. Panels are ≤ 2 sentences each; the walkthrough replaces prose for that flow, it does not add to it.

Prose and code:

- Skim sentences ≤ ~20 words, concrete, no forward references ("uses X so that Y", not "see below").
- Each change body: visual first, then at most two short paragraphs. Move definitional asides into term tooltips instead of parenthetical prose.
- Code panels are tightly scoped: the changed lines plus just enough context to read them — a single branch, signature, or data shape, never a whole function. For unified diffs, wrap added/removed lines in `<span class="add">` / `<span class="del">` inside a `<pre class="diff">` block, and choose `language-*` classes to match the source language.
- Do not invent design systems or heavy chrome. The template's typography-led, low-border style is intentional; match it. Do not pull in Mermaid or other diagram runtimes unless the user asks.

## Communication rules

- Optimize for reader understanding over exhaustiveness; do not collapse the doc into "file A changed X, file B changed Y" without naming the concept those edits serve.
- Do not use review language like "finding" or "severity" unless the user wants review mode.
- When inferring intent or tradeoffs, label it as an inference.
- If the change is tiny or the user wants a quick look, keep the doc short — fewer sections, fewer visuals. If they want deep teaching, go further into mechanisms and tradeoffs.
