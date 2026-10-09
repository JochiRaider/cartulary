import { createHash } from "node:crypto";
import path from "node:path";
import { readLocalFile, readLocalChunks } from "../runtime/secure-local-files.mjs";
import { validateSchemaSync } from "../contract/index.mjs";

const counters = new Set(["cpu_user_us", "cpu_system_us", "io_read_bytes", "io_write_bytes", "io_read_chars", "io_write_chars", "psi_some_us", "psi_full_us", "cgroup_cpu_us", "cgroup_throttled_us", "cgroup_throttled_periods", "cgroup_memory_high_events", "cgroup_memory_max_events", "cgroup_oom_events", "cgroup_oom_kill_events"]);

export function resourceProjection(runRoot, run) {
  let index;
  try { index = JSON.parse(readLocalFile(path.join(runRoot, "diagnostics/resource-index.json"), { maximum: 4 * 1024 ** 2 })); }
  catch (error) { if (error.code === "ENOENT") return { availability: "not_observed", mode: run.manifest.instrumentation.mode }; throw error; }
  validateSchemaSync("cartulary.harness_resource_index.v1", index);
  for (const key of ["run_id", "source_digest", "graph_digest"]) if (index[key] !== run.manifest[key]) throw new Error("resource identity mismatch");
  if (index.policy_digest !== run.manifest.instrumentation.policy_digest) throw new Error("resource policy mismatch");
  if (index.sample_digest === null) return { availability: "collector_failed", mode: run.manifest.instrumentation.mode };
  const roster = new Map(index.processes.map((entry) => [entry.process_ref, entry]));
  if (roster.size !== index.processes.length || new Set(index.processes.map((entry) => entry.identity_digest)).size !== roster.size) throw new Error("duplicate resource identity");
  for (const entry of roster.values()) if (entry.unit_id !== null && !run.registrations.has(entry.unit_id)) throw new Error("unknown resource unit");
  if (new Set(index.leases.map((lease) => lease.lease_ref)).size !== index.leases.length) throw new Error("duplicate lease correlation");
  for (const lease of index.leases) if (!run.registrations.has(lease.unit_id)) throw new Error("unknown lease unit");
  const hash = createHash("sha256"), scopes = new Map();
  let tail = "", bytes = 0, sequence = 0, previousTime = 0;
  let simultaneousTime = null, simultaneousRSS = 0, maximumSimultaneousRSS = 0, rssObserved = false;
  const consume = (line) => {
    if (!line || Buffer.byteLength(line) > 65536) throw new Error("invalid resource record length");
    const record = JSON.parse(line);
    validateSchemaSync("cartulary.harness_resource_sample.v1", record);
    if (record.availability !== "available" && Object.keys(record.metrics).length) throw new Error("unavailable scope contains measurements");
    if (record.seq !== ++sequence || record.elapsed_ms < previousTime) throw new Error("invalid resource ordering");
    previousTime = record.elapsed_ms;
    if (record.scope === "process" && !roster.has(record.scope_ref)) throw new Error("unknown sampled process");
    if (record.scope === "process" && record.identity_digest !== roster.get(record.scope_ref).identity_digest) throw new Error("sample start identity mismatch");
    if (!scopes.has(record.scope_ref)) {
      if (scopes.size >= 4120) throw new Error("resource scope limit exceeded");
      scopes.set(record.scope_ref, { scope_ref: record.scope_ref, scope: record.scope,
        unit_id: roster.get(record.scope_ref)?.unit_id ?? null, allocation_ref: roster.get(record.scope_ref)?.allocation_ref ?? null, samples: 0, unavailable: 0,
        resets: 0, metrics: {}, previous: new Map(), previousCPU: null, cpuDelta: 0, cpuInterval: 0 });
    }
    const aggregate = scopes.get(record.scope_ref);
    if (aggregate.scope !== record.scope) throw new Error("scope changed");
    aggregate.samples += 1;
    if (record.availability !== "available") { aggregate.unavailable += 1; aggregate.previous.clear(); }
    // An omitted counter is an observation gap too; never bridge it merely
    // because the producer returned other fields for this scope.
    for (const name of aggregate.previous.keys()) if (!Object.hasOwn(record.metrics, name)) aggregate.previous.delete(name);
    const user = record.metrics.cpu_user_us, system = record.metrics.cpu_system_us;
    if (user?.availability === "available" && system?.availability === "available") {
      const prior = aggregate.previousCPU;
      if (prior && prior.segment === record.segment && prior.identity === record.identity_digest && record.elapsed_ms > prior.time && user.value >= prior.user && system.value >= prior.system) {
        aggregate.cpuDelta += user.value - prior.user + system.value - prior.system;
        aggregate.cpuInterval += record.elapsed_ms - prior.time;
      }
      aggregate.previousCPU = { user: user.value, system: system.value, segment: record.segment, identity: record.identity_digest, time: record.elapsed_ms };
    } else aggregate.previousCPU = null;
    for (const [name, value] of Object.entries(record.metrics)) {
      const metric = aggregate.metrics[name] ??= { maximum: null, observed_delta: null, observed_interval_ms: 0, unavailable: 0 };
      if (value.availability !== "available") { metric.unavailable += 1; aggregate.previous.delete(name); continue; }
      metric.maximum = Math.max(metric.maximum ?? value.value, value.value);
      const prior = aggregate.previous.get(name);
      if (counters.has(name) && prior && prior.segment === record.segment && prior.identity === record.identity_digest) {
        if (value.value < prior.value) aggregate.resets += 1;
        else if (record.elapsed_ms > prior.time) {
          metric.observed_delta = (metric.observed_delta ?? 0) + value.value - prior.value;
          metric.observed_interval_ms += record.elapsed_ms - prior.time;
        }
      }
      aggregate.previous.set(name, { value: value.value, time: record.elapsed_ms, segment: record.segment, identity: record.identity_digest });
    }
    if (record.scope === "process") {
      if (simultaneousTime !== record.elapsed_ms) {
        maximumSimultaneousRSS = Math.max(maximumSimultaneousRSS, simultaneousRSS);
        simultaneousTime = record.elapsed_ms; simultaneousRSS = 0;
      }
      if (record.metrics.rss_bytes?.availability === "available") { rssObserved = true; simultaneousRSS += record.metrics.rss_bytes.value; }
    }
  };
  for (const chunk of readLocalChunks(path.join(runRoot, "diagnostics/resource-samples.ndjson"), { maximum: 64 * 1024 ** 2 })) {
    hash.update(chunk); bytes += chunk.length; tail += chunk.toString("utf8");
    for (let newline; (newline = tail.indexOf("\n")) >= 0;) { consume(tail.slice(0, newline)); tail = tail.slice(newline + 1); }
    if (Buffer.byteLength(tail) > 65536) throw new Error("resource line too long");
  }
  if (tail || bytes !== index.sample_bytes || sequence !== index.samples || `sha256:${hash.digest("hex")}` !== index.sample_digest) throw new Error("resource artifact does not close");
  maximumSimultaneousRSS = Math.max(maximumSimultaneousRSS, simultaneousRSS);
  const rows = [...scopes.values()].map(({ previous: _previous, previousCPU: _previousCPU, cpuDelta, cpuInterval, ...scope }) => {
    // Equal accumulated durations do not prove that user/system counters cover
    // the same windows. Derive utilization only from jointly observed pairs.
    return { ...scope, cpu_percent_one_core: cpuInterval > 0 ? 100 * cpuDelta / (cpuInterval * 1000) : null };
  }).sort((a, b) => (b.metrics.cpu_user_us?.observed_delta ?? 0) - (a.metrics.cpu_user_us?.observed_delta ?? 0) || a.scope_ref.localeCompare(b.scope_ref));
  const unitCPU = new Map();
  for (const scope of rows) {
    if (scope.scope !== "process" || scope.unit_id === null) continue;
    const user = scope.metrics.cpu_user_us?.maximum, system = scope.metrics.cpu_system_us?.maximum;
    if (user !== null && user !== undefined && system !== null && system !== undefined) unitCPU.set(scope.unit_id, (unitCPU.get(scope.unit_id) ?? 0) + user + system);
  }
  return { availability: "available", completeness: index.status, coverage: index.coverage,
    mode: run.manifest.instrumentation.mode, samples: sequence, sample_bytes: bytes,
    omitted_observations: index.omitted_observations, discovery_truncations: index.discovery_truncations,
    sampled_simultaneous_rss_sum_bytes: rssObserved ? maximumSimultaneousRSS : null,
    maximum_sample_sweep_ms: index.observer.maximum_sweep_ms,
    unit_cpu_lower_bounds: [...unitCPU].sort(([a], [b]) => a.localeCompare(b)).slice(0, 20).map(([unit_id, observed_lifetime_cpu_us]) => ({ unit_id, observed_lifetime_cpu_us })),
    unit_cpu_omitted: Math.max(0, unitCPU.size - 20),
    rss_sum_semantics: "sweep_subset_shared_pages_may_repeat",
    leases: index.leases.slice(0, 20), leases_omitted: Math.max(0, index.leases.length - 20),
    scopes: rows.slice(0, 20), scopes_omitted: Math.max(0, rows.length - 20), observer: index.observer };
}
