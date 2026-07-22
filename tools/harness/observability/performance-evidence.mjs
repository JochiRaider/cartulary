import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";

import { repoRoot } from "../contract/index.mjs";
import { canonicalJSONString, semanticJSONSHA256 } from "../test-catalog/semantic-json.mjs";
import { loadRetainedObservability, resolveExactRunDir } from "./observability.mjs";

export const backendFinalizerTarget = "backend-output-finalizer";
export const performanceEvidenceSchemaID = "cartulary.harness_performance_evidence_roots.v2";
export const performanceBaselineSchemaID = "cartulary.harness_public_target_duration_baselines.v2";

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

export function qualificationReasons(context, { allowMissingInvocationBoundary = false } = {}) {
  const reasons = new Set(context.contamination_reasons);
  if (!context.invocation_boundary_retained && !allowMissingInvocationBoundary) reasons.add("artifact_incomplete");
  if (context.source_state !== "clean") reasons.add("dirty_source");
  if (context.status !== "passed") reasons.add("failed_execution");
  if (context.interrupted) reasons.add("interrupted_execution");
  if (context.retry_count !== 0) reasons.add("retry_observed");
  if (context.warm_eligibility !== "eligible" && reasons.size === 0) reasons.add("external_activity");
  return [...reasons].sort((left, right) => left.localeCompare(right));
}

// Narrow fixture-facing projection retained for observability contract tests.
// Qualification and performance acceptance use explicit v2 bindings below.
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

function observationFor(record, target, timingSource, evidenceKind) {
  if (timingSource === "backend_report_collation_union") return nativeFinalizerTiming(record.runDir);
  if (timingSource === "public_invocation_envelope") {
    if (record.context.target !== target) throw new Error(`${target} public invocation timing requires a direct provider`);
    if (record.context.invocation_boundary_retained) {
      const { span } = rootSpan(record.retained);
      return {
        value: durationMs(span),
        start: record.context.started_at,
        end: record.context.ended_at,
      };
    }
    if (evidenceKind !== "retained_v1_reference_migration") {
      throw new Error(`${target} strict evidence has no invocation boundary`);
    }
    return nativeTargetTiming(record.runDir, target);
  }
  if (timingSource === "aggregate_scheduler_work_envelope") {
    if (record.context.target === target) return nativeTargetTiming(record.runDir, target);
    try {
      return exactAggregateSpan(record.retained, target);
    } catch (error) {
      if (evidenceKind !== "retained_v1_reference_migration") throw error;
      return nativeTargetTiming(record.runDir, target);
    }
  }
  throw new Error(`${target} has unsupported timing source ${timingSource}`);
}

