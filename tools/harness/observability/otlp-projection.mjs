import { createHash } from "node:crypto";
const harnessScope = "cartulary.harness.execution";
const sha256 = (value) => createHash("sha256").update(value).digest("hex");

function eventIntervals(run) {
  const intervals = [];
  for (const [unitID, event] of run.terminal) {
    if (["completed", "failed", "cancelled"].includes(event.event)) {
      const start = run.started.get(unitID);
      if (start !== undefined) {
        intervals.push({
          unit_id: unitID,
          start_ms: start,
          end_ms: event.monotonic_ms,
          status: event.status,
        });
      }
    }
  }
  return intervals.sort(
    (left, right) =>
      left.start_ms - right.start_ms || left.unit_id.localeCompare(right.unit_id),
  );
}

export function otlpProjection(run) {
  const startedNs = BigInt(Date.parse(run.manifest.started_at)) * 1_000_000n;
  const timestamp = (ms) => (startedNs + BigInt(ms) * 1_000_000n).toString();
  const attr = (key, value) => ({ key, value: { stringValue: value } });
  const resource = { attributes: [attr("service.name", "cartulary.harness")] };
  const attributes = [attr("cartulary.target", run.manifest.target), attr("cartulary.command", run.manifest.command_id), attr("cartulary.status", run.summary.status)];
  const traceId = sha256(`${run.manifest.run_id}:${run.manifest.source_digest}:trace`).slice(0, 32);
  const rootID = sha256(`${traceId}:invocation`).slice(0, 16);
  const spans = [{ traceId, spanId: rootID, name: "harness.invocation", kind: 1,
    startTimeUnixNano: timestamp(0), endTimeUnixNano: timestamp(run.summary.wall_duration_ms), attributes,
    status: { code: run.summary.status === "pass" ? 1 : 2 } },
    ...eventIntervals(run).map((interval) => ({ traceId, spanId: sha256(`${traceId}:${interval.unit_id}`).slice(0, 16), parentSpanId: rootID,
      // Raw unit IDs can contain test symbols: local correlation stays local.
      name: `harness.${run.registrations.get(interval.unit_id)?.kind ?? "unit"}`,
      kind: 1, startTimeUnixNano: timestamp(interval.start_ms), endTimeUnixNano: timestamp(interval.end_ms),
      attributes, status: { code: interval.status === "passed" ? 1 : 2 } }))];
  const values = {
    "invocation.duration": run.summary.wall_duration_ms,
    "dependency.critical_path": run.summary.actual_dependency_critical_path_ms,
    "scheduler.queue_wait": [...run.intervals.values()].reduce((sum, value) => sum + value.queue_ms, 0),
    "scheduler.resource_blocking": run.summary.timing_accounting.resource_blocking_ms,
    "invocation.unattributed": run.summary.timing_accounting.unattributed_ms,
  };
  const metrics = Object.entries(values).map(([name, value]) => ({ name: `cartulary.harness.${name}`, unit: "ms",
    gauge: { dataPoints: [{ asInt: String(value), timeUnixNano: timestamp(run.summary.wall_duration_ms), attributes }] } }));
  return { traceOTLP: { resourceSpans: [{ resource, scopeSpans: [{ scope: { name: harnessScope }, spans }] }] },
    metricsOTLP: { resourceMetrics: [{ resource, scopeMetrics: [{ scope: { name: harnessScope }, metrics }] }] } };
}

