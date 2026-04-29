import path from "node:path";
import {
  formatResourceMap as formatSchedulerResourceMap,
  resourceLimitSummary as schedulerResourceLimitSummary,
  resourceMapToObject as schedulerResourceMapToObject,
} from "./scheduler-resources.mjs";

function configuredProgressIntervalMs() {
  const raw = process.env.CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS;
  if (raw === undefined || raw === "") {
    return 10_000;
  }
  const value = Number(raw);
  if (!Number.isInteger(value) || value < 1) {
    throw new Error("CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS must be a positive integer");
  }
  return value;
}

export const schedulerProgressIntervalMs = configuredProgressIntervalMs();

export function verboseSchedulerOutput() {
  return process.env.VERBOSE === "1" || process.env.CI_VERBOSE === "1";
}

export function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

export function relToRepo(repoRoot, value) {
  if (!value) {
    return "";
  }
  const normalized = normalizePath(value);
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = normalizePath(path.relative(repoRoot, value));
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

export function schedulerTargetDir(repoRoot, target) {
  const runID = process.env.CARTULARY_TEST_RUN_ID || "adhoc";
  const configured = process.env.CARTULARY_TEST_RESULTS_DIR;
  const resultsRoot = configured
    ? path.isAbsolute(configured)
      ? configured
      : path.join(repoRoot, configured)
    : path.join(repoRoot, ".cartulary", "test-results");
  return path.join(resultsRoot, runID, target);
}

export function schedulerLogDir(repoRoot, target) {
  return path.join(schedulerTargetDir(repoRoot, target), "scheduler-logs");
}

export function formatResourceMap(values) {
  return formatSchedulerResourceMap(values);
}

export function resourceMapToObject(values) {
  return schedulerResourceMapToObject(values);
}

export function formatResourceList(values) {
  if (values.length === 0) {
    return "none";
  }
  return values.join(",");
}

export function sortedUnique(values) {
  return Array.from(new Set(values.filter(Boolean))).sort((left, right) => left.localeCompare(right));
}

export function schedulerWaitingOnForUnit(unit, completed) {
  return sortedUnique((unit.needs ?? []).filter((need) => !completed.has(need)));
}

export function schedulerWaitingOnForUnits(units, completed) {
  return sortedUnique(units.flatMap((unit) => schedulerWaitingOnForUnit(unit, completed)));
}

export function schedulerBlockedUnitRecords({
  dependencyBlocked = [],
  resourceBlocked = [],
  completed,
  blockedResourcesForUnit = () => [],
}) {
  const recordsByWorkUnit = new Map();
  const ensureRecord = (unit) => {
    const workUnit = unit.label ?? unit.target ?? unit.id ?? unit.aggregateTarget;
    if (!recordsByWorkUnit.has(workUnit)) {
      recordsByWorkUnit.set(workUnit, {
        work_unit: workUnit,
        waiting_on: [],
        blocked_resources: [],
      });
    }
    return recordsByWorkUnit.get(workUnit);
  };

  for (const unit of dependencyBlocked) {
    ensureRecord(unit).waiting_on = schedulerWaitingOnForUnit(unit, completed);
  }
  for (const unit of resourceBlocked) {
    ensureRecord(unit).blocked_resources = sortedUnique(blockedResourcesForUnit(unit));
  }

  return Array.from(recordsByWorkUnit.values()).sort((left, right) =>
    left.work_unit.localeCompare(right.work_unit),
  );
}

export function formatLabelList(values, limit = 3) {
  if (values.length === 0) {
    return "none";
  }
  const displayed = values.slice(0, limit);
  const suffix = values.length > limit ? `,+${values.length - limit}` : "";
  return `${displayed.join(",")}${suffix}`;
}

export function formatDurationMs(value) {
  if (!Number.isFinite(value)) {
    return "0.00s";
  }
  return `${(value / 1000).toFixed(2)}s`;
}

export function formatSlowestWork(work) {
  if (work.length === 0) {
    return "none";
  }
  return work.map((entry) => `${entry.label}:${formatDurationMs(entry.duration_ms)}`).join(",");
}

function countMapEntries(values) {
  if (values instanceof Map) {
    return Array.from(values.entries());
  }
  if (values && typeof values === "object" && !Array.isArray(values)) {
    return Object.entries(values);
  }
  return [];
}

export function formatCountMap(values) {
  const entries = countMapEntries(values).sort((left, right) => left[0].localeCompare(right[0]));
  if (entries.length === 0) {
    return "none";
  }
  return entries.map(([key, value]) => `${key}:${value}`).join(",");
}

export function schedulerUnitGroup(unit) {
  return unit.group ?? unit.aggregateTarget ?? unit.class ?? unit.target ?? unit.label ?? "work";
}

export function schedulerActiveGroups(units) {
  const groups = new Map();
  for (const unit of units) {
    const group = schedulerUnitGroup(unit);
    groups.set(group, (groups.get(group) ?? 0) + 1);
  }
  return groups;
}

export function schedulerBlockedBy({ reason = null, blockedResources = [] } = {}) {
  const values = new Set();
  if (reason) {
    for (const entry of String(reason).split(",")) {
      const normalized = entry.trim();
      if (normalized && normalized !== "none" && normalized !== "resources") {
        values.add(normalized);
      }
    }
  }
  for (const resource of blockedResources) {
    if (resource) {
      values.add(resource);
    }
  }
  return Array.from(values).sort((left, right) => left.localeCompare(right));
}

export function schedulerSlowestRunningRecord(runningUnits, startedAt, now = Date.now()) {
  let slowest = null;
  for (const unit of runningUnits) {
    const started = startedAt.get(unit.id ?? unit.target ?? unit.label);
    if (!Number.isFinite(started)) {
      continue;
    }
    const durationMs = Math.max(0, now - started);
    if (!slowest || durationMs > slowest.duration_ms || (durationMs === slowest.duration_ms && unit.label.localeCompare(slowest.label) < 0)) {
      slowest = { label: unit.label, duration_ms: durationMs };
    }
  }
  return slowest;
}

export function formatSlowestRunning(value) {
  if (!value) {
    return "none";
  }
  return `${value.label}:${formatDurationMs(value.duration_ms)}`;
}

export function schedulerProgressSnapshot({
  runningUnits,
  startedAt,
  now = Date.now(),
  reason = "none",
  blockedResources = [],
  waitingOn = [],
  unblocksAfter = "none",
}) {
  const activeGroups = schedulerActiveGroups(runningUnits);
  return {
    activeGroups,
    blockedBy: schedulerBlockedBy({ reason, blockedResources }),
    waitingOn,
    unblocksAfter: unblocksAfter && unblocksAfter !== "none" ? unblocksAfter : null,
    slowestRunning: schedulerSlowestRunningRecord(runningUnits, startedAt, now),
  };
}

export function schedulerProgressEventFields(snapshot) {
  return {
    active_groups: resourceMapToObject(snapshot.activeGroups),
    blocked_by: snapshot.blockedBy,
    unblocks_after: snapshot.unblocksAfter,
    slowest_running: snapshot.slowestRunning,
  };
}

export function resourceLimitSummary(resourceLimits, preferred = []) {
  return schedulerResourceLimitSummary(resourceLimits, preferred);
}

export function countBy(values, field) {
  const counts = new Map();
  for (const value of values) {
    counts.set(value[field], (counts.get(value[field]) ?? 0) + 1);
  }
  return Array.from(counts.entries())
    .sort((left, right) => String(left[0]).localeCompare(String(right[0])))
    .map(([key, count]) => `${key}:${count}`)
    .join(",");
}

export function topWeightedUnits(workUnits, limit = 5) {
  return [...workUnits]
    .sort((left, right) => right.weight - left.weight || left.label.localeCompare(right.label))
    .slice(0, limit)
    .map((unit) => `${unit.label}:${unit.weight}`)
    .join(",");
}

function bracketedFields(fields) {
  return fields.filter((field) => field !== null && field !== undefined && field !== "").join(" ");
}

export function schedulerTelemetryLine(prefix, target, event, fields) {
  const suffix = bracketedFields(fields);
  return `[${prefix}] ${target} ${event}${suffix ? ` ${suffix}` : ""}\n`;
}

export function writeSchedulerTelemetry(stream, prefix, target, event, fields) {
  stream.write(schedulerTelemetryLine(prefix, target, event, fields));
}

export function schedulerStartLine({
  prefix,
  target,
  workUnitCount,
  resourceLimits,
  preferredResources = [],
  finalizerCount = null,
  workUnits = [],
  artifacts = "",
  extraFields = [],
}) {
  const fields = [
    `work_units=${workUnitCount}`,
    finalizerCount === null ? null : `finalizers=${finalizerCount}`,
    `capacity={${resourceLimitSummary(resourceLimits, preferredResources)}}`,
  ];
  if (workUnits.length > 0 && workUnits.some((unit) => unit.class !== undefined)) {
    fields.push(`classes={${countBy(workUnits, "class")}}`);
  }
  if (workUnits.length > 0 && workUnits.some((unit) => unit.type !== undefined)) {
    fields.push(`types={${countBy(workUnits, "type")}}`);
  }
  if (workUnits.length > 0) {
    fields.push(`top_weighted=${topWeightedUnits(workUnits)}`);
  }
  if (artifacts) {
    fields.push(`artifacts=${artifacts}`);
  }
  fields.push(...extraFields);
  return schedulerTelemetryLine(prefix, target, "start", fields);
}

export function schedulerProgressLine({
  prefix,
  target,
  completed,
  total,
  running,
  pending,
  blocked,
  finalizing = null,
  activeGroups = new Map(),
  blockedBy = [],
  waitingOn = [],
  unblocksAfter = "none",
  slowestRunning = "none",
  artifacts = "",
}) {
  return schedulerTelemetryLine(prefix, target, "progress", [
    `completed_work_units=${completed}/${total}`,
    `running=${running}`,
    `pending=${pending}`,
    `blocked=${blocked}`,
    finalizing === null ? null : `finalizing=${finalizing}`,
    `active_groups=${formatCountMap(activeGroups)}`,
    `blocked_by=${formatResourceList(blockedBy)}`,
    waitingOn.length > 0 ? `waiting_on=${formatResourceList(waitingOn)}` : null,
    `unblocks_after=${unblocksAfter || "none"}`,
    `slowest_running=${formatSlowestRunning(slowestRunning)}`,
    artifacts ? `artifacts=${artifacts}` : null,
  ]);
}

export function schedulerNestedProgressLine({
  prefix,
  target,
  workUnit,
  nestedTarget,
  completed,
  total,
  running,
  pending,
  blocked,
  finalizing = null,
  activeGroups = {},
  blockedBy = [],
  waitingOn = [],
  unblocksAfter = null,
  slowestRunning = null,
  artifacts = "",
}) {
  return schedulerTelemetryLine(prefix, target, "nested-progress", [
    `work_unit=${workUnit}`,
    `nested_target=${nestedTarget}`,
    `completed_work_units=${completed}/${total}`,
    `running=${running}`,
    `pending=${pending}`,
    `blocked=${blocked}`,
    finalizing === null ? null : `finalizing=${finalizing}`,
    `active_groups=${formatCountMap(activeGroups)}`,
    `blocked_by=${formatResourceList(blockedBy)}`,
    waitingOn.length > 0 ? `waiting_on=${formatResourceList(waitingOn)}` : null,
    `unblocks_after=${unblocksAfter || "none"}`,
    `slowest_running=${formatSlowestRunning(slowestRunning)}`,
    artifacts ? `artifacts=${artifacts}` : null,
  ]);
}

export function schedulerSummaryLine({
  prefix,
  target,
  status,
  completed,
  total,
  failed,
  failureClass = null,
  skipped = 0,
  finalizerFailures = 0,
  slowest = [],
  artifacts = "",
}) {
  return schedulerTelemetryLine(prefix, target, "summary", [
    `status=${status}`,
    status === "pass" ? null : `failure_class=${failureClass ?? "helper"}`,
    `completed_work_units=${completed}/${total}`,
    `failed=${failed ?? "none"}`,
    skipped > 0 ? `skipped=${skipped}` : null,
    finalizerFailures > 0 ? `finalizer_failures=${finalizerFailures}` : null,
    `slowest=${formatSlowestWork(slowest)}`,
    artifacts ? `artifacts=${artifacts}` : null,
  ]);
}

export function schedulerDryRunLine({
  target,
  manifest,
  resourceLimits,
  preferredResources = [],
  workUnits = [],
  dependencies = null,
  finalizerCount = null,
  extraFields = [],
}) {
  const fields = [
    `manifest=${manifest}`,
    `resource_limits={${resourceLimitSummary(resourceLimits, preferredResources)}}`,
    `work_units=${workUnits.length}`,
  ];
  if (dependencies !== null) {
    fields.push(`dependencies=${dependencies}`);
  }
  if (workUnits.some((unit) => unit.class !== undefined)) {
    fields.push(`classes={${countBy(workUnits, "class")}}`);
  }
  if (workUnits.some((unit) => unit.type !== undefined)) {
    fields.push(`types={${countBy(workUnits, "type")}}`);
  }
  if (finalizerCount !== null) {
    fields.push(`finalizers=${finalizerCount}`);
  }
  fields.push(`top_weighted=${topWeightedUnits(workUnits)}`);
  fields.push(...extraFields);
  return `[DRY-RUN] ${target} ${bracketedFields(fields)}\n`;
}
