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
export function createResourceCollector({ runRoot, manifest, capacities, epoch, admissionRoot = hostAdmissionRoot }) {
  if (manifest.instrumentation.mode === "off") return null;
  const { policy } = instrumentationPolicy();
  let worker, adapter, stopped = false, failure = null, latest = null, sequence = 0;
  const acknowledgements = new Map();
  const pendingRegistrations = new Map();
  let pendingBytes = 0, droppedRegistrations = 0, terminalMessage = false;
  try {
    adapter = createLinuxResourceAdapter();
    worker = new Worker(new URL("./resource-collector-worker.mjs", import.meta.url), {
      workerData: { policy, runRoot, capacities, admissionRoot, epoch: String(epoch), identity: {
        run_id: manifest.run_id, source_digest: manifest.source_digest,
        graph_digest: manifest.graph_digest, policy_digest: manifest.instrumentation.policy_digest,
      } },
    });
    worker.on("message", (message) => {
      if (message.type === "registered") {
        pendingBytes -= pendingRegistrations.get(message.request) ?? 0;
        pendingRegistrations.delete(message.request);
      }
      if (message.type === "latest") latest = message.value;
      if (message.type === "paused") acknowledgements.get(message.request)?.();
      if (message.type === "stopped" || message.type === "failed") {
        terminalMessage = true;
        if (message.type === "failed") failure = "collector_failed";
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
      const done = () => { clearTimeout(timer); acknowledgements.delete(id); resolve(); };
      acknowledgements.set(id, done);
      timer = setTimeout(async () => { failure = "collector_failed"; await worker.terminate(); done(); }, policy.stop_ms);
      try { worker.postMessage({ type, request: id, omitted_registrations: droppedRegistrations }); }
      catch { failure = "collector_failed"; void worker.terminate(); done(); }
    });
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
    register(pid, unitID, { allocationRef = null } = {}) {
      if (stopped || failure || !worker || !Number.isSafeInteger(pid)) return;
      try { sendCorrelation({ type: "register", proof: adapter.proof(pid), unit_id: unitID, allocation_ref: allocationRef }); }
      catch { droppedRegistrations += 1; }
    },
    lease(record) { sendCorrelation({ type: "lease", record }); },
    async pause() {
      await request("pause");
      if (!failure) latest = { ...latest, availability: "paused_measurement", sampled_elapsed_ms: latest?.sampled_elapsed_ms ?? null, rss_bytes: null, process_count: null };
    },
    resume() {
      try { if (!stopped && !failure) worker?.postMessage({ type: "resume" }); }
      catch { failure = "collector_failed"; }
    },
    async stop() { if (stopped) return; stopped = true; await request("stop"); },
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
