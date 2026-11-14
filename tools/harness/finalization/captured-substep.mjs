import path from "node:path";
import { primaryPublicFailure, publicExitCodeForFailure, redactString, secureWriteFile } from "../contract/index.mjs";
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
  const failures = [];
  let phase = "capture";
  let stdout = "", stderr = "";
  let stdoutLog = null, stderrLog = null;
  try {
    result = await capture(options.command, options.args, {
      cwd: options.repoRoot, repoRoot: options.repoRoot, runRoot: options.runRoot,
      env: { ...options.env, ...context.environment }, launch,
      dependencies: JSON.parse(options.env.CARTULARY_HARNESS_RESOURCE_DEPENDENCIES ?? "null"),
      tailBytes: 256 * 1024,
    });
    const childFailure = context.read();
    if (childFailure) failures.push({ ...childFailure, phase: "child" });
    else if (result.signal) failures.push({ failure_class: "interrupted", failure_reason: "cancelled_or_interrupted", phase: "interruption" });
    phase = "diagnostic_publication";
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
    const failure = error instanceof CommandFailure
      ? { failure_class: error.failure_class, failure_reason: error.failure_reason }
      : { failure_class: "artifact", failure_reason: "artifact_error" };
    failures.push({ ...failure, phase });
    const diagnostic = phase === "diagnostic_publication" ? "finalization diagnostic publication failed"
      : redactString(error.message ?? "finalization substep capture failed");
    stderr = stderr ? `${stderr}\n${diagnostic}` : diagnostic;
  } finally {
    for (const [phase, release] of [["capture_release", () => result?.cleanup()], ["context_release", () => context.close()]]) {
      try { release(); }
      catch { failures.push({ failure_class: "harness", failure_reason: "cleanup_error", phase }); }
    }
  }
  // Failed unclassified child exits retain their native status. A release
  // failure is secondary to that outcome even without a typed child envelope.
  const onlyCleanup = failures.length > 0 && failures.every(failure => failure.failure_reason === "cleanup_error");
  const failure = onlyCleanup && result?.status !== 0 && result?.status != null ? null : primaryPublicFailure(failures);
  const status = failure ? publicExitCodeForFailure(failure, { signal: result?.signal, status: result?.status }) : result?.status ?? 1;
  launch.settle(result?.signal || failure?.failure_reason === "cancelled_or_interrupted" ? "cancelled" : status === 0 ? "passed" : "failed");
  return { status, failure, failures, stdout, stderr, stdoutLog, stderrLog };
}
