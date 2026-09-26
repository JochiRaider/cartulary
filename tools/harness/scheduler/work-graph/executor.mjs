import { execFileSync, spawn } from "node:child_process";
import { accessSync, constants, readFileSync, statSync } from "node:fs";
import path from "node:path";

import { publicExitCodeForFailure, validateSchemaSync } from "../../contract/index.mjs";
import { createCommandFailureContext } from "../../runtime/command-failure.mjs";

function nodeRuntimeError(message) {
  return Object.assign(new Error(message), {
    failure_class: "config",
    failure_reason: "configuration_error",
  });
}

export function resolveGraphNodeBinary({ cwd, nodeBin }) {
  if (typeof nodeBin !== "string" || nodeBin.trim() === "") {
    throw nodeRuntimeError("selected NODE_BIN is missing");
  }
  const resolved = path.resolve(cwd, nodeBin);
  try {
    if (!statSync(resolved).isFile()) {
      throw new Error("not a regular file");
    }
    accessSync(resolved, constants.X_OK);
  } catch {
    throw nodeRuntimeError("selected NODE_BIN is not an executable file");
  }
  return resolved;
}

export function assertGraphNodeLaunch(nodeBinary) {
  try {
    execFileSync(nodeBinary, ["--version"], { timeout: 5000, stdio: "ignore" });
  } catch {
    throw nodeRuntimeError("selected NODE_BIN could not be executed");
  }
}

export function withGraphNodeRuntime(environment, nodeBinary) {
  const directory = path.dirname(nodeBinary);
  const entries = String(environment.PATH ?? "").split(path.delimiter);
  while (entries[0] === directory) entries.shift();
  const suffix = entries.join(path.delimiter);
  return {
    ...environment,
    NODE_BIN: nodeBinary,
    PATH: suffix ? `${directory}${path.delimiter}${suffix}` : directory,
  };
}

const commandMaps = new Map();
function commandID(root, target) {
  if (!commandMaps.has(root)) {
    const surface = JSON.parse(readFileSync(path.join(root, "tools/task_surface_owner.json"), "utf8"));
    commandMaps.set(root, new Map(surface.targets.map((entry) => [entry.name, entry.command_id])));
  }
  return commandMaps.get(root).get(target);
}

const failureClasses = new Set([
  "artifact",
  "config",
  "harness",
  "infra",
  "interrupted",
  "product",
  "security",
  "timing",
  "unknown",
]);

function retainedFailureMarker(stdout, stderr) {
  const text = `${stdout}\n${stderr}`;
  const classes = [...text.matchAll(/(?:^|\s)failure_class=([a-z][a-z0-9_]*)/g)]
    .map((match) => match[1])
    .filter((value) => failureClasses.has(value));
  const reasons = [...text.matchAll(/(?:^|\s)(?:failure_reason|reason)=([a-z][a-z0-9_]*)/g)]
    .map((match) => match[1]);
  if (classes.length === 0) return null;
  return {
    failure_class: classes.at(-1),
    failure_reason: reasons.at(-1) ?? "unknown_failure",
  };
}

