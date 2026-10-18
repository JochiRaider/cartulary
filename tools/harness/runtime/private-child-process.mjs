import { spawn } from "../workspace/child-process.mjs";
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

import { createLaunchContext } from "./launch-context.mjs";
import { CommandFailure } from "./command-failure.mjs";
import { borrowRuntime } from "./session-client.mjs";
import { ownedProcess, stopOwnedProcess } from "./owned-process.mjs";

const componentPattern = /^[A-Za-z0-9_.-]+$/u;

function assertPrivate(stat, directory) {
  if (stat.uid !== process.getuid() || stat.isSymbolicLink() ||
      (directory ? !stat.isDirectory() : !stat.isFile() || stat.nlink !== 1) ||
      (stat.mode & 0o777) !== (directory ? 0o700 : 0o600)) {
    throw new Error("unsafe private child capture resource");
  }
}

function privateCaptureDirectory(runtime) {
  const directory = runtime.producerPath("child-captures");
  try { mkdirSync(directory, { mode: 0o700 }); }
  catch (error) { if (error.code !== "EEXIST") throw error; }
  assertPrivate(lstatSync(directory), true);
  return directory;
}

function captureFailure(reason, cause, stage = "cleanup") {
  const diagnostic = cause?.privateDiagnostic;
  const candidateCode = cause?.code ?? diagnostic?.code;
  const code = new Set(["ENOENT", "ESRCH", "EACCES", "EPERM", "EMFILE", "ENFILE", "ENOSPC", "EIO", "ETIMEDOUT", "ETXTBSY", "ENOBUFS"]).has(candidateCode) ? candidateCode : "unclassified";
  // Retain only closed categories. Native process identities, proof-file names,
  // request values and raw cause text remain private.
  const native = new Set(["unsafe", "unresolved", "deadline", "io", "internal", "transport"]).has(diagnostic?.reason) ? diagnostic.reason : "none";
  const errno = Number.isInteger(diagnostic?.errno) && diagnostic.errno > 0 && diagnostic.errno <= 4095 ? diagnostic.errno : "none";
  const transportPhase = new Set(["spawn", "decode", "identity", "outcome"]).has(diagnostic?.phase) ? diagnostic.phase : "none";
  const transportStatus = Number.isInteger(diagnostic?.status) && diagnostic.status >= 0 && diagnostic.status <= 255 ? diagnostic.status : "none";
  const transportSignal = new Set(["SIGABRT", "SIGBUS", "SIGFPE", "SIGHUP", "SIGILL", "SIGINT", "SIGKILL", "SIGPIPE", "SIGQUIT", "SIGSEGV", "SIGTERM", "SIGTRAP", "SIGXCPU", "SIGXFSZ"]).has(diagnostic?.signal) ? diagnostic.signal : "none";
  return new CommandFailure(
    `${reason === "cleanup_error" ? "private child capture cleanup failed" : "private child capture failed"} (stage=${stage}, code=${code}, native=${native}, errno=${errno}, transport_phase=${transportPhase}, transport_status=${transportStatus}, transport_signal=${transportSignal})`,
    { failure_class: reason === "cleanup_error" ? "harness" : "artifact", failure_reason: reason },
    { cause },
  );
}

