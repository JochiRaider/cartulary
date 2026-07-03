import { spawn } from "node:child_process";
import { createReadStream, createWriteStream } from "node:fs";
import { Transform } from "node:stream";

import { redactString } from "../../core/public-contract.mjs";

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

export function sanitizeMakeFlags(value) {
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
      reject(new Error(`${testOutputScript} ${args.join(" ")} exited ${closeStatus}`));
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

export function runCommand(repoRoot, command, args, logFile, env = process.env) {
  return new Promise((resolve) => {
    const log = createWriteStream(logFile);
    let settled = false;
    const finish = (status) => {
      if (settled) {
        return;
      }
      settled = true;
      log.end(() => resolve({ status }));
    };
    const child = spawn(command, args, {
      cwd: repoRoot,
      env,
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
