#!/usr/bin/env node

import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { writeCanonicalFixture } from "./retained-fixture.mjs";
import { mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import { validateSchemaSync } from "../../contract/index.mjs";
import { semanticJSONSHA256 } from "../../contract/index.mjs";
import {
  assertBaselineClosure,
  buildFromManifest,
  compareQualifiedBaselines,
  median,
  nearestRankP90,
  performanceRoster,
} from "../canonical-performance.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const surface = JSON.parse(readFileSync(path.join(root, "tools/task_surface_owner.json"), "utf8"));
const roster = performanceRoster(surface);
assert.deepEqual([...roster.keys()], surface.observability_policy.required_targets, "measurement must cover the authored public-target roster");
for (const target of ["test-evidence-audit", "benchmark-claim-check", "scheduler-event-order-drift", "scheduler-summary-timing-drift"]) {
  assert.equal(roster.has(target), false, "caller-selected evidence is not a fixed performance workload");
  assert.ok(surface.observability_policy.out_of_scope_targets.some((entry) => entry.target === target && entry.reason));
}
assert.equal(surface.observability_policy.measurement_profiles.some((profile) => Object.hasOwn(profile.canonical_inputs, "EVIDENCE_ROOTS_FILE")), false);
assert.deepEqual(
  [...roster.entries()].filter(([, policy]) => policy.gate === "required_improvement").map(([target]) => target),
  [
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
  ],
);
assert.equal(median([1, 9, 3, 5, 7]), 5);
assert.equal(nearestRankP90([1, 2, 3, 4, 5]), 5);

const evidenceFixture = {
  schema_id: "cartulary.harness_performance_evidence_roots.v3",
  mode: "comparison",
  reference_windows: [{
    window_id: "reference",
    provider_target: "check",
    cold_root: "cold",
    warmup_root: "warmup",
    measured_roots: ["r1", "r2", "r3", "r4", "r5"],
  }],
  reference_bindings: [{
    target: "check",
    window_id: "reference",
    timing_source: "public_invocation_envelope",
  }],
  candidate_windows: [{
    window_id: "candidate",
    provider_target: "check",
    cold_root: "cold-candidate",
    warmup_root: "warmup-candidate",
    measured_roots: ["c1", "c2", "c3", "c4", "c5"],
  }],
  candidate_bindings: [{
    target: "check",
    window_id: "candidate",
    timing_source: "public_invocation_envelope",
  }],
};
validateSchemaSync(evidenceFixture.schema_id, evidenceFixture);

const evidenceDirectory = mkdtempSync(path.join(os.tmpdir(), "cartulary-performance-paths-"));
try {
  const file = path.join(evidenceDirectory, "manifest.json");
  writeFileSync(file, JSON.stringify(evidenceFixture));
  symlinkSync(file, path.join(evidenceDirectory, "alias.json"));
  await assert.rejects(buildFromManifest(root, path.join(evidenceDirectory, "alias.json"), surface), /ELOOP/u);
  const forbiddenDirectory = "docs";
  await assert.rejects(buildFromManifest(root, path.join(evidenceDirectory, forbiddenDirectory, "manifest.json"), surface), /documentation/u);
  const outside = path.join(evidenceDirectory, "outside"); mkdirSync(outside);
  symlinkSync(outside, path.join(evidenceDirectory, "cold"));
  await assert.rejects(buildFromManifest(root, file, surface), /ENOTDIR|ELOOP/u);
  writeFileSync(file, '{"mode":"baseline","mode":"comparison"}');
  await assert.rejects(buildFromManifest(root, file, surface), /duplicate/iu);
} finally { rmSync(evidenceDirectory, { recursive: true, force: true }); }

function qualifiedBaseline({ candidate = false } = {}) {
  const targets = [...roster.entries()].map(([target, policy]) => {
    const required = policy.gate === "required_improvement";
    const value = candidate ? (required ? 80 : 101) : 100;
    const executionPolicy = { target, graph_digest: `sha256:${"a".repeat(64)}` };
    return {
      target,
      gate: policy.gate,
      command_id: policy.command_id,
      measurement_profile_id: policy.measurement_profile_id,
      canonical_inputs: policy.canonical_inputs,
      timing_source: "canonical_unit_interval_union",
      source_commit: (candidate ? "b" : "a").repeat(40),
      source_snapshot_sha256: (candidate ? "b" : "a").repeat(64),
      system_profile_sha256: "c".repeat(64),
      toolchain_profile_sha256: "d".repeat(64),
      workload_evidence_profile_sha256: "e".repeat(64),
      execution_policy: executionPolicy,
      execution_policy_sha256: semanticJSONSHA256(executionPolicy),
      ...(policy.allowed_policy_transition
        ? { allowed_policy_transition: policy.allowed_policy_transition }
        : {}),
      sample_provider_target: target,
      cold_root: `${target}-${candidate ? "candidate" : "reference"}-cold`,
      cold_ms: value + 10,
      warmup_root: `${target}-${candidate ? "candidate" : "reference"}-warmup`,
      sample_roots: [1, 2, 3, 4, 5].map((index) => `${target}-${candidate ? "c" : "r"}${index}`),
      sample_count: 5,
      samples_ms: [value, value, value, value, value],
      p50_ms: value,
      p90_ms: value,
      mad_ms: 0,
      timing_accounting: {
        critical_path_p50_ms: candidate && required ? 80 : 100,
        resource_blocking_p50_ms: 10,
        setup_p50_ms: 1,
        fixture_p50_ms: 2,
        execution_p50_ms: value - 5,
        collation_p50_ms: 1,
        wrapper_p50_ms: 1,
        unattributed_p50_ms: 0,
        process_count_p50: 3,
      },
    };
  });
  const sourceCommit = (candidate ? "b" : "a").repeat(40);
  const sourceSnapshot = (candidate ? "b" : "a").repeat(64);
  return {
    schema_id: "cartulary.harness_public_target_duration_baselines.v3",
    status: "qualified",
    timing_authority: "canonical_unit_events",
    sample_rule: "one_cold_one_discarded_warmup_five_warm_sixth_only_if_unstable",
    targets,
    internal_diagnostics: [],
    source_windows: [{
      source_commit: sourceCommit,
      source_snapshot_sha256: sourceSnapshot,
      targets: [...roster.keys()],
    }],
    public_entrypoint_portfolio: {
      target_count: targets.length,
      targets: [...roster.keys()],
      total_p50_ms: targets.reduce((sum, row) => sum + row.p50_ms, 0),
    },
    rejected_roots: [],
  };
}

const reference = qualifiedBaseline();
const candidate = qualifiedBaseline({ candidate: true });
assertBaselineClosure(reference, surface);
assertBaselineClosure(candidate, surface);
const comparison = compareQualifiedBaselines(reference, candidate, surface);
assert.deepEqual(comparison.failures, []);
assert.equal(comparison.rows.length, roster.size);
const reused = structuredClone(candidate);
reused.targets[0].sample_roots[0] = reference.targets[0].cold_root;
assert.throws(() => compareQualifiedBaselines(reference, reused, surface), /reuse an observation/u);
const forgedPolicy = structuredClone(reference);
forgedPolicy.targets[0].gate = forgedPolicy.targets[0].gate === "no_regression" ? "required_improvement" : "no_regression";
assert.throws(() => assertBaselineClosure(forgedPolicy, surface), /reviewed measurement policy/u);
const duplicate = structuredClone(reference);
duplicate.targets[0].warmup_root = duplicate.targets[0].cold_root;
assert.throws(() => assertBaselineClosure(duplicate, surface), /repeats an observation/u);
const fiveUnstable = structuredClone(reference);
fiveUnstable.targets[0].samples_ms = [1, 20, 40, 60, 80];
assert.throws(() => assertBaselineClosure(fiveUnstable, surface), /requires a sixth observation/u);

const regressed = structuredClone(candidate);
const row = regressed.targets.find((entry) => entry.gate === "no_regression");
row.samples_ms = [200, 200, 200, 200, 200];
row.p50_ms = 200;
row.p90_ms = 200;
regressed.public_entrypoint_portfolio.total_p50_ms = regressed.targets.reduce((sum, entry) => sum + entry.p50_ms, 0);
assert.ok(compareQualifiedBaselines(reference, regressed, surface).failures.includes(row.target));

function setSamples(baseline, target, samples) {
  const row = baseline.targets.find((entry) => entry.target === target);
  row.samples_ms = samples;
  row.sample_count = samples.length;
  row.sample_roots = samples.map((_, index) => `${row.cold_root}-measured-${index}`);
  row.p50_ms = median(samples);
  row.p90_ms = nearestRankP90(samples);
  row.mad_ms = median(samples.map((value) => Math.abs(value - row.p50_ms)));
  baseline.public_entrypoint_portfolio.total_p50_ms = baseline.targets.reduce((sum, entry) => sum + entry.p50_ms, 0);
}

const sensitiveReference = structuredClone(reference);
const sensitiveCandidate = structuredClone(candidate);
for (const baseline of [sensitiveReference, sensitiveCandidate]) {
  setSamples(baseline, "test-slice", [100, 100, 100, 100, 110]);
}
assert.throws(() => compareQualifiedBaselines(sensitiveReference, sensitiveCandidate, surface),
  /requires one matched sixth/u, "removing the two maxima together must not hide sensitivity to one reference maximum");
for (const baseline of [sensitiveReference, sensitiveCandidate]) {
  setSamples(baseline, "test-slice", [100, 100, 100, 100, 110, 110]);
}
assert.deepEqual(compareQualifiedBaselines(sensitiveReference, sensitiveCandidate, surface).failures, []);


// Complete retained fixtures exercise admission independently of current source
// reconstruction. Each mutation preserves unrelated bindings where necessary.
const hash = (value) => `sha256:${semanticJSONSHA256(value)}`;
const byteHash = (value) => `sha256:${createHash("sha256").update(value).digest("hex")}`;
const portfolioDirectory = mkdtempSync(path.join(os.tmpdir(), "cartulary-performance-context-"));
try {
  const write = (file, value) => writeFileSync(file, JSON.stringify(value));
  const policyBytes = readFileSync(path.join(root, "tools/harness_instrumentation_policy.json"), "utf8");
  const inputs = [
    { path: "tools/harness_instrumentation_policy.json", kind: "file", mode: "0644", byte_digest: byteHash(policyBytes) },
    { path: "tools/toolchain_pins.json", kind: "file", mode: "0644", byte_digest: byteHash("toolchain fixture") },
  ];
  const roots = ["cold", "warmup", "m1", "m2", "m3", "m4", "m5", "m6"];
  const writeRoot = (name, index) => {
    const dir = path.join(portfolioDirectory, name);
    const { manifest, summary } = writeCanonicalFixture(dir);
    const unit = { unit_id: "a", owner_id: "harness.evidence_accounting", kind: "runner",
      command: { executable: "true", args: [], environment: {} }, needs: [], resource_claims: { cpu: 1, process: 1 },
      shared_locks: [], exclusive_locks: [], fixture_lease: "none", service_dependencies: [], cache_policy: "none", timeout_ms: 1000,
      current_run_evidence_outputs: [], reusable_artifact_outputs: [], failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" }, estimated_work_ms: 10 };
    unit.semantic_digest = hash(unit);
    const graph = { schema_id: "cartulary.harness_work_graph.v5", selector: "fixture", units: [unit], graph_digest: hash({ units: [unit] }) };
    delete manifest.capability_snapshot.snapshot_digest;
    manifest.capability_snapshot.snapshot_digest = hash(manifest.capability_snapshot);
    Object.assign(manifest, { declared_inputs: { OWNER: "module.auth", PLAYWRIGHT_WORKERS: "3", VITEST_MAX_WORKERS: "4" }, source_state: "clean", cache_mode: index === 0 ? "cold" : "normal",
      started_at: new Date(Date.UTC(2026, 9, 16, 0, 0, index)).toISOString(), source_digest: hash({ schema_id: "cartulary.source_snapshot_digest.v2", entries: inputs }),
      toolchain_digest: inputs[1].byte_digest, system_digest: manifest.capability_snapshot.snapshot_digest, graph_digest: graph.graph_digest,
      instrumentation: { mode: "off", policy_digest: hash(JSON.parse(policyBytes)) } });
    const context = { schema_id: "cartulary.harness_execution_context.v1", manifest_digest: hash(manifest), graph,
      projections: { "test-slice": { command_id: manifest.command_id, unit_ids: ["a"] } }, source_inputs: inputs,
      host_profile: { platform: "linux", architecture: "x64", kernel_release: "fixture", cpu_models: ["fixture CPU"] }, instrumentation_policy_bytes: policyBytes };
    summary.artifact_refs.push("execution-context.json", "cleanup-results.json");
    write(path.join(dir, "run-manifest.json"), manifest);
    write(path.join(dir, "run-summary.json"), summary);
    write(path.join(dir, "execution-context.json"), context);
    write(path.join(dir, "cleanup-results.json"), { schema_id: "cartulary.harness_cleanup_results.v1", run_id: name, results: [] });
    write(path.join(dir, "retained-secret-scan.json"), { schema_id: "cartulary.harness_retained_secret_scan.v1", status: "pass", scanned_files: 12, policy_digest: byteHash("independent scanner fixture") });
    const targetFile = path.join(dir, "target-summaries/test-slice.json"), target = JSON.parse(readFileSync(targetFile));
    target.workload_digest = hash({ target: "test-slice", evidence_outputs: [] });
    write(targetFile, target);
  };
  roots.forEach((name, index) => writeRoot(name, index));
  const focusedSurface = structuredClone(surface);
  focusedSurface.observability_policy.required_targets = ["test-slice"];
  const evidenceFile = path.join(portfolioDirectory, "evidence.json");
  write(evidenceFile, { schema_id: "cartulary.harness_performance_evidence_roots.v3", mode: "baseline",
    reference_windows: [{ window_id: "fixture", provider_target: "test-slice", cold_root: "cold", warmup_root: "warmup", measured_roots: roots.slice(2, 7) }],
    reference_bindings: [{ target: "test-slice", window_id: "fixture", timing_source: "canonical_unit_interval_union" }] });
  const valid = await buildFromManifest(root, evidenceFile, focusedSurface);
  assert.equal(valid.reference.targets[0].p50_ms, 7);
  const evidence = readFileSync(evidenceFile);
  const setRetainedDuration = (name, duration) => {
    const dir = path.join(portfolioDirectory, name);
    const summaryFile = path.join(dir, "run-summary.json");
    const targetFile = path.join(dir, "target-summaries/test-slice.json");
    const eventsFile = path.join(dir, "unit-events.ndjson");
    const summary = JSON.parse(readFileSync(summaryFile));
    const target = JSON.parse(readFileSync(targetFile));
    const events = readFileSync(eventsFile, "utf8").trim().split("\n").map((line) => JSON.parse(line));
    const previous = target.inclusive_wall_ms;
    for (const event of events) {
      if (event.monotonic_ms >= 2 + previous) event.monotonic_ms += duration - previous;
    }
    summary.wall_duration_ms = duration + 3;
    summary.actual_dependency_critical_path_ms = duration + 2;
    summary.timing_accounting.execution_ms = duration;
    target.inclusive_wall_ms = duration;
    target.exclusive_wall_ms = duration;
    target.actual_dependency_critical_path_ms = duration + 2;
    target.timing_accounting.execution_ms = duration;
    write(summaryFile, summary); write(targetFile, target);
    writeFileSync(eventsFile, events.map((event) => JSON.stringify(event)).join("\n") + "\n");
  };
  try {
    const extended = JSON.parse(evidence);
    extended.reference_windows[0].measured_roots.push("m6");
    extended.reference_windows[0].stability_extension_reason = "high_mad";
    write(evidenceFile, extended);
    await assert.rejects(buildFromManifest(root, evidenceFile, focusedSurface), /unearned sixth/u);
    [6, 7, 7, 8, 9, 7].forEach((duration, index) => setRetainedDuration(`m${index + 1}`, duration));
    assert.equal((await buildFromManifest(root, evidenceFile, focusedSurface)).reference.targets[0].sample_count, 6);
    for (const reason of ["leave_one_out_gate_change", "high_mad_and_leave_one_out_gate_change"]) {
      extended.reference_windows[0].stability_extension_reason = reason;
      write(evidenceFile, extended);
      await assert.rejects(buildFromManifest(root, evidenceFile, focusedSurface), /reason does not match its first five/u);
    }
  } finally {
    for (const name of roots.slice(2)) setRetainedDuration(name, 7);
    writeFileSync(evidenceFile, evidence);
  }
  try {
    roots.forEach((name, index) => writeRoot(`candidate-${name}`, index));
    const paired = JSON.parse(evidence);
    paired.mode = "comparison";
    paired.candidate_windows = [{ window_id: "candidate", provider_target: "test-slice",
      cold_root: "candidate-cold", warmup_root: "candidate-warmup",
      measured_roots: roots.slice(2, 7).map((name) => `candidate-${name}`) }];
    paired.candidate_bindings = [{ ...paired.reference_bindings[0], window_id: "candidate" }];
    for (const prefix of ["", "candidate-"]) {
      [100, 100, 100, 100, 110, 110].forEach((duration, index) => setRetainedDuration(`${prefix}m${index + 1}`, duration));
    }
    write(evidenceFile, paired);
    await assert.rejects(buildFromManifest(root, evidenceFile, focusedSurface), /requires one matched sixth/u);
    paired.reference_windows[0].measured_roots.push("m6");
    paired.candidate_windows[0].measured_roots.push("candidate-m6");
    for (const window of [...paired.reference_windows, ...paired.candidate_windows]) {
      window.stability_extension_reason = "leave_one_out_gate_change";
    }
    write(evidenceFile, paired);
    assert.deepEqual((await buildFromManifest(root, evidenceFile, focusedSurface)).comparison.failures, []);
    paired.candidate_windows[0].stability_extension_reason = "high_mad";
    write(evidenceFile, paired);
    await assert.rejects(buildFromManifest(root, evidenceFile, focusedSurface), /reason does not match its first five/u);
  } finally {
    for (const name of roots.slice(2)) setRetainedDuration(name, 7);
    writeFileSync(evidenceFile, evidence);
  }
  try {
    const invalid = JSON.parse(evidence);
    invalid.reference_bindings[0].timing_source = "public_invocation_envelope";
    write(evidenceFile, invalid);
    await assert.rejects(buildFromManifest(root, evidenceFile, focusedSurface), /graph duration is not public invocation duration/u);
  } finally { writeFileSync(evidenceFile, evidence); }
  const originalInputs = roots.map((name) => {
    const mf = path.join(portfolioDirectory, name, "run-manifest.json"), cf = path.join(portfolioDirectory, name, "execution-context.json");
    return { mf, cf, mb: readFileSync(mf), cb: readFileSync(cf) };
  });
  try {
    for (const { mf, cf, mb, cb } of originalInputs) {
      const manifest = JSON.parse(mb), context = JSON.parse(cb);
      manifest.declared_inputs.VITEST_MAX_WORKERS = "5";
      context.manifest_digest = hash(manifest);
      write(mf, manifest); write(cf, context);
    }
    await assert.rejects(buildFromManifest(root, evidenceFile, focusedSurface), /non-canonical inputs VITEST_MAX_WORKERS/u);
  } finally {
    for (const { mf, cf, mb, cb } of originalInputs) { writeFileSync(mf, mb); writeFileSync(cf, cb); }
  }
  const tamper = async (name, file, mutate, expected) => {
    const target = path.join(portfolioDirectory, "cold", file), bytes = readFileSync(target);
    try { const value = JSON.parse(bytes); mutate(value); write(target, value);
      await assert.rejects(buildFromManifest(root, evidenceFile, focusedSurface), expected, name);
    } finally { writeFileSync(target, bytes); }
  };
  await tamper("unit semantics", "execution-context.json", (x) => { x.graph.units[0].command.executable = "false"; }, /graph.*binding/u);
  await tamper("source preimage", "execution-context.json", (x) => { x.source_inputs[0].mode = "0755"; }, /source snapshot/u);
  await tamper("policy bytes", "execution-context.json", (x) => { x.instrumentation_policy_bytes += " "; }, /instrumentation bytes/u);
  await tamper("workload", "target-summaries/test-slice.json", (x) => { x.workload_digest = `sha256:${"f".repeat(64)}`; }, /target workload/u);
  await tamper("timing redistribution", "run-summary.json", (x) => { x.timing_accounting.setup_ms++; x.timing_accounting.execution_ms--; }, /timing accounting/u);
  await tamper("host substitution", "execution-context.json", (x) => { x.host_profile.cpu_models = ["different CPU"]; }, /host profiles/u);
  await tamper("missing context reference", "run-summary.json", (x) => { x.artifact_refs = x.artifact_refs.filter((v) => v !== "execution-context.json"); }, /omits execution-context/u);
  await tamper("unknown context", "execution-context.json", (x) => { x.private_bindings = {}; }, /schema|additional/u);
  await tamper("cleanup failure", "cleanup-results.json", (x) => { x.results = [{ sequence: 1, operation: "runtime_close", unit_id: null, fixture_lease_id: null, outcome: "blocked", failure_class: null, failure_reason: null, failures: [], artifact_refs: [] }]; }, /cleanup is incomplete/u);
  await tamper("failed retained scan", "retained-secret-scan.json", (x) => { x.status = "fail"; }, /schema|constant/u);
  await tamper("mixed scan policies", "retained-secret-scan.json", (x) => { x.policy_digest = `sha256:${"b".repeat(64)}`; }, /mixes retained scan policies/u);
  for (const [name, mutate, expected] of [
    ["dirty source", (x) => { x.source_state = "dirty"; }, /dirty_source/u],
    ["toolchain substitution", (x) => { x.toolchain_digest = `sha256:${"f".repeat(64)}`; }, /toolchain.*binding/u],
    ["capacity substitution", (x) => { x.system_digest = `sha256:${"f".repeat(64)}`; }, /system capacity/u],
    ["reused observation", (x) => { x.run_id = "warmup"; }, /identity do not close/u],
    ["out of order", (x) => { x.started_at = "2027-01-01T00:00:00.000Z"; }, /overlapping or out of order/u],
  ]) {
    const mf = path.join(portfolioDirectory, "cold/run-manifest.json"), cf = path.join(portfolioDirectory, "cold/execution-context.json");
    const mb = readFileSync(mf), cb = readFileSync(cf);
    try { const m = JSON.parse(mb), c = JSON.parse(cb); mutate(m); c.manifest_digest = hash(m); write(mf, m); write(cf, c);
      await assert.rejects(buildFromManifest(root, evidenceFile, focusedSurface), expected, name);
    } finally { writeFileSync(mf, mb); writeFileSync(cf, cb); }
  }
} finally { rmSync(portfolioDirectory, { recursive: true, force: true }); }
const unstable = qualifiedBaseline();
const unstableRow = unstable.targets[0];
unstableRow.samples_ms = [1, 20, 40, 60, 80, 100];
unstableRow.sample_roots.push("sixth"); unstableRow.sample_count = 6;
assert.throws(() => assertBaselineClosure(unstable, surface), /unstable after six/u);
