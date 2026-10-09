import { resourceCapacities } from "../scheduler/work-graph/capability.mjs";

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

export function unitIntervals(events) {
  const starts = new Map();
  const waits = new Map();
  const intervals = new Map();
  for (const event of events) {
    if (event.event === "wait_started") waits.set(event.unit_id, event.monotonic_ms);
    if (event.event === "started") starts.set(event.unit_id, event.monotonic_ms);
    if (["completed", "failed", "cancelled", "skipped"].includes(event.event) && waits.has(event.unit_id)) {
      intervals.set(event.unit_id, {
        start: starts.get(event.unit_id) ?? event.monotonic_ms,
        queue_ms: (starts.get(event.unit_id) ?? event.monotonic_ms) - waits.get(event.unit_id),
        end: event.monotonic_ms,
      });
    }
  }
  return intervals;
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
  events,
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
  const started = new Map();
  const fixtureAcquired = new Map();
  const intervals = [];
  const resourceWaitStarted = new Map();
  const resourceWaitIntervals = [];
  let cleanupStarted = null;
  let firstUnitStarted = null;
  let processCount = 0;
  for (const event of events) {
    if (event.event === "started" && byID.has(event.unit_id)) {
      started.set(event.unit_id, event.monotonic_ms);
      firstUnitStarted ??= event.monotonic_ms;
      processCount += 1;
    }
    if (event.event === "fixture_acquired" && byID.has(event.unit_id)) {
      fixtureAcquired.set(event.unit_id, event.monotonic_ms);
    }
    if (
      event.event === "wait_started" &&
      byID.has(event.unit_id) &&
      !resourceWaitStarted.has(event.unit_id)
    ) {
      resourceWaitStarted.set(event.unit_id, event.monotonic_ms);
    }
    if (event.event === "wait_ended" && resourceWaitStarted.has(event.unit_id)) {
      resourceWaitIntervals.push({
        start: resourceWaitStarted.get(event.unit_id),
        end: event.monotonic_ms,
      });
      resourceWaitStarted.delete(event.unit_id);
    }
    if (event.event === "cleanup_started") cleanupStarted = event.monotonic_ms;
    if (
      ["completed", "failed", "cancelled"].includes(event.event) &&
      started.has(event.unit_id)
    ) {
      const unit = byID.get(event.unit_id);
      const start = started.get(event.unit_id);
      const acquired = fixtureAcquired.get(event.unit_id);
      if (acquired !== undefined && acquired > start) {
        intervals.push({ start, end: acquired, bucket: "fixture" });
      }
      const executionStart = acquired ?? start;
      const bucket = ["finalizer", "projection"].includes(unit.kind)
        ? "collation"
        : ["artifact", "readiness"].includes(unit.kind)
          ? "setup"
          : "execution";
      if (event.monotonic_ms > executionStart) {
        intervals.push({ start: executionStart, end: event.monotonic_ms, bucket });
      }
    }
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

export function projectResourcePressure(events, graph, snapshot) {
  const byID = new Map(graph.units.map((unit) => [unit.unit_id, unit]));
  const capacities = Object.fromEntries(resourceCapacities(snapshot));
  const active = new Map();
  const saturationStart = new Map();
  const saturationMs = new Map();
  const peak = new Map();
  const admitted = new Set();
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
  for (const event of events) {
    const unit = byID.get(event.unit_id);
    if (!unit) continue;
    if (event.event === "admitted") {
      admitted.add(event.unit_id);
      for (const [resource, amount] of Object.entries(unit.resource_claims)) add(resource, amount, event.monotonic_ms);
    }
    if (
      ["completed", "failed", "cancelled"].includes(event.event) &&
      admitted.delete(event.unit_id)
    ) {
      for (const [resource, amount] of Object.entries(unit.resource_claims)) add(resource, -amount, event.monotonic_ms);
    }
  }
  const completedAt = events.at(-1)?.monotonic_ms ?? 0;
  for (const [resource, startedAt] of saturationStart) {
    saturationMs.set(resource, (saturationMs.get(resource) ?? 0) + completedAt - startedAt);
  }
  const waits = events.filter((event) => event.event === "wait_started");
  return {
    requested_capacity: Object.fromEntries(
      Object.keys(capacities).sort().map((resource) => [resource, Math.max(...graph.units.map((unit) => unit.resource_claims[resource] ?? 0))]),
    ),
    resolved_capacity: capacities,
    peak_use: Object.fromEntries([...peak.entries()].sort(([left], [right]) => left.localeCompare(right))),
    saturation_ms: Object.fromEntries([...saturationMs.entries()].sort(([left], [right]) => left.localeCompare(right))),
    wait_events: waits.length,
    blocked_units: [...new Set(waits.map((event) => event.unit_id))].sort(),
    resource_holders: [...new Set(waits.flatMap((event) => event.blocking_unit_ids ?? []))].sort(),
  };
}
