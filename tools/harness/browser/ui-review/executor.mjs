import { fork } from "../../workspace/child-process.mjs";
import { fileURLToPath } from "node:url";
import path from "node:path";
import { readFileSync, readdirSync } from "node:fs";
import { acquireHostAdmission, processIdentity } from "../../runtime/host-admission.mjs";
import { atomicLocalFile, readLocalFile, removePrivateTree } from "../../runtime/secure-local-files.mjs";
import { ArtifactStore } from "./artifact-store.mjs";
import { ownedProcess, stopOwnedProcess } from "./ownership.mjs";
import { workClaims } from "./work-policy.mjs";
import { ReviewFailure, limits, failureMappings } from "./contract.mjs";

// Linux process-tree RSS, including native allocations; sampled by the parent
// while the worker is busy or stalled. The worker also reports kernel maxRSS.
export function processTreeRSS(pid) {
  const processes = [];
  for (const entry of readdirSync("/proc")) {
    if (!/^\d+$/u.test(entry)) continue;
    try {
      const fields = readFileSync(`/proc/${entry}/stat`, "utf8").split(") ").at(-1).split(" ");
      processes.push({ pid: Number(entry), parent: Number(fields[1]), rss: Number(fields[21]) * 4096 });
    } catch (error) { if (!["ENOENT", "ESRCH"].includes(error.code)) throw error; }
  }
  const children = (parent) => processes.filter((entry) => entry.parent === parent).reduce((sum, entry) => sum + entry.rss + children(entry.pid), 0);
  return (processes.find((entry) => entry.pid === pid)?.rss ?? 0) + children(pid);
}

/** A single deadline and one admitted child own the complete data operation.
 * IPC carries identities and staged paths only. The coordinator adopts the
 * store index after exit, and only while cancellation still permits publication.
 */