export async function runPrivateCapturedProcess(command, args, options) {
  const launch = options.launch ?? createLaunchContext({ repoRoot: options.repoRoot, environment: options.env, producer: options.producer ?? "private_capture" });
  let spawned = false;
  const descriptors = new Set();
  const files = new Set();
  let directory, runtime, processProof, processCompletion;
  let interrupted = false;
  let stage = "input";
  const interrupt = () => { interrupted = true; if (processProof) void stopOwnedProcess(processProof).catch(() => {}); };
  const drain = async () => {
    if (!processProof) return;
    await stopOwnedProcess(processProof);
    await processCompletion.catch(() => {});
    runtime.record({ kind: "helper_process", target: processProof, state: "released" });
    processProof = undefined;
  };
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
    stage = "runtime_borrow";
    runtime = borrowRuntime({
      repoRoot: options.repoRoot,
      runRoot: options.runRoot,
      environment: options.env,
    });
    stage = "capture_directory";
    directory = mkdtempSync(path.join(privateCaptureDirectory(runtime), "capture-"));
    const component = path.basename(directory);
    if (!componentPattern.test(component) || component.length > 128) {
      throw new Error("private child capture requires a safe bounded component");
    }
    assertPrivate(lstatSync(directory), true);
    const stdoutPath = path.join(directory, "stdout");
    const stderrPath = path.join(directory, "stderr");
    stage = "stdout_open";
    const stdoutFD = createStream(stdoutPath);
    stage = "stderr_open";
    const stderrFD = createStream(stderrPath);
    stage = "child_environment";
    const childEnvironment = { ...options.env, ...launch.environment, ...runtime.producerEnvironment() };
    const failureContext = childEnvironment.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
    if (failureContext && JSON.parse(failureContext).invocation_id !== launch.identity.invocation_id) delete childEnvironment.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
    stage = "spawn";
    const child = spawn(command, args, {
      cwd: options.cwd,
      env: childEnvironment,
      detached: options.detached ?? true,
      stdio: ["ignore", stdoutFD, stderrFD],
    });
    if (child.pid) { spawned = true; launch.spawned(child.pid); }
    // Even a failed asynchronous spawn must reach close before releasing resources.
    const completion = new Promise((resolve, reject) => {
      let spawnError;
      child.once("error", (error) => { spawnError = error; });
      child.once("close", (status, signal) => {
        if (spawnError) reject(spawnError);
        else resolve({ status, signal });
      });
    });
    // Register close/error before handing acquisition to a caller. A failed
    // publication callback must drain the resource it has just acquired.
    completion.catch(() => {});
    processCompletion = completion;
    if (child.pid) {
      stage = "process_proof";
      try { processProof = ownedProcess(child.pid); }
      catch (error) { if (!["ENOENT", "ESRCH"].includes(error.code)) throw error; }
      stage = "process_registration";
      if (processProof) runtime.record({ kind: "helper_process", target: processProof, depends_on: options.dependencies });
    }
    process.on("SIGINT", interrupt); process.on("SIGTERM", interrupt);
    if (interrupted) interrupt();
    if (child.pid && options.onSpawn) {
      stage = "spawn_callback";
      const proof = ownedProcess(child.pid);
      try { options.onSpawn(child); }
      catch (error) {
        let drained = true;
        try { await stopOwnedProcess(proof); }
        catch (cleanupError) { drained = false; (error.cleanupFailures ??= []).push(cleanupError); }
        if (drained) await completion.catch(() => {});
        throw error;
      }
    }
    stage = "completion";
    const outcome = await completion;
    stage = "process_drain";
    await drain();
    stage = "launch_settlement";
    if (!options.launch) launch.settle(outcome.signal ? "cancelled" : outcome.status === 0 ? "passed" : "failed");
    stage = "stream_close";
    close(stdoutFD);
    close(stderrFD);
    stage = "stdout_read";
    const stdout = boundedTail(stdoutPath, tailBytes);
    stage = "stderr_read";
    const stderr = boundedTail(stderrPath, tailBytes);
    // Ownership transfers only after both validated reads have completed.
    return { ...outcome, launch: launch.identity, stderr, stderrPath, stdout, stdoutPath, cleanup };
  } catch (error) {
    try { await drain(); } catch (cleanupError) { (error.cleanupFailures ??= []).push(cleanupError); }
    launch.settle(spawned ? "failed" : "spawn_failed");
    const failure = captureFailure("artifact_error", error, stage);
    try { cleanup(); }
    catch (cleanupError) {
      failure.message += "; cleanup_error during release";
      failure.cause = new AggregateError([error, cleanupError]);
    }
    throw failure;
  } finally {
    process.off("SIGINT", interrupt); process.off("SIGTERM", interrupt);
  }
}
