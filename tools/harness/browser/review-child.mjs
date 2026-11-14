import { createLaunchContext } from "../runtime/launch-context.mjs";
import { CommandFailure, createCommandFailureContext } from "../runtime/command-failure.mjs";
import { runOwnedCommand } from "../runtime/owned-command.mjs";

// Parent-owned cancellation and exact private diagnostic identity survive Make.
// Child output never participates in classification.
export async function reviewChild({ root, command, args, environment, signal, output = "ignore", onChildProcess = () => () => {}, commandID, context }) {
  const launch = createLaunchContext({ repoRoot: root, environment, unitID: `review:${context.subject_id}`, commandID, producer: "browser_review" });
  if (signal?.aborted) { launch.settle("cancelled"); signal.throwIfAborted(); }
  let diagnostic, primary, outcome, release = () => {};
  try {
    diagnostic = launch.identity.command_id ? createCommandFailureContext({ repoRoot: root, environment, launch }) : null;
    outcome = await runOwnedCommand(command, args, { cwd: root, environment: { ...environment, ...launch.environment, ...diagnostic?.environment }, signal, output,
      onReady: proof => { launch.spawned(proof.pid); release = onChildProcess(proof.pid); } });
    const ownerFailure = diagnostic?.read();
    if (signal?.aborted) primary = signal.reason;
    else if (ownerFailure) primary = new CommandFailure("classified preparation child failure", outcome.code === 0 ? { failure_class: "harness", failure_reason: "scheduler_accounting_error" } : ownerFailure);
    else if (outcome.error || outcome.code !== 0) primary = new CommandFailure("unclassified preparation child failure", outcome.error?.failure_class ? outcome.error : { failure_class: outcome.error?.code === "ENOENT" ? "config" : "unknown", failure_reason: outcome.error?.code === "ENOENT" ? "configuration_error" : "unknown_failure" }, { cause: outcome.error });
    if ((signal?.aborted || ownerFailure) && outcome.error && primary !== outcome.error) (primary.cleanupFailures ??= []).push(outcome.error);
  } catch (error) { primary ??= error; }
  finally {
    // The callback retires registration only after the native owner has settled
    // descendants and reaped the workload. It has no physical cleanup authority.
    for (const cleanup of [...(outcome?.settled ? [release] : []), () => diagnostic?.close()]) {
      try { await cleanup(); }
      catch (error) { primary ??= new CommandFailure("preparation release failed", { failure_class: "harness", failure_reason: "cleanup_error" }); (primary.cleanupFailures ??= []).push(error); }
    }
    launch.settle(signal?.aborted ? "cancelled" : !outcome?.started ? "spawn_failed" : primary ? "failed" : "passed");
  }
  if (primary) {
    for (const [key, value] of Object.entries(context ?? {})) primary[key] ??= value;
    throw primary;
  }
}
