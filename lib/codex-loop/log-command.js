const fs = require("fs");
const path = require("path");
const { spawnSync } = require("child_process");

function runLogCommand(config, argv = process.argv.slice(2)) {
  try {
    main(config, argv);
  } catch (error) {
    console.error(`${config.commandName}: ${error.message}`);
    process.exitCode = 1;
  }
}

function main(config, argv) {
  const { command, runSelector, phasePrefix, options } = parseArgs(config, argv);
  const root = repoRoot();
  const runs = readRuns(root, config.loopName);

  if (command === "list") {
    printList(config, runs, options.json);
    return;
  }

  const run = selectRun(runs, runSelector);
  if (command === "path") {
    printPath(config, run, options.json);
  } else if (command === "show") {
    printShow(config, run, options.json);
  } else if (command === "transcript") {
    printTranscript(config, run, phasePrefix, options.json);
  } else if (command === "events") {
    printEvents(config, run, phasePrefix, options.json);
  } else {
    throw new Error(`unknown command: ${command}`);
  }
}

function usage(config) {
  console.log(`Usage: ${config.commandName} [command] [run-id|latest] [phase-prefix] [options]

Commands:
  list                         List saved ${config.loopName} runs.
  path [run-id|latest]         Print a run directory path.
  show [run-id|latest]         Show a run summary. Default command.
  transcript [run-id|latest] [phase-prefix]
                               Show readable transcript logs.
  events [run-id|latest] [phase-prefix]
                               Show useful JSONL events for a run or phase.

Options:
  --json                       Emit JSON for show/list/path/transcript/events.
  --help                       Show this help.

Examples:
${config.examples.map((example) => `  ${example}`).join("\n")}`);
}

function parseArgs(config, argv) {
  const options = { json: false };
  const positional = [];

  for (const arg of argv) {
    if (arg === "--help" || arg === "-h") {
      usage(config);
      process.exit(0);
    }
    if (arg === "--json") {
      options.json = true;
      continue;
    }
    if (arg.startsWith("--")) {
      throw new Error(`unknown option: ${arg}`);
    }
    positional.push(arg);
  }

  const knownCommands = new Set(["list", "path", "show", "transcript", "events"]);
  let command = "show";
  if (positional[0] && knownCommands.has(positional[0])) {
    command = positional.shift();
  }

  return {
    command,
    runSelector: positional[0] || "latest",
    phasePrefix: positional[1] || "",
    options,
  };
}

function runGit(args) {
  const result = spawnSync("git", args, {
    encoding: "utf8",
    maxBuffer: 10 * 1024 * 1024,
  });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    const detail = (result.stderr || result.stdout || "").trim();
    throw new Error(`git ${args.join(" ")} failed${detail ? `: ${detail}` : ""}`);
  }
  return result.stdout.trim();
}

function repoRoot() {
  return runGit(["rev-parse", "--show-toplevel"]);
}

function logsRoot(root, loopName) {
  const gitDir = runGit(["rev-parse", "--git-dir"]);
  const absoluteGitDir = path.isAbsolute(gitDir) ? gitDir : path.join(root, gitDir);
  return path.join(absoluteGitDir, loopName);
}

function readRuns(root, loopName) {
  const dir = logsRoot(root, loopName);
  if (!fs.existsSync(dir)) return [];

  return fs
    .readdirSync(dir, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => {
      const runDir = path.join(dir, entry.name);
      const stat = fs.statSync(runDir);
      return {
        id: entry.name,
        path: runDir,
        mtimeMs: stat.mtimeMs,
        mtime: stat.mtime.toISOString(),
      };
    })
    .sort((a, b) => b.mtimeMs - a.mtimeMs);
}

function selectRun(runs, selector) {
  if (!runs.length) return null;
  if (!selector || selector === "latest") return runs[0];

  const exact = runs.find((run) => run.id === selector);
  if (exact) return exact;

  const prefixMatches = runs.filter((run) => run.id.startsWith(selector));
  if (prefixMatches.length === 1) return prefixMatches[0];
  if (prefixMatches.length > 1) {
    throw new Error(`run selector is ambiguous: ${selector}`);
  }

  throw new Error(`run not found: ${selector}`);
}

function readJsonIfExists(file) {
  if (!fs.existsSync(file)) return null;
  try {
    return JSON.parse(fs.readFileSync(file, "utf8"));
  } catch (error) {
    return { parseError: error.message, path: file };
  }
}

function phaseFiles(run) {
  return fs
    .readdirSync(run.path)
    .filter((name) => name.endsWith("-parsed.json"))
    .sort()
    .map((name) => {
      const phase = name.replace(/-parsed\.json$/, "");
      return {
        phase,
        parsedPath: path.join(run.path, name),
        jsonlPath: path.join(run.path, `${phase}.jsonl`),
        stderrPath: path.join(run.path, `${phase}.stderr.log`),
        transcriptPath: path.join(run.path, `${phase}.transcript.log`),
        lastMessagePath: path.join(run.path, `${phase}-last-message.json`),
      };
    });
}