function loadWindowRecord(root, window) {
  const runDir = resolveExactRunDir(root);
  const retained = loadRetainedObservability(runDir);
  const context = retained.context;
  if (context.target !== window.provider_target) {
    throw new Error(`${context.run_id} provider ${context.target} does not match ${window.provider_target}`);
  }
  if (window.evidence_kind === "strict_current" && context.schema_id !== "cartulary.harness_execution_context.v2") {
    throw new Error(`${context.run_id} strict evidence must use execution-context v2`);
  }
  const reasons = qualificationReasons(context, {
    allowMissingInvocationBoundary: window.evidence_kind === "retained_v1_reference_migration",
  });
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

function targetStatistics(binding, window, warmup, samples) {
  const target = binding.target;
  const records = [warmup, ...samples];
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
  const values = allObservations.slice(1).map((observation) => observation.value);
  const baselineMedian = median(values);
  const mad = median(values.map((sample) => Math.abs(sample - baselineMedian)));
  const executionPolicy = contract.execution_policy ?? {
    retained_v1_execution_policy_sha256: contract.execution_policy_sha256,
  };
  if (contract.execution_policy && semanticJSONSHA256(contract.execution_policy) !== contract.execution_policy_sha256) {
    throw new Error(`${target} retained execution-policy projection digest mismatch`);
  }
  const gate = target === backendFinalizerTarget || contract.performance_gates.some((gateName) => gateName.endsWith("_improvement"))
    ? "required_improvement"
    : "no_regression";
  const allowedPolicyTransition = contract.allowed_policy_transition ?? currentAllowedPolicyTransition(target);
  return {
    target,
    gate,
    command_id: contract.command_id,
    measurement_profile_id: contract.measurement_profile_id,
    canonical_inputs: contract.canonical_inputs,
    timing_source: binding.timing_source,
    evidence_kind: window.evidence_kind,
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
    warmup_root: rootRef(warmup.runDir, warmup.context),
    sample_roots: samples.map((record) => rootRef(record.runDir, record.context)),
    samples_ms: values,
    median_ms: baselineMedian,
    mad_ms: mad,
    no_regression_limit_ms: baselineMedian + Math.max(1000, 3 * mad, baselineMedian * 0.05),
    required_improvement_ms: Math.max(1000, 3 * mad, baselineMedian * 0.1),
  };
}

function publicTargets() {
  const manifest = readJSON(path.join(repoRoot, "tools", "task_surface_manifest.json"));
  return [...manifest.observability_policy.required_targets].sort((left, right) => left.localeCompare(right));
}

function currentAllowedPolicyTransition(target) {
  const manifest = readJSON(path.join(repoRoot, "tools", "task_surface_manifest.json"));
  const contractTarget = target === backendFinalizerTarget ? "backend-unit" : target;
  const binding = manifest.observability_policy.target_measurement_profiles
    .find((entry) => entry.target === contractTarget);
  const profile = manifest.observability_policy.measurement_profiles
    .find((entry) => entry.profile_id === binding?.profile_id);
  return profile?.allowed_policy_transition;
}

export function buildQualifiedBaseline(windows, bindings, {
  rejectedRoots = [],
  role = "reference",
} = {}) {
  assertUnique(windows, "window_id", `${role} windows`);
  assertUnique(bindings, "target", `${role} bindings`);
  const windowsByID = new Map(windows.map((window) => [window.window_id, window]));
  const loadedWindows = new Map();
  for (const window of windows) {
    if (role === "candidate" && window.evidence_kind !== "strict_current") {
      throw new Error("candidate evidence cannot use retained-v1 migration mode");
    }
    const roots = [window.warmup_root, ...window.measured_roots];
    if (new Set(roots).size !== 3) throw new Error(`${window.window_id} duplicates evidence roots`);
    loadedWindows.set(window.window_id, {
      warmup: loadWindowRecord(window.warmup_root, window),
      samples: window.measured_roots.map((root) => loadWindowRecord(root, window)),
    });
  }
  const targets = bindings.map((binding) => {
    const window = windowsByID.get(binding.window_id);
    if (!window) throw new Error(`${binding.target} binds unknown window ${binding.window_id}`);
    const loaded = loadedWindows.get(binding.window_id);
    return targetStatistics(binding, window, loaded.warmup, loaded.samples);
  }).sort((left, right) => left.target.localeCompare(right.target));
  const required = publicTargets();
  const actualPublic = targets.map((row) => row.target).filter((target) => required.includes(target));
  if (!sameJSON(actualPublic, required)) throw new Error("baseline does not contain the exact 48-target public inventory");
  for (const requiredInternal of ["release-browser-readiness", backendFinalizerTarget]) {
    if (!targets.some((row) => row.target === requiredInternal)) throw new Error(`baseline is missing ${requiredInternal}`);
  }
  if (role === "candidate") {
    const commits = new Set(targets.map((row) => row.source_commit));
    const snapshots = new Set(targets.map((row) => row.source_snapshot_sha256));
    if (commits.size !== 1 || snapshots.size !== 1) throw new Error("candidate windows must share one clean frozen commit and snapshot");
  }
  return {
    schema_id: performanceBaselineSchemaID,
    status: "qualified",
    qualification: windows.every((window) => window.evidence_kind === "strict_current")
      ? "strict_current"
      : "composite_reference_migration",
    targets,
    public_entrypoint_portfolio: {
      target_count: required.length,
      targets: required,
      total_median_ms: targets.filter((row) => required.includes(row.target)).reduce((total, row) => total + row.median_ms, 0),
    },
    rejected_roots: [...rejectedRoots].sort((left, right) => left.root.localeCompare(right.root)),
  };
}

export function comparePerformanceWindows(manifest) {
  return compareQualifiedBaselines(
    buildQualifiedBaseline(manifest.reference_windows, manifest.reference_bindings, {
      rejectedRoots: manifest.reference_rejected_roots ?? [],
      role: "reference",
    }),
    buildQualifiedBaseline(manifest.candidate_windows, manifest.candidate_bindings, {
      rejectedRoots: manifest.candidate_rejected_roots ?? [],
      role: "candidate",
    }),
  );
}

function nominallyEligible(context) {
  return context.source_state === "clean" &&
    context.status === "passed" &&
    !context.interrupted &&
    context.retry_count === 0 &&
    context.warm_eligibility === "eligible" &&
    context.contamination_reasons.length === 0;
}

function retainedContextFile(runDir) {
  return path.join(runDir, "_shared", "harness-observability", "execution-context.json");
}

function migrationCandidate(record, target) {
  const contractTarget = target === backendFinalizerTarget ? "backend-unit" : target;
  const contract = record.contracts.get(contractTarget);
  if (!contract) return null;
  if (target !== backendFinalizerTarget && contract.observation_eligibility === "direct_only" && record.context.target !== target) {
    return null;
  }
  const timingSource = target === backendFinalizerTarget
    ? "backend_report_collation_union"
    : record.context.target === target
      ? "public_invocation_envelope"
      : "aggregate_scheduler_work_envelope";
  try {
    observationFor(record, target, timingSource, "retained_v1_reference_migration");
  } catch {
    return null;
  }
  const bundle = rootSpan(record.retained).bundle;
  const breadth = new Set(bundle.spans.filter((span) => span.phase === "target" && span.status === "OK").map((span) => span.name)).size;
  return {
    record,
    target,
    timingSource,
    breadth,
    compatibility: canonicalJSONString({
      provider: record.context.target,
      commit: record.context.commit,
      source_snapshot_sha256: record.context.source_snapshot_sha256,
      host_profile_sha256: record.context.host_profile_sha256,
      capacity_profile_sha256: record.context.capacity_profile_sha256,
      toolchain_profile_sha256: record.context.toolchain_profile_sha256,
      command_id: contract.command_id,
      measurement_profile_id: contract.measurement_profile_id,
      canonical_inputs: contract.canonical_inputs,
      workload_evidence_profile_sha256: contract.workload_evidence_profile_sha256,
      execution_policy_sha256: contract.execution_policy_sha256,
      timing_source: timingSource,
    }),
  };
}

function selectMigrationTriple(target, records) {
  const groups = new Map();
  for (const record of records) {
    const candidate = migrationCandidate(record, target);
    if (!candidate) continue;
    const values = groups.get(candidate.compatibility) ?? [];
    values.push(candidate);
    groups.set(candidate.compatibility, values);
  }
  const triples = [];
  for (const values of groups.values()) {
    values.sort((left, right) => Date.parse(left.record.context.started_at) - Date.parse(right.record.context.started_at));
    if (values.length < 3) continue;
    triples.push(values.slice(-3));
  }
  triples.sort((left, right) => {
    const leftDirect = left[0].record.context.target === target ? 0 : 1;
    const rightDirect = right[0].record.context.target === target ? 0 : 1;
    return leftDirect - rightDirect ||
      left[0].breadth - right[0].breadth ||
      Date.parse(right[2].record.context.started_at) - Date.parse(left[2].record.context.started_at) ||
      left[0].record.context.target.localeCompare(right[0].record.context.target);
  });
  if (triples.length === 0) throw new Error(`${target} has no compatible retained warm-up plus two-sample window`);
  return triples[0];
}

export function buildRetainedV1ReferenceManifest(resultsRoot, { windowStartedAt, windowEndedAt }) {
  const root = path.resolve(resultsRoot);
  const started = Date.parse(windowStartedAt);
  const ended = Date.parse(windowEndedAt);
  if (!Number.isFinite(started) || !Number.isFinite(ended) || ended <= started) throw new Error("invalid retained audit window");
  const inspected = [];
  for (const entry of readdirSync(root, { withFileTypes: true }).sort((left, right) => left.name.localeCompare(right.name))) {
    if (!entry.isDirectory()) continue;
    const runDir = path.join(root, entry.name);
    const file = retainedContextFile(runDir);
    if (!existsSync(file)) continue;
    const context = readJSON(file);
    const timestamp = Date.parse(context.started_at);
    if (timestamp >= started && timestamp <= ended) inspected.push({ runDir, context });
  }
  const nominal = inspected.filter(({ context }) => nominallyEligible(context));
  const strictRejected = [];
  const migrationRecords = [];
  let strictAccepted = 0;
  for (const item of nominal) {
    try {
      const retained = loadRetainedObservability(item.runDir);
      const strictReasons = qualificationReasons(retained.context);
      if (strictReasons.length === 0) strictAccepted += 1;
      else strictRejected.push({ root: `retained:${retained.context.run_id}`, reasons: strictReasons });
      const migrationReasons = qualificationReasons(retained.context, { allowMissingInvocationBoundary: true });
      if (migrationReasons.length === 0) {
        migrationRecords.push({
          runDir: item.runDir,
          retained,
          context: retained.context,
          contracts: contractsByTarget(retained.context),
        });
      }
    } catch {
      strictRejected.push({ root: `retained:${item.context.run_id}`, reasons: ["artifact_incomplete"] });
    }
  }
  const targets = [...publicTargets(), "release-browser-readiness", backendFinalizerTarget]
    .sort((left, right) => left.localeCompare(right));
  const windowsBySignature = new Map();
  const bindings = [];
  for (const target of targets) {
    const triple = selectMigrationTriple(target, migrationRecords);
    const signature = canonicalJSONString({
      provider: triple[0].record.context.target,
      roots: triple.map((candidate) => candidate.record.runDir),
    });
    let window = windowsBySignature.get(signature);
    if (!window) {
      const windowID = `reference-${String(windowsBySignature.size + 1).padStart(2, "0")}`;
      window = {
        window_id: windowID,
        provider_target: triple[0].record.context.target,
        evidence_kind: "retained_v1_reference_migration",
        warmup_root: triple[0].record.runDir,
        measured_roots: triple.slice(1).map((candidate) => candidate.record.runDir),
      };
      windowsBySignature.set(signature, window);
    }
    bindings.push({ target, window_id: window.window_id, timing_source: triple[0].timingSource });
  }
  return {
    schema_id: performanceEvidenceSchemaID,
    mode: "baseline",
    reference_windows: [...windowsBySignature.values()],
    reference_bindings: bindings,
    reference_rejected_roots: strictRejected.sort((left, right) => left.root.localeCompare(right.root)),
    reference_audit: {
      window_started_at: new Date(started).toISOString(),
      window_ended_at: new Date(ended).toISOString(),
      inspected_contexts: inspected.length,
      nominally_eligible_contexts: nominal.length,
      strict_accepted_contexts: strictAccepted,
      strict_rejected_contexts: strictRejected.length,
    },
  };
}

function validatePolicyTransition(target, baselineRow, candidateRow) {
  const transition = baselineRow.allowed_policy_transition;
  if (transition !== candidateRow.allowed_policy_transition) {
    throw new Error(`${target} baseline and candidate have mismatched policy transition contracts`);
  }
  if (transition === undefined) {
    const matches = baselineRow.evidence_kind === "retained_v1_reference_migration"
      ? baselineRow.execution_policy_sha256 === candidateRow.execution_policy_sha256
      : sameJSON(baselineRow.execution_policy, candidateRow.execution_policy);
    if (!matches) {
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
    "serial_to_topology_dag",
    "browser_stack_capacity_1_to_2",
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
  if (transition === "serial_to_topology_dag") {
    const profiles = {
      small_check: [{ host_cpu: 1, host_io: 1, process: 1 }, 1, undefined],
      cpu_analysis: [{ host_cpu: 4, host_io: 1, process: 1 }, "host_cpu", undefined],
      script: [{ host_cpu: 2, host_io: 2, process: 1 }, "host_cpu", undefined],
      artifact_generation: [{ host_cpu: 2, host_io: 4, process: 1 }, "host_cpu", undefined],
      build: [{ host_cpu: 6, host_io: 3, process: 1 }, "host_cpu", undefined],
      service_validation: [{ host_cpu: 2, host_io: 4, process: 1 }, 1, undefined],
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
    };
    const expectedSteps = {
      lint: [
        ["lint-go", [], "cpu_analysis"],
        ["lint-biome", [], "cpu_analysis"],
        ["frontend-import-boundary-check", [], "small_check"],
        ["backend-module-boundary-check", [], "small_check"],
        ["lint-scripts", [], "script"],
        ["lint-markdown", [], "script"],
        ["lint-shell", [], "script"],
        ["frontend-typecheck", [], "cpu_analysis"],
      ],
      ci: [
        ["check", [], "nested_check", 100],
        ["harness-contract", ["check"], "cpu_analysis"],
        ["go-gosec-audit", ["check"], "cpu_analysis"],
        ["deployable-shape", ["check"], "small_check"],
        ["duration-baseline-drift-suite", ["check"], "artifact_generation"],
      ],
      "release-check": [
        ["check", [], "nested_check", 100],
        ["harness-contract", ["check"], "cpu_analysis"],
        ["go-gosec-audit", ["check"], "cpu_analysis"],
        ["license-report", ["check"], "artifact_generation"],
        ["sbom", ["license-report"], "artifact_generation"],
        ["seaweedfs-compatibility", ["check"], "service_validation"],
        ["seaweedfs-migration-preservation", ["check"], "service_validation"],
        ["seaweedfs-release-gate", ["seaweedfs-compatibility", "seaweedfs-migration-preservation", "license-report", "sbom"], "small_check"],
        ["build", ["check"], "build"],
        ["deployable-shape", ["build"], "small_check"],
        ["release-browser-readiness", ["check"], "nested_service_validation"],
        ["release-readiness-evidence", ["harness-contract", "go-gosec-audit", "license-report", "sbom", "seaweedfs-compatibility", "seaweedfs-migration-preservation", "seaweedfs-release-gate", "build", "deployable-shape", "release-browser-readiness"], "small_check"],
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
      stage_capacities: { visual: 1, accessibility: 1 },
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
}

export function compareQualifiedBaselines(baseline, candidate) {
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
      "gate", "command_id", "measurement_profile_id", "timing_source", "host_profile_sha256",
      "capacity_profile_sha256", "toolchain_profile_sha256", "workload_evidence_profile_sha256",
    ]) {
      if (baselineRow[field] !== candidateRow[field]) throw new Error(`${target} baseline and candidate have mismatched ${field}`);
    }
    if (!sameJSON(baselineRow.canonical_inputs, candidateRow.canonical_inputs)) {
      throw new Error(`${target} baseline and candidate have mismatched canonical inputs`);
    }
    validatePolicyTransition(target, baselineRow, candidateRow);
    const limit = baselineRow.gate === "required_improvement"
      ? baselineRow.median_ms - baselineRow.required_improvement_ms
      : baselineRow.no_regression_limit_ms;
    const passed = candidateRow.median_ms <= limit;
    if (!passed) failures.push(target);
    rows.push({
      target,
      gate: baselineRow.gate,
      baseline_median_ms: baselineRow.median_ms,
      baseline_mad_ms: baselineRow.mad_ms,
      candidate_median_ms: candidateRow.median_ms,
      limit_ms: limit,
      status: passed ? "pass" : "fail",
    });
  }
  const portfolio = {
    target_count: baseline.public_entrypoint_portfolio.target_count,
    baseline_total_ms: baseline.public_entrypoint_portfolio.total_median_ms,
    candidate_total_ms: candidate.public_entrypoint_portfolio.total_median_ms,
    delta_ms: candidate.public_entrypoint_portfolio.total_median_ms - baseline.public_entrypoint_portfolio.total_median_ms,
    status: candidate.public_entrypoint_portfolio.total_median_ms < baseline.public_entrypoint_portfolio.total_median_ms ? "pass" : "fail",
  };
  if (portfolio.status === "fail") failures.push("public_entrypoint_portfolio_total_ms");
  return {
    baseline,
    candidate,
    rows,
    portfolio,
    failures,
    rejected_roots: { baseline: baseline.rejected_roots, candidate: candidate.rejected_roots },
  };
}
