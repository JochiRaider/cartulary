import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";

import { repoRoot } from "../contract/index.mjs";
import { canonicalJSONString, semanticJSONSHA256 } from "../test-catalog/semantic-json.mjs";
import { loadRetainedObservability, resolveExactRunDir } from "./observability.mjs";

export const backendFinalizerTarget = "backend-output-finalizer";
export const performanceEvidenceSchemaID = "cartulary.harness_performance_evidence_roots.v3";
export const performanceBaselineSchemaID = "cartulary.harness_public_target_duration_baselines.v3";

function sameJSON(left, right) {
  return canonicalJSONString(left) === canonicalJSONString(right);
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

export function median(values) {
  if (values.length === 0) throw new Error("median requires at least one value");
  const sorted = [...values].sort((left, right) => left - right);
  const middle = Math.floor(sorted.length / 2);
  return sorted.length % 2 === 0
    ? (sorted[middle - 1] + sorted[middle]) / 2
    : sorted[middle];
}

export function nearestRankP90(values) {
  if (values.length === 0) throw new Error("p90 requires at least one value");
  const sorted = [...values].sort((left, right) => left - right);
  return sorted[Math.ceil(sorted.length * 0.9) - 1];
}

function durationMs(span) {
  return Number((BigInt(span.end_time_unix_nano) - BigInt(span.start_time_unix_nano)) / 1_000_000n);
}

function isoDurationMs(start, end, label) {
  const value = Date.parse(end) - Date.parse(start);
  if (!Number.isFinite(value) || value <= 0) throw new Error(`${label} has invalid timing boundaries`);
  return value;
}

export function intervalUnionMs(spans) {
  const intervals = spans
    .map((span) => {
      if (span.start_time_unix_nano !== undefined) {
        return [
          Number(BigInt(span.start_time_unix_nano) / 1_000_000n),
          Number(BigInt(span.end_time_unix_nano) / 1_000_000n),
        ];
      }
      return [Date.parse(span.start_time), Date.parse(span.end_time)];
    })
    .filter(([start, end]) => Number.isFinite(start) && Number.isFinite(end) && end > start)
    .sort((left, right) => left[0] - right[0] || left[1] - right[1]);
  let total = 0;
  let current = null;
  for (const interval of intervals) {
    if (!current || interval[0] > current[1]) {
      if (current) total += current[1] - current[0];
      current = [...interval];
    } else {
      current[1] = Math.max(current[1], interval[1]);
    }
  }
  if (current) total += current[1] - current[0];
  return total;
}

function rootRef(runDir, context) {
  const relative = path.relative(repoRoot, runDir).replaceAll("\\", "/");
  return relative !== "" && relative !== ".." && !relative.startsWith("../")
    ? relative
    : `retained:${context.run_id}`;
}

function contractsByTarget(context) {
  const result = new Map();
  for (const contract of context.measurement_contracts) {
    if (result.has(contract.target)) throw new Error(`retained context duplicates measurement contract ${contract.target}`);
    result.set(contract.target, contract);
  }
  return result;
}

export function qualificationReasons(context) {
  const reasons = new Set(context.contamination_reasons);
  if (!context.invocation_boundary_retained) reasons.add("artifact_incomplete");
  if (context.source_state !== "clean") reasons.add("dirty_source");
  if (context.status !== "passed") reasons.add("failed_execution");
  if (context.interrupted) reasons.add("interrupted_execution");
  if (context.retry_count !== 0) reasons.add("retry_observed");
  if (context.warm_eligibility !== "eligible" && reasons.size === 0) reasons.add("external_activity");
  return [...reasons].sort((left, right) => left.localeCompare(right));
}

// Narrow fixture-facing projection retained for observability contract tests.
// Qualification and performance acceptance use explicit v3 bindings below.
export function collectRetainedObservations(retained, runDir) {
  const context = retained.context;
  const reasons = qualificationReasons(context);
  const observations = new Map();
  if (reasons.length > 0) return { runDir, context, observations, reasons };
  const contracts = contractsByTarget(context);
  const { bundle, span } = rootSpan(retained);
  const rootContract = contracts.get(context.target);
  if (!rootContract) return { runDir, context, observations, reasons: ["missing_command_id"] };
  observations.set(context.target, [{ value: durationMs(span), context, contract: rootContract, root: runDir }]);
  for (const targetSpan of bundle.spans.filter((item) => item.phase === "target" && item.status === "OK" && item.span_id !== span.span_id)) {
    const contract = contracts.get(targetSpan.name);
    if (!contract || observations.has(targetSpan.name)) continue;
    if (contract.observation_eligibility === "direct_only") continue;
    observations.set(targetSpan.name, [{ value: durationMs(targetSpan), context, contract, root: runDir }]);
  }
  return { runDir, context, observations, reasons };
}

function timingFile(runDir, target) {
  return path.join(runDir, target, "target-timing.json");
}

function nativeSummary(runDir, target) {
  for (const name of ["tool-run-summary.json", "scheduler-summary.json", "target-summary.json"]) {
    const file = path.join(runDir, target, name);
    if (!existsSync(file)) continue;
    const summary = readJSON(file);
    const status = summary.status;
    if (status !== "pass" && status !== "passed") throw new Error(`${target} native summary is not terminal-successful`);
    const start = summary.started_at ?? summary.start_time;
    const end = summary.completed_at ?? summary.end_time;
    return {
      value: isoDurationMs(start, end, `${target} native summary`),
      start,
      end,
    };
  }
  throw new Error(`${target} has no terminal native summary`);
}

function nativeTargetTiming(runDir, target) {
  const file = timingFile(runDir, target);
  if (!existsSync(file)) return nativeSummary(runDir, target);
  const timing = readJSON(file);
  if (timing.target !== target || timing.status !== "pass") throw new Error(`${target} native timing is not terminal-successful`);
  const spans = (timing.buckets ?? [])
    .flatMap((bucket) => bucket.spans ?? [])
    .filter((span) => span.status === "pass");
  if (spans.length > 0) {
    const value = intervalUnionMs(spans);
    if (value <= 0) throw new Error(`${target} native timing interval union is empty`);
    return {
      value,
      start: spans.map((span) => span.start_time).sort()[0],
      end: spans.map((span) => span.end_time).sort().at(-1),
    };
  }
  return {
    value: isoDurationMs(timing.start_time, timing.end_time, `${target} native timing`),
    start: timing.start_time,
    end: timing.end_time,
  };
}

function nativeFinalizerTiming(runDir) {
  const file = timingFile(runDir, "backend-unit");
  if (!existsSync(file)) throw new Error("backend-unit native report-collation timing is missing");
  const timing = readJSON(file);
  if (timing.target !== "backend-unit" || timing.status !== "pass") {
    throw new Error("backend-unit native report-collation timing is not terminal-successful");
  }
  const bucket = timing.buckets.find((entry) => entry.name === "report_collation");
  const spans = bucket?.spans?.filter((span) => span.status === "pass") ?? [];
  const value = intervalUnionMs(spans);
  if (value <= 0) throw new Error("backend-unit native report-collation interval union is empty");
  return {
    value,
    start: spans.map((span) => span.start_time).sort()[0],
    end: spans.map((span) => span.end_time).sort().at(-1),
  };
}

function rootSpan(retained) {
  if (retained.built.length !== 1) throw new Error("retained root must contain exactly one invocation");
  const bundle = retained.built[0].result.bundle;
  const span = bundle.spans.find((item) => item.span_id === bundle.root_span_id);
  if (!span || span.status !== "OK") throw new Error("retained invocation root is not terminal-successful");
  return { bundle, span };
}

function exactAggregateSpan(retained, target) {
  const { bundle } = rootSpan(retained);
  const matches = bundle.spans.filter((span) => span.phase === "target" && span.name === target && span.status === "OK");
  const unique = new Map(matches.map((span) => [
    `${span.start_time_unix_nano}:${span.end_time_unix_nano}`,
    span,
  ]));
  if (unique.size !== 1) throw new Error(`${target} aggregate timing is not exact-once`);
  const span = [...unique.values()][0];
  return {
    value: durationMs(span),
    start: new Date(Number(BigInt(span.start_time_unix_nano) / 1_000_000n)).toISOString(),
    end: new Date(Number(BigInt(span.end_time_unix_nano) / 1_000_000n)).toISOString(),
  };
}

function schedulerEventFiles(root) {
  const files = [];
  const visit = (current) => {
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const candidate = path.join(current, entry.name);
      if (entry.isDirectory()) {
        visit(candidate);
      } else if (entry.isFile() && entry.name === "scheduler-events.jsonl") {
        files.push(candidate);
      }
    }
  };
  visit(root);
  return files.sort((left, right) => left.localeCompare(right));
}