export async function executeWork({ store, identity, signal, capacities, parentLease = null, deadline = performance.now() + limits.operation, command, request, operationID, observe, onResource = () => {}, onMeasurement = () => {}, workerURL = import.meta.url }) {
  const controller = new AbortController(), combined = AbortSignal.any([signal, controller.signal]);
  const timer = setTimeout(() => controller.abort(new ReviewFailure("operation_expired")), Math.max(1, deadline - performance.now()));
  const kind = command === "ui-review-report" ? "report" : "raster";
  let lease, work, child, proof, exited, failure, peak = 0;
  try {
    combined.throwIfAborted();
    try { lease = await acquireHostAdmission({ claims: workClaims[kind], capacities, parent: parentLease, signal: combined, timeoutMs: Math.min(limits.lock, Math.max(0, deadline - performance.now())) }); }
    catch (cause) { combined.throwIfAborted(); throw new ReviewFailure("capacity_exceeded", { cause }); }
    combined.throwIfAborted(); work = store.beginWork(operationID);
    let inputBytes = 0;
    const stage = (name, bytes) => {
      combined.throwIfAborted();
      if (!/^[a-z][a-z0-9.-]*$/u.test(name) || inputBytes + bytes.length >= work.capacity) throw new ReviewFailure("capacity_exceeded");
      inputBytes += bytes.length; atomicLocalFile(path.join(work.root, "inputs", name), bytes);
    };
    const observation = command === "ui-capture" && request.source === "page" ? await observe(stage) : null;
    combined.throwIfAborted();
    child = fork(fileURLToPath(workerURL), ["--worker"], { detached: true, execArgv: ["--expose-gc"], serialization: "json", stdio: ["ignore", "ignore", "ignore", "ipc"], env: { PATH: "/usr/bin:/bin", LANG: "C.UTF-8", TZ: "UTC", UV_THREADPOOL_SIZE: "1", MALLOC_ARENA_MAX: "2" } });
    let response;
    exited = new Promise((resolve) => { child.once("exit", resolve); child.once("error", resolve); });
    const abort = () => child.kill("SIGKILL");
    combined.addEventListener("abort", abort, { once: true });
    const sample = setInterval(() => {
      peak = Math.max(peak, processTreeRSS(child.pid));
      if (peak > workClaims[kind].memory_mb * 1048576) controller.abort(new ReviewFailure("capacity_exceeded"));
    }, 20);
    try {
      proof = ownedProcess(child.pid); await onResource({ kind: "helper_process", target: proof });
      combined.throwIfAborted();
      try { await lease.bind(processIdentity(child.pid)); }
      catch (cause) { combined.throwIfAborted(); throw new ReviewFailure("capacity_exceeded", { cause }); }
      combined.throwIfAborted();
      child.once("message", (value) => { response = value; });
      child.send({ identity, root: work.root, capacity: work.capacity - inputBytes, index: work.index, command, request, operationID, observation }, (error) => { if (error) { failure ??= new ReviewFailure("analysis_failed"); child.kill("SIGKILL"); } });
      await exited;
      combined.throwIfAborted();
      if (failure) throw failure;
      if (response?.error) { const error = new ReviewFailure(Object.hasOwn(failureMappings, response.error) ? response.error : "analysis_failed"); if (response.cleanup) error.cleanupFailures = [new ReviewFailure("cleanup_failed")]; throw error; }
      if (!response?.result) throw new ReviewFailure("analysis_failed");
      peak = Math.max(peak, response.peakBytes);
      if (!Number.isSafeInteger(peak) || peak > workClaims[kind].memory_mb * 1048576) throw new ReviewFailure("capacity_exceeded");
      await stopOwnedProcess(proof, { graceMS: 0 }); await onResource({ kind: "helper_process", target: proof, state: "released" }); proof = null;
      removePrivateTree(path.join(work.root, "inputs"));
      combined.throwIfAborted(); work.commit(response.store);
      return response.result;
    } finally { clearInterval(sample); combined.removeEventListener("abort", abort); }
  } catch (error) { failure = error; throw error; }
  finally {
    clearTimeout(timer);
    const cleanup = [];
    try { if (child && proof) { child.kill("SIGKILL"); await exited; await stopOwnedProcess(proof, { graceMS: 0 }); await onResource({ kind: "helper_process", target: proof, state: "released" }); proof = null; } } catch (error) { cleanup.push(error); }
    // Never delete inputs or surrender capacity while a child may still use it.
    if (!proof) {
      try { work?.rollback(); } catch (error) { cleanup.push(error); }
      try { await lease?.release(); } catch (error) { cleanup.push(error); }
    }
    onMeasurement({ kind, peak_bytes: peak, limit_bytes: workClaims[kind].memory_mb * 1048576, reaped: !proof });
    if (cleanup.length) {
      if (failure) (failure.cleanupFailures ??= []).push(...cleanup);
      else throw Object.assign(new ReviewFailure("cleanup_failed"), { cleanupFailures: cleanup });
    }
  }
}

if (process.argv[2] === "--worker") {
  process.once("disconnect", () => process.exit(15));
  process.once("message", async (job) => {
    const reply = (value) => process.send(value, () => process.exit(0));
    try {
      const sharp = (await import("sharp")).default; sharp.concurrency(1); sharp.cache(false);
      const signal = new AbortController().signal;
      const store = new ArtifactStore({ ...job.identity, signal, privatePath: (...parts) => path.join(job.root, ...parts), capacity: job.capacity, index: job.index });
      const owner = await import(job.command === "ui-capture" ? "./capture.mjs" : job.command === "ui-analyze" ? "./analysis.mjs" : "./report.mjs");
      const result = await owner.execute({ ...job.identity, store, observation: job.observation, readStage: (name) => readLocalFile(path.join(job.root, "inputs", name), { maximum: limits.png }) }, job.request, job.operationID);
      reply({ result, store: store.workResult(), peakBytes: process.resourceUsage().maxRSS * 1024 });
    } catch (error) { reply({ error: error instanceof ReviewFailure ? error.diagnostic : "analysis_failed", cleanup: Boolean(error.cleanupFailures?.length) }); }
  });
}
