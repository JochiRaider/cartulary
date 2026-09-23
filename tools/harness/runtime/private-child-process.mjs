import { spawn } from "node:child_process";
import {
  closeSync,
  constants,
  fstatSync,
  lstatSync,
  mkdirSync,
  mkdtempSync,
  openSync,
  readSync,
  rmdirSync,
  unlinkSync,
} from "node:fs";
import path from "node:path";

import { CommandFailure } from "./command-failure.mjs";
import { borrowSuiteRuntime } from "./suite-runtime.mjs";

const componentPattern = /^[A-Za-z0-9_.-]+$/u;

function assertPrivate(stat, directory) {
  if (stat.uid !== process.getuid() || stat.isSymbolicLink() ||
      (directory ? !stat.isDirectory() : !stat.isFile() || stat.nlink !== 1) ||
      (stat.mode & 0o777) !== (directory ? 0o700 : 0o600)) {
    throw new Error("unsafe private child capture resource");
  }
}

function privateCaptureDirectory(runtime) {
  const directory = runtime.privatePath("child-captures");
  try { mkdirSync(directory, { mode: 0o700 }); }
  catch (error) { if (error.code !== "EEXIST") throw error; }
  assertPrivate(lstatSync(directory), true);
  return directory;
}

function captureFailure(reason, cause) {
  return new CommandFailure(
    reason === "cleanup_error" ? "private child capture cleanup failed" : "private child capture failed",
    { failure_class: reason === "cleanup_error" ? "harness" : "artifact", failure_reason: reason },
    { cause },
  );
}

export async function runPrivateCapturedProcess(command, args, options) {
  const descriptors = new Set();
  const files = new Set();
  let directory;
  const close = (descriptor) => {
    closeSync(descriptor);
    descriptors.delete(descriptor);
  };
  // Failed releases stay owned for retry. Never recurse into borrowed storage.
  const cleanup = () => {
    const failures = [];
    for (const descriptor of descriptors) {
      try { close(descriptor); }
      catch (error) { failures.push(error); }
    }
    for (const file of files) {
      try { unlinkSync(file); files.delete(file); }
      catch (error) {
        if (error.code === "ENOENT") files.delete(file);
        else failures.push(error);
      }
    }
    if (directory) {
      try { rmdirSync(directory); directory = undefined; }
      catch (error) {
        if (error.code === "ENOENT") directory = undefined;
        else failures.push(error);
      }
    }
    if (failures.length) throw captureFailure("cleanup_error", new AggregateError(failures));
  };
  const createStream = (file) => {
    const descriptor = openSync(file, constants.O_WRONLY | constants.O_CREAT | constants.O_EXCL | constants.O_NOFOLLOW, 0o600);
    descriptors.add(descriptor);
    files.add(file);
    assertPrivate(fstatSync(descriptor), false);
    return descriptor;
  };
  const boundedTail = (file, limitBytes) => {
    assertPrivate(lstatSync(file), false);
    const descriptor = openSync(file, constants.O_RDONLY | constants.O_NOFOLLOW);
    descriptors.add(descriptor);
    const stat = fstatSync(descriptor);
    assertPrivate(stat, false);
    const length = Math.min(stat.size, limitBytes);
    const buffer = Buffer.alloc(length);
    let offset = 0;
    while (offset < length) {
      const count = readSync(descriptor, buffer, offset, length - offset, stat.size - length + offset);
      if (count === 0) break;
      offset += count;
    }
    close(descriptor);
    return buffer.subarray(0, offset).toString("utf8");
  };
  try {
    const tailBytes = options.tailBytes ?? 256 * 1024;
    if (!Number.isSafeInteger(tailBytes) || tailBytes < 1024 || tailBytes > 1024 * 1024) {
      throw new Error("private child diagnostic tail must be between 1 KiB and 1 MiB");
    }
    const runtime = borrowSuiteRuntime({
      repoRoot: options.repoRoot,
      runRoot: options.runRoot,
      environment: options.env,
    });
    directory = mkdtempSync(path.join(privateCaptureDirectory(runtime), "capture-"));
    const component = path.basename(directory);
    if (!componentPattern.test(component) || component.length > 128) {
      throw new Error("private child capture requires a safe bounded component");
    }
    assertPrivate(lstatSync(directory), true);
    const stdoutPath = path.join(directory, "stdout");
    const stderrPath = path.join(directory, "stderr");
    const stdoutFD = createStream(stdoutPath);
    const stderrFD = createStream(stderrPath);
    const child = spawn(command, args, {
      cwd: options.cwd,
      env: options.env,
      detached: options.detached ?? false,
      stdio: ["ignore", stdoutFD, stderrFD],
    });
    // Even a failed asynchronous spawn must reach close before releasing resources.
    const outcome = await new Promise((resolve, reject) => {
      let spawnError;
      child.once("error", (error) => { spawnError = error; });
      child.once("close", (status, signal) => {
        if (spawnError) reject(spawnError);
        else resolve({ status, signal });
      });
    });
    close(stdoutFD);
    close(stderrFD);
    const stdout = boundedTail(stdoutPath, tailBytes);
    const stderr = boundedTail(stderrPath, tailBytes);
    // Ownership transfers only after both validated reads have completed.
    return { ...outcome, stderr, stderrPath, stdout, stdoutPath, cleanup };
  } catch (error) {
    const failure = captureFailure("artifact_error", error);
    try { cleanup(); }
    catch (cleanupError) {
      failure.message += "; cleanup_error during release";
      failure.cause = new AggregateError([error, cleanupError]);
    }
    throw failure;
  }
}
