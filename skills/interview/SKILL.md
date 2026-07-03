---
name: interview
description: "Interview the user to recover the intention their prompt left out and to clarify ambiguous points before doing the task. Use only when the user explicitly asks to be interviewed or asks for clarifying questions about their request; do not self-trigger on ambiguous prompts. Recovers intent and constraints; does not front-load technical decisions."
---

# Interview

A prompt is a compressed version of what the user actually wants. The motivation, scope boundaries, quality bar, and success criteria often live only in their head. Recover that omitted intention by asking, instead of guessing it or papering over it with technical decisions.

The interview succeeds when you can state the user's intent in a way they would endorse. It does not need to produce a full spec, and it must not turn into an upfront technical design session.

## Core Job

- Recover the intent the prompt left out: motivation, desired outcome, scope, constraints, success criteria, priorities.
- Resolve ambiguities whose answers change what you would build or do.
- Confirm the recovered intent with the user, then carry it into the task.

## Calibrate the Interview

An explicit invocation does not mean every part of the request is ambiguous. Scale the interview to the actual ambiguity and the cost of misunderstanding:

- Focus questions where interpretations genuinely diverge: multiple reasonable readings that lead to meaningfully different work, or missing motivation that would change the approach.
- Go deeper where misreading would be expensive: large changes, hard-to-reverse actions, outward-facing results.
- If the request is already well specified, say so and go straight to a short readback with at most a question or two.
- Do not manufacture questions to justify the invocation.

## What to Ask About

Intent-level questions are the job:

- Motivation: what prompted this now; what problem it solves; what happens if it is not done.
- Outcome: what done looks like; who or what consumes the result; how the user will judge it.
- Scope: what is in and out; this instance or the whole class; what must not change.
- Constraints: hard requirements, compatibility, quality bar, appetite (quick patch versus proper fix).
- Priorities: which way to lean when tradeoffs bite.
- Hidden context: prior attempts, known landmines, strong opinions the user already holds.

Technical decisions are not the job:

- Do not ask the user to pre-decide stacks, libraries, patterns, naming, file layout, or algorithms.
- Decide those during the work with normal judgment, or ask just-in-time when one genuinely blocks progress.
- Exception: ask upfront when a technical choice is really the user's to own — externally visible, hard to reverse, or one they signaled an opinion about.

## How to Ask

1. Look first, then ask. Extract everything the prompt, conversation, and obvious context already answer. Never ask what you can look up.
2. Anchor briefly. Open with one or two sentences on what you already understood and where the load-bearing gaps are, so questions read as grounded rather than generic.
3. Ask in small rounds of 2-4 questions, highest-leverage first. A question earns its slot only if its answer changes what you would do.
4. Pick the form per question:
   - Enumerable answers: use the agent's structured question tool when one exists; offer genuinely different options, put your best-guess default first, and keep an escape hatch for "none of these".
   - Open-ended intent, such as "what prompted this?": ask in plain language. Do not force options onto motivation questions; pre-baked options bias the answer.
   - No question widget available: ask numbered plain-text questions.
5. Ask one thing per question, phrased neutrally. Do not lead toward the answer you hope for.
6. Follow surprises. An answer that contradicts your assumption matters more than your planned next question.

## When to Stop

Keep interviewing until confident, not until a question count runs out. Confident means you could state the user's goal, motivation, scope, and success criteria and expect them to endorse the statement.

- Continue when an answer surprised you, two answers conflict, or the readback got corrected.
- Stop when new answers no longer change your understanding and the remaining unknowns are implementation-level or cheap to reverse.

## Readback, Then Proceed

1. Restate the recovered intent compactly: goal, motivation, in scope, out of scope, success criteria.
2. Label explicit assumptions for anything deliberately left open.
3. Ask for confirmation or correction. On correction, adjust and re-confirm only the corrected part.
4. On confirmation, proceed with the original task carrying the confirmed intent. When the user asked only for the interview, the confirmed readback is the deliverable.

Prefer the user's own words in the readback; introduce new terms only when they sharpen the statement.

## Non-Goals

- Not an upfront technical design session or a requirements waterfall.
- Not a stall: do not interview instead of working when a safe default exists.
- Not permission-seeking; "should I proceed?" is not an interview question.
- Not a completeness quiz; do not walk every question domain by checklist.

## Example Triggers

- "$interview I want a dashboard for my team."
- "Interview me about this feature before you start."
- "Ask me questions until you understand what I actually want."
