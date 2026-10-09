import { parentPort, workerData } from "node:worker_threads";
import { readlinkSync, watch } from "node:fs";
import { createHash } from "node:crypto";
import path from "node:path";
import { createAtomicNDJSONWriter } from "../scheduler/work-graph/atomic-ndjson.mjs";
import { atomicLocalFile, readLocalFile, privateDirectory } from "../runtime/secure-local-files.mjs";
import { acquireHostAdmission, processIdentity, processIdentityAlive } from "../runtime/host-admission.mjs";
import { createLinuxResourceAdapter, availabilityFor } from "./resource-linux.mjs";
import { validateSchemaSync } from "../contract/index.mjs";

const { policy, identity, runRoot, epoch, capacities, admissionRoot } = workerData;
const elapsed = () => Number((process.hrtime.bigint() - BigInt(epoch)) / 1_000_000n);
const directory = privateDirectory(path.join(runRoot, "diagnostics"));
const writer = createAtomicNDJSONWriter(path.join(directory, "resource-samples.ndjson"), JSON.stringify);
const hash = createHash("sha256");
const cpuStart = process.threadCpuUsage();
// /proc/TID/stat identifies this OS thread, unlike process.pid. A killed worker
// must not leave a fence alive until its still-running harness process exits.
const observerIdentity = processIdentity(Number(readlinkSync("/proc/thread-self").split("/").at(-1)));
const processes = new Map();
const leases = new Map();
let watcher, watchTimer, quietWaiting = false;
let adapter, timer, busy = null, stopped = false, paused = false, truncated = false;
let samples = 0, bytes = 0, sweeps = 0, omitted = 0, discoveryTruncations = 0, failures = 0;
let segment = 0;
let lastContext = -Infinity, maximumSweepMs = 0, collectorHeapPeak = 0;
let status = "complete", latest = null, pendingAdmission = null;
try { adapter = createLinuxResourceAdapter(); } catch { status = "unsupported"; }

