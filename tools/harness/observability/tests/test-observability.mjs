#!/usr/bin/env node

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import { createServer } from "node:http";
import {
  chmodSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";

import {
  captureExecutionContext,
  deterministicBytes,
  loadRetainedObservability,
  reconstructObservability,
} from "../observability.mjs";
import { validateSchemaSync } from "../../contract/index.mjs";
import {
  collectRetainedObservations,
  compareQualifiedBaselines,
  intervalUnionMs,
  median,
  qualificationReasons,
} from "../performance-evidence.mjs";
import {
  deliver,
  exporterTimeoutMs,
  exportRetainedObservability,
  headersFromFile,
  loadExporterInput,
  signalURL,
  validatedEndpoint,
} from "../otel-export-cli.mjs";

const rootDir = path.resolve(import.meta.dirname, "../../../..");
const fixtureRoot = mkdtempSync(path.join(rootDir, ".cartulary", "test-results", "observability-contract-"));
const checkCLI = path.join(rootDir, "tools", "harness", "observability", "observability-check-cli.mjs");
const exporterCLI = path.join(rootDir, "tools", "harness", "observability", "otel-export-cli.mjs");
const performanceCheckCLI = path.join(rootDir, "tools", "harness", "observability", "performance-check-cli.mjs");
const publicBaselinesCLI = path.join(rootDir, "tools", "harness", "observability", "public-target-baselines-cli.mjs");

function writeJSON(file, value) {
  mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`, { mode: 0o600 });
}

function writeJSONL(file, values) {
  mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  writeFileSync(file, `${values.map((value) => JSON.stringify(value)).join("\n")}\n`, { mode: 0o600 });
}

function schedulerState(workUnitID, ordinal, dependencies, claims, eligibility, start, terminal, terminalState, waitReason, blockingResources = []) {
  return {
    work_unit_id: workUnitID,
    manifest_ordinal: ordinal,
    priority: 100 - ordinal,
    dependencies,
    resource_claims: claims,
    eligibility_monotonic_ms: eligibility,
    started_monotonic_ms: start,
    terminal_monotonic_ms: terminal,
    terminal_state: terminalState,
    wait_reason: waitReason,
    blocking_resources: blockingResources,
    blocking_units: [],
  };
}

function schedulerEvent(seq, event, monotonicMs, detail = {}) {
  const states = [
    schedulerState("alpha", 1, [], { process: 1 }, 0, 500, 3000, "passed", "capacity"),
    schedulerState("beta", 2, ["alpha"], { browser_stack: 1 }, 3000, 3500, 7000, "passed", "resources", ["browser_stack"]),
    schedulerState("summary", 3, ["beta"], { process: 1 }, 7000, 8000, 9000, "passed", "capacity"),
  ];
  return {
    schema_id: "cartulary.scheduler_event.v7",
    target: "harness-contract",
    scheduler_kind: "sequence",
    seq,
    event,
    monotonic_ms: monotonicMs,
    emitted_at: new Date(Date.parse("2026-07-20T12:00:00.000Z") + monotonicMs).toISOString(),
    pending: 0,
    running: 0,
    total_work_units: 3,
    blocked: 0,
    completed: event === "scheduler-finish" ? 3 : 0,
    blocked_reason: null,
    blocked_resources: [],
    waiting_on: [],
    blocked_units: [],
    active_resource_claims: {},
    resource_limits: { process: 1, browser_stack: 1 },
    resource_limit_sources: { process: "schedule", browser_stack: "schedule" },
    dependency_edges: [{ from: "alpha", to: "beta" }, { from: "beta", to: "summary" }],
    work_unit_states: states,
    ...detail,
  };
}

function sourceFixture(runDir) {
  const targetDir = path.join(runDir, "harness-contract");
  writeJSON(path.join(targetDir, "target-summary.json"), {
    target: "harness-contract",
    status: "pass",
    start_time: "2026-07-20T12:00:00.000Z",
    end_time: "2026-07-20T12:00:10.000Z",
    wall_duration_ms: 10_000,
    critical_path_wall_duration_ms: 9_000,
    children: { expected: ["json-shape-check", "generated-artifact-policy-check", "toolchain-drift"] },
  });
  writeJSON(path.join(targetDir, "tool-run-summary.json"), {
    target: "harness-contract",
    status: "pass",
    hostile_fixture_values: [
      "secret-user:secret-pass",
      "/home/private/source/file.go",
      "SELECT secret FROM credentials",
    ],
  });
  writeJSON(path.join(runDir, "alpha", "target-summary.json"), {
    target: "json-shape-check",
    status: "pass",
    start_time: "2026-07-20T12:00:01.000Z",
    end_time: "2026-07-20T12:00:04.000Z",
    children: { expected: ["toolchain-drift"] },
  });
  writeJSON(path.join(runDir, "beta", "target-summary.json"), {
    target: "generated-artifact-policy-check",
    status: "pass",
    start_time: "2026-07-20T12:00:04.000Z",
    end_time: "2026-07-20T12:00:07.000Z",
  });
  writeJSON(path.join(runDir, "gamma", "target-summary.json"), {
    target: "toolchain-drift",
    status: "pass",
    start_time: "2026-07-20T12:00:02.000Z",
    end_time: "2026-07-20T12:00:03.000Z",
  });
  writeJSON(path.join(targetDir, "step-summary.json"), {
    target: "harness-contract",
    label: "fixture-runner",
    runner: "node",
    status: "pass",
    start_time: "2026-07-20T12:00:01.000Z",
    end_time: "2026-07-20T12:00:04.000Z",
  });
  writeJSON(path.join(targetDir, "timing-spans", "service.json"), {
    label: "fixture-service",
    bucket: "service_wait",
    status: "pass",
    start_time: "2026-07-20T12:00:00.250Z",
    end_time: "2026-07-20T12:00:00.750Z",
  });
  writeJSON(path.join(targetDir, "timing-spans", "finalizer.json"), {
    label: "fixture-finalizer",
    bucket: "report_collation",
    status: "pass",
    start_time: "2026-07-20T12:00:08.000Z",
    end_time: "2026-07-20T12:00:09.000Z",
  });
  writeJSONL(path.join(targetDir, "scheduler-events.jsonl"), [
    schedulerEvent(1, "scheduler-start", 0),
    schedulerEvent(2, "start", 500, { work_unit: "alpha", work_unit_id: "alpha", work_unit_class: "test" }),
    schedulerEvent(3, "finish", 3000, { work_unit: "alpha", work_unit_id: "alpha", status: 0 }),
    schedulerEvent(4, "tick", 3000, { blocked: 1, blocked_reason: "resources", blocked_resources: ["browser_stack"] }),
    schedulerEvent(5, "start", 3500, { work_unit: "beta", work_unit_id: "beta", work_unit_class: "test" }),
    schedulerEvent(6, "finish", 7000, { work_unit: "beta", work_unit_id: "beta", status: 0 }),
    schedulerEvent(7, "finalize-start", 8000, { finalizer: "summary", finalizer_id: "summary" }),
    schedulerEvent(8, "finalize-finish", 9000, { finalizer: "summary", finalizer_id: "summary", status: 0 }),
    schedulerEvent(9, "scheduler-finish", 9000),
  ]);
  return targetDir;
}

function toolRootWithNestedTargetFixture(runDir) {
  writeJSON(path.join(runDir, "agent-finalize", "tool-run-summary.json"), {
    target: "agent-finalize",
    status: "pass",
    started_at: "2026-07-20T12:00:00.000Z",
    completed_at: "2026-07-20T12:00:10.000Z",
  });
  writeJSON(path.join(runDir, "json-shape-check", "target-summary.json"), {
    target: "json-shape-check",
    status: "pass",
    start_time: "2026-07-20T12:00:01.000Z",
    end_time: "2026-07-20T12:00:09.000Z",
  });
}

function invocationBoundaryToolRootFixture(runDir) {
  writeJSON(path.join(runDir, "_shared", "harness-invocation-start.json"), {
    schema_id: "cartulary.harness_invocation_start.v1",
    run_id: path.basename(runDir),
    target: "agent-finalize",
    started_at: "2026-07-20T11:59:58.000Z",
    invocation_edges: [{ parent_target: "agent-finalize", child_target: "json-shape-check" }],
  });
  writeJSON(path.join(runDir, "agent-finalize", "tool-run-summary.json"), {
    target: "agent-finalize",
    status: "pass",
    started_at: "2026-07-20T12:00:00.000Z",
    completed_at: "2026-07-20T12:00:10.000Z",
  });
  writeJSON(path.join(runDir, "json-shape-check", "target-summary.json"), {
    target: "json-shape-check",
    status: "pass",
    start_time: "2026-07-20T11:59:59.000Z",
    end_time: "2026-07-20T12:00:09.000Z",
  });
}

function performanceArtifact({
  standardMedian = 13_000,
  improvementMedian = 17_000,
  transitionPolicy = "b".repeat(64),
} = {}) {
  const target = ({
    name,
    gate,
    medianMs,
    policy = "a".repeat(64),
    transition,
  }) => ({
    target: name,
    gate,
    command_id: `cartulary.harness.command.${name.replaceAll("-", "_")}.v1`,
    measurement_profile_id: `${name.replaceAll("-", "_")}_profile`,
    canonical_inputs: {},
    workload_evidence_profile_sha256: "c".repeat(64),
    execution_policy_sha256: policy,
    ...(transition ? { allowed_policy_transition: transition } : {}),
    sample_provider_target: name,
    sample_roots: ["root-1", "root-2", "root-3"],
    samples_ms: [medianMs, medianMs, medianMs],
    median_ms: medianMs,
    mad_ms: 0,
    no_regression_limit_ms: name === "standard-target" ? 13_000 : medianMs + 1000,
    required_improvement_ms: name === "improvement-target" ? 3000 : 1000,
  });
  return {
    schema_id: "cartulary.harness_public_target_duration_baselines.v1",
    status: "qualified",
    source_commit: "abcdef1",
    source_snapshot_sha256: "d".repeat(64),
    profile_digests: {
      host: "e".repeat(64),
      capacity: "f".repeat(64),
      workload: "1".repeat(64),
      toolchain: "2".repeat(64),
    },
    targets: [
      target({ name: "standard-target", gate: "no_regression", medianMs: standardMedian }),
      target({ name: "improvement-target", gate: "required_improvement", medianMs: improvementMedian }),
      target({
        name: "transition-target",
        gate: "no_regression",
        medianMs: 10_000,
        policy: transitionPolicy,
        transition: "serial_to_topology_dag",
      }),
    ],
    rejected_roots: [],
  };
}

function digestTree(dir) {
  const result = new Map();
  function visit(current) {
    for (const entry of readdirSync(current, { withFileTypes: true }).sort((left, right) => left.name.localeCompare(right.name))) {
      const file = path.join(current, entry.name);
      if (entry.isDirectory()) visit(file);
      else result.set(path.relative(dir, file), createHash("sha256").update(readFileSync(file)).digest("hex"));
    }
  }
  visit(dir);
  return result;
}

function verifyRun(runDir) {
  const first = reconstructObservability(runDir);
  const beforeReadOnly = digestTree(runDir);
  const retained = loadRetainedObservability(runDir);
  assert.ok(
    retained.context.measurement_contracts.some((contract) => contract.target === "release-browser-readiness"),
    "execution context must retain check-internal performance measurement subjects",
  );
  assert.deepEqual(digestTree(runDir), beforeReadOnly, "retained validation must be read-only");
  assert.equal(deterministicBytes(first), deterministicBytes(retained));
  const invocation = retained.built[0].result;
  assert.equal(invocation.hotspot.actual_dependency_critical_path_ms, 9000);
  assert.equal(invocation.hotspot.queue_wait_ms, 2000);
  assert.equal(invocation.hotspot.resource_blocking_ms, 500);
  assert.deepEqual(invocation.hotspot.resource_blocking_by_resource, { browser_stack: 500 });
  assert.equal(invocation.hotspot.finalization_union_ms, 1000, "overlapping finalizers must be unioned");
  assert.equal(invocation.traceOTLP.resourceSpans[0].resource.attributes[0].value.stringValue, "cartulary.harness");
  assert(invocation.metricsOTLP.resourceMetrics[0].scopeMetrics[0].metrics.every((metric) => metric.name.startsWith("cartulary.harness.")));
  const gamma = invocation.bundle.spans.find((span) => span.name === "toolchain-drift" && span.phase === "target");
  const alpha = invocation.bundle.spans.find((span) => span.name === "json-shape-check" && span.phase === "target");
  assert.equal(gamma.parent_span_id, alpha.span_id, "target parentage must use explicit retained summaries");
  const rendered = deterministicBytes(retained);
  for (const forbidden of ["secret-user:secret-pass", "/home/private/source/file.go", "SELECT secret FROM credentials", os.homedir()]) {
    assert(!rendered.includes(forbidden), `diagnostics exposed hostile value ${forbidden}`);
  }
  return retained;
}

try {
  const runDir = path.join(fixtureRoot, "20260720T120000Z-p1");
  const targetDir = sourceFixture(runDir);
  const retained = verifyRun(runDir);
  const eligibleRetained = {
    ...retained,
    context: {
      ...retained.context,
      contamination_reasons: [],
      source_state: "clean",
      status: "passed",
      interrupted: false,
      retry_count: 0,
      warm_eligibility: "eligible",
      invocation_boundary_retained: true,
    },
  };
  const retainedObservations = collectRetainedObservations(eligibleRetained, runDir);
  assert.equal(retainedObservations.observations.get("harness-contract")?.length, 1);
  assert.equal(retainedObservations.observations.get("harness-contract")?.[0].value, 10_000);

  const toolRootRunDir = path.join(fixtureRoot, "20260720T120011Z-p2");
  toolRootWithNestedTargetFixture(toolRootRunDir);
  captureExecutionContext(toolRootRunDir, { target: "agent-finalize", status: "passed" });
  const toolRootResult = reconstructObservability(toolRootRunDir);
  assert.equal(toolRootResult.context.target, "agent-finalize");
  assert.equal(toolRootResult.built[0].root.target, "agent-finalize");
  assert.ok(
    toolRootResult.built[0].result.bundle.spans.some(
      (span) => span.phase === "target" && span.name === "json-shape-check",
    ),
  );

  const invocationBoundaryRunDir = path.join(fixtureRoot, "20260720T120022Z-p3");
  invocationBoundaryToolRootFixture(invocationBoundaryRunDir);
  captureExecutionContext(invocationBoundaryRunDir, { target: "agent-finalize", status: "passed" });
  const invocationBoundaryResult = reconstructObservability(invocationBoundaryRunDir);
  const invocationBoundaryContext = invocationBoundaryResult.context;
  assert.equal(invocationBoundaryContext.invocation_boundary_retained, true);
  assert.deepEqual(invocationBoundaryContext.invocation_edges, [
    { parent_target: "agent-finalize", child_target: "json-shape-check" },
  ]);
  const invocationRootSpan = invocationBoundaryResult.built[0].result.bundle.spans.find(
    (span) => span.span_id === invocationBoundaryResult.built[0].result.bundle.root_span_id,
  );
  const invocationChildSpan = invocationBoundaryResult.built[0].result.bundle.spans.find(
    (span) => span.phase === "target" && span.name === "json-shape-check",
  );
  assert.equal(
    Number((BigInt(invocationRootSpan.end_time_unix_nano) - BigInt(invocationRootSpan.start_time_unix_nano)) / 1_000_000n),
    12_000,
    "the root span must retain prerequisite time from public preflight",
  );
  assert.equal(invocationChildSpan.parent_span_id, invocationRootSpan.span_id);

  assert.equal(median([9000, 1000, 5000]), 5000);
  assert.equal(intervalUnionMs([
    { start_time_unix_nano: "0", end_time_unix_nano: "5000000000" },
    { start_time_unix_nano: "3000000000", end_time_unix_nano: "7000000000" },
    { start_time_unix_nano: "9000000000", end_time_unix_nano: "10000000000" },
  ]), 8000);
  const baselinePerformance = performanceArtifact({
    standardMedian: 10_000,
    improvementMedian: 20_000,
    transitionPolicy: "a".repeat(64),
  });
  const boundaryPerformance = performanceArtifact();
  assert.deepEqual(compareQualifiedBaselines(
    baselinePerformance,
    boundaryPerformance,
  ).failures, [], "exact performance boundaries must pass");
  const failedPerformance = performanceArtifact({
    standardMedian: 13_001,
    improvementMedian: 17_001,
  });
  assert.deepEqual(compareQualifiedBaselines(
    baselinePerformance,
    failedPerformance,
  ).failures, ["improvement-target", "standard-target"]);
  assert.throws(
    () => compareQualifiedBaselines(
      baselinePerformance,
      performanceArtifact({ transitionPolicy: "a".repeat(64) }),
    ),
    /did not apply declared policy transition/u,
  );
  const mismatchedHost = performanceArtifact();
  for (const field of ["host", "capacity", "workload", "toolchain"]) {
    const mismatchedEnvironment = structuredClone(mismatchedHost);
    mismatchedEnvironment.profile_digests[field] = "3".repeat(64);
    assert.throws(
      () => compareQualifiedBaselines(baselinePerformance, mismatchedEnvironment),
      new RegExp(`mismatched ${field} profile`, "u"),
    );
  }
  for (const [field, value, message] of [
    ["gate", "required_improvement", /mismatched gate/u],
    ["command_id", "cartulary.harness.command.different.v1", /mismatched command_id/u],
    ["measurement_profile_id", "different_profile", /mismatched measurement_profile_id/u],
    ["workload_evidence_profile_sha256", "4".repeat(64), /mismatched workload_evidence_profile_sha256/u],
    ["canonical_inputs", { OWNER: "different" }, /mismatched canonical inputs/u],
  ]) {
    const mismatchedContract = performanceArtifact();
    mismatchedContract.targets[0][field] = value;
    assert.throws(
      () => compareQualifiedBaselines(baselinePerformance, mismatchedContract),
      message,
    );
  }
  const undeclaredPolicyChange = performanceArtifact();
  undeclaredPolicyChange.targets[0].execution_policy_sha256 = "5".repeat(64);
  assert.throws(
    () => compareQualifiedBaselines(baselinePerformance, undeclaredPolicyChange),
    /undeclared execution-policy change/u,
  );
  const duplicateTarget = performanceArtifact();
  duplicateTarget.targets.push(structuredClone(duplicateTarget.targets[0]));
  assert.throws(
    () => compareQualifiedBaselines(baselinePerformance, duplicateTarget),
    /duplicates target identities/u,
  );
  const missingTarget = performanceArtifact();
  missingTarget.targets.pop();
  assert.throws(
    () => compareQualifiedBaselines(baselinePerformance, missingTarget),
    /target inventories differ/u,
  );
  const baselineRootsManifest = {
    schema_id: "cartulary.harness_performance_evidence_roots.v1",
    mode: "baseline",
    baseline_roots: ["root-1", "root-2", "root-3"],
  };
  validateSchemaSync("cartulary.harness_performance_evidence_roots.v1", baselineRootsManifest);
  validateSchemaSync("cartulary.harness_performance_evidence_roots.v1", {
    ...baselineRootsManifest,
    mode: "comparison",
    candidate_roots: ["candidate-1", "candidate-2", "candidate-3"],
  });
  assert.throws(
    () => validateSchemaSync("cartulary.harness_performance_evidence_roots.v1", {
      ...baselineRootsManifest,
      candidate_roots: ["candidate-1"],
    }),
  );
  assert.throws(
    () => validateSchemaSync("cartulary.harness_performance_evidence_roots.v1", {
      ...baselineRootsManifest,
      mode: "comparison",
    }),
  );
  const baselineManifestFile = path.join(fixtureRoot, "baseline-roots.json");
  const comparisonManifestFile = path.join(fixtureRoot, "comparison-roots.json");
  writeJSON(baselineManifestFile, baselineRootsManifest);
  writeJSON(comparisonManifestFile, {
    ...baselineRootsManifest,
    mode: "comparison",
    candidate_roots: ["candidate-1", "candidate-2", "candidate-3"],
  });
  assert.equal(spawnSync(process.execPath, [performanceCheckCLI], { encoding: "utf8" }).status, 2);
  assert.equal(spawnSync(process.execPath, [publicBaselinesCLI], { encoding: "utf8" }).status, 2);
  assert.equal(
    spawnSync(process.execPath, [performanceCheckCLI, "--evidence-roots-file", baselineManifestFile], { encoding: "utf8" }).status,
    2,
  );
  assert.equal(
    spawnSync(process.execPath, [publicBaselinesCLI, "--evidence-roots-file", comparisonManifestFile], { encoding: "utf8" }).status,
    2,
  );
  const eligibleContext = {
    invocation_boundary_retained: true,
    contamination_reasons: [],
    source_state: "clean",
    status: "passed",
    interrupted: false,
    retry_count: 0,
    warm_eligibility: "eligible",
  };
  assert.ok(
    qualificationReasons({ ...eligibleContext, invocation_boundary_retained: false }).includes("artifact_incomplete"),
  );
  for (const [field, value, reason] of [
    ["source_state", "dirty", "dirty_source"],
    ["status", "failed", "failed_execution"],
    ["interrupted", true, "interrupted_execution"],
    ["retry_count", 1, "retry_observed"],
    ["warm_eligibility", "ineligible", "external_activity"],
  ]) {
    assert.ok(qualificationReasons({ ...eligibleContext, [field]: value }).includes(reason));
  }

  const beforeCheck = digestTree(runDir);
  const exactCheck = spawnSync(process.execPath, [checkCLI, "--results-dir", runDir], { encoding: "utf8" });
  assert.equal(exactCheck.status, 0, exactCheck.stderr);
  assert.match(exactCheck.stdout, /read_only=1/u);
  assert.deepEqual(digestTree(runDir), beforeCheck, "explicit check must not write selected evidence");
  const selectedCheck = spawnSync(
    process.execPath,
    [checkCLI, "--results-dir", fixtureRoot, "--run-id", path.basename(runDir)],
    { encoding: "utf8" },
  );
  assert.equal(selectedCheck.status, 0, selectedCheck.stderr);
  const ambiguousCheck = spawnSync(process.execPath, [checkCLI, "--results-dir", fixtureRoot], { encoding: "utf8" });
  assert.equal(ambiguousCheck.status, 2);
  assert.match(ambiguousCheck.stderr, /reason=usage_error/u);

  let ordinaryFetchCalls = 0;
  const originalFetch = globalThis.fetch;
  globalThis.fetch = async () => {
    ordinaryFetchCalls += 1;
    throw new Error("ordinary reconstruction attempted network egress");
  };
  try {
    reconstructObservability(runDir, { write: false });
  } finally {
    globalThis.fetch = originalFetch;
  }
  assert.equal(ordinaryFetchCalls, 0, "ordinary observability reconstruction must never export");

  assert.equal(signalURL(validatedEndpoint("https://collector.example/base/"), "traces").href, "https://collector.example/base/v1/traces");
  assert.equal(signalURL(validatedEndpoint("http://127.0.0.1:4318/v1/metrics"), "traces").href, "http://127.0.0.1:4318/v1/traces");
  assert.equal(exporterTimeoutMs, 10_000);
  for (const invalidEndpoint of [
    "http://collector.example",
    "https://user:pass@collector.example",
    "https://collector.example?secret=1",
    "https://%65xample.com",
    "http://127.999.0.1",
  ]) {
    assert.throws(() => validatedEndpoint(invalidEndpoint), /HARNESS_OTLP_ENDPOINT/u);
  }

  const headersFile = path.join(fixtureRoot, "headers.json");
  writeJSON(headersFile, { Authorization: "secret-export-token", "x-tenant": "fixture" });
  chmodSync(headersFile, 0o600);
  assert.deepEqual(headersFromFile(headersFile), {
    Authorization: "secret-export-token",
    "x-tenant": "fixture",
  });
  chmodSync(headersFile, 0o644);
  assert.throws(() => headersFromFile(headersFile), /0600/u);
  chmodSync(headersFile, 0o600);
  for (const [name, value] of [
    ["headers-array.json", []],
    ["headers-forbidden.json", { Host: "collector.example" }],
    ["headers-newline.json", { "x-value": "unsafe\nvalue" }],
    ["headers-large-value.json", { "x-value": "é".repeat(2049) }],
    ["headers-count.json", Object.fromEntries(Array.from({ length: 33 }, (_, index) => [`x-${index}`, "value"]))],
  ]) {
    const file = path.join(fixtureRoot, name);
    writeJSON(file, value);
    chmodSync(file, 0o600);
    assert.throws(() => headersFromFile(file), /HARNESS_OTLP_HEADERS_FILE/u);
  }
  const headerSymlink = path.join(fixtureRoot, "headers-link.json");
  symlinkSync(headersFile, headerSymlink);
  assert.throws(() => headersFromFile(headerSymlink), /0600/u);

  const exportInput = loadExporterInput({
    resultsDir: runDir,
    runID: "",
    endpoint: "https://collector.example/base",
    headersFile,
  });
  const exportSourceBefore = digestTree(runDir);
  const requests = [];
  const exportResult = await exportRetainedObservability(exportInput, {
    fetchImpl: async (url, options) => {
      requests.push({ url: String(url), options });
      return { ok: true, redirected: false, status: 200 };
    },
  });
  assert.deepEqual(exportResult, { invocations: 1, signals: 2 });
  assert.deepEqual(requests.map((request) => request.url), [
    "https://collector.example/base/v1/traces",
    "https://collector.example/base/v1/metrics",
  ]);
  assert.equal(requests.every((request) => request.options.redirect === "error"), true);
  assert.equal(JSON.parse(requests[0].options.body).resourceSpans.length, 1);
  assert.equal(JSON.parse(requests[1].options.body).resourceMetrics.length, 1);
  assert.deepEqual(digestTree(runDir), exportSourceBefore, "export must not modify selected evidence");

  const receiverRequests = [];
  const receiver = createServer((request, response) => {
    const chunks = [];
    request.on("data", (chunk) => chunks.push(chunk));
    request.on("end", () => {
      receiverRequests.push({
        path: request.url,
        body: JSON.parse(Buffer.concat(chunks).toString("utf8")),
      });
      response.writeHead(204);
      response.end();
    });
  });
  await new Promise((resolve, reject) => {
    receiver.once("error", reject);
    receiver.listen(0, "127.0.0.1", resolve);
  });
  try {
    const address = receiver.address();
    const localInput = {
      ...exportInput,
      endpoint: validatedEndpoint(`http://127.0.0.1:${address.port}/collector`),
    };
    assert.deepEqual(await exportRetainedObservability(localInput), { invocations: 1, signals: 2 });
  } finally {
    await new Promise((resolve, reject) => receiver.close((error) => error ? reject(error) : resolve()));
  }
  assert.deepEqual(receiverRequests.map((request) => request.path), [
    "/collector/v1/traces",
    "/collector/v1/metrics",
  ]);
  assert.equal(receiverRequests[0].body.resourceSpans.length, 1);
  assert.equal(receiverRequests[1].body.resourceMetrics.length, 1);
  await assert.rejects(
    deliver(
      new URL("https://collector.example/v1/traces"),
      { resourceSpans: [] },
      {},
      100,
      async () => ({ ok: false, redirected: true, status: 302 }),
    ),
    /redirect/u,
  );
  await assert.rejects(
    deliver(
      new URL("https://collector.example/v1/traces"),
      { resourceSpans: [] },
      {},
      100,
      async () => ({ ok: false, redirected: false, status: 503 }),
    ),
    /HTTP 503/u,
  );
  await assert.rejects(
    deliver(
      new URL("https://collector.example/v1/traces"),
      { resourceSpans: [] },
      {},
      25,
      async (_url, options) => new Promise((_resolve, reject) => {
        const guard = setTimeout(() => reject(new Error("timeout signal was not delivered")), 250);
        options.signal.addEventListener("abort", () => {
          clearTimeout(guard);
          reject(options.signal.reason);
        }, { once: true });
      }),
    ),
    /timeout|abort/iu,
  );
  const invalidExport = spawnSync(
    process.execPath,
    [exporterCLI, "--results-dir", fixtureRoot, "--endpoint", "https://collector.example"],
    { encoding: "utf8" },
  );
  assert.equal(invalidExport.status, 2);
  assert.match(invalidExport.stderr, /reason=configuration_error/u);
  const failedExport = spawnSync(
    process.execPath,
    [exporterCLI, "--results-dir", runDir, "--endpoint", "http://127.0.0.1:1", "--headers-file", headersFile],
    { encoding: "utf8", timeout: 5_000 },
  );
  assert.equal(failedExport.status, 1);
  assert.match(failedExport.stderr, /reason=tool_diagnostic_failure/u);
  assert(!failedExport.stderr.includes("secret-export-token"), "export failure output leaked a header value");

  const traceRef = retained.index.invocations[0].artifacts.trace_bundle;
  const traceFile = path.join(runDir, traceRef.path);
  const traceBytes = readFileSync(traceFile);
  writeFileSync(traceFile, `${traceBytes.toString("utf8")} `, { mode: 0o600 });
  assert.throws(() => loadRetainedObservability(runDir), /digest mismatch/u);
  const tamperedTree = digestTree(runDir);
  const tamperedCheck = spawnSync(process.execPath, [checkCLI, "--results-dir", runDir], { encoding: "utf8" });
  assert.equal(tamperedCheck.status, 11);
  assert.match(tamperedCheck.stderr, /reason=artifact_error/u);
  assert.deepEqual(digestTree(runDir), tamperedTree, "failed explicit check must not repair tampered evidence");
  writeFileSync(traceFile, traceBytes, { mode: 0o600 });

  const sourceFile = path.join(targetDir, "step-summary.json");
  const sourceBytes = readFileSync(sourceFile);
  writeFileSync(sourceFile, `${sourceBytes.toString("utf8")} `, { mode: 0o600 });
  assert.throws(() => loadRetainedObservability(runDir), /independent reconstruction|source artifact digest mismatch/u);
  writeFileSync(sourceFile, sourceBytes, { mode: 0o600 });

  const rootSummaryFile = path.join(targetDir, "target-summary.json");
  const rootSummary = JSON.parse(readFileSync(rootSummaryFile, "utf8"));
  const betaSummaryFile = path.join(runDir, "beta", "target-summary.json");
  const betaSummary = JSON.parse(readFileSync(betaSummaryFile, "utf8"));
  writeJSON(betaSummaryFile, { ...betaSummary, children: { expected: ["toolchain-drift"] } });
  assert.throws(() => reconstructObservability(runDir, { write: false }), /ambiguous explicit summary parents/u);
  writeJSON(betaSummaryFile, betaSummary);

  writeJSON(rootSummaryFile, { ...rootSummary, start_time: "malformed-clock" });
  assert.throws(() => reconstructObservability(runDir, { write: false }), /invalid clock/u);
  writeJSON(rootSummaryFile, rootSummary);

  const externalRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-observability-owner-root-"));
  try {
    const externalRun = path.join(externalRoot, "external-run");
    sourceFixture(externalRun);
    const external = verifyRun(externalRun);
    assert(!deterministicBytes(external).includes(externalRoot), "run-relative evidence must not leak external absolute roots");
  } finally {
    rmSync(externalRoot, { recursive: true, force: true });
  }

  process.stdout.write("observability reconstruction PASS provenance=1 graph=1 union=1 tamper=1 external_root=1 read_only=1 export=1 egress=0\n");
} finally {
  rmSync(fixtureRoot, { recursive: true, force: true });
}
