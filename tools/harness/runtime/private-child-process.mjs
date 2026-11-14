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
import { runOwnedCommand } from "./owned-command.mjs";

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
  const outputFailure = ["limit", "io"].includes(cause?.captureFailure) ? cause.captureFailure : "none";
  return new CommandFailure(
    `${reason === "cleanup_error" ? "private child capture cleanup failed" : "private child capture failed"} (stage=${stage}, code=${code}, native=${native}, errno=${errno}, transport_phase=${transportPhase}, transport_status=${transportStatus}, transport_signal=${transportSignal}, output_failure=${outputFailure})`,
    { failure_class: reason === "cleanup_error" ? "harness" : reason === "cancelled_or_interrupted" ? "interrupted" : "artifact", failure_reason: reason },
    { cause },
  );
}

export async function runPrivateCapturedProcess(command, args, options) {
  const launch = options.launch ?? createLaunchContext({ repoRoot: options.repoRoot, environment: options.env, producer: options.producer ?? "private_capture" });
  let spawned = false;
  const descriptors = new Set();
  const files = new Set();
  let directory, runtime, processProof, registered = false, settled = true;
  const cancellation = new AbortController();
  const signal = options.signal ?? cancellation.signal;
  let stage = "input", acquisitionStage;
  const interrupt = () => cancellation.abort();
  const retire = () => {
    if (!settled || !registered) return;
    runtime.record({ kind: "helper_process", target: processProof, state: "released" });
    registered = false;
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
    if (["detached", "onSpawn", "onReady"].some(key => Object.hasOwn(options, key))) throw new Error("retired private capture ownership option");
    if (signal.aborted) throw new CommandFailure("private child capture interrupted", { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted" }, { cause: signal.reason });
    const tailBytes = options.tailBytes ?? 256 * 1024;
    if (!Number.isSafeInteger(tailBytes) || tailBytes < 1024 || tailBytes > 1024 * 1024) {
      throw new Error("private child diagnostic tail must be between 1 KiB and 1 MiB");
    }
    const captureLimitBytes = options.captureLimitBytes ?? 256 * 1024 * 1024;
    if (!Number.isSafeInteger(captureLimitBytes) || captureLimitBytes < 1 || captureLimitBytes > 256 * 1024 * 1024) {
      throw new Error("private child capture limit must be between 1 byte and 256 MiB");
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
    if (!options.signal) { process.on("SIGINT", interrupt); process.on("SIGTERM", interrupt); }
    settled = false;
    const result = await runOwnedCommand(command, args, {
      cwd: options.cwd, environment: childEnvironment, output: { stdout: stdoutFD, stderr: stderrFD }, captureLimitBytes, signal,
      onReady: proof => {
        spawned = true; launch.spawned(proof.pid); processProof = proof;
        try {
          stage = "process_registration";
          runtime.record({ kind: "helper_process", target: proof, depends_on: options.dependencies });
          registered = true;
        } catch (error) { acquisitionStage = stage; throw error; }
      },
    });
    settled = result.settled;
    stage = acquisitionStage ?? "completion";
    if (result.error) {
      try { retire(); } catch (error) { (result.error.cleanupFailures ??= []).push(error); }
      throw result.error;
    }
    stage = "process_drain";
    retire();
    const outcome = { status: result.signal ? null : result.code, signal: result.signal };
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
    const reason = ["cleanup_error", "cancelled_or_interrupted"].includes(error.failure_reason) ? error.failure_reason : "artifact_error";
    launch.settle(reason === "cancelled_or_interrupted" ? "cancelled" : spawned ? "failed" : "spawn_failed");
    const failure = captureFailure(reason, error, stage);
    if (error.cleanupFailures?.length) failure.cleanupFailures = [...error.cleanupFailures];
    failure.processSettled = settled;
    try {
      if (settled) cleanup();
      else {
        // These are only the parent's duplicate descriptors. Keep the output
        // files and runtime process registration until recovery settles writers.
        const failures = [];
        for (const descriptor of descriptors) {
          try { close(descriptor); } catch (cleanupError) { failures.push(cleanupError); }
        }
        if (failures.length) throw new AggregateError(failures);
      }
    }
    catch (cleanupError) {
      failure.message += "; cleanup_error during release";
      failure.cause = new AggregateError([error, cleanupError]);
    }
    throw failure;
  } finally {
    process.off("SIGINT", interrupt); process.off("SIGTERM", interrupt);
  }
}
