import { spawn } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../../contract/index.mjs";

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
    fixtureLease,
    inheritProcessEnvironment = true,
    outputLimitBytes = 1048576,
  } = {},
) {
  return new Promise((resolve) => {
    const childEnvironment = {
      ...(inheritProcessEnvironment ? process.env : {}),
      ...environment,
      ...(fixtureLease?.allocation?.environment ?? {}),
      ...(fixtureLease?.resource?.environment ?? {}),
      ...unit.command.environment,
    };
    const child = spawn(unit.command.executable, unit.command.args, {
      cwd,
      env: childEnvironment,
      detached: process.platform !== "win32",
      stdio: ["ignore", "pipe", "pipe"],
    });
    let stdout = "";
    let stderr = "";
    let cancelled = false;
    let timedOut = false;
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
    };
    const onAbort = () => terminate("cancelled");
    signal?.addEventListener("abort", onAbort, { once: true });
    if (signal?.aborted) onAbort();
    const timeout = setTimeout(() => terminate("timeout"), unit.timeout_ms);
    timeout.unref?.();
    child.on("error", (error) => {
      clearTimeout(timeout);
      signal?.removeEventListener("abort", onAbort);
      resolve({
        status: "failed",
        failure_class: "infra",
        failure_reason: "service_start_error",
        error,
        stdout,
        stderr,
      });
    });
    child.on("close", (code, closeSignal) => {
      clearTimeout(timeout);
      signal?.removeEventListener("abort", onAbort);
      let failure = code === 0 && !cancelled
        ? {}
        : classifyFailure(unit, code, stdout, stderr, { timedOut, cancelled });
      if (code !== 0 && !timedOut && !cancelled && unit.kind === "lifecycle") {
        failure = retainedLifecycleFailure(unit, cwd, childEnvironment);
      }
      resolve({
        status: cancelled ? "cancelled" : code === 0 ? "passed" : "failed",
        ...failure,
        exit_code: code,
        signal: closeSignal,
        stdout,
        stderr,
      });
    });
  });
}
