import { createHash } from "node:crypto";
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import { repoRoot } from "../contract/index.mjs";
import { compareRunMetadata } from "./run-comparison.mjs";
import { resourceProjection } from "./resource-projection.mjs";
import { intervalUnion } from "../evidence-accounting/index.mjs";
import { readLocalFile } from "../runtime/secure-local-files.mjs";
import { validateSchemaSync } from "../contract/index.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";

export const harnessScope = "cartulary.harness.execution";
function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function sha256(value) {
  return createHash("sha256").update(value).digest("hex");
}

function policy() {
  const owner = readJSON(path.join(repoRoot, "tools/task_surface_owner.json"));
  return owner.observability_policy;
}

export function observabilityRequiredTarget(target) {
  return policy().required_targets.includes(target);
}

export function executionProfileDigests() {
  const host = JSON.stringify({
    architecture: process.arch,
    available_parallelism: os.availableParallelism(),
    logical_cpus: os.cpus().length,
    platform: process.platform,
  });
  return {
    host: sha256(host),
    capacity: sha256(
      readFileSync(path.join(repoRoot, "tools/scheduler_resource_registry.json")),
    ),
    workload: sha256(
      readFileSync(path.join(repoRoot, "tools/harness_work_graph_owner.json")),
    ),
    toolchain: sha256(readFileSync(path.join(repoRoot, "tools/toolchain_pins.json"))),
  };
}

function isCanonicalRun(dir) {
  return ["run-manifest.json", "run-summary.json", "unit-events.ndjson"].every(
    (name) => existsSync(path.join(dir, name)),
  );
}

export function resolveRunDir(resultsDir, runID = "", { allowNewest = true } = {}) {
  const selected = path.resolve(repoRoot, resultsDir);
  if (!existsSync(selected)) throw new Error("RESULTS_DIR does not exist");
  if (runID) {
    const candidate = path.basename(selected) === runID ? selected : path.join(selected, runID);
    if (!isCanonicalRun(candidate)) {
      throw new Error("RUN_ID does not identify canonical retained evidence");
    }
    return candidate;
  }
  if (isCanonicalRun(selected)) return selected;
  if (!allowNewest) {
    throw new Error("RUN_ID is required when RESULTS_DIR names a result root");
  }
  const candidates = readdirSync(selected, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => path.join(selected, entry.name))
    .filter(isCanonicalRun)
    .sort(
      (left, right) =>
        statSync(right).mtimeMs - statSync(left).mtimeMs || right.localeCompare(left),
    );
  if (candidates.length === 0) {
    throw new Error("RESULTS_DIR contains no canonical retained harness run");
  }
  return candidates[0];
}

export function resolveExactRunDir(resultsDir, runID = "") {
  return resolveRunDir(resultsDir, runID, { allowNewest: false });
}

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

async function retainedProjection(runDir) {
  const run = await validateCanonicalRun(runDir);
  const sourceDigests = [
    run.manifest.source_digest,
    run.manifest.toolchain_digest,
    run.manifest.system_digest,
    run.manifest.graph_digest,
  ];
  const result = otlpProjection(run);
  return {
    run,
    index: {
      schema_id: "cartulary.harness_canonical_observability.v1",
      status: "complete",
      invocations: [
        {
          target: run.manifest.target,
          run_id: run.manifest.run_id,
          source_digests: sourceDigests,
        },
      ],
    },
    built: [{ result }],
  };
}

export async function captureExecutionContext(runDir, metadata = {}) {
  const retained = await retainedProjection(runDir);
  return { manifest: retained.run.manifest, metadata };
}

export async function loadRetainedExecutionContext(runDir) {
  return captureExecutionContext(runDir);
}

export async function reconstructObservability(runDir) {
  return retainedProjection(runDir);
}

export async function loadRetainedObservability(runDir) {
  return retainedProjection(runDir);
}

export function writePartialObservability(_runDir, error) {
  return {
    status: "partial",
    diagnostic: `${error instanceof Error ? error.name : "Error"}:canonical-evidence-unavailable`,
  };
}

export async function finalizeObservabilitySafely(runDir) {
  try {
    const retained = await retainedProjection(runDir);
    return { status: "complete", retained };
  } catch (error) {
    // Step and target wrappers can finish before the graph scheduler publishes
    // its atomic evidence set. They must not create a second timing authority.
    return {
      status: "skipped",
      diagnostic: `${error instanceof Error ? error.name : "Error"}:canonical-evidence-pending`,
    };
  }
}

export function deterministicBytes(result) {
  return `${JSON.stringify(result, null, 2)}\n`;
}

