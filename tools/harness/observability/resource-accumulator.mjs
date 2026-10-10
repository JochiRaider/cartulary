import { instrumentationSignals } from "../contract/index.mjs";

export function createResourceAccumulator({ signals = instrumentationSignals } = {}) {
  const scopes = new Map();
  let simultaneousTime = null, simultaneousRSS = 0, maximumSimultaneousRSS = 0, rssObserved = false;
  const consume = (record, roster) => {
    if (!scopes.has(record.scope_ref)) {
      if (scopes.size >= 4120) throw new Error("resource scope limit exceeded");
      scopes.set(record.scope_ref, { scope_ref: record.scope_ref, scope: record.scope,
        unit_id: roster.get(record.scope_ref)?.unit_id ?? null, allocation_ref: roster.get(record.scope_ref)?.allocation_ref ?? null, samples: 0, unavailable: 0,
        resets: 0, metrics: {}, previous: new Map(), previousCPU: null, cpuDelta: 0, cpuInterval: 0, attributedCPU: null });
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
        const delta = user.value - prior.user + system.value - prior.system;
        aggregate.cpuDelta += delta;
        if (prior.time >= (roster.get(record.scope_ref)?.attribution_start_ms ?? Infinity)) aggregate.attributedCPU = (aggregate.attributedCPU ?? 0) + delta;
        aggregate.cpuInterval += record.elapsed_ms - prior.time;
      }
      aggregate.previousCPU = { user: user.value, system: system.value, segment: record.segment, identity: record.identity_digest, time: record.elapsed_ms };
    } else aggregate.previousCPU = null;
    for (const [name, value] of Object.entries(record.metrics)) {
      const metric = aggregate.metrics[name] ??= { maximum: null, observed_delta: null, observed_interval_ms: 0, unavailable: 0 };
      if (value.availability !== "available") { metric.unavailable += 1; aggregate.previous.delete(name); continue; }
      metric.maximum = Math.max(metric.maximum ?? value.value, value.value);
      const prior = aggregate.previous.get(name);
      if (signals[name]?.aggregation === "segmented_delta" && prior && prior.segment === record.segment && prior.identity === record.identity_digest) {
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
  function finish() {
    maximumSimultaneousRSS = Math.max(maximumSimultaneousRSS, simultaneousRSS);
    const rows = [...scopes.values()].map(({ previous: _previous, previousCPU: _previousCPU, cpuDelta, cpuInterval, ...scope }) => {
      // Equal accumulated durations do not prove that user/system counters cover
      // the same windows. Derive utilization only from jointly observed pairs.
      return { ...scope, cpu_percent_one_core: cpuInterval > 0 ? 100 * cpuDelta / (cpuInterval * 1000) : null };
    }).sort((a, b) => (b.metrics.cpu_user_us?.observed_delta ?? 0) - (a.metrics.cpu_user_us?.observed_delta ?? 0) || a.scope_ref.localeCompare(b.scope_ref));
    const unitCPU = new Map();
    for (const scope of rows) {
      if (scope.scope !== "process" || scope.unit_id === null) continue;
      if (scope.attributedCPU !== null) unitCPU.set(scope.unit_id, (unitCPU.get(scope.unit_id) ?? 0) + scope.attributedCPU);
    }
    for (const scope of rows) delete scope.attributedCPU;
    return { rows, unitCPU, sampledSimultaneousRSS: rssObserved ? maximumSimultaneousRSS : null };
  }
  return { consume, finish };
}
