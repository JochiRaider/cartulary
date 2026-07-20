import { spawn } from "node:child_process";
import { createReadStream, createWriteStream } from "node:fs";
import { Transform } from "node:stream";

import { redactString } from "../contract/index.mjs";

function redactionTransform() {
  return new Transform({
    transform(chunk, _encoding, callback) {
      callback(null, redactString(chunk.toString()));
    },
  });
}

export function isDryRunFromMakeFlags(env = process.env) {
  const flags = ` ${env.MAKEFLAGS ?? ""} `;
  return flags.includes(" n") || flags.includes(" --just-print") || flags.includes(" --dry-run");
}

function sanitizeMakeFlags(value) {
  if (!value) {
    return "";
  }
  return value
    .split(/\s+/)
    .filter(Boolean)
    .filter((entry) => !entry.startsWith("--jobserver-auth="))
    .filter((entry) => !entry.startsWith("--jobserver-fds="))
    .filter((entry) => !entry.startsWith("--jobserver-style="))
    .filter((entry) => !entry.startsWith("-j"))
    .join(" ");
}

export function makeChildEnv(env = process.env) {
  const childEnv = { ...env };
  for (const name of ["MAKEFLAGS", "MFLAGS"]) {
    const sanitized = sanitizeMakeFlags(childEnv[name]);
    if (sanitized) {
      childEnv[name] = sanitized;
    } else {
      delete childEnv[name];
    }
  }
  return childEnv;
}

export function sanitizeLogName(value) {
  return value.replace(/[^A-Za-z0-9._-]+/g, "-");
}

export function runLifecycle(repoRoot, testOutputScript, args, stream = process.stdout, env = process.env) {
  return new Promise((resolve, reject) => {
    let closeStatus = null;
    let stdoutEnded = false;
    let stderrEnded = false;
    let settled = false;
    const maybeSettle = () => {
      if (settled || closeStatus === null || !stdoutEnded || !stderrEnded) {
        return;
      }
      settled = true;
      if (closeStatus === 0) {
        resolve();
        return;
      }
      const error = new Error(`${testOutputScript} ${args.join(" ")} exited ${closeStatus}`);
      error.exitCode = closeStatus;
      reject(error);
    };
    const command = testOutputScript.endsWith(".mjs")
      ? env.NODE_BIN || process.env.NODE_BIN || process.execPath
      : testOutputScript;
    const commandArgs = testOutputScript.endsWith(".mjs") ? [testOutputScript, ...args] : args;
    const child = spawn(command, commandArgs, {
      cwd: repoRoot,
      env,
      stdio: ["ignore", "pipe", "pipe"],
    });
    child.stdout.pipe(redactionTransform()).pipe(stream, { end: false });
    child.stderr.pipe(redactionTransform()).pipe(process.stderr, { end: false });
    child.stdout.on("end", () => {
      stdoutEnded = true;
      maybeSettle();
    });
    child.stderr.on("end", () => {
      stderrEnded = true;
      maybeSettle();
    });
    child.on("error", (error) => {
      if (settled) {
        return;
      }
      settled = true;
      reject(error);
    });
    child.on("close", (status) => {
      closeStatus = status ?? 1;
      maybeSettle();
    });
  });
}

function commandOptions(value) {
  if (
    value &&
    typeof value === "object" &&
    (Object.hasOwn(value, "env") || Object.hasOwn(value, "signal") || Object.hasOwn(value, "timeoutMs"))
  ) {
    return {
      env: value.env ?? process.env,
      signal: value.signal ?? null,
      timeoutMs: value.timeoutMs ?? 0,
    };
  }
  return { env: value ?? process.env, signal: null, timeoutMs: 0 };
}

function terminateChild(child, signal) {
  if (child.exitCode !== null || child.signalCode !== null) {
    return;
  }
  try {
    if (process.platform === "win32") {
      child.kill(signal);
    } else {
      process.kill(-child.pid, signal);
    }
  } catch (error) {
    if (error?.code !== "ESRCH") {
      child.kill(signal);
    }
  }
}

export function runCommand(repoRoot, command, args, logFile, rawOptions = process.env) {
  const options = commandOptions(rawOptions);
  return new Promise((resolve) => {
    const log = createWriteStream(logFile);
    let settled = false;
    let terminationReason = null;
    let terminationStatus = null;
    let killTimer = null;
    let timeout = null;
    const finish = (status) => {
      if (settled) {
        return;
      }
      settled = true;
      if (timeout) clearTimeout(timeout);
      if (killTimer) clearTimeout(killTimer);
      options.signal?.removeEventListener("abort", abort);
      log.end(() => resolve({
        status: terminationStatus ?? status,
        terminationReason,
      }));
    };
    const terminate = (status, reason, signal = "SIGTERM") => {
      if (terminationStatus !== null) return;
      terminationStatus = status;
      terminationReason = reason;
      log.write(`[scheduler] ${reason}; forwarding ${signal}\n`);
      terminateChild(child, signal);
      killTimer = setTimeout(() => terminateChild(child, "SIGKILL"), 5_000);
      killTimer.unref?.();
    };
    const abort = () => {
      const reason = options.signal?.reason;
      terminate(
        Number.isInteger(reason?.exitCode) ? reason.exitCode : 15,
        reason?.reason ?? "cancelled_or_interrupted",
        reason?.signal ?? "SIGTERM",
      );
    };
    const child = spawn(command, args, {
      cwd: repoRoot,
      env: options.env,
      detached: process.platform !== "win32",
      stdio: ["ignore", "pipe", "pipe"],
    });
    child.stdout.pipe(redactionTransform()).pipe(log, { end: false });
    child.stderr.pipe(redactionTransform()).pipe(log, { end: false });
    child.on("error", (error) => {
      log.write(`${redactString(error.message)}\n`);
      finish(127);
    });
    child.on("close", (status) => {
      finish(status ?? 1);
    });
    if (options.signal) {
      if (options.signal.aborted) abort();
      else options.signal.addEventListener("abort", abort, { once: true });
    }
    if (Number.isFinite(options.timeoutMs) && options.timeoutMs > 0) {
      timeout = setTimeout(
        () => terminate(13, `scheduler watchdog exceeded ${options.timeoutMs}ms`),
        options.timeoutMs,
      );
      timeout.unref?.();
    }
  });
}

export async function replayLog(file, stream) {
  await new Promise((resolve, reject) => {
    const reader = createReadStream(file);
    reader.on("error", reject);
    reader.on("end", resolve);
    reader.pipe(redactionTransform()).pipe(stream, { end: false });
  });
}