function summarizeRun(config, run) {
  const finalSummary = readJsonIfExists(path.join(run.path, "final-summary.json"));
  const phases = phaseFiles(run).map((phase) => ({
    phase: phase.phase,
    parsed: readJsonIfExists(phase.parsedPath),
    files: {
      parsed: phase.parsedPath,
      jsonl: fs.existsSync(phase.jsonlPath) ? phase.jsonlPath : null,
      stderr: fs.existsSync(phase.stderrPath) ? phase.stderrPath : null,
      transcript: fs.existsSync(phase.transcriptPath) ? phase.transcriptPath : null,
      lastMessage: fs.existsSync(phase.lastMessagePath) ? phase.lastMessagePath : null,
    },
  }));

  const summary = {
    id: run.id,
    path: run.path,
    mtime: run.mtime,
    finalSummary,
    phases,
  };

  if (config.includeStart) {
    summary.start = readJsonIfExists(path.join(run.path, "start.json"));
  }

  return summary;
}

function printList(config, runs, asJson) {
  if (asJson) {
    console.log(JSON.stringify({ runs }, null, 2));
    return;
  }
  if (!runs.length) {
    console.log(noRunsMessage(config));
    return;
  }
  for (const run of runs) {
    console.log(`${run.id}\t${run.mtime}\t${run.path}`);
  }
}

function printPath(config, run, asJson) {
  if (!run) {
    printNoRuns(config, asJson);
    return;
  }
  if (asJson) {
    console.log(JSON.stringify({ id: run.id, path: run.path }, null, 2));
  } else {
    console.log(run.path);
  }
}

function printShow(config, run, asJson) {
  if (!run) {
    printNoRuns(config, asJson);
    return;
  }

  const summary = summarizeRun(config, run);
  if (asJson) {
    console.log(JSON.stringify(summary, null, 2));
    return;
  }

  console.log(`Run: ${summary.id}`);
  console.log(`Path: ${summary.path}`);
  console.log(`Updated: ${summary.mtime}`);

  if (summary.finalSummary) {
    config.printFinalSummary(summary.finalSummary);
  } else {
    console.log("Final summary: missing");
  }

  if (summary.phases.length) {
    console.log("");
    console.log("Phases:");
    for (const phase of summary.phases) {
      const parsed = phase.parsed || {};
      const status = parsed.status || parsed.parseError || "unknown";
      const text = parsed.summary ? ` - ${parsed.summary}` : "";
      console.log(`- ${phase.phase}: ${status}${text}`);
    }
  }
}

function printEvents(config, run, phasePrefix, asJson) {
  if (!run) {
    printNoRuns(config, asJson);
    return;
  }

  const files = fs
    .readdirSync(run.path)
    .filter((name) => name.endsWith(".jsonl"))
    .filter((name) => !phasePrefix || name.startsWith(phasePrefix))
    .sort();

  const events = [];
  for (const file of files) {
    const fullPath = path.join(run.path, file);
    const lines = fs.readFileSync(fullPath, "utf8").split(/\r?\n/);
    for (const line of lines) {
      if (!line.trim()) continue;
      let event;
      try {
        event = JSON.parse(line);
      } catch {
        continue;
      }

      const item = event.item || {};
      if (item.type === "agent_message") {
        events.push({ file, type: "agent_message", text: item.text || "" });
      } else if (item.type === "command_execution") {
        events.push({
          file,
          type: "command_execution",
          command: item.command || "",
          status: item.status || "",
        });
      } else if (event.type === "turn.completed") {
        events.push({ file, type: "turn.completed", usage: event.usage || null });
      } else if (event.type === "turn.failed" || event.type === "error") {
        events.push({ file, type: event.type, message: event.message || event.error || "" });
      }
    }
  }

  if (asJson) {
    console.log(JSON.stringify({ id: run.id, path: run.path, phasePrefix, events }, null, 2));
    return;
  }

  if (!events.length) {
    console.log("No matching JSONL events found.");
    return;
  }
  for (const event of events) {
    if (event.type === "agent_message") {
      console.log(`\n[${event.file}] agent_message\n${event.text}`);
    } else if (event.type === "command_execution") {
      console.log(`[${event.file}] command ${event.status}: ${event.command}`);
    } else if (event.type === "turn.completed") {
      console.log(`[${event.file}] turn.completed ${JSON.stringify(event.usage || {})}`);
    } else {
      console.log(`[${event.file}] ${event.type}: ${event.message}`);
    }
  }
}

function printTranscript(config, run, phasePrefix, asJson) {
  if (!run) {
    printNoRuns(config, asJson);
    return;
  }

  const transcripts = fs
    .readdirSync(run.path)
    .filter((name) => name.endsWith(".transcript.log"))
    .filter((name) => !phasePrefix || name.startsWith(phasePrefix))
    .sort()
    .map((file) => {
      const fullPath = path.join(run.path, file);
      return {
        file,
        path: fullPath,
        text: fs.readFileSync(fullPath, "utf8"),
      };
    });

  if (asJson) {
    console.log(JSON.stringify({ id: run.id, path: run.path, phasePrefix, transcripts }, null, 2));
    return;
  }

  if (!transcripts.length) {
    console.log("No matching transcript logs found.");
    return;
  }

  for (const transcript of transcripts) {
    console.log(`\n== ${transcript.file} ==`);
    process.stdout.write(transcript.text);
    if (!transcript.text.endsWith("\n")) {
      console.log("");
    }
  }
}

function printNoRuns(config, asJson) {
  const message = noRunsMessage(config);
  if (asJson) {
    console.log(JSON.stringify({ error: message }, null, 2));
  } else {
    console.log(message);
  }
}

function noRunsMessage(config) {
  return `No ${config.loopName} runs found in this repository.`;
}

module.exports = {
  runLogCommand,
};
