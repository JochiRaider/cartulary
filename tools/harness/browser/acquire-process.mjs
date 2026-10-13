// Physical browser producer supervision shared by the native fixture adapter.
import { spawn } from "../workspace/child-process.mjs";
import { createLaunchContext } from "../runtime/launch-context.mjs";
import { createCommandFailureContext, CommandFailure } from "../runtime/command-failure.mjs";
import { ownedProcess, stopOwnedProcess } from "../runtime/owned-process.mjs";

export async function acquireProcess(command, args, { cwd, environment, signal, onChildProcess = () => () => {}, onReaped = () => {} }) {
  const launch = createLaunchContext({ repoRoot: cwd, environment, unitID: environment.CARTULARY_WORK_UNIT_ID ?? "review:browser_stack", producer: "browser_acquisition", commandID: environment.CARTULARY_PREPARATION_POLICY === "installed_only" ? "cartulary.harness.command.ui_review.v1" : "cartulary.harness.command.browser_design_review.v1" });
  let child, proof, diagnostic, released = () => {}, primary, killDeadline;
  const abort = () => {
    try { if (child?.pid) process.kill(-child.pid, "SIGTERM"); } catch (error) { if (error.code !== "ESRCH") primary ??= error; }
    killDeadline ??= setTimeout(() => { try { if (child?.pid) process.kill(-child.pid, "SIGKILL"); } catch (error) { if (error.code !== "ESRCH") primary ??= error; } }, 2000);
  };
  try {
    signal?.throwIfAborted();
    diagnostic = createCommandFailureContext({ repoRoot: cwd, environment, launch });
    // The caller owns the complete child environment, including intentional
    // removal of ambient parent-graph and browser bindings.
    child = spawn(command, args, { cwd, env: { ...environment, ...launch.environment, ...diagnostic?.environment }, stdio: "ignore", detached: true });
    const closed = new Promise((resolve) => {
      child.once("error", (error) => { primary ??= error; });
      child.once("close", resolve);
    });
    try {
      if (child.pid) { launch.spawned(child.pid); proof = ownedProcess(child.pid); released = onChildProcess(child.pid, launch.identity.invocation_id); }
    } catch (error) { primary = error; abort(); }
    signal?.addEventListener("abort", abort, { once: true });
    if (signal?.aborted) abort();
    const status = await closed;
    const failure = diagnostic?.read();
    primary = signal?.aborted ? signal.reason : primary ?? (failure ? new CommandFailure("browser acquisition failed", status === 0 ? { failure_class: "harness", failure_reason: "scheduler_accounting_error" } : failure) : status === 0 ? null : new Error("unclassified browser acquisition failure"));
  } catch (error) { primary ??= error; }
  finally {
    clearTimeout(killDeadline);
    signal?.removeEventListener("abort", abort);
    for (const cleanup of [async () => {
      if (proof) await stopOwnedProcess(proof, { environment });
      if (!child?.pid || proof) onReaped();
    }, released, () => diagnostic?.close()]) {
      try { await cleanup(); } catch (error) { primary ??= new CommandFailure("acquisition cleanup failed", { failure_class: "harness", failure_reason: "cleanup_error" }); (primary.cleanupFailures ??= []).push(error); }
    }
  }
  launch.settle(signal?.aborted ? "cancelled" : !child?.pid ? "spawn_failed" : primary ? "failed" : "passed");
  if (primary) throw primary;
}