function schedulerUnitTiming(runDir, target, { required }) {
  const occurrences = [];
  for (const file of schedulerEventFiles(runDir)) {
    const events = readFileSync(file, "utf8")
      .split("\n")
      .filter((line) => line !== "")
      .map((line, index) => {
        try {
          return JSON.parse(line);
        } catch (error) {
          throw new Error(`${target} scheduler event ${path.relative(runDir, file)}:${index + 1} is invalid JSON`, {
            cause: error,
          });
        }
      })
      .filter((event) => event.work_unit_id === target && (event.event === "start" || event.event === "finish"));
    if (events.length === 0) continue;
    const starts = events.filter((event) => event.event === "start");
    const finishes = events.filter((event) => event.event === "finish");
    if (starts.length !== 1 || finishes.length !== 1) {
      throw new Error(`${target} canonical scheduler unit does not have exactly one start/finish pair`);
    }
    const start = starts[0];
    const finish = finishes[0];
    if (finish.status !== 0) throw new Error(`${target} canonical scheduler unit is not terminal-successful`);
    if (start.target !== finish.target) throw new Error(`${target} canonical scheduler unit crosses scheduler streams`);
    const startMs = Date.parse(start.emitted_at);
    const finishMs = Date.parse(finish.emitted_at);
    if (!Number.isFinite(startMs) || !Number.isFinite(finishMs) || finishMs <= startMs) {
      throw new Error(`${target} canonical scheduler unit has invalid timing boundaries`);
    }
    occurrences.push({
      value: finishMs - startMs,
      start: new Date(startMs).toISOString(),
      end: new Date(finishMs).toISOString(),
    });
  }
  if (occurrences.length === 0 && !required) return null;
  if (occurrences.length !== 1) {
    throw new Error(`${target} canonical scheduler unit is not exact-once`);
  }
  return occurrences[0];
}

export function canonicalSchedulerUnitTiming(runDir, target) {
  return schedulerUnitTiming(runDir, target, { required: true });
}

