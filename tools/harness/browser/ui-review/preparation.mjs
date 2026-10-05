import { fork } from "../../workspace/child-process.mjs";
import { fileURLToPath } from "node:url";
import { borrowSuiteRuntime } from "../../runtime/suite-runtime.mjs";
import { privateDirectory } from "../../runtime/secure-local-files.mjs";
import { recoverReviewPreparation, runPreparedReview } from "../review-preparation.mjs";
import { ReviewFailure, preparationFailure, failureRecord, failureFromRecord } from "./failure.mjs";
import { boundedCleanup, ownedProcess, recordResource, recoveryResources, stopOwnedProcess } from "./ownership.mjs";
import { repoRoot } from "./policy.mjs";
import { cleanupErrors, cleanupFailure } from "../../scheduler/fixture-broker/cleanup-lifecycle.mjs";

export function cleanupRecords(errors, phase) {
  const records = [...new Set(errors.flatMap((error) => cleanupErrors(error)))].map((error) => failureRecord(preparationFailure(cleanupFailure(error), {
    ...phase, phase: "cleanup", condition: "child_failed", recovery_id: "exact_stop",
  })));
  if (records.length <= 64) return records;
  return [...records.slice(0, 63), failureRecord(new ReviewFailure("diagnostic_invalid", {
    context: { ...phase, phase: "cleanup", condition: "child_failed", recovery_id: "exact_stop" },
  }))];
}

