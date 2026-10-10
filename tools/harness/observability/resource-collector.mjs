import { Worker } from "node:worker_threads";
import { readFileSync } from "node:fs";
import { createLinuxResourceAdapter } from "./resource-linux.mjs";
import { semanticJSONDigest, validateSchemaSync } from "../contract/index.mjs";
import { hostAdmissionRoot } from "../runtime/host-admission.mjs";

export function instrumentationPolicy() {
  const policy = JSON.parse(readFileSync(new URL("../../harness_instrumentation_policy.json", import.meta.url)));
  validateSchemaSync(policy.schema_id, policy);
  return { policy, digest: semanticJSONDigest(policy) };
}
export function createResourceCollector({ runRoot, manifest, capacities, epoch, admissionRoot = hostAdmissionRoot, workerFactory = (url, options) => new Worker(url, options), adapterFactory = createLinuxResourceAdapter, timers = { setTimeout, clearTimeout }, onRelationship = () => {} }) {
  if (manifest.instrumentation.mode === "off") return null;
  const { policy } = instrumentationPolicy();
  let stopCompletion, controlVersion = 0;
  let worker, adapter, stopped = false, failure = null, latest = null, sequence = 0;
  const acknowledgements = new Map();
  const pendingRegistrations = new Map();
  let pendingBytes = 0, droppedRegistrations = 0, terminalMessage = false;
  try {
    adapter = adapterFactory();
    worker = workerFactory(new URL("./resource-collector-worker.mjs", import.meta.url), {
      workerData: { policy, runRoot, capacities, admissionRoot, epoch: String(epoch), identity: {
        run_id: manifest.run_id, source_digest: manifest.source_digest,
        graph_digest: manifest.graph_digest, policy_digest: manifest.instrumentation.policy_digest,
      } },
    });
    worker.on("message", (message) => {
      if (message.type === "relationship") { try { onRelationship(message.record); } catch { droppedRegistrations += 1; } }
      if (message.type === "registered") {
        pendingBytes -= pendingRegistrations.get(message.request) ?? 0;
        pendingRegistrations.delete(message.request);
      }
      if (message.type === "latest" && !stopped) latest = message.value;
      if (message.type === "paused") acknowledgements.get(message.request)?.();
      if (message.type === "stopped" || message.type === "failed") {
        terminalMessage = true;
        if (message.type === "failed") { failure = "collector_failed"; for (const done of acknowledgements.values()) done(); }
        acknowledgements.get("stop")?.();
      }
    });
    worker.on("error", () => { failure = "collector_failed"; for (const done of acknowledgements.values()) done(); });
    worker.on("exit", () => { if (!terminalMessage) failure = "collector_failed"; for (const done of acknowledgements.values()) done(); });
  } catch { failure = "unsupported"; }
  async function request(type) {
    if (!worker) return;
    if (failure) { await worker.terminate(); return; }
    const id = type === "stop" ? "stop" : ++sequence;
    let timer;
    await new Promise((resolve) => {
      const done = () => { timers.clearTimeout(timer); acknowledgements.delete(id); resolve(); };
      acknowledgements.set(id, done);
      timer = timers.setTimeout(() => { failure = "collector_failed"; done(); }, policy.stop_ms);
      try { worker.postMessage({ type, request: id, omitted_registrations: droppedRegistrations }); }
      catch { failure = "collector_failed"; done(); }
    });
    if (failure || type === "stop") await worker.terminate();
  }
  function sendCorrelation(message) {
    if (stopped || failure || !worker) return;
    try {
      message.request = ++sequence;
      const size = Buffer.byteLength(JSON.stringify(message));
      if (pendingBytes + size > policy.maximum_pending_bytes) { droppedRegistrations += 1; return; }
      pendingRegistrations.set(message.request, size); pendingBytes += size;
      worker.postMessage(message);
    } catch { droppedRegistrations += 1; }
  }
  const result = {
    register(pid, unitID, { allocationRef = null, invocationID = null } = {}) {
      if (stopped || failure || !worker || !Number.isSafeInteger(pid)) return;
      try { sendCorrelation({ type: "register", proof: adapter.proof(pid), unit_id: unitID, allocation_ref: allocationRef, invocation_id: invocationID }); }
      catch { droppedRegistrations += 1; }
    },
    registerProof(proof, unitID, { allocationRef = null, invocationID = null } = {}) {
      sendCorrelation({ type: "register", proof, unit_id: unitID, allocation_ref: allocationRef, invocation_id: invocationID });
    },
    async pause() {
      if (stopped) return stopCompletion;
      const version = ++controlVersion;
      await request("pause");
      if (!failure && !stopped && version === controlVersion) latest = { ...latest, availability: "paused_measurement", sampled_elapsed_ms: latest?.sampled_elapsed_ms ?? null, rss_bytes: null, process_count: null };
    },
    resume() {
      controlVersion += 1;
      try { if (!stopped && !failure) worker?.postMessage({ type: "resume" }); }
      catch { failure = "collector_failed"; }
    },
    stop() {
      if (stopCompletion) return stopCompletion;
      stopped = true; controlVersion += 1;
      stopCompletion = request("stop");
      return stopCompletion;
    },
    snapshot() {
      if (failure) return { coverage: "observed_partial", omitted_observations: null, discovery_truncations: null, failed_sweeps: null, schema_id: "cartulary.harness_resource_live.v1", clock: "graph_process_monotonic", availability: failure, sampled_elapsed_ms: null, rss_bytes: null, process_count: null };
      const value = latest ?? { availability: "not_observed", sampled_elapsed_ms: null, rss_bytes: null, process_count: null };
      const age = value.sampled_elapsed_ms === null ? null : Math.max(0, Number((process.hrtime.bigint() - epoch) / 1_000_000n) - value.sampled_elapsed_ms);
      return { omitted_observations: null, discovery_truncations: null, failed_sweeps: null, ...value, coverage: "observed_partial", schema_id: "cartulary.harness_resource_live.v1", clock: "graph_process_monotonic", age_ms: age, stale: age !== null && age > policy.stale_ms && value.availability !== "paused_measurement" };
    },
  };
  result.register(process.pid, null);
  return result;
}
