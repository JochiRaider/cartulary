import path from "node:path";
import { publicExitCodeForFailure, redactString, secureWriteFile } from "../contract/index.mjs";
import { CommandFailure, createCommandFailureContext } from "../runtime/command-failure.mjs";
import { createLaunchContext } from "../runtime/launch-context.mjs";
import { runPrivateCapturedProcess } from "../runtime/private-child-process.mjs";

// Raw child output lives only in the runtime's private capture. Publish bounded,
// redacted diagnostics before releasing it, including on unsuccessful substeps.
export async function captureFinalizationSubstep(options, dependencies = {}) {
  const capture = dependencies.capture ?? runPrivateCapturedProcess;
  const launch = (dependencies.createLaunch ?? createLaunchContext)({
    repoRoot: options.repoRoot, environment: options.env, producer: "private_capture",
  });
  const context = (dependencies.createFailureContext ?? createCommandFailureContext)({
    repoRoot: options.repoRoot, environment: options.env, launch,
  });
  let result;
  let failure;
  let cleanupFailed = false;
  let stdout = "", stderr = "";
  let stdoutLog = null, stderrLog = null;
  try {
    result = await capture(options.command, options.args, {
      cwd: options.repoRoot, repoRoot: options.repoRoot, runRoot: options.runRoot,
      env: { ...options.env, ...context.environment }, launch,
      dependencies: JSON.parse(options.env.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES ?? "null"),
      tailBytes: 256 * 1024,
    });
    failure = context.read();
    stdout = redactString(result.stdout ?? "");
    stderr = redactString(result.stderr ?? "");
    for (const [name, value] of [["stdout", stdout], ["stderr", stderr]]) {
      if (!value) continue;
      const file = path.join(options.logDirectory, `${name}.log`);
      secureWriteFile(file, value);
      if (name === "stdout") stdoutLog = file;
      else stderrLog = file;
    }
  } catch (error) {
    failure = error instanceof CommandFailure
      ? { failure_class: error.failure_class, failure_reason: error.failure_reason }
      : { failure_class: "artifact", failure_reason: "artifact_error" };
    stderr = redactString(error.message ?? "finalization substep capture failed");
  } finally {
    for (const release of [() => result?.cleanup(), () => context.close()]) {
      try { release(); }
      catch { cleanupFailed = true; }
    }
  }
  const cleanupFailure = cleanupFailed ? { failure_class: "harness", failure_reason: "cleanup_error" } : null;
  if (!failure && result?.status === 0 && cleanupFailure) failure = cleanupFailure;
  const status = failure ? publicExitCodeForFailure(failure) : result?.status ?? 1;
  launch.settle(result?.signal ? "cancelled" : status === 0 ? "passed" : "failed");
  return { status, failure, cleanupFailure, stdout, stderr, stdoutLog, stderrLog };
}
