import path from "node:path";

import { repoRoot } from "../contract/index.mjs";
import { loadRetainedObservability, resolveExactRunDir } from "./observability.mjs";

export const backendFinalizerTarget = "backend-output-finalizer";

export function median(values) {
  const sorted = [...values].sort((left, right) => left - right);
  return sorted[Math.floor(sorted.length / 2)];
}

function durationMs(span) {
  return Number((BigInt(span.end_time_unix_nano) - BigInt(span.start_time_unix_nano)) / 1_000_000n);
}

export function intervalUnionMs(spans) {
  const intervals = spans
    .map((span) => [
      Number(BigInt(span.start_time_unix_nano) / 1_000_000n),
      Number(BigInt(span.end_time_unix_nano) / 1_000_000n),
    ])
    .filter(([start, end]) => end > start)
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
    if (result.has(contract.target)) {
      throw new Error(`retained context duplicates measurement contract ${contract.target}`);
    }
    result.set(contract.target, contract);
  }
  return result;
}

export function qualificationReasons(context) {
  const reasons = new Set(context.contamination_reasons);
  if (context.source_state !== "clean") reasons.add("dirty_source");
  if (context.status !== "passed") reasons.add("failed_execution");
  if (context.interrupted) reasons.add("interrupted_execution");
  if (context.retry_count !== 0) reasons.add("retry_observed");
  if (context.warm_eligibility !== "eligible" && reasons.size === 0) {
    reasons.add("external_activity");
  }
  return [...reasons].sort((left, right) => left.localeCompare(right));
}

function addObservation(result, target, value, context, contract, runDir) {
  if (!Number.isFinite(value) || value <= 0) {
    throw new Error(`${context.run_id} has invalid ${target} duration`);
  }
  if (result.has(target)) {
    throw new Error(`${context.run_id} contains duplicate target observation ${target}`);
  }
  result.set(target, [{ value, context, contract, root: runDir }]);
}

export function collectRootObservations(root) {
  const runDir = resolveExactRunDir(root);
  const retained = loadRetainedObservability(runDir);
  return collectRetainedObservations(retained, runDir);
}

export function collectRetainedObservations(retained, runDir) {
  const context = retained.context;
  const reasons = qualificationReasons(context);
  const result = new Map();
  const contracts = contractsByTarget(context);
  if (reasons.length > 0) {
    return { runDir, context, observations: result, reasons };
  }
  for (const invocation of retained.built) {
    const rootSpan = invocation.result.bundle.spans.find(
      (span) => span.span_id === invocation.result.bundle.root_span_id,
    );
    if (!rootSpan || rootSpan.status !== "OK") {
      reasons.push("failed_execution");
      continue;
    }
    const rootContract = contracts.get(context.target);
    if (!rootContract) {
      reasons.push("missing_command_id");
      continue;
    }
    addObservation(
      result,
      context.target,
      durationMs(rootSpan),
      context,
      rootContract,
      runDir,
    );
    for (const span of invocation.result.bundle.spans.filter(
      (item) => item.span_id !== rootSpan.span_id && item.phase === "target" && item.status === "OK",
    )) {
      const contract = contracts.get(span.name);
      if (!contract) continue;
      if (contract.observation_eligibility === "direct_only" && context.target !== span.name) {
        continue;
      }
      addObservation(result, span.name, durationMs(span), context, contract, runDir);
    }
    const backend = context.target === "backend-unit"
      ? rootSpan
      : invocation.result.bundle.spans.find(
        (span) => span.phase === "target" && span.name === "backend-unit",
      );
    if (backend) {
      const descendants = new Set([backend.span_id]);
      let changed = true;
      while (changed) {
        changed = false;
        for (const span of invocation.result.bundle.spans) {
          if (span.parent_span_id && descendants.has(span.parent_span_id) && !descendants.has(span.span_id)) {
            descendants.add(span.span_id);
            changed = true;
          }
        }
      }
      const finalizers = invocation.result.bundle.spans.filter(
        (span) => span.phase === "finalizer" && descendants.has(span.parent_span_id),
      );
      const value = intervalUnionMs(finalizers);
      if (value > 0) {
        const contract = contracts.get("backend-unit");
        addObservation(result, backendFinalizerTarget, value, context, contract, runDir);
      }
    }
  }
  return {
    runDir,
    context,
    observations: reasons.length === 0 ? result : new Map(),
    reasons: [...new Set(reasons)].sort((left, right) => left.localeCompare(right)),
  };
}

function annotateProviderBreadth(record) {
  const breadth = record.observations.size;
  for (const values of record.observations.values()) {
    for (const sample of values) {
      sample.provider_target = record.context.target;
      sample.provider_breadth = breadth;
    }
  }
}

