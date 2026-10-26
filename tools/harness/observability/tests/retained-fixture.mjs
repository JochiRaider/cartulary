import { readFileSync } from "node:fs";
import { createHash } from "node:crypto";
import { canonicalJSONString } from "../../contract/index.mjs";
import path from "node:path";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";
import { executionRecordsDigest } from "../../runtime/execution-observations.mjs";
const digest = `sha256:${"a".repeat(64)}`;
export const buildEvidenceUnit = (id, rows = []) => ({ unit_id: id, owner_id: "harness.evidence_accounting", kind: "runner", command: { executable: "true", args: [], environment: {} }, needs: [], resource_claims: { cpu: 1, process: 1 }, fixture_lease: "none", postgres_isolation: "none", renderer_requirement: "none", physical_allocations: {}, service_dependencies: [], cache_policy: "none", timeout_ms: 1000, shared_locks: [], exclusive_locks: [], reusable_artifact_outputs: [], row_bindings: rows.map(({ row_id, runner }) => ({ row_id, runner })), current_run_evidence_outputs: [...rows.map((row) => `rows/${row.row_id}.json`), `unit-results/${id}.json`], failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" }, estimated_work_ms: 10 });
// Minimal reader fixture: one independent unit with a literal 0..10 ms timeline.
// Caller-owned units/projections stay explicit; this is not an execution simulator.
export function buildReaderEvidence({ runID, target = "test-slice", commandID,
  units = [buildEvidenceUnit("a")], projections, outcomes = {} } = {}) {
  if (units.length !== 1 || units[0].needs.length) throw new Error("reader fixture requires one independent unit and explicit evidence for larger graphs");
  units = structuredClone(units);
  const selected = units[0];
  const unitID = selected.unit_id;
  const outcome = outcomes[unitID] ?? "passed";
  const projectionInputs = structuredClone(projections ?? { [target]: { command_id: commandID, unit_ids: [unitID] } });
  const wait = { wait_reason: "rank", blocking_resources: [], blocking_unit_ids: [] };
  const events = [
    ["run_started", 0], ["queued", 0, { unit_kind: selected.kind, row_ids: selected.row_bindings.map((row) => row.row_id) }],
    ["eligible", 0], ["wait_started", 0, wait], ["wait_ended", 2, wait],
    ["admitted", 2], ["started", 2], ["completed", 9],
    ["cleanup_started", 9], ["cleanup_completed", 10], ["run_completed", 10],
  ].map(([event, monotonic_ms, extra = {}], index) => ({ schema_id: "cartulary.harness_unit_event.v4",
    seq: index + 1, monotonic_ms, event, unit_id: /^(run_|cleanup_)/u.test(event) ? "harness:run" : unitID,
    status: event === "cleanup_completed" ? "passed" : ["completed", "run_completed"].includes(event) ? outcome : "running",
    needs: [], resource_claims: /^(run_|cleanup_)/u.test(event) ? {} : { cpu: 1, process: 1 }, service_dependencies: [], ...extra }));
  const capability = { schema_id: "cartulary.harness_capability_snapshot.v1", cpu_tokens: 2, memory_bytes: 1024,
    process_slots: 2, io_tokens: 1, postgres_lanes: 1, object_store_lanes: 1,
    writable_volume: true, port_lanes: 1, services: {}, sources: {}, snapshot_digest: digest };
  const manifest = { schema_id: "cartulary.harness_run_manifest.v3", run_id: runID, target,
    command_id: commandID, declared_inputs: {}, source_commit: "a".repeat(40), source_state: "dirty",
    source_digest: digest, toolchain_digest: digest, system_digest: digest, graph_digest: digest, capability_snapshot: capability, go_worker_budget: 2,
    cache_mode: "off", instrumentation: { mode: "basic", policy_digest: digest }, started_at: "2026-10-09T00:00:00.000Z" };
  const timing = { setup_ms: 2, fixture_ms: 0, execution_ms: 7, collation_ms: 0, wrapper_ms: 1, unattributed_ms: 0, resource_blocking_ms: 0, process_count: 1 };
  const summary = { schema_id: "cartulary.harness_run_summary.v1", run_id: runID, target, status: outcome === "passed" ? "pass" : outcome === "cancelled" ? "cancelled" : "fail",
    failure_class: outcome === "passed" ? null : "product", failure_reason: outcome === "passed" ? null : "test_assertion_failure",
    unit_counts: { total: 1, passed: outcome === "passed" ? 1 : 0, failed: outcome === "failed" ? 1 : 0, skipped: 0, cancelled: outcome === "cancelled" ? 1 : 0 },
    wall_duration_ms: 10, critical_path: [unitID], actual_dependency_critical_path_ms: 9, timing_accounting: timing,
    resource_pressure: {
      requested_capacity: { browser_stack: 0, cpu: 1, io: 0, memory_mb: 0, object_store: 0, port_lane: 0, postgres: 0, process: 1, service_stack: 0, volume: 0 },
      resolved_capacity: { browser_stack: 1, cpu: 2, io: 1, memory_mb: 1, object_store: 1, port_lane: 1, postgres: 1, process: 2, service_stack: 1, volume: 1 },
      peak_use: { cpu: 1, process: 1 }, saturation_ms: {}, wait_events: 1, blocked_units: [], resource_holders: [],
    }, cache: {}, artifact_refs: Object.keys(projectionInputs).map((name) => `target-summaries/${name}.json`) };
  const projection = { schema_id: "cartulary.harness_target_summary.v1", target, command_id: manifest.command_id,
    status: summary.status, failure_class: summary.failure_class, failure_reason: summary.failure_reason, workload_digest: digest, unit_ids: [unitID],
    inclusive_wall_ms: 7, exclusive_wall_ms: 7, actual_dependency_critical_path_ms: 9,
    timing_accounting: { ...timing, setup_ms: 0, wrapper_ms: 0 }, children: [], evidence_refs: [] };
  const files = {
    "run-manifest.json": manifest, "unit-events.ndjson": events, "run-summary.json": summary,
    ...Object.fromEntries(Object.entries(projectionInputs).map(([name, value]) => [`target-summaries/${name}.json`, {
      ...structuredClone(projection), ...value, target: name,
    }])),
  };
  const records = [{ kind: "activity", local_start_ms: null, local_end_ms: null, anchors: [], ref: "activity:example", activity: "report_parse", unit_id: unitID, invocation_id: null,
    allocation_ref: null, clock: "node_monotonic", clock_identity: "clock:example", resolution_ms: 0.000001, duration_ms: 5,
    availability: "available", outcome: "failed" }];
  files["diagnostics/execution-index.json"] = { schema_id: "cartulary.harness_execution_index.v5", host_admission: { status: "complete", omitted_records: 0, records: [], records_digest: executionRecordsDigest([]) }, allocation_lifecycle: { status: "complete", omitted_records: 0, omitted_refs: [], records: [], records_digest: executionRecordsDigest([]) }, graph_clock: null,
    run_id: manifest.run_id, source_digest: digest, graph_digest: digest, policy_digest: digest, status: "complete", omitted_records: 0,
    omitted_refs: [], records, records_digest: executionRecordsDigest(records) };
  return { manifest, summary, events, files };
}

const hash = (value) => `sha256:${createHash("sha256").update(canonicalJSONString(value)).digest("hex")}`;
const bytesHash = (value) => `sha256:${createHash("sha256").update(value).digest("hex")}`;

// Loading authored identities is separate from construction and persistence.
export function loadFixtureIdentity(target = "test-slice") {
  const surface = JSON.parse(readFileSync(new URL("../../../task_surface_owner.json", import.meta.url), "utf8"));
  const commandID = surface.targets.find((entry) => entry.name === target)?.command_id;
  if (!commandID) throw new Error(`fixture target has no authored command identity: ${target}`);
  const policyBytes = readFileSync(new URL("../../../harness_instrumentation_policy.json", import.meta.url), "utf8");
  return { commandID, policyBytes,
    sourceInputs: [
      { path: "tools/harness_instrumentation_policy.json", kind: "file", mode: "0644", byte_digest: bytesHash(policyBytes) },
      { path: "tools/toolchain_pins.json", kind: "file", mode: "0644", byte_digest: bytesHash("fixture toolchain") },
    ],
    host: { platform: "linux", architecture: "x64", kernel_release: "fixture", cpu_models: ["fixture"] },
  };
}

export function buildRowResult(row, outcome = "passed") {
  return { schema_id: "cartulary.harness_row_result.v6", row_id: row.row_id, runner: row.runner, terminal_state: outcome,
    duration_ms: 1, exit_code: outcome === "passed" ? 0 : 1,
    failure_class: outcome === "passed" ? null : "product", failure_reason: outcome === "passed" ? null : "test_assertion_failure", failure_diagnostic: null,
    started_at: "2026-10-16T00:00:00.000Z", finished_at: "2026-10-16T00:00:00.001Z", wall_duration_ms: 1 };
}

export function buildUnitResult(unit, outcome = "passed") {
  return { schema_id: "cartulary.harness_unit_result.v1", unit_id: unit.unit_id, semantic_digest: unit.semantic_digest,
    status: outcome, exit_code: outcome === "passed" ? 0 : 1, signal: null,
    failure_class: outcome === "passed" ? null : "product", failure_reason: outcome === "passed" ? null : "test_assertion_failure",
    evidence_outputs: [...unit.current_run_evidence_outputs], missing_outputs: [] };
}

export function buildCleanupResults(runID, units) {
  return { schema_id: "cartulary.harness_cleanup_results.v2", run_id: runID,
    results: units.map((unit, i) => ({ sequence: i + 1, operation: "host_release", unit_id: unit.unit_id,
      fixture_lease_id: null, outcome: "completed", failure_class: null, failure_reason: null, failures: [], artifact_refs: [], diagnostic: null })) };
}

// Complete admissible evidence is built once, before serialization or corruption.
// No validator, retained reader or audit is involved in construction.
export function buildCompleteEvidence({ runID, target = "test-slice", identity, rows = [],
  units = [buildEvidenceUnit("a", rows)], projections, outcomes = {},
  declaredInputs = target === "test-slice" ? { OWNER: "harness.evidence_accounting", SCOPE: "all" } : {},
  cacheMode = "off", startedAt = "2026-10-09T00:00:00.000Z" }) {
  const { commandID, policyBytes, sourceInputs, host } = structuredClone(identity);
  if (units.some((unit) => Object.hasOwn(unit, "semantic_digest"))) throw new Error("complete fixture construction expects unsealed unit definitions");
  const selected = structuredClone(units).map((unit) => ({ ...unit, semantic_digest: hash(unit) }));
  const projectionInputs = structuredClone(projections ?? { [target]: { command_id: commandID, unit_ids: selected.map((unit) => unit.unit_id) } });
  const fixture = buildReaderEvidence({ runID, target, commandID, units: selected, projections: projectionInputs, outcomes });
  const { manifest, summary, files } = fixture;
  const graph = { schema_id: "cartulary.harness_work_graph.v8", selector: "fixture", units: selected, graph_digest: hash({ units: selected }) };
  delete manifest.capability_snapshot.snapshot_digest;
  manifest.capability_snapshot.snapshot_digest = hash(manifest.capability_snapshot);
  Object.assign(manifest, { declared_inputs: structuredClone(declaredInputs), cache_mode: cacheMode, started_at: startedAt,
    source_digest: hash({ schema_id: "cartulary.source_snapshot_digest.v2", entries: sourceInputs }),
    toolchain_digest: sourceInputs.find((input) => input.path === "tools/toolchain_pins.json").byte_digest,
    system_digest: manifest.capability_snapshot.snapshot_digest, graph_digest: graph.graph_digest,
    instrumentation: { mode: "off", policy_digest: hash(JSON.parse(policyBytes)) } });
  const context = { schema_id: "cartulary.harness_execution_context.v3", manifest_digest: hash(manifest), graph,
    projections: projectionInputs, source_inputs: sourceInputs, host_profile: host, instrumentation_policy_bytes: policyBytes };
  files["execution-context.json"] = context;
  files["cleanup-results.json"] = buildCleanupResults(runID, selected);
  for (const unit of selected) files[`unit-results/${unit.unit_id}.json`] = buildUnitResult(unit, outcomes[unit.unit_id]);
  for (const row of rows) {
    const unit = selected.find((entry) => entry.row_bindings.some((binding) => binding.row_id === row.row_id));
    files[`rows/${row.row_id}.json`] = buildRowResult(row, outcomes[unit.unit_id]);
  }
  files["retained-secret-scan.json"] = { schema_id: "cartulary.harness_retained_secret_scan.v1", status: "pass", scanned_files: 12, policy_digest: bytesHash("independent scanner fixture") };
  for (const [name, selection] of Object.entries(projectionInputs)) {
    const outputs = selected.filter((unit) => selection.unit_ids.includes(unit.unit_id)).flatMap((unit) => unit.current_run_evidence_outputs);
    Object.assign(files[`target-summaries/${name}.json`], { evidence_refs: outputs,
      workload_digest: hash({ target: name, evidence_outputs: outputs.filter((file) => file.startsWith("rows/")).sort() }) });
  }
  delete files["diagnostics/execution-index.json"];
  summary.artifact_refs = Object.keys(files).sort();
  return { ...fixture, context };
}

// Serialization deliberately preserves caller-supplied bytes and identities.
// It never repairs, hashes, reseals, validates or reloads a specimen.
export function writeEvidenceFile(directory, relative, value) {
  const bytes = typeof value === "string" ? value : relative.endsWith(".ndjson")
    ? value.map((record) => JSON.stringify(record)).join("\n") + "\n" : JSON.stringify(value);
  atomicLocalFile(path.join(directory, relative), bytes, { replace: true });
}

export function persistEvidence(directory, fixture) {
  for (const [relative, value] of Object.entries(fixture.files)) writeEvidenceFile(directory, relative, value);
}
