import { spawn } from "node:child_process";
import {
  chmodSync,
  closeSync,
  lstatSync,
  mkdirSync,
  openSync,
  readSync,
  rmSync,
} from "node:fs";
import path from "node:path";

import { borrowSuiteRuntime } from "./suite-runtime.mjs";

const captureIDPattern = /^[A-Za-z0-9_.-]+$/u;

function privateCaptureDirectory(runtime) {
  const directory = runtime.privatePath("child-captures");
  mkdirSync(directory, { recursive: true, mode: 0o700 });
  chmodSync(directory, 0o700);
  const stat = lstatSync(directory);
  if (!stat.isDirectory() || stat.isSymbolicLink() || (stat.mode & 0o777) !== 0o700) {
    throw new Error("private child capture directory is not owner-only");
  }
  return directory;
}

function boundedTail(file, limitBytes) {
  const descriptor = openSync(file, "r");
  try {
    const stat = lstatSync(file);
    const length = Math.min(stat.size, limitBytes);
    const buffer = Buffer.alloc(length);
    if (length > 0) readSync(descriptor, buffer, 0, length, stat.size - length);
    return buffer.toString("utf8");
  } finally {
    closeSync(descriptor);
  }
}

export async function runPrivateCapturedProcess(command, args, options) {
  const captureID = String(options.captureID ?? "");
  if (!captureIDPattern.test(captureID) || captureID.length > 128) {
    throw new Error("private child capture requires a safe bounded identity");
  }
  const tailBytes = options.tailBytes ?? 256 * 1024;
  if (!Number.isSafeInteger(tailBytes) || tailBytes < 1024 || tailBytes > 1024 * 1024) {
    throw new Error("private child diagnostic tail must be between 1 KiB and 1 MiB");
  }
  const runtime = borrowSuiteRuntime({
    repoRoot: options.repoRoot,
    runRoot: options.runRoot,
    environment: options.env,
  });
  const directory = privateCaptureDirectory(runtime);
  const stdoutPath = path.join(directory, `${captureID}.stdout`);
  const stderrPath = path.join(directory, `${captureID}.stderr`);
  const stdoutFD = openSync(stdoutPath, "wx", 0o600);
  const stderrFD = openSync(stderrPath, "wx", 0o600);
  let child;
  try {
    child = spawn(command, args, {
      cwd: options.cwd,
      env: options.env,
      detached: options.detached ?? false,
      stdio: ["ignore", stdoutFD, stderrFD],
    });
  } catch (error) {
    closeSync(stdoutFD);
    closeSync(stderrFD);
    rmSync(stdoutPath, { force: true });
    rmSync(stderrPath, { force: true });
    throw error;
  }
  closeSync(stdoutFD);
  closeSync(stderrFD);
  let outcome;
  try {
    outcome = await new Promise((resolve, reject) => {
      child.once("error", reject);
      child.once("close", (status, signal) => resolve({ status, signal }));
    });
  } catch (error) {
    rmSync(stdoutPath, { force: true });
    rmSync(stderrPath, { force: true });
    throw error;
  } finally {
    child.removeAllListeners("error");
    child.removeAllListeners("close");
  }
  return {
    ...outcome,
    stderr: boundedTail(stderrPath, tailBytes),
    stderrPath,
    stdout: boundedTail(stdoutPath, tailBytes),
    stdoutPath,
    cleanup() {
      rmSync(stdoutPath, { force: false });
      rmSync(stderrPath, { force: false });
    },
  };
}
