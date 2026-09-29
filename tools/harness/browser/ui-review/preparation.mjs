import { fork } from "node:child_process";
import { fileURLToPath } from "node:url";
import { borrowSuiteRuntime } from "../../runtime/suite-runtime.mjs";
import { privateDirectory } from "../../runtime/secure-local-files.mjs";
import { recoverReviewPreparation, runPreparedReview } from "../review-preparation.mjs";
import { ReviewFailure } from "./failure.mjs";
import { boundedCleanup, ownedProcess, recordResource, recoveryResources, stopOwnedProcess } from "./ownership.mjs";
import { repoRoot } from "./policy.mjs";

export function prepareReview({ input, runtime, runID, signal, onProcess = async () => {} }) {
  signal.throwIfAborted();
  const ready = Promise.withResolvers(), done = Promise.withResolvers();
  privateDirectory(runtime.privatePath("lifecycle", runID));
  const child = fork(fileURLToPath(import.meta.url), ["--prepare"], { detached: true, execArgv: [], stdio: ["ignore", "ignore", "ignore", "ipc"] });
  const target = ownedProcess(child.pid);
  try { recordResource(runtime, { kind: "preparation_process", target }); }
  catch (error) { child.kill("SIGKILL"); throw error; }
  let outcome, prepared, checking, submitted = false;
  const stop = () => {
    if (!submitted) { outcome = { failure: "interrupted" }; child.kill("SIGKILL"); }
    else if (child.connected) child.send({ stop: true });
  };
  signal.addEventListener("abort", stop, { once: true });
  child.on("message", (message) => {
    if (message.ready) {
      prepared = true;
      ready.resolve({ ...message.ready, check: async () => {
        if (!child.connected) throw new ReviewFailure("session_lost");
        checking ??= Promise.withResolvers(); child.send({ check: true });
        return boundedCleanup(() => checking.promise, 10000);
      } });
    }
    if (message.checked !== undefined) {
      if (message.checked) checking?.resolve(); else checking?.reject(new ReviewFailure("session_lost"));
      checking = null;
    }
    if (message.done) outcome = message;
  });
  child.once("error", () => { ready.reject(new ReviewFailure("startup_failed")); });
  child.once("exit", async () => {
    signal.removeEventListener("abort", stop);
    try { await stopOwnedProcess(target); recordResource(runtime, { kind: "preparation_process", target, state: "released" }); }
    catch { outcome = { ...outcome, cleanup_failed: true }; }
    checking?.reject(new ReviewFailure("session_lost"));
    const error = new ReviewFailure(outcome?.failure ?? "session_lost");
    if (outcome?.cleanup_failed || !outcome) error.cleanupFailures = [new ReviewFailure("cleanup_failed")];
    if (!prepared) ready.reject(error);
    if (outcome && !outcome.failure && !outcome.cleanup_failed) done.resolve(); else done.reject(error);
  });
  done.promise.catch(() => {}); ready.promise.catch(() => {});
  Promise.resolve().then(() => onProcess(target)).then(() => {
    signal.throwIfAborted();
    submitted = true;
    child.send({ input, runID, runtime: { root: runtime.root, leaseID: runtime.leaseID, runRoot: runtime.privatePath("lifecycle", runID) } });
  }).catch(() => { outcome = { failure: signal.aborted ? "interrupted" : "startup_failed" }; child.kill("SIGKILL"); });
  return { ready: ready.promise, done: done.promise, stop: async () => {
    stop();
    try { await boundedCleanup(() => done.promise, 250000); }
    catch (error) { await stopOwnedProcess(target); throw error; }
  } };
}

if (process.argv[2] === "--prepare") {
  const { input, runID, runtime: owner } = await new Promise((resolve) => process.once("message", resolve));
  const runtime = borrowSuiteRuntime({ repoRoot, runRoot: owner.runRoot, environment: {
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: owner.root,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: owner.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID,
  } });
  const abort = new AbortController(), hold = Promise.withResolvers(); let prepared;
  const stop = () => { abort.abort(new ReviewFailure("interrupted")); hold.resolve(); };
  process.once("disconnect", stop); process.on("SIGTERM", stop); process.on("SIGINT", stop);
  process.on("message", async (message) => {
    if (message.stop) stop();
    if (message.check) { let checked = false; try { await prepared.check(); checked = true; } catch {} if (process.connected) process.send({ checked }); }
  });
  let outcome = { done: true, failure: null, cleanup_failed: false };
  try {
    await runPreparedReview({ environment: { ...process.env, REVIEW_PROFILE: input.REVIEW_PROFILE }, signal: abort.signal,
      target: "ui-review", runID, runRoot: owner.runRoot, runtime, provision: false, retainDetail: false, writeOutput: () => {},
      onOwnedResource: (resource) => recordResource(runtime, resource),
      onChildProcess: (pid) => {
        const resource = { kind: "helper_process", target: ownedProcess(pid) };
        recordResource(runtime, resource);
        return async () => { await stopOwnedProcess(resource.target); recordResource(runtime, { ...resource, state: "released" }); };
      },
      onReady: (value) => { abort.signal.throwIfAborted(); prepared = value; if (process.connected) process.send({ ready: { attached: value.attached, privateDirectory: value.privateDirectory, runRoot: value.runRoot } }); },
      hold: () => hold.promise });
  } catch (error) {
    outcome = { done: true, failure: abort.signal.aborted ? null : "startup_failed", cleanup_failed: Boolean(error.cleanupFailures?.length) };
  }
  if (!outcome.cleanup_failed) {
    try { await recoverReviewPreparation({ runtime, resources: recoveryResources(runtime), onReleased: (resource) => recordResource(runtime, { ...resource, state: "released" }) }); }
    catch { outcome.cleanup_failed = true; }
  }
  if (process.connected) process.send(outcome, () => process.disconnect());
}