function qualifiedWindow(roots) {
  const records = [];
  const rejectedRoots = [];
  const seenRoots = new Set();
  for (const root of roots) {
    const runDir = resolveExactRunDir(root);
    if (seenRoots.has(runDir)) throw new Error(`performance evidence duplicates retained root ${path.basename(runDir)}`);
    seenRoots.add(runDir);
    const record = collectRootObservations(runDir);
    annotateProviderBreadth(record);
    records.push(record);
    if (record.reasons.length > 0) {
      rejectedRoots.push({ root: rootRef(record.runDir, record.context), reasons: record.reasons });
    }
  }
  const accepted = records.filter((record) => record.reasons.length === 0);
  if (accepted.length === 0) throw new Error("performance window has no accepted retained roots");
  const contractDigest = accepted[0].context.measurement_contracts_sha256;
  if (accepted.some((record) => record.context.measurement_contracts_sha256 !== contractDigest)) {
    throw new Error("performance window has mismatched retained measurement contracts");
  }
  const policy = contractsByTarget(accepted[0].context);
  const observations = new Map();
  for (const record of accepted) {
    for (const [target, values] of record.observations) {
      const retained = observations.get(target) ?? [];
      retained.push(...values);
      observations.set(target, retained);
    }
  }
  return { policy, observations, records, rejectedRoots };
}

function gateFor(contract) {
  return contract.performance_gates.some((gate) => gate.endsWith("_improvement"))
    ? "required_improvement"
    : "no_regression";
}

function selectTargetSamples(target, samples) {
  const byProvider = new Map();
  for (const sample of samples) {
    const retained = byProvider.get(sample.provider_target) ?? [];
    retained.push(sample);
    byProvider.set(sample.provider_target, retained);
  }
  for (const [provider, values] of byProvider) {
    if (values.length !== 3) {
      throw new Error(`${target} provider ${provider} has ${values.length} observations instead of three`);
    }
  }
  const candidates = [...byProvider.entries()]
    .sort(([leftProvider, left], [rightProvider, right]) => {
      if (leftProvider === target && rightProvider !== target) return -1;
      if (rightProvider === target && leftProvider !== target) return 1;
      return left[0].provider_breadth - right[0].provider_breadth || leftProvider.localeCompare(rightProvider);
    });
  if (candidates.length === 0) return [];
  return candidates[0][1];
}

function assertEqualSamples(target, samples, field, label = field) {
  if (samples.some((sample) => sample[field] !== samples[0][field])) {
    throw new Error(`${target} observations have mismatched ${label}`);
  }
}

function targetStatistics(target, contract, allSamples, gate = gateFor(contract)) {
  const samples = selectTargetSamples(target, allSamples);
  if (samples.length !== 3) {
    throw new Error(`${target} requires exactly three accepted observations; found ${samples.length}`);
  }
  const contexts = samples.map((sample) => sample.context);
  const contracts = samples.map((sample) => sample.contract);
  for (const field of [
    "commit",
    "source_snapshot_sha256",
    "host_profile_sha256",
    "capacity_profile_sha256",
    "toolchain_profile_sha256",
    "execution_policy_sha256",
  ]) {
    assertEqualSamples(target, contexts, field);
  }
  for (const field of [
    "command_id",
    "measurement_profile_id",
    "observation_eligibility",
    "workload_evidence_profile_sha256",
    "execution_policy_sha256",
  ]) {
    assertEqualSamples(target, contracts, field);
  }
  const canonical = JSON.stringify(contracts[0].canonical_inputs);
  if (contracts.some((sample) => JSON.stringify(sample.canonical_inputs) !== canonical)) {
    throw new Error(`${target} observations have mismatched canonical inputs`);
  }
  if (target !== backendFinalizerTarget) {
    for (const sample of samples.filter((item) => item.context.target === target)) {
      if (sample.context.command_id !== contract.command_id) {
        throw new Error(`${target} direct observation has mismatched command ID`);
      }
      if (JSON.stringify(sample.context.canonical_inputs) !== canonical) {
        throw new Error(`${target} direct observation has noncanonical inputs`);
      }
      if (sample.context.workload_evidence_profile_sha256 !== contract.workload_evidence_profile_sha256) {
        throw new Error(`${target} direct observation has mismatched workload evidence profile`);
      }
    }
  }
  for (let index = 1; index < samples.length; index += 1) {
    if (Date.parse(samples[index].context.started_at) <= Date.parse(samples[index - 1].context.started_at)) {
      throw new Error(`${target} observations are not in strictly increasing order`);
    }
  }
  const values = samples.map((sample) => sample.value);
  const baselineMedian = median(values);
  const mad = median(values.map((sample) => Math.abs(sample - baselineMedian)));
  return {
    target,
    gate,
    command_id: contract.command_id,
    measurement_profile_id: contract.measurement_profile_id,
    canonical_inputs: contract.canonical_inputs,
    workload_evidence_profile_sha256: contract.workload_evidence_profile_sha256,
    execution_policy_sha256: contract.execution_policy_sha256,
    ...(contract.allowed_policy_transition === undefined
      ? {}
      : { allowed_policy_transition: contract.allowed_policy_transition }),
    sample_provider_target: samples[0].provider_target,
    sample_roots: samples.map((sample) => rootRef(sample.root, sample.context)),
    samples_ms: values,
    median_ms: baselineMedian,
    mad_ms: mad,
    no_regression_limit_ms: baselineMedian + Math.max(1000, 3 * mad, baselineMedian * 0.05),
    required_improvement_ms: Math.max(1000, 3 * mad, baselineMedian * 0.1),
  };
}

