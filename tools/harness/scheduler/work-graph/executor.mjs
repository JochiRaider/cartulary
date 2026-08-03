import { spawn } from "node:child_process";
import path from "node:path";

const failureClasses = new Set([
  "artifact",
  "config",
  "harness",
  "infra",
  "interrupted",
  "product",
  "security",
  "timing",
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
    failure_reason: reasons.at(-1) ?? "execution_failure",
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
  if (code === 3) return { failure_class: "infra", failure_reason: "infrastructure_error" };
  if (code === 10) return { failure_class: "product", failure_reason: "test_assertion_failure" };
  if (code === 11) return { failure_class: "harness", failure_reason: "execution_failure" };
  if (code === 13) return { failure_class: "timing", failure_reason: "timeout_failure" };
  return path.basename(unit.command.executable) === "make"
    ? { failure_class: "harness", failure_reason: "execution_failure" }
    : { failure_class: "product", failure_reason: "test_assertion_failure" };
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
    const child = spawn(unit.command.executable, unit.command.args, {
      cwd,
      env: {
        ...(inheritProcessEnvironment ? process.env : {}),
        ...environment,
        ...(fixtureLease?.allocation?.environment ?? {}),
        ...(fixtureLease?.resource?.environment ?? {}),
        ...unit.command.environment,
      },
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
        failure_reason: "process_start_error",
        error,
        stdout,
        stderr,
      });
    });
    child.on("close", (code, closeSignal) => {
      clearTimeout(timeout);
      signal?.removeEventListener("abort", onAbort);
      const failure = code === 0 && !cancelled
        ? {}
        : classifyFailure(unit, code, stdout, stderr, { timedOut, cancelled });
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
