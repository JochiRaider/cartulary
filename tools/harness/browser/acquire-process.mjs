// Physical browser producer supervision shared by the native fixture adapter.
import { createLaunchContext } from "../runtime/launch-context.mjs";
import { createCommandFailureContext, CommandFailure } from "../runtime/command-failure.mjs";
import { runOwnedCommand } from "../runtime/owned-command.mjs";

export async function acquireProcess(command, args, { cwd, environment, signal, onChildProcess = () => () => {}, onReaped = () => {} }) {
  const launch = createLaunchContext({ repoRoot: cwd, environment, unitID: environment.CARTULARY_WORK_UNIT_ID ?? "review:browser_stack", producer: "browser_acquisition", commandID: environment.CARTULARY_PREPARATION_POLICY === "installed_only" ? "cartulary.harness.command.ui_review.v1" : "cartulary.harness.command.browser_design_review.v1" });
  let outcome = { started: false, settled: true }, diagnostic, released = () => {}, primary;
  try {
    signal?.throwIfAborted();
    diagnostic = createCommandFailureContext({ repoRoot: cwd, environment, launch });
    // The caller owns the complete environment, including intentional removal
    // of ambient parent-graph and browser bindings.
    outcome = undefined;
    outcome = await runOwnedCommand(command, args, { cwd, environment: { ...environment, ...launch.environment, ...diagnostic?.environment }, signal,
      onReady: proof => { launch.spawned(proof.pid); released = onChildProcess(proof.pid, launch.identity.invocation_id); } });
    const failure = diagnostic?.read();
    primary = signal?.aborted ? signal.reason : outcome.error ?? (failure ? new CommandFailure("browser acquisition failed", outcome.code === 0 ? { failure_class: "harness", failure_reason: "scheduler_accounting_error" } : failure) : outcome.code === 0 ? null : new Error("unclassified browser acquisition failure"));
    if (signal?.aborted && outcome.error && primary !== outcome.error) (primary.cleanupFailures ??= []).push(outcome.error);
  } catch (error) { primary ??= error; }
  finally {
    // Failed native settlement retains the process journal for recovery.
    const cleanups = outcome?.settled ? [onReaped, released] : [];
    cleanups.push(() => diagnostic?.close());
    for (const cleanup of cleanups) {
      try { await cleanup(); } catch (error) { primary ??= new CommandFailure("acquisition cleanup failed", { failure_class: "harness", failure_reason: "cleanup_error" }); (primary.cleanupFailures ??= []).push(error); }
    }
  }
  launch.settle(signal?.aborted ? "cancelled" : !outcome?.started ? "spawn_failed" : primary ? "failed" : "passed");
  if (primary) throw primary;
}