export function prepareReview({ input, runtime, runID, signal, onProcess = async () => {}, workerModule = import.meta.url, stopProcess = stopOwnedProcess }) {
  signal.throwIfAborted();
  const ready = Promise.withResolvers(), done = Promise.withResolvers();
  privateDirectory(runtime.privatePath("lifecycle", runID));
  const child = fork(fileURLToPath(workerModule), ["--prepare"], { detached: true, execArgv: [], stdio: ["ignore", "ignore", "ignore", "ipc"] });
  let target;
  if (child.pid) target = ownedProcess(child.pid);
  try { if (target) recordResource(runtime, { kind: "preparation_process", target }); }
  catch (error) { child.kill("SIGKILL"); throw error; }
  let outcome, prepared, checking, invalid = false, submitted = false, stopSent = false;
  let phase = { phase: "prerequisites", subject_id: "preparation_child", condition: "unknown", recovery_id: "inspect_failure" };
  const failure = (error) => failureRecord(preparationFailure(error, phase));
  const stop = () => {
    if (stopSent) return; stopSent = true;
    if (!submitted) { outcome = { failure: failure(signal.reason ?? new ReviewFailure("interrupted")) }; child.kill("SIGKILL"); }
    else if (child.connected) child.send({ stop: true });
  };
  signal.addEventListener("abort", stop, { once: true });
  child.on("message", (message) => {
    if (invalid || (outcome && message?.checked !== undefined)) return;
    try {
      const keys = Object.keys(message).sort().join(",");
      if (!["phase", "ready", "checked", "cleanup_failures,done,failure"].includes(keys)) throw new Error("invalid preparation message");
      if ((keys === "phase" && (!message.phase || typeof message.phase !== "object")) || (keys === "ready" && (!message.ready || typeof message.ready !== "object"))) throw new Error("invalid preparation payload");
      if (outcome || (message.ready && prepared)) throw new Error("duplicate preparation outcome");
      if (keys === "cleanup_failures,done,failure") {
        if (message.done !== true || !Array.isArray(message.cleanup_failures) || message.cleanup_failures.length > 64) throw new Error("invalid preparation outcome");
        for (const record of message.cleanup_failures) failureFromRecord(record);
      }
      if (keys === "checked" && typeof message.checked !== "boolean") throw new Error("invalid health outcome");
      if (message.phase) {
        phase = failureFromRecord(failureRecord(new ReviewFailure("preparation_failed", { context: message.phase }))).context;
      }
      if (message.done && message.failure !== null) failureFromRecord(message.failure);
    } catch {
      invalid = true;
      outcome = { failure: failure(new ReviewFailure("diagnostic_invalid", { context: phase })), cleanup_failures: [] };
      child.kill("SIGTERM"); return;
    }
    if (message.ready && !signal.aborted) {
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
  child.once("error", (error) => { if (!invalid) outcome = { failure: failure(error) }; });
  child.once("close", async (status) => {
    if (status !== 0 && !signal.aborted && outcome && !outcome.failure) outcome = { failure: failure(new ReviewFailure("diagnostic_invalid", { context: phase })), cleanup_failures: outcome.cleanup_failures };
    signal.removeEventListener("abort", stop);
    const secondary = outcome?.cleanup_failures?.map(failureFromRecord) ?? [];
    try { if (target) { await stopProcess(target); recordResource(runtime, { kind: "preparation_process", target, state: "released" }); } }
    catch (error) { secondary.push(...cleanupRecords([error], phase).map(failureFromRecord)); }
    checking?.reject(new ReviewFailure("session_lost"));
    if (!outcome) secondary.push(new ReviewFailure("cleanup_failed", { context: { ...phase, phase: "cleanup", condition: "child_failed", recovery_id: "exact_stop" } }));
    const error = outcome?.failure ? failureFromRecord(outcome.failure) : (outcome ? secondary[0] : null) ?? preparationFailure(signal.aborted ? signal.reason : new ReviewFailure(prepared ? "session_lost" : "preparation_failed", { context: phase }), phase);
    if (secondary.length) error.cleanupFailures = secondary;
    if (!prepared) ready.reject(error);
    if (outcome && !outcome.failure && !secondary.length) done.resolve(); else done.reject(error);
  });
  done.promise.catch(() => {}); ready.promise.catch(() => {});
  Promise.resolve().then(() => target && onProcess(target)).then(() => {
    signal.throwIfAborted();
    submitted = true;
    child.send({ input, runID, runtime: { root: runtime.root, leaseID: runtime.leaseID, runRoot: runtime.privatePath("lifecycle", runID) } });
  }).catch((error) => { if (!invalid) outcome = { failure: failure(signal.aborted ? signal.reason : error) }; child.kill("SIGKILL"); });
  return { ready: ready.promise, done: done.promise, stop: async () => {
    stop();
    try { await boundedCleanup(() => done.promise, 250000); }
    catch (error) {
      try { if (target) await stopProcess(target); }
      catch (cleanupError) { (error.cleanupFailures ??= []).push(...cleanupRecords([cleanupError], phase).map(failureFromRecord)); }
      throw error;
    }
  } };
}

if (process.argv[2] === "--prepare") {
  const { input, runID, runtime: owner } = await new Promise((resolve) => process.once("message", resolve));
  let runtime;
  const abort = new AbortController(), hold = Promise.withResolvers(); let prepared;
  const stop = () => { abort.abort(new ReviewFailure("interrupted")); hold.resolve(); };
  process.once("disconnect", stop); process.on("SIGTERM", stop); process.on("SIGINT", stop);
  process.on("message", async (message) => {
    if (message.stop) stop();
    if (message.check) { let checked = false; try { await prepared.check(); checked = true; } catch {} if (process.connected) process.send({ checked }); }
  });
  let outcome = { done: true, failure: null, cleanup_failures: [] };
  try {
    runtime = borrowSuiteRuntime({ repoRoot, runRoot: owner.runRoot, environment: {
    CARTULARY_HARNESS_SUITE_RUNTIME_ROOT: owner.root,
    CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID: owner.leaseID,
    CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID: runID,
  } });
    await runPreparedReview({ environment: { ...process.env, REVIEW_PROFILE: input.REVIEW_PROFILE }, signal: abort.signal,
      target: "ui-review", runID, runRoot: owner.runRoot, runtime, preparationPolicy: "installed_only", onPhase: (phase) => { if (process.connected) process.send({ phase }); }, retainDetail: false, writeOutput: () => {},
      onChildProcess: (pid) => {
        const resource = { kind: "helper_process", target: ownedProcess(pid) };
        recordResource(runtime, resource);
        return async () => { await stopOwnedProcess(resource.target); recordResource(runtime, { ...resource, state: "released" }); };
      },
      onReady: (value) => { abort.signal.throwIfAborted(); prepared = value; if (process.connected) process.send({ ready: { attached: value.attached, privateDirectory: value.privateDirectory, runRoot: value.runRoot } }); },
      hold: () => hold.promise });
  } catch (error) {
    outcome = { done: true, failure: abort.signal.aborted && error === abort.signal.reason ? null : failureRecord(preparationFailure(error)),
      cleanup_failures: cleanupRecords(error.cleanupFailures ?? [], { subject_id: "preparation_child" }) };
  }
  if (runtime && !outcome.cleanup_failures.length) {
    try { await recoverReviewPreparation({ runtime, resources: recoveryResources(runtime), onReleased: (resource) => recordResource(runtime, { ...resource, state: "released" }) }); }
    catch (error) { outcome.cleanup_failures.push(...cleanupRecords([error], { subject_id: "preparation_child" })); }
  }
  if (process.connected) process.send(outcome, () => process.disconnect());
}
