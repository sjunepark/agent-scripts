#!/usr/bin/env node

const REMINDER =
  "Before stopping, update the active repo-local progress Markdown file with any material decisions, completed work, validation state, blockers, and the next step. If no progress file exists but this is long-running or unattended work, create a concise PROGRESS.md or PLAN-<slug>.md; keep it brief, then stop.";

function output(value) {
  process.stdout.write(`${JSON.stringify(value)}\n`);
}

async function readStdin() {
  const chunks = [];
  for await (const chunk of process.stdin) chunks.push(chunk);
  return Buffer.concat(chunks).toString("utf8");
}

async function handleStatus() {
  console.log("Progress reminder hook: enabled");
  console.log("Set CODEX_PROGRESS_REMINDER_DISABLED=1 to make the hook no-op.");
}

async function handleHook() {
  if (process.env.CODEX_PROGRESS_REMINDER_DISABLED === "1") {
    output({ continue: true });
    return;
  }

  const raw = await readStdin();
  if (!raw.trim()) {
    output({ continue: true });
    return;
  }

  let input;
  try {
    input = JSON.parse(raw);
  } catch {
    output({ continue: true });
    return;
  }

  if (input?.hook_event_name !== "Stop" || input?.stop_hook_active === true) {
    output({ continue: true });
    return;
  }

  output({ decision: "block", reason: REMINDER });
}

if (process.argv.length > 2) {
  const action = process.argv[2];
  if (action === "status") {
    handleStatus();
  } else {
    console.error("Usage: progress-reminder.mjs [status]");
    process.exitCode = 2;
  }
} else {
  handleHook().catch(() => output({ continue: true }));
}
