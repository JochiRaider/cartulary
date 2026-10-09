export function intervalUnion(intervals) {
  const sorted = intervals
    .filter((interval) => interval.end >= interval.start)
    .sort((left, right) => left.start - right.start || left.end - right.end);
  let total = 0;
  let active = null;
  for (const interval of sorted) {
    if (!active || interval.start > active.end) {
      if (active) total += active.end - active.start;
      active = { ...interval };
    } else {
      active.end = Math.max(active.end, interval.end);
    }
  }
  if (active) total += active.end - active.start;
  return total;
}

export function actualCriticalPath(graph, intervals, selectedUnitIDs = null) {
  const selected = selectedUnitIDs ? new Set(selectedUnitIDs) : null;
  const units = selected
    ? graph.units.filter((unit) => selected.has(unit.unit_id))
    : graph.units;
  const byID = new Map(units.map((unit) => [unit.unit_id, unit]));
  const memo = new Map();
  function pathTo(unitID) {
    if (memo.has(unitID)) return memo.get(unitID);
    const unit = byID.get(unitID);
    const predecessors = unit.needs.filter((dependency) => byID.has(dependency)).map(pathTo);
    const prior = predecessors.sort(
      (left, right) => right.duration - left.duration || left.ids.join("\0").localeCompare(right.ids.join("\0")),
    )[0] ?? { duration: 0, ids: [] };
    const interval = intervals.get(unitID);
    const result = {
      duration: prior.duration + (interval ? interval.end - interval.start + interval.queue_ms : 0),
      ids: [...prior.ids, unitID],
    };
    memo.set(unitID, result);
    return result;
  }
  return units
    .map((unit) => pathTo(unit.unit_id))
    .sort(
      (left, right) => right.duration - left.duration || left.ids.join("\0").localeCompare(right.ids.join("\0")),
    )[0]?.ids ?? [];
}

export function canonicalTimingAccounting(
  facts,
  graph,
  durationMs,
  { includeRunEnvelope = true, selectedUnitIDs = null } = {},
) {
  const selected = selectedUnitIDs ? new Set(selectedUnitIDs) : null;
  const byID = new Map(
    graph.units
      .filter((unit) => !selected || selected.has(unit.unit_id))
      .map((unit) => [unit.unit_id, unit]),
  );
  const intervals = [];
  const resourceWaitIntervals = [...facts.waits].filter(([id]) => byID.has(id)).map(([, wait]) => wait);
  const cleanupStarted = facts.cleanupStarted;
  let firstUnitStarted = null;
  let processCount = 0;
  for (const [id, start] of facts.starts) {
    if (!byID.has(id)) continue;
    firstUnitStarted = firstUnitStarted === null ? start : Math.min(firstUnitStarted, start);
    processCount += 1;
    const terminal = facts.terminals.get(id);
    if (!terminal || !["completed", "failed", "cancelled"].includes(terminal.event)) continue;
    const unit = byID.get(id), acquired = facts.fixtureAcquired.get(id);
    if (acquired !== undefined && acquired > start) intervals.push({ start, end: acquired, bucket: "fixture" });
    const executionStart = acquired ?? start;
    const bucket = ["finalizer", "projection"].includes(unit.kind) ? "collation"
      : ["artifact", "readiness"].includes(unit.kind) ? "setup" : "execution";
    if (terminal.monotonic_ms > executionStart) intervals.push({ start: executionStart, end: terminal.monotonic_ms, bucket });
  }
  if (includeRunEnvelope && (firstUnitStarted ?? 0) > 0) {
    intervals.push({ start: 0, end: firstUnitStarted, bucket: "setup" });
  }
  if (includeRunEnvelope && cleanupStarted !== null && durationMs > cleanupStarted) {
    intervals.push({ start: cleanupStarted, end: durationMs, bucket: "wrapper" });
  }
  const boundaries = [...new Set([
    ...(includeRunEnvelope ? [0, durationMs] : []),
    ...intervals.flatMap((interval) => [interval.start, interval.end]),
  ])].sort((left, right) => left - right);
  const totals = {
    setup_ms: 0,
    fixture_ms: 0,
    execution_ms: 0,
    collation_ms: 0,
    wrapper_ms: 0,
    unattributed_ms: 0,
  };
  const precedence = ["collation", "fixture", "execution", "setup", "wrapper"];
  for (let index = 1; index < boundaries.length; index += 1) {
    const start = boundaries[index - 1];
    const end = boundaries[index];
    if (end <= start) continue;
    const active = new Set(
      intervals
        .filter((interval) => interval.start < end && interval.end > start)
        .map((interval) => interval.bucket),
    );
    const bucket = precedence.find((candidate) => active.has(candidate));
    if (bucket) totals[`${bucket}_ms`] += end - start;
    else if (includeRunEnvelope) totals.unattributed_ms += end - start;
  }
  return {
    ...totals,
    resource_blocking_ms: intervalUnion(resourceWaitIntervals),
    process_count: processCount,
  };
}

export function projectResourcePressure(facts, graph, capacities) {
  const byID = new Map(graph.units.map((unit) => [unit.unit_id, unit]));
  const active = new Map();
  const saturationStart = new Map();
  const saturationMs = new Map();
  const peak = new Map();
  const add = (resource, amount, now) => {
    const before = active.get(resource) ?? 0;
    const after = before + amount;
    active.set(resource, after);
    peak.set(resource, Math.max(peak.get(resource) ?? 0, after));
    const capacity = capacities[resource];
    if (before < capacity && after >= capacity) saturationStart.set(resource, now);
    if (before >= capacity && after < capacity && saturationStart.has(resource)) {
      saturationMs.set(resource, (saturationMs.get(resource) ?? 0) + now - saturationStart.get(resource));
      saturationStart.delete(resource);
    }
  };
  const boundaries = [];
  for (const [id, admission] of facts.admissions) {
    boundaries.push({ id, time: admission.time, seq: admission.seq, direction: 1 });
    const terminal = facts.terminals.get(id);
    if (terminal && ["completed", "failed", "cancelled"].includes(terminal.event)) {
      boundaries.push({ id, time: terminal.monotonic_ms, seq: terminal.seq, direction: -1 });
    }
  }
  boundaries.sort((a, b) => a.seq - b.seq);
  for (const boundary of boundaries) {
    const unit = byID.get(boundary.id);
    if (!unit) continue;
    for (const [resource, amount] of Object.entries(unit.resource_claims)) add(resource, boundary.direction * amount, boundary.time);
  }
  const completedAt = facts.finalMonotonicMs;
  for (const [resource, startedAt] of saturationStart) {
    saturationMs.set(resource, (saturationMs.get(resource) ?? 0) + completedAt - startedAt);
  }
  const waits = [...facts.waits.values()];
  return {
    requested_capacity: Object.fromEntries(
      Object.keys(capacities).sort().map((resource) => [resource, Math.max(...graph.units.map((unit) => unit.resource_claims[resource] ?? 0))]),
    ),
    resolved_capacity: capacities,
    peak_use: Object.fromEntries([...peak.entries()].sort(([left], [right]) => left.localeCompare(right))),
    saturation_ms: Object.fromEntries([...saturationMs.entries()].sort(([left], [right]) => left.localeCompare(right))),
    wait_events: waits.length,
    blocked_units: [...facts.waits.keys()].sort(),
    resource_holders: [...new Set(waits.flatMap((event) => event.blocking_unit_ids ?? []))].sort(),
  };
}