function classifyFailure(unit, code, stdout, stderr, { timedOut, cancelled }) {
  if (timedOut) return { failure_class: "timing", failure_reason: "timeout_failure" };
  if (cancelled) return { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted" };
  const retained = retainedFailureMarker(stdout, stderr);
  if (retained) return retained;
  if (code === 2 && path.basename(unit.command.executable) !== "make") {
    return { failure_class: "config", failure_reason: "configuration_error" };
  }
  if (code === 3) return { failure_class: "infra", failure_reason: "preflight_error" };
  if (code === 10) return { failure_class: "product", failure_reason: "test_assertion_failure" };
  if (code === 11) return { failure_class: "artifact", failure_reason: "artifact_error" };
  if (code === 13) return { failure_class: "timing", failure_reason: "timeout_failure" };
  if (unit.kind === "lifecycle") {
    return { failure_class: "harness", failure_reason: "fixture_error" };
  }
  return path.basename(unit.command.executable) === "make"
    ? { failure_class: "harness", failure_reason: "child_target_failure" }
    : { failure_class: "product", failure_reason: "test_assertion_failure" };
}

function retainedLifecycleFailure(unit, cwd, environment) {
  if (unit.kind !== "lifecycle") return null;
  const output = unit.current_run_evidence_outputs.find((candidate) =>
    candidate.endsWith(".attempt.json")
  );
  const resultsRoot = environment.CARTULARY_TEST_RESULTS_DIR;
  const runID = environment.CARTULARY_TEST_RUN_ID;
  if (!output || !resultsRoot || !runID) {
    return { failure_class: "artifact", failure_reason: "artifact_error" };
  }
  try {
    const runRoot = path.resolve(cwd, resultsRoot, runID);
    const artifact = path.resolve(runRoot, output);
    if (!artifact.startsWith(`${runRoot}${path.sep}`)) {
      throw new Error("attempt path escapes run root");
    }
    const attempt = JSON.parse(readFileSync(artifact, "utf8"));
    validateSchemaSync(attempt.schema_id, attempt);
    if (attempt.schema_id !== "cartulary.browser_reset_attempt.v1" || attempt.status !== "fail") {
      throw new Error("lifecycle attempt is not a terminal failure");
    }
    return {
      failure_class: attempt.failure_class,
      failure_reason: attempt.failure_reason,
    };
  } catch {
    return { failure_class: "artifact", failure_reason: "artifact_error" };
  }
}

function terminateOwnedProcess(child, signal) {
  if (!child.pid || child.exitCode !== null) return;
  try {
    if (process.platform !== "win32") process.kill(-child.pid, signal);
    else child.kill(signal);
  } catch {
    try {
      child.kill(signal);
    } catch {
      // The process may have exited between the state check and signal.
    }
  }
}

export function executeUnitProcess(
  unit,
  {
    cwd,
    signal,
    environment = {},
    nodeBinary,
    fixtureLease,
    inheritProcessEnvironment = true,
    outputLimitBytes = 1048576,
  } = {},
) {
  return new Promise((resolve) => {
    let selectedNodeBinary;
    try {
      const binding = nodeBinary ?? environment.NODE_BIN;
      if (binding || unit.command.executable === "node") {
        selectedNodeBinary = resolveGraphNodeBinary({ cwd, nodeBin: binding });
      }
    } catch (error) {
      resolve({
        status: "failed",
        failure_class: "config",
        failure_reason: "configuration_error",
        exit_code: 2,
        stdout: "",
        stderr: `${error.message}\n`,
      });
      return;
    }
    const childEnvironment = {
      ...(inheritProcessEnvironment ? process.env : {}),
      ...environment,
      ...(fixtureLease?.allocation?.environment ?? {}),
      ...(fixtureLease?.resource?.environment ?? {}),
      ...unit.command.environment,
    };
    // Fresh identity belongs to this process invocation, never to the semantic
    // graph or a previous child. Unit results remain scheduler-owned.
    delete childEnvironment.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
    const diagnosticTarget = unit.command.environment.CARTULARY_TEST_TARGET;
    const diagnosticCommand = diagnosticTarget && childEnvironment.CARTULARY_HARNESS_SUITE_RUNTIME_ROOT
      ? commandID(cwd, diagnosticTarget) : null;
    const diagnostic = diagnosticCommand ? createCommandFailureContext({
      repoRoot: cwd, environment: childEnvironment, unitID: unit.unit_id, commandID: diagnosticCommand,
    }) : null;
    Object.assign(childEnvironment, diagnostic?.environment);
    if (selectedNodeBinary) {
      Object.assign(childEnvironment, withGraphNodeRuntime(childEnvironment, selectedNodeBinary));
    }
    const child = spawn(
      unit.command.executable === "node" ? selectedNodeBinary : unit.command.executable,
      unit.command.args,
      {
        cwd,
        env: childEnvironment,
        detached: process.platform !== "win32",
        stdio: ["ignore", "pipe", "pipe"],
      },
    );
    let stdout = "";
    let stderr = "";
    let cancelled = false;
    let timedOut = false;
    let spawnFailed = false;
    let killDeadline;
    const append = (current, chunk) =>
      `${current}${chunk}`.slice(-outputLimitBytes);
    child.stdout.on("data", (chunk) => {
      stdout = append(stdout, chunk);
    });
    child.stderr.on("data", (chunk) => {
      stderr = append(stderr, chunk);
    });
    const terminate = (reason) => {
      if (reason === "cancelled") cancelled = true;
      if (reason === "timeout") timedOut = true;
      terminateOwnedProcess(child, "SIGTERM");
      killDeadline ??= setTimeout(() => terminateOwnedProcess(child, "SIGKILL"), 2000);
      killDeadline.unref?.();
    };
    const onAbort = () => terminate("cancelled");
    signal?.addEventListener("abort", onAbort, { once: true });
    if (signal?.aborted) onAbort();
    const timeout = setTimeout(() => terminate("timeout"), unit.timeout_ms);
    timeout.unref?.();
    child.on("error", (error) => {
      spawnFailed = true;
      clearTimeout(timeout);
      clearTimeout(killDeadline);
      diagnostic?.close();
      signal?.removeEventListener("abort", onAbort);
      const nodeLaunch = unit.command.executable === "node";
      resolve({
        status: "failed",
        failure_class: nodeLaunch ? "config" : "infra",
        failure_reason: nodeLaunch ? "configuration_error" : "service_start_error",
        exit_code: nodeLaunch ? 2 : publicExitCodeForFailure({
          failure_class: "infra",
          failure_reason: "service_start_error",
        }),
        stderr: nodeLaunch ? "selected NODE_BIN could not be launched\n" : stderr,
        ...(nodeLaunch ? {} : { error }),
        stdout,
      });
    });
    child.on("close", (code, closeSignal) => {
      if (spawnFailed) return;
      clearTimeout(timeout);
      clearTimeout(killDeadline);
      signal?.removeEventListener("abort", onAbort);
      let failure = code === 0 && !cancelled && !timedOut
        ? {}
        : classifyFailure(unit, code, stdout, stderr, { timedOut, cancelled });
      if (code !== 0 && !timedOut && !cancelled && unit.kind === "lifecycle") {
        failure = retainedLifecycleFailure(unit, cwd, childEnvironment);
      }
      const diagnosed = diagnostic?.read();
      diagnostic?.close();
      if (diagnosed && !timedOut && !cancelled) {
        failure = code === 0
          ? { failure_class: "harness", failure_reason: "scheduler_accounting_error" }
          : diagnosed;
      }
      resolve({
        status: cancelled ? "cancelled" : code === 0 && !diagnosed && !timedOut ? "passed" : "failed",
        ...failure,
        exit_code: diagnosed || timedOut || cancelled ? publicExitCodeForFailure(failure, { signal: closeSignal }) : code,
        signal: closeSignal,
        stdout,
        stderr,
      });
    });
  });
}
