import { spawn } from "../workspace/child-process.mjs";
import { createLaunchContext } from "../runtime/launch-context.mjs";
import { CommandFailure, createCommandFailureContext } from "../runtime/command-failure.mjs";

// Parent-owned cancellation and exact private diagnostic identity survive Make.
// Child output never participates in classification.
export async function reviewChild({ root, command, args, environment, signal, output = "ignore", onChildProcess = () => () => {}, commandID, context, timeoutMs = 300000 }) {
  const launch = createLaunchContext({ repoRoot: root, environment, unitID: `review:${context.subject_id}`, commandID, producer: "browser_review" });
  if (signal?.aborted) { launch.settle("cancelled"); signal.throwIfAborted(); }
  const diagnostic = commandID ? createCommandFailureContext({ repoRoot: root, environment, launch }) : null;
  let primary;
  try {
    await new Promise((resolve, reject) => {
      const child = spawn(command, args, { cwd: root, env: { ...environment, ...launch.environment, ...diagnostic?.environment }, stdio: output, detached: true });
      if (child.pid) launch.spawned(child.pid);
      let release = () => {}, spawnError, expired = false;
      const terminate = () => { try { process.kill(-child.pid, "SIGTERM"); } catch (error) { if (error.code !== "ESRCH") spawnError ??= error; } };
      const kill = () => { try { process.kill(-child.pid, "SIGKILL"); } catch (error) { if (error.code !== "ESRCH") spawnError ??= error; } };
      let hardStop;
      const stop = () => { terminate(); hardStop ??= setTimeout(kill, 2000); };
      try { if (child.pid) release = onChildProcess(child.pid); }
      catch (error) { spawnError = error; kill(); }
      const timer = setTimeout(() => { expired = true; stop(); }, timeoutMs);
      signal?.addEventListener("abort", stop, { once: true });
      if (signal?.aborted) stop();
      child.once("error", (error) => { spawnError ??= error; });
      child.once("close", async (status) => {
        clearTimeout(timer); clearTimeout(hardStop); signal?.removeEventListener("abort", stop);
        const ownerFailure = diagnostic?.read();
        let failure;
        if (signal?.aborted) failure = signal.reason;
        else if (expired) failure = new CommandFailure("preparation deadline expired", { failure_class: "timing", failure_reason: "timeout_failure" });
        else if (ownerFailure) failure = new CommandFailure("classified preparation child failure", status === 0 ? { failure_class: "harness", failure_reason: "scheduler_accounting_error" } : ownerFailure);
        else if (spawnError || status !== 0) failure = new CommandFailure("unclassified preparation child failure", { failure_class: spawnError?.code === "ENOENT" ? "config" : "unknown", failure_reason: spawnError?.code === "ENOENT" ? "configuration_error" : "unknown_failure" }, { cause: spawnError });
        try { await release(); }
        catch (error) { failure ??= new CommandFailure("preparation release failed", { failure_class: "harness", failure_reason: "cleanup_error" }); (failure.cleanupFailures ??= []).push(error); }
        launch.settle(signal?.aborted ? "cancelled" : expired ? "timeout" : spawnError && !child.pid ? "spawn_failed" : failure ? "failed" : "passed");
        if (failure) reject(failure); else resolve();
      });
    });
  } catch (error) {
    primary = error;
    for (const [key, value] of Object.entries(context ?? {})) error[key] ??= value;
    throw error;
  } finally {
    try { diagnostic?.close(); }
    catch (error) {
      if (primary) (primary.cleanupFailures ??= []).push(error);
      else throw new CommandFailure("diagnostic cleanup failed", { failure_class: "harness", failure_reason: "cleanup_error", ...context }, { cause: error });
    }
  }
}
