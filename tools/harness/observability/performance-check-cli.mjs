#!/usr/bin/env node

import { readFileSync } from "node:fs";
import path from "node:path";

import { repoRoot, validateSchemaSync } from "../contract/index.mjs";
import { reconstructObservability, resolveRunDir } from "./observability.mjs";

const requiredImprovementTargets = new Set([
  "backend-unit",
  "backend-output-finalizer",
  "release-browser-readiness",
  "release-check",
]);

function usage() {
  process.stderr.write("usage: performance-check-cli.mjs --evidence-roots-file <manifest.json>\n");
  process.exit(2);
}

function median(values) {
  const sorted = [...values].sort((left, right) => left - right);
  return sorted[Math.floor(sorted.length / 2)];
}

const observationCache = new Map();

function observations(root) {
  const runDir = resolveRunDir(root);
  if (observationCache.has(runDir)) return observationCache.get(runDir);
  const result = reconstructObservability(runDir, { write: false });
  const observationsByTarget = new Map();
  let startedAt = Number.POSITIVE_INFINITY;
  for (const { result: invocation } of result.built) {
    const rootSpan = invocation.bundle.spans.find((span) => span.span_id === invocation.bundle.root_span_id);
    if (!rootSpan || rootSpan.status !== "OK") throw new Error(`${path.basename(runDir)} is not a successful uncontaminated observation`);
    startedAt = Math.min(startedAt, Number(BigInt(invocation.bundle.start_time_unix_nano) / 1_000_000n));
    const byID = new Map(invocation.bundle.spans.map((span) => [span.span_id, span]));
    for (const span of invocation.bundle.spans.filter((item) => item.phase === "target" && item.status === "OK")) {
      const values = observationsByTarget.get(span.name) ?? [];
      values.push(Number((BigInt(span.end_time_unix_nano) - BigInt(span.start_time_unix_nano)) / 1_000_000n));
      observationsByTarget.set(span.name, values);
    }
    const backend = invocation.bundle.spans.find((span) => span.phase === "target" && span.name === "backend-unit");
    if (backend) {
      const finalizer = invocation.bundle.spans
        .filter((span) => span.phase === "finalizer" && (span.parent_span_id === backend.span_id || byID.get(span.parent_span_id)?.name === "backend-unit"))
        .reduce((total, span) => total + Number((BigInt(span.end_time_unix_nano) - BigInt(span.start_time_unix_nano)) / 1_000_000n), 0);
      if (finalizer > 0) observationsByTarget.set("backend-output-finalizer", [finalizer]);
    }
    const browser = invocation.bundle.spans
      .filter((span) => ["browser-e2e-support", "browser-e2e-visual", "browser-e2e-a11y"].includes(span.name));
    if (browser.length > 0) {
      const start = browser.reduce((lowest, span) => BigInt(span.start_time_unix_nano) < lowest ? BigInt(span.start_time_unix_nano) : lowest, BigInt(browser[0].start_time_unix_nano));
      const end = browser.reduce((highest, span) => BigInt(span.end_time_unix_nano) > highest ? BigInt(span.end_time_unix_nano) : highest, BigInt(browser[0].end_time_unix_nano));
      observationsByTarget.set("release-browser-readiness", [Number((end - start) / 1_000_000n)]);
    }
  }
  const record = { runDir, startedAt, observationsByTarget, profileDigests: result.index.profile_digests };
  observationCache.set(runDir, record);
  return record;
}

function windowValues(roots, target, label, profile) {
  if (roots.length !== 3) throw new Error(`${target} ${label} window requires exactly three consecutive warm roots`);
  const records = roots.map(observations);
  const profileFields = {
    host: "host_profile_digest",
    capacity: "capacity_profile_digest",
    workload: "workload_digest",
    toolchain: "toolchain_digest",
  };
  for (const record of records) {
    for (const [retainedName, manifestName] of Object.entries(profileFields)) {
      if (record.profileDigests[retainedName] !== profile[manifestName]) {
        throw new Error(`${target} ${label} root ${path.basename(record.runDir)} has a mismatched ${manifestName}`);
      }
    }
  }
  for (let index = 1; index < records.length; index += 1) {
    if (records[index].startedAt <= records[index - 1].startedAt) {
      throw new Error(`${target} ${label} roots are not in strictly increasing observation order`);
    }
  }
  return records.map((record) => {
    const values = record.observationsByTarget.get(target) ?? [];
    if (values.length !== 1 || !Number.isFinite(values[0]) || values[0] <= 0) {
      throw new Error(`${target} ${label} root ${path.basename(record.runDir)} has ${values.length} valid observations`);
    }
    return values[0];
  });
}

