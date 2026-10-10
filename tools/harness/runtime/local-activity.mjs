import { randomUUID } from "node:crypto";
import { readLaunchContext } from "./launch-context.mjs";
import { executionObservation, resumeExecutionObservation } from "./execution-observations.mjs";

const clockIdentity = `clock:${randomUUID()}`;
export const nodeActivityClock = () => ({ clock: "node_monotonic", clock_identity: clockIdentity,
  resolution_ms: 0.000001, monotonic_ms: Number(process.hrtime.bigint()) / 1e6 });
const unavailable = { clock: "unavailable", clock_identity: null, resolution_ms: null, monotonic_ms: null };
export function activityClock(read) { try { return read(); } catch { return unavailable; } }
export function activityTiming(start, end) {
  if (!start?.clock_identity || start.clock_identity !== end?.clock_identity || start.clock !== end.clock ||
      !Number.isFinite(start.monotonic_ms) || !Number.isFinite(end.monotonic_ms) || end.monotonic_ms < start.monotonic_ms) {
    return { clock: "unavailable", clock_identity: null, resolution_ms: null, duration_ms: null, availability: "unavailable" };
  }
  return { clock: start.clock, clock_identity: start.clock_identity, resolution_ms: start.resolution_ms,
    duration_ms: end.monotonic_ms - start.monotonic_ms, availability: "available" };
}
export function startActivity(options, read = nodeActivityClock) {
  let launch;
  try { launch = options.launch ?? readLaunchContext(options.environment); } catch { launch = null; }
  const initial = { kind: "activity", ref: `activity:${randomUUID()}`, activity: options.activity,
    unit_id: launch?.unit_id ?? null, invocation_id: launch?.invocation_id ?? null,
    allocation_ref: options.allocationRef ?? options.environment.CARTULARY_HARNESS_ALLOCATION_REF ?? null,
    outcome: "incomplete", clock: "unavailable", clock_identity: null, resolution_ms: null,
    duration_ms: null, availability: "not_observed" };
  const observation = executionObservation(options, initial);
  return { initial, token: observation.token ?? null, start: activityClock(read) };
}
export function finishActivity(options, pending, outcome, read = nodeActivityClock) {
  const record = { ...pending.initial, ...activityTiming(pending.start, activityClock(read)), outcome };
  resumeExecutionObservation(options, pending.token).update(record);
  return record;
}
export function beginActivity(options, read = nodeActivityClock) {
  const pending = startActivity(options, read);
  let settled = false;
  return { finish(outcome) { if (settled) return null; settled = true; return finishActivity(options, pending, outcome, read); } };
}