export async function performanceExplanation(runDir, target = "", { comparisonDir = "", comparison = "equivalent" } = {}) {
  const run = await validateCanonicalRun(runDir);
  const comparisonResult = comparisonDir ? compareRunMetadata(run, await validateCanonicalRun(comparisonDir), comparison) : null;
  const summary = run.targetSummaries.get(target || run.manifest.target);
  if (!summary) throw new Error("selected target has no canonical target summary");
  const members = new Set(summary.unit_ids);
  const phases = [...run.phases.values()].filter((phase) => members.has(phase.unit_id));
  const byPhase = new Map();
  for (const phase of phases) {
    if (!byPhase.has(phase.phase)) byPhase.set(phase.phase, []);
    if (phase.end_ms !== null) byPhase.get(phase.phase).push({ start: phase.start_ms, end: phase.end_ms });
  }
  const selectedIntervals = [...run.intervals].filter(([id]) => members.has(id));
  const units = selectedIntervals.map(([id, interval]) => ({ unit_id: id, kind: run.registrations.get(id).kind,
    queue_ms: interval.queue_ms, execution_ms: interval.end - interval.start,
    on_invocation_dependency_path: run.summary.critical_path.includes(id) }))
    .sort((a, b) => (b.queue_ms + b.execution_ms) - (a.queue_ms + a.execution_ms) || a.unit_id.localeCompare(b.unit_id));
  let envelope = null;
  try {
    envelope = JSON.parse(readLocalFile(path.join(runDir, "diagnostics/invocation-envelope.json"), { maximum: 65536 }));
    validateSchemaSync("cartulary.harness_invocation_envelope.v1", envelope);
    for (const key of ["run_id", "source_digest", "graph_digest"]) if (envelope[key] !== run.manifest[key]) throw new Error("envelope identity mismatch");
    let end = 0;
    for (const interval of envelope.intervals) { if (interval.start_ms !== end || interval.end_ms < end) throw new Error("invalid envelope intervals"); end = interval.end_ms; }
  } catch (error) { if (error.code !== "ENOENT") throw error; }
  const covered = intervalUnion(phases.filter((phase) => phase.end_ms !== null).map((phase) => ({ start: phase.start_ms, end: phase.end_ms })));
  return {
    comparison: comparisonResult, schema_id: "cartulary.harness_performance_explanation.v1", run_id: run.manifest.run_id,
    target: summary.target, status: summary.status, canonical_wall_ms: run.summary.wall_duration_ms,
    inclusive_wall_ms: summary.inclusive_wall_ms, exclusive_wall_ms: summary.exclusive_wall_ms,
    dependency_path_ms: summary.actual_dependency_critical_path_ms,
    invocation_path_queue_ms: run.summary.critical_path.reduce((sum, id) => sum + (run.intervals.get(id)?.queue_ms ?? 0), 0),
    phase_union_ms: Object.fromEntries([...byPhase].sort(([a], [b]) => a.localeCompare(b)).map(([name, values]) => [name, intervalUnion(values)])),
    canonical_time_outside_selected_phases_ms: Math.max(0, run.summary.wall_duration_ms - covered),
    phase_overlap_ms: [...byPhase.values()].reduce((sum, values) => sum + intervalUnion(values), 0) - covered,
    units: units.slice(0, 20), units_omitted: Math.max(0, units.length - 20),
    reservations: { ...run.summary.resource_pressure,
      blocked_units: run.summary.resource_pressure.blocked_units.slice(0, 20),
      resource_holders: run.summary.resource_pressure.resource_holders.slice(0, 20),
      blocked_units_omitted: Math.max(0, run.summary.resource_pressure.blocked_units.length - 20),
      resource_holders_omitted: Math.max(0, run.summary.resource_pressure.resource_holders.length - 20) }, resources: resourceProjection(runDir, run), envelope,
    limitations: ["unit_execution_includes_harness_work", "phase_categories_may_overlap", "reservations_are_not_consumption", "resource_coverage_is_observed_partial", "guest_is_not_windows_host"],
  };
}

export async function printObservabilityPerformance(runDir, target = "", { json = false, resourcesOnly = false, comparisonDir = "", comparison = "equivalent" } = {}) {
  const value = await performanceExplanation(runDir, target, { comparisonDir, comparison });
  validateSchemaSync(value.schema_id, value);
  const serialized = `${JSON.stringify(value)}\n`;
  if (Buffer.byteLength(serialized) > 262144) throw new Error("performance explanation exceeds byte limit");
  if (json) process.stdout.write(serialized);
  else {
    const lines = [`target=${value.target} status=${value.status} canonical_wall_ms=${value.canonical_wall_ms} inclusive_wall_ms=${value.inclusive_wall_ms} exclusive_wall_ms=${value.exclusive_wall_ms} dependency_path_ms=${value.dependency_path_ms}`];
    if (!resourcesOnly) {
      lines.push(`phase_union_ms=${JSON.stringify(value.phase_union_ms)} uncovered_ms=${value.canonical_time_outside_selected_phases_ms} overlap_ms=${value.phase_overlap_ms}`);
      for (const unit of value.units) lines.push(`[UNIT] ${unit.unit_id} queue_ms=${unit.queue_ms} execution_ms=${unit.execution_ms} invocation_path=${unit.on_invocation_dependency_path}`);
    }
    if (value.comparison) lines.push(`[COMPARISON] ${JSON.stringify(value.comparison)}`);
    const resources = value.resources;
    lines.push(`[RESOURCES] availability=${resources.availability} coverage=${resources.coverage ?? "not_observed"} completeness=${resources.completeness ?? "not_observed"} samples=${resources.samples ?? "not_observed"}`);
    for (const unit of resources.unit_cpu_lower_bounds ?? []) lines.push(`[UNIT-CPU] ${unit.unit_id} observed_lifetime_cpu_us=${unit.observed_lifetime_cpu_us} coverage=lower_bound`);
    for (const scope of resources.scopes ?? []) lines.push(`[SCOPE] ${scope.scope_ref} scope=${scope.scope} unit=${scope.unit_id ?? "context_or_harness"} cpu_percent_one_core=${scope.cpu_percent_one_core ?? "not_observed"} sampled_rss_max_bytes=${scope.metrics.rss_bytes?.maximum ?? "not_observed"}`);
    lines.push(`limits=${value.limitations.join(",")}`);
    const selected = []; let bytes = 0;
    for (const line of lines) { if (selected.length >= 119 || bytes + Buffer.byteLength(line) + 1 > 7900) break; selected.push(line); bytes += Buffer.byteLength(line) + 1; }
    if (selected.length < lines.length) selected.push(`[OMITTED] lines=${lines.length - selected.length}`);
    process.stdout.write(`${selected.join("\n")}\n`);
  }
  return value;
}