function observationFor(record, target, timingSource, evidenceKind) {
  if (timingSource === "backend_report_collation_union") return nativeFinalizerTiming(record.runDir);
  if (timingSource === "public_invocation_envelope") {
    if (record.context.target !== target) throw new Error(`${target} public invocation timing requires a direct provider`);
    if (!record.context.invocation_boundary_retained || evidenceKind !== "strict_current") {
      throw new Error(`${target} current evidence has no invocation boundary`);
    }
    const { span } = rootSpan(record.retained);
    return {
      value: durationMs(span),
      start: record.context.started_at,
      end: record.context.ended_at,
    };
  }
  if (timingSource === "canonical_unit_interval_union") {
    if (existsSync(timingFile(record.runDir, target)) || record.context.target === target) {
      return nativeTargetTiming(record.runDir, target);
    }
    const schedulerTiming = schedulerUnitTiming(record.runDir, target, { required: false });
    if (schedulerTiming) return schedulerTiming;
    return exactAggregateSpan(record.retained, target);
  }
  if (timingSource === "canonical_timing_bucket_union") {
    if (target !== backendFinalizerTarget) {
      throw new Error(`${target} has no registered canonical timing bucket`);
    }
    return nativeFinalizerTiming(record.runDir);
  }
  throw new Error(`${target} has unsupported timing source ${timingSource}`);
}

function spanInterval(span) {
  return [
    Number(BigInt(span.start_time_unix_nano) / 1_000_000n),
    Number(BigInt(span.end_time_unix_nano) / 1_000_000n),
  ];
}

function exclusiveTimingAccounting(record) {
  const { bundle, span: root } = rootSpan(record.retained);
  const [rootStart, rootEnd] = spanInterval(root);
  const bucketByPhase = new Map([
    ["scheduler_wait", "setup"],
    ["service", "fixture"],
    ["target", "execution"],
    ["sequence_step", "execution"],
    ["scheduler_work", "execution"],
    ["runner", "execution"],
    ["artifact", "collation"],
    ["finalizer", "collation"],
    ["unattributed", "wrapper"],
  ]);
  const intervals = bundle.spans
    .filter((item) => item.span_id !== root.span_id && bucketByPhase.has(item.phase))
    .map((item) => {
      const [start, end] = spanInterval(item);
      return {
        start: Math.max(rootStart, start),
        end: Math.min(rootEnd, end),
        bucket: bucketByPhase.get(item.phase),
      };
    })
    .filter((item) => item.end > item.start);
  const boundaries = [...new Set([
    rootStart,
    rootEnd,
    ...intervals.flatMap((item) => [item.start, item.end]),
  ])].sort((left, right) => left - right);
  const totals = { setup: 0, fixture: 0, execution: 0, collation: 0, wrapper: 0 };
  const precedence = ["collation", "fixture", "execution", "setup", "wrapper"];
  for (let index = 1; index < boundaries.length; index += 1) {
    const start = boundaries[index - 1];
    const end = boundaries[index];
    if (end <= start || start < rootStart || end > rootEnd) continue;
    const active = new Set(intervals
      .filter((item) => item.start < end && item.end > start)
      .map((item) => item.bucket));
    const bucket = precedence.find((name) => active.has(name)) ?? "wrapper";
    totals[bucket] += end - start;
  }
  const hotspot = record.retained.built[0].result.hotspot;
  const processCount = bundle.spans.filter((item) =>
    item.phase === "runner" || item.phase === "scheduler_work" || item.phase === "sequence_step").length;
  return {
    critical_path_ms: hotspot.actual_dependency_critical_path_ms,
    resource_blocking_ms: hotspot.resource_blocking_ms,
    setup_ms: totals.setup,
    fixture_ms: totals.fixture,
    execution_ms: totals.execution,
    collation_ms: totals.collation,
    wrapper_ms: totals.wrapper,
    process_count: processCount,
    unattributed_ms: 0,
  };
}

function timingAccounting(records) {
  const samples = records.map((record) => exclusiveTimingAccounting(record));
  const p50 = (field) => median(samples.map((sample) => sample[field]));
  return {
    critical_path_p50_ms: p50("critical_path_ms"),
    resource_blocking_p50_ms: p50("resource_blocking_ms"),
    setup_p50_ms: p50("setup_ms"),
    fixture_p50_ms: p50("fixture_ms"),
    execution_p50_ms: p50("execution_ms"),
    collation_p50_ms: p50("collation_ms"),
    wrapper_p50_ms: p50("wrapper_ms"),
    process_count_p50: p50("process_count"),
    unattributed_p50_ms: p50("unattributed_ms"),
  };
}

function loadWindowRecord(root, window) {
  const runDir = resolveExactRunDir(root);
  const retained = loadRetainedObservability(runDir);
  const context = retained.context;
  if (context.target !== window.provider_target) {
    throw new Error(`${context.run_id} provider ${context.target} does not match ${window.provider_target}`);
  }
  if (window.evidence_kind !== "strict_current" || context.schema_id !== "cartulary.harness_execution_context.v2") {
    throw new Error(`${context.run_id} strict evidence must use execution-context v2`);
  }
  const reasons = qualificationReasons(context);
  if (reasons.length > 0) throw new Error(`${context.run_id} is not qualified: ${reasons.join(",")}`);
  return { runDir, retained, context, contracts: contractsByTarget(context) };
}

function assertUnique(items, key, label) {
  const seen = new Set();
  for (const item of items) {
    const identity = item[key];
    if (seen.has(identity)) throw new Error(`${label} duplicates ${identity}`);
    seen.add(identity);
  }
}