export function buildQualifiedBaseline(roots) {
  const window = qualifiedWindow(roots);
  const targets = [];
  for (const [target, contract] of [...window.policy].sort(([left], [right]) => left.localeCompare(right))) {
    targets.push(targetStatistics(target, contract, window.observations.get(target) ?? []));
  }
  const finalizerSamples = window.observations.get(backendFinalizerTarget) ?? [];
  targets.push(targetStatistics(
    backendFinalizerTarget,
    window.policy.get("backend-unit"),
    finalizerSamples,
    "required_improvement",
  ));
  targets.sort((left, right) => left.target.localeCompare(right.target));
  const contexts = window.records.filter((record) => record.reasons.length === 0).map((record) => record.context);
  for (const field of [
    "commit",
    "source_snapshot_sha256",
    "host_profile_sha256",
    "capacity_profile_sha256",
    "toolchain_profile_sha256",
    "measurement_contracts_sha256",
    "workload_contracts_sha256",
  ]) {
    if (contexts.some((context) => context[field] !== contexts[0][field])) {
      throw new Error(`baseline observations have mismatched ${field}`);
    }
  }
  return {
    schema_id: "cartulary.harness_public_target_duration_baselines.v1",
    status: "qualified",
    source_commit: contexts[0].commit,
    source_snapshot_sha256: contexts[0].source_snapshot_sha256,
    profile_digests: {
      host: contexts[0].host_profile_sha256,
      capacity: contexts[0].capacity_profile_sha256,
      workload: contexts[0].workload_contracts_sha256,
      toolchain: contexts[0].toolchain_profile_sha256,
    },
    targets,
    rejected_roots: window.rejectedRoots,
  };
}

export function comparePerformanceWindows(baselineRoots, candidateRoots) {
  return compareQualifiedBaselines(
    buildQualifiedBaseline(baselineRoots),
    buildQualifiedBaseline(candidateRoots),
  );
}

function sameJSON(left, right) {
  return JSON.stringify(left) === JSON.stringify(right);
}

export function compareQualifiedBaselines(baseline, candidate) {
  for (const field of ["host", "capacity", "workload", "toolchain"]) {
    if (baseline.profile_digests[field] !== candidate.profile_digests[field]) {
      throw new Error(`baseline and candidate have mismatched ${field} profile`);
    }
  }
  const baselineTargets = new Map(baseline.targets.map((row) => [row.target, row]));
  const candidateTargets = new Map(candidate.targets.map((row) => [row.target, row]));
  if (baselineTargets.size !== baseline.targets.length) {
    throw new Error("baseline duplicates target identities");
  }
  if (candidateTargets.size !== candidate.targets.length) {
    throw new Error("candidate duplicates target identities");
  }
  const expectedTargets = [...baselineTargets.keys()].sort((left, right) => left.localeCompare(right));
  const actualTargets = [...candidateTargets.keys()].sort((left, right) => left.localeCompare(right));
  if (!sameJSON(expectedTargets, actualTargets)) {
    throw new Error("baseline and candidate target inventories differ");
  }
  const rows = [];
  const failures = [];
  for (const target of expectedTargets) {
    const baselineRow = baselineTargets.get(target);
    const candidateRow = candidateTargets.get(target);
    for (const field of [
      "gate",
      "command_id",
      "measurement_profile_id",
      "workload_evidence_profile_sha256",
    ]) {
      if (baselineRow[field] !== candidateRow[field]) {
        throw new Error(`${target} baseline and candidate have mismatched ${field}`);
      }
    }
    if (!sameJSON(baselineRow.canonical_inputs, candidateRow.canonical_inputs)) {
      throw new Error(`${target} baseline and candidate have mismatched canonical inputs`);
    }
    const transition = baselineRow.allowed_policy_transition;
    if (transition !== candidateRow.allowed_policy_transition) {
      throw new Error(`${target} baseline and candidate have mismatched policy transition contracts`);
    }
    if (transition === undefined) {
      if (baselineRow.execution_policy_sha256 !== candidateRow.execution_policy_sha256) {
        throw new Error(`${target} has an undeclared execution-policy change`);
      }
    } else if (baselineRow.execution_policy_sha256 === candidateRow.execution_policy_sha256) {
      throw new Error(`${target} did not apply declared policy transition ${transition}`);
    }
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
  return {
    baseline,
    candidate,
    rows,
    failures,
    rejected_roots: {
      baseline: baseline.rejected_roots,
      candidate: candidate.rejected_roots,
    },
  };
}
