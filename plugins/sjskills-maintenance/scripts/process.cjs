"use strict";

const fs = require("node:fs");
const path = require("node:path");
const { spawn } = require("node:child_process");
const LIMIT = 1024 * 1024;

function nativeExecutable(name, env = process.env) {
  const filename = process.platform === "win32" ? `${name}.exe` : name;
  for (const directory of (env.PATH || env.Path || "").split(path.delimiter)) {
    if (!path.isAbsolute(directory)) continue;
    const candidate = path.join(directory, filename);
    let fd;
    try {
      const resolved = fs.realpathSync(candidate);
      fd = fs.openSync(resolved, "r");
      if (!fs.fstatSync(fd).isFile()) continue;
      const bytes = Buffer.alloc(4);
      if (fs.readSync(fd, bytes, 0, 4, 0) !== 4) continue;
      // Reject source-building and package-manager wrappers without executing them.
      const magic = bytes.toString("hex");
      if (process.platform === "win32" ? magic.startsWith("4d5a") :
        ["cffaedfe", "cefaedfe", "feedfacf", "feedface", "cafebabe", "bebafeca", "7f454c46"].includes(magic)) {
        fs.accessSync(resolved, fs.constants.X_OK);
        return resolved;
      }
    } catch { /* An absent or unreadable PATH entry is not a runnable prerequisite. */ }
    finally { if (fd !== undefined) fs.closeSync(fd); }
    // The first existing command must not be bypassed to run a different install.
    if (fs.existsSync(candidate)) break;
  }
  throw new Error("native executable unavailable");
}

function runProcess(executable, args, { cwd, env = process.env, signal, limit = LIMIT }) {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) return reject(new Error("cancelled"));
    const windows = process.platform === "win32";
    const command = windows ? path.join(env.SystemRoot || "C:\\Windows", "System32", "WindowsPowerShell", "v1.0", "powershell.exe") : executable;
    const arguments_ = windows ? ["-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", path.join(__dirname, "windows-process.ps1")] : args;
    const child = spawn(command, arguments_, {
      cwd, env: windows ? { ...env, SJSKILLS_CHECK_PROCESS: JSON.stringify({ executable, arguments: args }) } : env,
      windowsHide: true, detached: !windows,
      stdio: ["ignore", "pipe", "pipe"],
    });
    let size = 0, stdout = [], failure, stopping;
    const killGroup = () => {
      if (!child.pid) return;
      try { process.kill(-child.pid, "SIGKILL"); } catch (error) { if (error.code !== "ESRCH") failure = error; }
    };
    const stop = () => {
      failure ||= new Error("cancelled or output limit exceeded");
      if (stopping || !child.pid) return;
      stopping = new Promise((done) => {
        if (process.platform !== "win32") {
          killGroup();
          done();
        } else {
          const killer = spawn(path.join(env.SystemRoot || "C:\\Windows", "System32", "taskkill.exe"),
            ["/PID", String(child.pid), "/T", "/F"], { windowsHide: true, stdio: "ignore" });
          const timer = setTimeout(() => { killer.kill(); child.kill(); }, 2000);
          killer.once("error", () => { clearTimeout(timer); child.kill(); done(); });
          killer.once("close", () => { clearTimeout(timer); done(); });
        }
      });
    };
    signal?.addEventListener("abort", stop, { once: true });
    for (const stream of [child.stdout, child.stderr]) stream.on("data", (chunk) => {
      size += chunk.length;
      if (size > limit) stop();
      else if (stream === child.stdout && !failure) stdout.push(chunk);
    });
    child.once("error", (error) => { failure = error; });
    // Parent exit ends the owned work even when descendants closed their pipes.
    if (!windows) child.once("exit", killGroup);
    child.once("close", async (code) => {
      signal?.removeEventListener("abort", stop);
      await stopping;
      if (failure) reject(failure);
      else resolve({ code, stdout: Buffer.concat(stdout).toString("utf8") });
    });
  });
}

module.exports = { nativeExecutable, runProcess, LIMIT };