function assertSame(target, values, label) {
  if (values.some((value) => !sameJSON(value, values[0]))) throw new Error(`${target} window has mismatched ${label}`);
}

const requiredImprovementTargets = new Set([
  "agent-finalize",
  "browser-e2e",
  "browser-e2e-a11y",
  "browser-e2e-measurement",
  "browser-e2e-stateful",
  "browser-e2e-visual",
  "browser-e2e-webserver-backed",
  "check",
  "ci",
  "go-vulncheck",
  "harness-contract",
  "release-check",
  "test",
  "test-fast",
]);

function targetStatistics(binding, window, cold, warmup, samples, { diagnostic = false } = {}) {
  const target = binding.target;
  const records = [cold, warmup, ...samples];
  const contracts = records.map((record) => record.contracts.get(target === backendFinalizerTarget ? "backend-unit" : target));
  if (contracts.some((contract) => contract === undefined)) throw new Error(`${target} has no retained measurement contract`);
  const contexts = records.map((record) => record.context);
  for (const field of [
    "commit", "source_snapshot_sha256", "host_profile_sha256", "capacity_profile_sha256", "toolchain_profile_sha256",
  ]) assertSame(target, contexts.map((context) => context[field]), field);
  for (const field of [
    "command_id", "measurement_profile_id", "canonical_inputs", "workload_evidence_profile_sha256", "execution_policy_sha256",
  ]) assertSame(target, contracts.map((contract) => contract[field]), field);
  const allObservations = records.map((record) => observationFor(record, target, binding.timing_source, window.evidence_kind));
  for (let index = 1; index < allObservations.length; index += 1) {
    if (Date.parse(allObservations[index].start) <= Date.parse(allObservations[index - 1].start)) {
      throw new Error(`${target} window is not in strictly increasing order`);
    }
  }
  const contract = contracts[0];
  const values = allObservations.slice(2).map((observation) => observation.value);
  const baselineMedian = median(values);
  const mad = median(values.map((sample) => Math.abs(sample - baselineMedian)));
  if (!contract.execution_policy) {
    throw new Error(`${target} retained measurement contract has no normalized execution-policy projection`);
  }
  const executionPolicy = contract.execution_policy;
  if (semanticJSONSHA256(contract.execution_policy) !== contract.execution_policy_sha256) {
    throw new Error(`${target} retained execution-policy projection digest mismatch`);
  }
  const gate = diagnostic
    ? "diagnostic_only"
    : requiredImprovementTargets.has(target) ? "required_improvement" : "no_regression";
  const allowedPolicyTransition = contract.allowed_policy_transition ?? currentAllowedPolicyTransition(target);
  return {
    target,
    gate,
    command_id: contract.command_id,
    measurement_profile_id: contract.measurement_profile_id,
    canonical_inputs: contract.canonical_inputs,
    timing_source: binding.timing_source,
    source_commit: contexts[0].commit,
    source_snapshot_sha256: contexts[0].source_snapshot_sha256,
    host_profile_sha256: contexts[0].host_profile_sha256,
    capacity_profile_sha256: contexts[0].capacity_profile_sha256,
    toolchain_profile_sha256: contexts[0].toolchain_profile_sha256,
    workload_evidence_profile_sha256: contract.workload_evidence_profile_sha256,
    execution_policy: executionPolicy,
    execution_policy_sha256: contract.execution_policy_sha256,
    ...(allowedPolicyTransition === undefined ? {} : { allowed_policy_transition: allowedPolicyTransition }),
    sample_provider_target: window.provider_target,
    cold_root: rootRef(cold.runDir, cold.context),
    cold_ms: allObservations[0].value,
    warmup_root: rootRef(warmup.runDir, warmup.context),
    sample_roots: samples.map((record) => rootRef(record.runDir, record.context)),
    sample_count: values.length,
    samples_ms: values,
    p50_ms: baselineMedian,
    p90_ms: nearestRankP90(values),
    mad_ms: mad,
    timing_accounting: timingAccounting(samples),
  };
}

function publicTargets() {
  const manifest = readJSON(path.join(repoRoot, "tools", "task_surface_manifest.json"));
  return [...manifest.observability_policy.required_targets].sort((left, right) => left.localeCompare(right));
}

export function sourceWindowRoster(rows) {
  const windows = new Map();
  for (const row of rows) {
    const key = `${row.source_commit}\u0000${row.source_snapshot_sha256}`;
    const window = windows.get(key) ?? {
      source_commit: row.source_commit,
      source_snapshot_sha256: row.source_snapshot_sha256,
      targets: [],
    };
    window.targets.push(row.target);
    windows.set(key, window);
  }
  return [...windows.values()]
    .map((window) => ({
      ...window,
      targets: [...window.targets].sort((left, right) => left.localeCompare(right)),
    }))
    .sort((left, right) => left.source_commit.localeCompare(right.source_commit) ||
      left.source_snapshot_sha256.localeCompare(right.source_snapshot_sha256));
}

function currentAllowedPolicyTransition(target) {
  const manifest = readJSON(path.join(repoRoot, "tools", "task_surface_manifest.json"));
  const contractTarget = target === backendFinalizerTarget ? "backend-unit" : target;
  const binding = manifest.observability_policy.target_measurement_profiles
    .find((entry) => entry.target === contractTarget);
  const profile = manifest.observability_policy.measurement_profiles
    .find((entry) => entry.profile_id === binding?.profile_id);
  return binding?.allowed_policy_transition ?? profile?.allowed_policy_transition;
}

