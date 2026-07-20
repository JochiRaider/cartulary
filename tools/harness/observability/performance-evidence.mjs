import { readFileSync } from "node:fs";
import path from "node:path";

import { repoRoot } from "../contract/index.mjs";
import { loadRetainedObservability, resolveExactRunDir } from "./observability.mjs";

const syntheticTargets = new Set([
  "backend-output-finalizer",
  "release-browser-readiness",
]);

export function median(values) {
  const sorted = [...values].sort((left, right) => left - right);
  return sorted[Math.floor(sorted.length / 2)];
}

function durationMs(span) {
  return Number((BigInt(span.end_time_unix_nano) - BigInt(span.start_time_unix_nano)) / 1_000_000n);
}

function intervalUnionMs(spans) {
  const intervals = spans
    .map((span) => [
      Number(BigInt(span.start_time_unix_nano) / 1_000_000n),
      Number(BigInt(span.end_time_unix_nano) / 1_000_000n),
    ])
    .filter(([start, end]) => end > start)
    .sort((left, right) => left[0] - right[0] || left[1] - right[1]);
  let total = 0;
  let current;
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

function measurementPolicy() {
  const manifest = JSON.parse(readFileSync(path.join(repoRoot, "tools", "task_surface_manifest.json"), "utf8"));
  const profiles = new Map(manifest.observability_policy.measurement_profiles.map((profile) => [profile.profile_id, profile]));
  return new Map(manifest.observability_policy.target_measurement_profiles.map((binding) => [
    binding.target,
    { ...binding, profile: profiles.get(binding.profile_id) },
  ]));
}

function addObservation(result, target, value, context, root) {
  const values = result.get(target) ?? [];
  values.push({ value, context, root });
  result.set(target, values);
}

export function collectRootObservations(root) {
  const runDir = resolveExactRunDir(root);
  const retained = loadRetainedObservability(runDir);
  const context = retained.context;
  if (context.warm_eligibility !== "eligible" || context.contamination_reasons.length > 0) {
    throw new Error(`${path.basename(runDir)} is not warm-eligible: ${context.contamination_reasons.join(",") || "unknown"}`);
  }
  const result = new Map();
  for (const invocation of retained.built) {
    const rootSpan = invocation.result.bundle.spans.find((span) => span.span_id === invocation.result.bundle.root_span_id);
    if (!rootSpan || rootSpan.status !== "OK") throw new Error(`${path.basename(runDir)} contains a non-successful invocation`);
    for (const span of invocation.result.bundle.spans.filter((item) => item.phase === "target" && item.status === "OK")) {
      addObservation(result, span.name, durationMs(span), context, runDir);
    }
    const backend = invocation.result.bundle.spans.find((span) => span.phase === "target" && span.name === "backend-unit");
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
      const finalizers = invocation.result.bundle.spans.filter((span) => span.phase === "finalizer" && descendants.has(span.parent_span_id));
      const value = intervalUnionMs(finalizers);
      if (value > 0) addObservation(result, "backend-output-finalizer", value, context, runDir);
    }
    const browser = invocation.result.bundle.spans.filter((span) =>
      new Set(["browser-e2e-support", "browser-e2e-visual", "browser-e2e-a11y"]).has(span.name));
    if (browser.length > 0) {
      const start = browser.reduce((minimum, span) => BigInt(span.start_time_unix_nano) < minimum ? BigInt(span.start_time_unix_nano) : minimum, BigInt(browser[0].start_time_unix_nano));
      const end = browser.reduce((maximum, span) => BigInt(span.end_time_unix_nano) > maximum ? BigInt(span.end_time_unix_nano) : maximum, BigInt(browser[0].end_time_unix_nano));
      addObservation(result, "release-browser-readiness", Number((end - start) / 1_000_000n), context, runDir);
    }
  }
  return { runDir, context, observations: result };
}

function qualifyingWindows(roots) {
  const records = roots.map(collectRootObservations);
  const policy = measurementPolicy();
  const observations = new Map();
  const rejectedRoots = [];
  for (const record of records) {
    for (const [target, values] of record.observations) {
      if (!policy.has(target) && !syntheticTargets.has(target)) continue;
      const retained = observations.get(target) ?? [];
      retained.push(...values);
      observations.set(target, retained);
    }
  }
  return { policy, observations, records, rejectedRoots };
}

function targetStatistics(target, binding, samples) {
  if (samples.length !== 3) throw new Error(`${target} requires exactly three accepted observations; found ${samples.length}`);
  const values = samples.map((sample) => sample.value);
  const baselineMedian = median(values);
  const mad = median(values.map((sample) => Math.abs(sample - baselineMedian)));
  return {
    target,
    gate: binding.profile.performance_gate,
    command_id: samples[0].context.command_id,
    measurement_profile_id: binding.profile_id,
    execution_policy_sha256: samples[0].context.execution_policy_sha256,
    sample_roots: samples.map((sample) => path.relative(repoRoot, sample.root).replaceAll("\\", "/")),
    samples_ms: values,
    median_ms: baselineMedian,
    mad_ms: mad,
    no_regression_limit_ms: baselineMedian + Math.max(1000, 3 * mad, baselineMedian * 0.05),
    required_improvement_ms: Math.max(1000, 3 * mad, baselineMedian * 0.1),
  };
}

export function buildQualifiedBaseline(roots) {
  const window = qualifyingWindows(roots);
  const targets = [];
  for (const [target, binding] of [...window.policy].sort(([left], [right]) => left.localeCompare(right))) {
    const samples = window.observations.get(target) ?? [];
    targets.push(targetStatistics(target, binding, samples));
  }
  for (const target of syntheticTargets) {
    const samples = window.observations.get(target) ?? [];
    if (samples.length === 0) continue;
    const binding = target === "release-browser-readiness"
      ? window.policy.get(target)
      : { profile_id: "backend_improvement", profile: { performance_gate: "required_improvement" } };
    targets.push(targetStatistics(target, binding, samples));
  }
  targets.sort((left, right) => left.target.localeCompare(right.target));
  const contexts = targets.flatMap((target) => target.sample_roots.map((root) => window.records.find((record) => path.relative(repoRoot, record.runDir).replaceAll("\\", "/") === root)?.context)).filter(Boolean);
  if (contexts.length === 0) throw new Error("baseline window has no accepted observations");
  for (const field of ["commit", "source_snapshot_sha256", "host_profile_sha256", "capacity_profile_sha256", "toolchain_profile_sha256"]) {
    if (contexts.some((context) => context[field] !== contexts[0][field])) throw new Error(`baseline observations have mismatched ${field}`);
  }
  return {
    schema_id: "cartulary.harness_public_target_duration_baselines.v1",
    status: "qualified",
    source_commit: contexts[0].commit,
    source_snapshot_sha256: contexts[0].source_snapshot_sha256,
    profile_digests: {
      host: contexts[0].host_profile_sha256,
      capacity: contexts[0].capacity_profile_sha256,
      workload: contexts[0].workload_evidence_profile_sha256,
      toolchain: contexts[0].toolchain_profile_sha256,
    },
    targets,
    rejected_roots: window.rejectedRoots,
  };
}

export function comparePerformanceWindows(baselineRoots, candidateRoots) {
  return { baseline: qualifyingWindows(baselineRoots), candidate: qualifyingWindows(candidateRoots) };
}