function quietHost() {
  try {
    const state = JSON.parse(readLocalFile(path.join(admissionRoot, "state.json"), { maximum: 4 * 1024 ** 2, allowReplacement: true }));
    return state.leases.some((lease) => lease.mode === "exclusive" && !lease.released && processIdentityAlive(lease.process));
  } catch (error) { return error.code !== "ENOENT"; }
}
function register(proof, unitID, owner, allocationRef = null) {
  if (processes.has(proof.identity)) {
    // A direct lifecycle registration is more precise than discovery ancestry.
    if (owner === "registered" || processes.get(proof.identity).attribution !== "registered") Object.assign(processes.get(proof.identity), { unit_id: unitID, attribution: owner, allocation_ref: allocationRef });
    return;
  }
  if (processes.size >= policy.maximum_lifetime_processes) { omitted += 1; truncated = true; return; }
  processes.set(proof.identity, { proof, process_ref: `process:${processes.size + 1}`,
    unit_id: unitID, attribution: owner, allocation_ref: allocationRef, gone: false });
}
async function write(sample) {
  validateSchemaSync(sample.schema_id, sample);
  const serialized = `${JSON.stringify(sample)}\n`, size = Buffer.byteLength(serialized);
  if (size > policy.maximum_record_bytes || bytes + size > policy.maximum_sample_bytes) {
    omitted += 1; truncated = true; return;
  }
  // One awaited write at a time: pending bytes never exceed one bounded record.
  await writer.write(sample);
  hash.update(serialized); bytes += size; samples += 1;
}
function stopWatching() { watcher?.close(); watcher = null; clearTimeout(watchTimer); quietWaiting = false; }
function waitForQuietEnd() {
  quietWaiting = true;
  try {
    watcher ??= watch(admissionRoot, { persistent: false }, (_event, name) => {
      if (name !== "state.json" || stopped || paused) return;
      clearTimeout(watchTimer);
      watchTimer = setTimeout(() => {
        if (!quietHost()) { stopWatching(); schedule(); }
      }, 50);
    });
    watcher.on("error", () => { failures += 1; stopWatching(); });
  } catch { failures += 1; /* Remain quiescent if the wakeup channel is unavailable. */ }
}
async function sweep() {
  if (stopped || paused || !adapter || truncated) return;
  if (quietHost()) { waitForQuietEnd(); segment += 1; parentPort.postMessage({ type: "latest", value: { omitted_observations: omitted, discovery_truncations: discoveryTruncations, failed_sweeps: failures, availability: "paused_measurement", sampled_elapsed_ms: latest?.sampled_elapsed_ms ?? null, rss_bytes: null, process_count: null } }); return; }
  let lease;
  pendingAdmission = new AbortController();
  try {
    // A bounded shared observation fence closes the check/admission race. It has
    // zero test claims and exists only while reading counters, never continuously.
    lease = await acquireHostAdmission({ mode: "shared", claims: {}, capacities,
      timeoutMs: 250, signal: pendingAdmission.signal, ownerIdentity: observerIdentity, root: admissionRoot });
    if (paused || stopped) return;
    const start = performance.now(), deadline = start + policy.sweep_budget_ms;
    const discovered = adapter.discover(policy.maximum_discovery_entries, deadline);
    if (discovered.truncated) discoveryTruncations += 1;
    const byPID = new Map(discovered.found.map((proof) => [proof.pid, proof]));
    const active = new Map([...processes.values()].filter((entry) => !entry.gone).map((entry) => [entry.proof.pid, entry]));
    for (const proof of discovered.found) {
      if (processes.get(proof.identity)?.attribution === "registered") continue;
      let ancestor = proof, seen = new Set();
      while (ancestor && !seen.has(ancestor.pid)) {
        seen.add(ancestor.pid);
        const owned = active.get(ancestor.stat.parent);
        const observedParent = byPID.get(ancestor.stat.parent);
        if (owned && observedParent?.identity === owned.proof.identity && BigInt(proof.start) >= BigInt(owned.proof.start)) {
          register(proof, owned.unit_id, "observed_descendant", owned.allocation_ref); break;
        }
        ancestor = observedParent;
      }
    }
    const sampledAt = elapsed();
    let live = 0, rss = 0, completeRSS = true;
    for (const entry of processes.values()) {
      if (entry.gone) continue;
      if (live >= policy.maximum_live_processes || performance.now() >= deadline) { omitted += 1; completeRSS = false; continue; }
      let sample;
      try { sample = adapter.processSample(entry.proof); }
      catch (error) { sample = { availability: availabilityFor(error), metrics: {} }; }
      if (sample.availability === "process_gone") entry.gone = true;
      else if (sample.availability === "available") live += 1;
      const memory = sample.metrics.rss_bytes;
      if (!entry.gone && memory?.availability === "available") rss += memory.value;
      else if (!entry.gone) completeRSS = false;
      await write({ schema_id: "cartulary.harness_resource_sample.v1", seq: samples + 1,
        elapsed_ms: sampledAt, segment, identity_digest: entry.proof.identity, scope: "process", scope_ref: entry.process_ref,
        availability: sample.availability, metrics: sample.metrics });
    }
    if (sampledAt - lastContext >= policy.context_cadence_ms && performance.now() < deadline) {
      lastContext = sampledAt;
      for (const record of adapter.context()) {
        if (performance.now() >= deadline) { omitted += 1; break; }
        await write({ schema_id: "cartulary.harness_resource_sample.v1", seq: samples + 1,
          elapsed_ms: sampledAt, segment, identity_digest: record.identity_digest, scope: record.scope, scope_ref: record.scope_ref,
          availability: "available", metrics: record.metrics });
      }
    }
    maximumSweepMs = Math.max(maximumSweepMs, performance.now() - start);
    collectorHeapPeak = Math.max(collectorHeapPeak, process.memoryUsage().heapUsed);
    sweeps += 1;
    latest = { omitted_observations: omitted, discovery_truncations: discoveryTruncations, failed_sweeps: failures, availability: live > 0 ? "available" : "not_observed", sampled_elapsed_ms: sampledAt,
      rss_bytes: live > 0 && completeRSS ? rss : null, process_count: live > 0 ? live : null };
    parentPort.postMessage({ type: "latest", value: latest });
  } catch {
    failures += 1;
    parentPort.postMessage({ type: "latest", value: {
      availability: latest?.availability ?? "not_observed",
      sampled_elapsed_ms: latest?.sampled_elapsed_ms ?? null,
      rss_bytes: latest?.rss_bytes ?? null, process_count: latest?.process_count ?? null,
      omitted_observations: omitted, discovery_truncations: discoveryTruncations, failed_sweeps: failures,
    } });
  }
  finally { pendingAdmission = null; await lease?.release(); }
}
function schedule() {
  clearTimeout(timer);
  if (stopped || paused || quietWaiting || !adapter || truncated) return;
  timer = setTimeout(() => {
    busy = sweep().catch(() => { failures += 1; }).finally(() => { busy = null; schedule(); });
  }, policy.process_cadence_ms);
}
async function finish() {
  stopped = true; stopWatching(); clearTimeout(timer); pendingAdmission?.abort();
  await busy;
  const shutdownStart = performance.now();
  let sampleDigest = null;
  try { await writer.close(); sampleDigest = `sha256:${hash.digest("hex")}`; }
  catch { status = "collector_failed"; await writer.abort(); }
  const cpu = process.threadCpuUsage(cpuStart);
  const index = {
    schema_id: "cartulary.harness_resource_index.v1", ...identity,
    status: status === "collector_failed" || status === "unsupported" ? status : truncated ? "truncated" : (failures || omitted || discoveryTruncations) ? "partial" : status,
    clock: "graph_process_monotonic", coverage: "observed_partial",
    samples, sample_bytes: bytes, sample_digest: sampleDigest,
    omitted_observations: omitted, discovery_truncations: discoveryTruncations,
    failed_sweeps: failures, sweeps,
    processes: [...processes.values()].map((entry) => ({ process_ref: entry.process_ref,
      identity_digest: entry.proof.identity, unit_id: entry.unit_id, attribution: entry.attribution, allocation_ref: entry.allocation_ref })),
    leases: [...leases.values()],
    observer: { cpu_user_us: cpu.user, cpu_system_us: cpu.system,
      heap_peak_bytes: collectorHeapPeak, read_bytes: adapter?.bytesRead ?? 0,
      write_bytes: bytes, maximum_sweep_ms: maximumSweepMs,
      shutdown_ms: performance.now() - shutdownStart, gate_cpu: "not_observed" },
  };
  validateSchemaSync(index.schema_id, index);
  const serialized = `${JSON.stringify(index)}\n`;
  if (Buffer.byteLength(serialized) > policy.maximum_metadata_bytes) throw new Error("resource index limit");
  atomicLocalFile(path.join(directory, "resource-index.json"), serialized);
  parentPort.postMessage({ type: "stopped" }); parentPort.close();
}
parentPort.on("message", async (message) => {
  if (stopped) return;
  if (message.type === "register") {
    if (adapter) register(message.proof, message.unit_id, "registered", message.allocation_ref);
    parentPort.postMessage({ type: "registered", request: message.request });
  }
  if (message.type === "lease") {
    if (leases.size < policy.maximum_lifetime_processes) leases.set(message.record.lease_ref, message.record);
    else { omitted += 1; truncated = true; }
    parentPort.postMessage({ type: "registered", request: message.request });
  }
  if (message.type === "pause") {
    segment += 1; paused = true; stopWatching(); clearTimeout(timer); pendingAdmission?.abort(); await busy;
    parentPort.postMessage({ type: "paused", request: message.request });
  }
  if (message.type === "resume") { paused = false; schedule(); }
  if (message.type === "stop") {
    omitted += message.omitted_registrations ?? 0;
    truncated ||= (message.omitted_registrations ?? 0) > 0;
    try { await finish(); } catch { parentPort.postMessage({ type: "failed" }); parentPort.close(); }
  }
});
schedule();