export function buildQualifiedBaseline(windows, bindings, {
  internalWindows = [],
  internalBindings = [],
  rejectedRoots = [],
  role = "reference",
} = {}) {
  assertUnique(windows, "window_id", `${role} windows`);
  assertUnique(bindings, "target", `${role} bindings`);
  assertUnique(internalWindows, "window_id", `${role} internal windows`);
  assertUnique(internalBindings, "target", `${role} internal bindings`);
  const loadRows = (selectedWindows, selectedBindings, { diagnostic = false } = {}) => {
    const windowsByID = new Map(selectedWindows.map((window) => [window.window_id, window]));
    const loadedWindows = new Map();
    for (const window of selectedWindows) {
      if (window.evidence_kind !== "strict_current") {
        throw new Error(`${window.window_id} must use strict-current evidence`);
      }
      if (window.measured_roots.length !== 5 && window.measured_roots.length !== 6) {
        throw new Error(`${window.window_id} must contain five or six measured roots`);
      }
      const roots = [window.cold_root, window.warmup_root, ...window.measured_roots];
      if (new Set(roots).size !== roots.length) throw new Error(`${window.window_id} duplicates evidence roots`);
      loadedWindows.set(window.window_id, {
        cold: loadWindowRecord(window.cold_root, window),
        warmup: loadWindowRecord(window.warmup_root, window),
        samples: window.measured_roots.map((root) => loadWindowRecord(root, window)),
      });
    }
    const referencedWindows = new Set();
    const rows = selectedBindings.map((binding) => {
      const window = windowsByID.get(binding.window_id);
      if (!window) throw new Error(`${binding.target} binds unknown window ${binding.window_id}`);
      referencedWindows.add(binding.window_id);
      const loaded = loadedWindows.get(binding.window_id);
      const row = targetStatistics(binding, window, loaded.cold, loaded.warmup, loaded.samples, { diagnostic });
      if (row.sample_count === 5 && row.mad_ms > row.p50_ms * 0.1) {
        throw new Error(`${binding.target} requires a sixth measured observation because MAD exceeds ten percent of p50`);
      }
      return row;
    }).sort((left, right) => left.target.localeCompare(right.target));
    const unusedWindows = [...windowsByID.keys()].filter((windowID) => !referencedWindows.has(windowID));
    if (unusedWindows.length > 0) throw new Error(`${role} contains unused windows: ${unusedWindows.join(",")}`);
    return rows;
  };
  const targets = loadRows(windows, bindings);
  const internalDiagnostics = loadRows(internalWindows, internalBindings, { diagnostic: true });
  const required = publicTargets();
  const actualPublic = targets.map((row) => row.target);
  if (!sameJSON(actualPublic, required)) {
    throw new Error(`baseline does not contain the exact ${required.length}-target public inventory`);
  }
  const publicSet = new Set(required);
  for (const row of internalDiagnostics) {
    if (publicSet.has(row.target)) throw new Error(`${row.target} is public and cannot be an internal diagnostic`);
  }
  const totalP50 = targets.reduce((total, row) => total + row.p50_ms, 0);
  return {
    schema_id: performanceBaselineSchemaID,
    status: "qualified",
    qualification: "strict_current",
    targets,
    internal_diagnostics: internalDiagnostics,
    source_windows: sourceWindowRoster(targets),
    public_entrypoint_portfolio: {
      target_count: required.length,
      targets: required,
      total_p50_ms: totalP50,
    },
    rejected_roots: [...rejectedRoots].sort((left, right) => left.root.localeCompare(right.root)),
  };
}

export function comparePerformanceWindows(manifest) {
  return compareQualifiedBaselines(
    buildQualifiedBaseline(manifest.reference_windows, manifest.reference_bindings, {
      internalWindows: manifest.reference_internal_windows ?? [],
      internalBindings: manifest.reference_internal_bindings ?? [],
      rejectedRoots: manifest.reference_rejected_roots ?? [],
      role: "reference",
    }),
    buildQualifiedBaseline(manifest.candidate_windows, manifest.candidate_bindings, {
      internalWindows: manifest.candidate_internal_windows ?? [],
      internalBindings: manifest.candidate_internal_bindings ?? [],
      rejectedRoots: manifest.candidate_rejected_roots ?? [],
      role: "candidate",
    }),
  );
}