try {
  const args = process.argv.slice(2);
  if (args.length !== 2 || args[0] !== "--evidence-roots-file") usage();
  const manifestFile = path.resolve(repoRoot, args[1]);
  const manifest = JSON.parse(readFileSync(manifestFile, "utf8"));
  validateSchemaSync("cartulary.harness_performance_evidence_roots.v1", manifest);
  for (const field of ["host_profile_digest", "capacity_profile_digest", "workload_digest", "toolchain_digest"]) {
    if (manifest.baseline[field] !== manifest.candidate[field]) throw new Error(`baseline and candidate ${field} differ`);
  }
  const targetPolicies = new Map();
  for (const targetPolicy of manifest.targets) {
    if (targetPolicies.has(targetPolicy.target)) throw new Error(`performance manifest duplicates target ${targetPolicy.target}`);
    targetPolicies.set(targetPolicy.target, targetPolicy);
  }
  const taskSurface = JSON.parse(readFileSync(path.join(repoRoot, "tools", "task_surface_manifest.json"), "utf8"));
  const coveredTargets = [...new Set([
    ...taskSurface.observability_policy.required_targets,
    ...requiredImprovementTargets,
  ])].sort((left, right) => left.localeCompare(right));
  const manifestTargets = [...targetPolicies.keys()].sort((left, right) => left.localeCompare(right));
  if (JSON.stringify(coveredTargets) !== JSON.stringify(manifestTargets)) {
    throw new Error("performance manifest target inventory differs from the authored observability policy");
  }
  for (const target of requiredImprovementTargets) {
    if (targetPolicies.get(target)?.gate !== "required_improvement") {
      throw new Error(`performance manifest must cover ${target} with required_improvement`);
    }
  }
  const failures = [];
  const rows = [...targetPolicies.values()].sort((left, right) => left.target.localeCompare(right.target)).map((targetPolicy) => {
    const target = targetPolicy.target;
    const baselineSamples = windowValues(targetPolicy.baseline_roots, target, "baseline", manifest.baseline);
    const candidateSamples = windowValues(targetPolicy.candidate_roots, target, "candidate", manifest.candidate);
    const baselineMedian = median(baselineSamples);
    const candidateMedian = median(candidateSamples);
    const mad = median(baselineSamples.map((sample) => Math.abs(sample - baselineMedian)));
    const noRegressionLimit = baselineMedian + Math.max(1000, 3 * mad, baselineMedian * 0.05);
    const requiredImprovement = Math.max(1000, 3 * mad, baselineMedian * 0.1);
    const gate = targetPolicy.gate;
    if (requiredImprovementTargets.has(target) !== (gate === "required_improvement")) {
      throw new Error(`${target} has an invalid performance gate ${gate}`);
    }
    const pass = gate === "required_improvement"
      ? candidateMedian <= baselineMedian - requiredImprovement
      : candidateMedian <= noRegressionLimit;
    if (!pass) failures.push(target);
    return { target, gate, baseline_median_ms: baselineMedian, baseline_mad_ms: mad, candidate_median_ms: candidateMedian, limit_ms: gate === "required_improvement" ? baselineMedian - requiredImprovement : noRegressionLimit, status: pass ? "pass" : "fail" };
  });
  for (const row of rows) {
    process.stdout.write(`[PERFORMANCE] target=${row.target} gate=${row.gate} status=${row.status} baseline_median_ms=${row.baseline_median_ms} baseline_mad_ms=${row.baseline_mad_ms} candidate_median_ms=${row.candidate_median_ms} limit_ms=${row.limit_ms}\n`);
  }
  if (failures.length > 0) {
    process.stderr.write(`harness-performance-check FAIL failures=${[...new Set(failures)].sort().join(",")}\n`);
    process.exit(13);
  }
  process.stdout.write(`harness-performance-check PASS targets=${rows.length}\n`);
} catch (error) {
  process.stderr.write(`harness-performance-check FAIL diagnostic=${error instanceof Error ? error.message : String(error)}\n`);
  process.exit(13);
}