function validatePolicyTransition(target, baselineRow, candidateRow) {
  const transition = baselineRow.allowed_policy_transition;
  if (transition !== candidateRow.allowed_policy_transition) {
    throw new Error(`${target} baseline and candidate have mismatched policy transition contracts`);
  }
  if (transition === undefined) {
    if (!sameJSON(baselineRow.execution_policy, candidateRow.execution_policy)) {
      throw new Error(`${target} has an undeclared execution-policy change`);
    }
    return;
  }
  if (sameJSON(baselineRow.execution_policy, candidateRow.execution_policy)) {
    throw new Error(`${target} did not apply declared policy transition ${transition}`);
  }
  const candidateTarget = candidateRow.execution_policy?.target;
  if (candidateTarget !== target && !(target === backendFinalizerTarget && candidateTarget === "backend-unit")) {
    throw new Error(`${target} candidate policy projection has the wrong target identity`);
  }
  const knownTransitions = new Set([
    "backend_grouped_capture_and_parallel_report_emission",
    "backend_process_shard_consolidation",
    "serial_to_topology_dag",
    "browser_stack_capacity_1_to_2",
    "browser_webserver_stage_capacity_1_to_2",
  ]);
  if (!knownTransitions.has(transition)) throw new Error(`${target} has unknown policy transition ${transition}`);
  if (transition === "backend_grouped_capture_and_parallel_report_emission") {
    const backend = candidateRow.execution_policy?.backend_unit;
    if (!sameJSON(backend?.capture_grouping?.dimensions, [
      "package_selection",
      "runtime_binaries",
      "runtime_profile",
      "resource_profile",
      "fixture_profile",
      "fixture_policy",
      "fixture_budget",
      "isolation_policy",
      "evidence_class",
    ]) || backend?.capture_grouping?.raw_selectors !== "isolated" ||
      backend?.worker_pool?.formula !== "min(group_count,clamp(floor(available_parallelism/4),1,8))" ||
      backend?.worker_pool?.child_gomaxprocs !== "max(1,floor(available_parallelism/workers))" ||
      backend?.report_projection?.physical_report_parse !== "once_per_physical_report" ||
      backend?.report_projection?.emission !== "bounded_reusable_host_derived_pool" ||
      backend?.report_projection?.owner_accounting_context !== "once_per_target") {
      throw new Error(`${target} candidate policy does not implement the exact backend transition`);
    }
  }
  if (transition === "backend_process_shard_consolidation") {
    const backend = candidateRow.execution_policy?.backend_process;
    if (!sameJSON(backend?.exact_symbol_shard_profile, {
      max_symbols: 16,
      max_estimated_test_work_ms: 24_000,
    }) || backend?.capture_pool?.workers !==
      "min(group_count,clamp(floor(available_parallelism/4),1,8))" ||
      backend?.capture_pool?.child_gomaxprocs !== "available_parallelism") {
      throw new Error(`${target} candidate policy does not implement the exact backend-process transition`);
    }
  }
  if (transition === "serial_to_topology_dag") {
    const profiles = {
      small_check: [{ host_cpu: 1, host_io: 1, process: 1 }, 1, undefined],
      cpu_analysis: [{ host_cpu: 4, host_io: 1, process: 1 }, "host_cpu", undefined],
      script: [{ host_cpu: 2, host_io: 2, process: 1 }, "host_cpu", undefined],
      artifact_generation: [{ host_cpu: 2, host_io: 4, process: 1 }, "host_cpu", undefined],
      build: [{ host_cpu: 6, host_io: 3, process: 1 }, "host_cpu", undefined],
      service_validation: [{ host_cpu: 2, host_io: 4, process: 1 }, 1, undefined],
      security_analysis: [
        { host_cpu: "limit", host_io: 1, process: 1 },
        "host_cpu",
        undefined,
      ],
      nested_check: [
        { host_cpu: "limit", host_io: "limit", process: 1 },
        1,
        "sequence_to_check",
      ],
      nested_service_validation: [
        { host_cpu: 2, host_io: 4, process: 1 },
        1,
        "sequence_to_service_backed",
      ],
      nested_browser_validation: [
        { host_cpu: 4, host_io: 4, process: 1 },
        1,
        "sequence_to_service_backed",
      ],
    };
    const expectedSteps = {
      lint: [
        ["lint-go", [], "cpu_analysis"],
        ["lint-biome", [], "cpu_analysis"],
        ["frontend-import-boundary-check", [], "small_check"],
        ["backend-module-boundary-check", [], "small_check"],
        ["lint-scripts", [], "script"],
        ["lint-shell", [], "script"],
        ["frontend-typecheck", [], "cpu_analysis"],
      ],
      ci: [
        ["check", [], "nested_check", 100],
        ["harness-contract", ["check"], "cpu_analysis"],
        ["go-gosec-audit", ["check"], "security_analysis"],
        ["deployable-shape", ["check"], "small_check"],
      ],
      "release-check": [
        ["check", [], "nested_check", 100],
        ["harness-contract", ["check"], "cpu_analysis"],
        ["go-gosec-audit", ["check"], "security_analysis"],
        ["license-report", ["check"], "artifact_generation"],
        ["sbom", ["license-report"], "artifact_generation"],
        ["seaweedfs-compatibility", ["check"], "service_validation"],
        ["seaweedfs-release-gate", ["seaweedfs-compatibility", "license-report", "sbom"], "small_check"],
        ["build", ["check"], "build"],
        ["deployable-shape", ["build"], "small_check"],
        ["release-browser-readiness", ["check"], "nested_browser_validation"],
        ["release-readiness-evidence", ["harness-contract", "go-gosec-audit", "license-report", "sbom", "seaweedfs-compatibility", "seaweedfs-release-gate", "build", "deployable-shape", "release-browser-readiness"], "small_check"],
      ],
    };
    const expectedTargetSteps = expectedSteps[target];
    const sequence = candidateRow.execution_policy?.sequence;
    if (!expectedTargetSteps || !sequence) {
      throw new Error(`${target} candidate policy does not implement the exact sequence transition`);
    }
    const projection = {
      execution_mode: sequence.execution_mode,
      max_jobs: sequence.max_jobs,
      capacity_profile: sequence.capacity_profile,
      resource_limits: sequence.resource_limits,
      steps: (sequence.steps ?? []).map((step) => ({
        target: step.target,
        needs: step.needs,
        resource_profile: step.resource_profile,
        resource_claims: step.resource_claims,
        make_jobs: step.make_jobs,
        ...(step.forwarding === undefined ? {} : { forwarding: step.forwarding }),
        priority: step.priority,
      })),
    };
    const expected = {
      execution_mode: "dag",
      max_jobs: 8,
      capacity_profile: "sequence_adaptive",
      resource_limits: { host_cpu: "auto", host_io: "auto", process: "auto" },
      steps: expectedTargetSteps.map(([stepTarget, needs, profileName, priority = 0]) => {
        const [resourceClaims, makeJobs, forwarding] = profiles[profileName];
        return {
          target: stepTarget,
          needs,
          resource_profile: profileName,
          resource_claims: resourceClaims,
          make_jobs: makeJobs,
          ...(forwarding === undefined ? {} : { forwarding }),
          priority,
        };
      }),
    };
    if (!sameJSON(projection, expected)) {
      throw new Error(`${target} candidate policy does not implement the exact sequence transition`);
    }
  }
  if (transition === "browser_stack_capacity_1_to_2") {
    const releaseBrowser = candidateRow.execution_policy?.release_browser;
    const expected = {
      browser_stack_capacity: 2,
      parent_sequence_profile: "nested_browser_validation",
      forwarded_limits: { go_cpu: 4, go_io: 4 },
      stage_capacities: { visual: 1, accessibility: 1 },
      retained_session_claims: ["browser_stack", "browser_stage_lane"],
      released_after_startup: ["process"],
      sessions: [
        {
          browser_session_group: "browser-a11y-network-flow-claimed",
          browser_stage: "a11y",
          runtime_profile_id: "network_flow_claimed",
          browser_session_isolation_reason: "claimed Network Flow accessibility evidence requires immutable startup-only extension configuration",
        },
        {
          browser_session_group: "browser-e2e-a11y-default",
          browser_stage: "a11y",
          runtime_profile_id: "default",
          browser_session_isolation_reason: "accessibility evidence owns an isolated fixture session",
        },
        {
          browser_session_group: "browser-e2e-support-default",
          browser_stage: "support",
          runtime_profile_id: "default",
          browser_session_isolation_reason: "release support evidence owns an isolated fixture session",
        },
        {
          browser_session_group: "browser-e2e-visual-default",
          browser_stage: "visual",
          runtime_profile_id: "default",
          browser_session_isolation_reason: "visual evidence owns an isolated fixture and snapshot session",
        },
        {
          browser_session_group: "browser-visual-network-flow-claimed",
          browser_stage: "visual",
          runtime_profile_id: "network_flow_claimed",
          browser_session_isolation_reason: "claimed Network Flow visual evidence requires immutable startup-only extension configuration",
        },
      ],
    };
    if (!sameJSON(releaseBrowser, expected) ||
      candidateRow.execution_policy?.service_backed_schedule?.resource_limits?.browser_stack !== 2) {
      throw new Error(`${target} candidate policy does not implement the exact browser-capacity transition`);
    }
  }
  if (transition === "browser_webserver_stage_capacity_1_to_2") {
    const webserver = candidateRow.execution_policy?.browser_webserver;
    const expected = {
      stage_capacity: 2,
      browser_group_resource_claims: { go_cpu: 2, go_io: 1, process: 1 },
      sessions: [
        {
          browser_session_group: "browser-e2e-webserver-backed-default",
          runtime_profile_id: "default",
        },
        {
          browser_session_group: "browser-functional-network-flow-claimed",
          runtime_profile_id: "network_flow_claimed",
          browser_session_isolation_reason: "claimed Network Flow evidence requires immutable startup-only extension configuration",
        },
      ],
    };
    if (!sameJSON(webserver, expected)) {
      throw new Error(`${target} candidate policy does not implement the exact browser-webserver transition`);
    }
  }
}

export function compareQualifiedBaselines(baseline, candidate) {
  const assertArtifactClosure = (artifact, label) => {
    const names = artifact.targets.map((row) => row.target);
    const sortedNames = [...names].sort((left, right) => left.localeCompare(right));
    if (!sameJSON(names, sortedNames) || new Set(names).size !== names.length) {
      throw new Error(`${label} target rows must be unique and sorted`);
    }
    const portfolio = artifact.public_entrypoint_portfolio;
    if (portfolio.target_count !== names.length || !sameJSON(portfolio.targets, names)) {
      throw new Error(`${label} public roster, rows, and count do not close`);
    }
    const total = artifact.targets.reduce((sum, row) => sum + row.p50_ms, 0);
    if (portfolio.total_p50_ms !== total) throw new Error(`${label} public portfolio sum does not close`);
    const internalNames = artifact.internal_diagnostics.map((row) => row.target);
    if (new Set(internalNames).size !== internalNames.length || internalNames.some((name) => names.includes(name))) {
      throw new Error(`${label} internal diagnostics are not separate from public rows`);
    }
    if (!sameJSON(artifact.source_windows, sourceWindowRoster(artifact.targets))) {
      throw new Error(`${label} source-window roster does not close`);
    }
    for (const row of [...artifact.targets, ...artifact.internal_diagnostics]) {
      if (row.sample_count !== row.samples_ms.length || row.sample_count !== row.sample_roots.length) {
        throw new Error(`${label} ${row.target} sample cardinality does not close`);
      }
      const p50 = median(row.samples_ms);
      const p90 = nearestRankP90(row.samples_ms);
      const mad = median(row.samples_ms.map((sample) => Math.abs(sample - p50)));
      if (row.p50_ms !== p50 || row.p90_ms !== p90 || row.mad_ms !== mad) {
        throw new Error(`${label} ${row.target} statistics do not close`);
      }
    }
  };
  assertArtifactClosure(baseline, "baseline");
  assertArtifactClosure(candidate, "candidate");
  const baselineTargets = new Map(baseline.targets.map((row) => [row.target, row]));
  const candidateTargets = new Map(candidate.targets.map((row) => [row.target, row]));
  if (baselineTargets.size !== baseline.targets.length) throw new Error("baseline duplicates target identities");
  if (candidateTargets.size !== candidate.targets.length) throw new Error("candidate duplicates target identities");
  const expectedTargets = [...baselineTargets.keys()].sort((left, right) => left.localeCompare(right));
  const actualTargets = [...candidateTargets.keys()].sort((left, right) => left.localeCompare(right));
  if (!sameJSON(expectedTargets, actualTargets)) throw new Error("baseline and candidate target inventories differ");
  const rows = [];
  const failures = [];
  for (const target of expectedTargets) {
    const baselineRow = baselineTargets.get(target);
    const candidateRow = candidateTargets.get(target);
    for (const field of [
      "gate", "measurement_profile_id", "timing_source", "host_profile_sha256",
      "capacity_profile_sha256", "toolchain_profile_sha256", "workload_evidence_profile_sha256",
    ]) {
      if (baselineRow[field] !== candidateRow[field]) throw new Error(`${target} baseline and candidate have mismatched ${field}`);
    }
    if (baselineRow.command_id !== candidateRow.command_id) {
      const pattern = /^(cartulary\.harness\.command\.[a-z][a-z0-9_]*\.v)([1-9][0-9]*)$/u;
      const before = pattern.exec(baselineRow.command_id);
      const after = pattern.exec(candidateRow.command_id);
      if (!before || !after || before[1] !== after[1] || Number(after[2]) !== Number(before[2]) + 1) {
        throw new Error(`${target} command_id is not the immediate reviewed successor`);
      }
    }
    if (!sameJSON(baselineRow.canonical_inputs, candidateRow.canonical_inputs)) {
      throw new Error(`${target} baseline and candidate have mismatched canonical inputs`);
    }
    validatePolicyTransition(target, baselineRow, candidateRow);
    if (baselineRow.samples_ms.length !== candidateRow.samples_ms.length) {
      throw new Error(`${target} reference and candidate sample counts differ`);
    }
    const evaluate = (beforeValues, afterValues) => {
      const beforeP50 = median(beforeValues);
      const beforeP90 = nearestRankP90(beforeValues);
      const beforeMAD = median(beforeValues.map((sample) => Math.abs(sample - beforeP50)));
      const afterP50 = median(afterValues);
      const afterP90 = nearestRankP90(afterValues);
      const afterMAD = median(afterValues.map((sample) => Math.abs(sample - afterP50)));
      const band = 3 * Math.max(beforeMAD, afterMAD, 1);
      const p90Passed = afterP90 <= beforeP90 + band;
      const passed = baselineRow.gate === "required_improvement"
        ? beforeP50 - afterP50 > band && p90Passed
        : p90Passed;
      return { passed, band, beforeP50, beforeP90, afterP50, afterP90 };
    };
    const result = evaluate(baselineRow.samples_ms, candidateRow.samples_ms);
    const leaveOneOutChanges = baselineRow.samples_ms.some((_, index) => {
      const before = baselineRow.samples_ms.filter((__, sampleIndex) => sampleIndex !== index);
      const after = candidateRow.samples_ms.filter((__, sampleIndex) => sampleIndex !== index);
      return evaluate(before, after).passed !== result.passed;
    });
    const highMAD = baselineRow.mad_ms > baselineRow.p50_ms * 0.1 ||
      candidateRow.mad_ms > candidateRow.p50_ms * 0.1;
    if (baselineRow.sample_count === 5 && (highMAD || leaveOneOutChanges)) {
      throw new Error(`${target} requires one matched sixth observation`);
    }
    if (baselineRow.sample_count === 6 && (highMAD || leaveOneOutChanges)) {
      throw new Error(`${target} remains unstable after six measured observations`);
    }
    const passed = result.passed;
    if (!passed) failures.push(target);
    rows.push({
      target,
      gate: baselineRow.gate,
      baseline_p50_ms: baselineRow.p50_ms,
      baseline_p90_ms: baselineRow.p90_ms,
      baseline_mad_ms: baselineRow.mad_ms,
      candidate_p50_ms: candidateRow.p50_ms,
      candidate_p90_ms: candidateRow.p90_ms,
      candidate_mad_ms: candidateRow.mad_ms,
      variability_band_ms: result.band,
      status: passed ? "pass" : "fail",
    });
  }
  const portfolio = {
    target_count: baseline.public_entrypoint_portfolio.target_count,
    baseline_total_p50_ms: baseline.public_entrypoint_portfolio.total_p50_ms,
    candidate_total_p50_ms: candidate.public_entrypoint_portfolio.total_p50_ms,
    delta_p50_ms: candidate.public_entrypoint_portfolio.total_p50_ms - baseline.public_entrypoint_portfolio.total_p50_ms,
    status: candidate.public_entrypoint_portfolio.total_p50_ms < baseline.public_entrypoint_portfolio.total_p50_ms
      ? "improved"
      : "increased_or_equal",
  };
  return {
    baseline,
    candidate,
    rows,
    portfolio,
    failures,
    rejected_roots: { baseline: baseline.rejected_roots, candidate: candidate.rejected_roots },
  };
}
