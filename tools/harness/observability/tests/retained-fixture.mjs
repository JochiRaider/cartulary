import { readFileSync, rmSync } from "node:fs";
import { createHash } from "node:crypto";
import { canonicalJSONString } from "../../contract/index.mjs";
import path from "node:path";
import { atomicLocalFile } from "../../runtime/secure-local-files.mjs";
import { executionRecordsDigest } from "../../runtime/execution-observations.mjs";
const digest = `sha256:${"a".repeat(64)}`;
const unit = (id) => ({ unit_id: id, owner_id: "harness.evidence_accounting", kind: "runner", command: { executable: "true", args: [], environment: {} }, needs: [], resource_claims: { cpu: 1, process: 1 }, fixture_lease: "none", postgres_isolation: "none", renderer_requirement: "none", physical_allocations: {}, service_dependencies: [], cache_policy: "none", timeout_ms: 1000, row_bindings: [], current_run_evidence_outputs: [], failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" }, estimated_work_ms: 10 });
export function writeCanonicalFixture(dir) {
  const graph = { units: [unit("a")] };
  const wait = { wait_reason: "rank", blocking_resources: [], blocking_unit_ids: [] };
  const events = [
    ["run_started", 0], ["queued", 0, { unit_kind: "runner", row_ids: [] }],
    ["eligible", 0], ["wait_started", 0, wait], ["wait_ended", 2, wait],
    ["admitted", 2], ["started", 2], ["completed", 9],
    ["cleanup_started", 9], ["cleanup_completed", 10], ["run_completed", 10],
  ].map(([event, monotonic_ms, extra = {}], index) => ({ schema_id: "cartulary.harness_unit_event.v4",
    seq: index + 1, monotonic_ms, event, unit_id: /^(run_|cleanup_)/u.test(event) ? "harness:run" : "a",
    status: ["completed", "run_completed", "cleanup_completed"].includes(event) ? "passed" : "running",
    needs: [], resource_claims: /^(run_|cleanup_)/u.test(event) ? {} : { cpu: 1, process: 1 }, service_dependencies: [], ...extra }));
  const capability = { schema_id: "cartulary.harness_capability_snapshot.v1", cpu_tokens: 2, memory_bytes: 1024,
    process_slots: 2, io_tokens: 1, postgres_lanes: 1, object_store_lanes: 1,
    writable_volume: true, port_lanes: 1, services: {}, sources: {}, snapshot_digest: digest };
  const manifest = { schema_id: "cartulary.harness_run_manifest.v3", run_id: path.basename(dir), target: "test-slice",
    command_id: "cartulary.harness.command.test_slice.v2", declared_inputs: {}, source_commit: "a".repeat(40), source_state: "dirty",
    source_digest: digest, toolchain_digest: digest, system_digest: digest, graph_digest: digest, capability_snapshot: capability, go_worker_budget: 2,
    cache_mode: "off", instrumentation: { mode: "basic", policy_digest: digest }, started_at: "2026-10-09T00:00:00.000Z" };
  const timing = { setup_ms: 2, fixture_ms: 0, execution_ms: 7, collation_ms: 0, wrapper_ms: 1, unattributed_ms: 0, resource_blocking_ms: 0, process_count: 1 };
  const summary = { schema_id: "cartulary.harness_run_summary.v1", run_id: path.basename(dir), target: "test-slice", status: "pass",
    failure_class: null, failure_reason: null, unit_counts: { total: 1, passed: 1, failed: 0, skipped: 0, cancelled: 0 },
    wall_duration_ms: 10, critical_path: ["a"], actual_dependency_critical_path_ms: 9, timing_accounting: timing,
    resource_pressure: {
      requested_capacity: { browser_stack: 0, cpu: 1, io: 0, memory_mb: 0, object_store: 0, port_lane: 0, postgres: 0, process: 1, service_stack: 0, volume: 0 },
      resolved_capacity: { browser_stack: 1, cpu: 2, io: 1, memory_mb: 1, object_store: 1, port_lane: 1, postgres: 1, process: 2, service_stack: 1, volume: 1 },
      peak_use: { cpu: 1, process: 1 }, saturation_ms: {}, wait_events: 1, blocked_units: [], resource_holders: [],
    }, cache: {}, artifact_refs: ["target-summaries/test-slice.json"] };
  const target = { schema_id: "cartulary.harness_target_summary.v1", target: "test-slice", command_id: manifest.command_id,
    status: "pass", failure_class: null, failure_reason: null, workload_digest: digest, unit_ids: ["a"],
    inclusive_wall_ms: 7, exclusive_wall_ms: 7, actual_dependency_critical_path_ms: 9,
    timing_accounting: { ...timing, setup_ms: 0, wrapper_ms: 0 }, children: [], evidence_refs: [] };
  atomicLocalFile(path.join(dir, "run-manifest.json"), JSON.stringify(manifest));
  atomicLocalFile(path.join(dir, "unit-events.ndjson"), events.map((event) => JSON.stringify(event)).join("\n") + "\n");
  atomicLocalFile(path.join(dir, "run-summary.json"), JSON.stringify(summary));
  atomicLocalFile(path.join(dir, "target-summaries/test-slice.json"), JSON.stringify(target));
  const records = [{ kind: "activity", local_start_ms: null, local_end_ms: null, anchors: [], ref: "activity:example", activity: "report_parse", unit_id: "a", invocation_id: null,
    allocation_ref: null, clock: "node_monotonic", clock_identity: "clock:example", resolution_ms: 0.000001, duration_ms: 5,
    availability: "available", outcome: "failed" }];
  atomicLocalFile(path.join(dir, "diagnostics/execution-index.json"), JSON.stringify({ schema_id: "cartulary.harness_execution_index.v5", host_admission: { status: "complete", omitted_records: 0, records: [], records_digest: executionRecordsDigest([]) }, allocation_lifecycle: { status: "complete", omitted_records: 0, omitted_refs: [], records: [], records_digest: executionRecordsDigest([]) }, graph_clock: null,
    run_id: manifest.run_id, source_digest: digest, graph_digest: digest, policy_digest: digest, status: "complete", omitted_records: 0,
    omitted_refs: [], records, records_digest: executionRecordsDigest(records) }));
  return { manifest, summary, events };
}

// Complete immutable evidence, independent of the current checkout except for
// the typed instrumentation policy whose exact bytes are retained as a preimage.
export function writeCompleteFixture(dir) {
  const { manifest, summary, events } = writeCanonicalFixture(dir);
  const hash = (value) => `sha256:${createHash("sha256").update(canonicalJSONString(value)).digest("hex")}`;
  const bytesHash = (value) => `sha256:${createHash("sha256").update(value).digest("hex")}`;
  const write = (relative, value) => atomicLocalFile(path.join(dir, relative), JSON.stringify(value), { replace: true });
  const policyBytes = readFileSync(new URL("../../../harness_instrumentation_policy.json", import.meta.url), "utf8");
  const inputs = [
    { path: "tools/harness_instrumentation_policy.json", kind: "file", mode: "0644", byte_digest: bytesHash(policyBytes) },
    { path: "tools/toolchain_pins.json", kind: "file", mode: "0644", byte_digest: bytesHash("fixture toolchain") },
  ];
  const selected = { ...unit("a"), shared_locks: [], exclusive_locks: [], reusable_artifact_outputs: [] };
  selected.semantic_digest = hash(selected);
  const graph = { schema_id: "cartulary.harness_work_graph.v8", selector: "fixture", units: [selected], graph_digest: hash({ units: [selected] }) };
  delete manifest.capability_snapshot.snapshot_digest;
  manifest.capability_snapshot.snapshot_digest = hash(manifest.capability_snapshot);
  Object.assign(manifest, { declared_inputs: { OWNER: "harness.evidence_accounting", SCOPE: "all" },
    source_digest: hash({ schema_id: "cartulary.source_snapshot_digest.v2", entries: inputs }), toolchain_digest: inputs[1].byte_digest,
    system_digest: manifest.capability_snapshot.snapshot_digest, graph_digest: graph.graph_digest,
    instrumentation: { mode: "off", policy_digest: hash(JSON.parse(policyBytes)) } });
  const context = { schema_id: "cartulary.harness_execution_context.v3", manifest_digest: hash(manifest), graph,
    projections: { "test-slice": { command_id: manifest.command_id, unit_ids: ["a"] } }, source_inputs: inputs,
    host_profile: { platform: "linux", architecture: "x64", kernel_release: "fixture", cpu_models: ["fixture"] }, instrumentation_policy_bytes: policyBytes };
  summary.artifact_refs.push("execution-context.json", "cleanup-results.json", "unit-results/a.json", "retained-secret-scan.json");
  write("run-manifest.json", manifest); write("run-summary.json", summary); write("execution-context.json", context);
  write("cleanup-results.json", { schema_id: "cartulary.harness_cleanup_results.v2", run_id: manifest.run_id,
    results: [{ sequence: 1, operation: "host_release", unit_id: "a", fixture_lease_id: null, outcome: "completed", failure_class: null, failure_reason: null, failures: [], artifact_refs: [], diagnostic: null }] });
  write("unit-results/a.json", { schema_id: "cartulary.harness_unit_result.v1", unit_id: "a", semantic_digest: selected.semantic_digest,
    status: "passed", exit_code: 0, signal: null, failure_class: null, failure_reason: null, evidence_outputs: [], missing_outputs: [] });
  write("retained-secret-scan.json", { schema_id: "cartulary.harness_retained_secret_scan.v1", status: "pass", scanned_files: 12, policy_digest: bytesHash("fixture scan") });
  const target = JSON.parse(readFileSync(path.join(dir, "target-summaries/test-slice.json")));
  target.workload_digest = hash({ target: "test-slice", evidence_outputs: [] });
  write("target-summaries/test-slice.json", target);
  rmSync(path.join(dir, "diagnostics/execution-index.json"));
  return { manifest, summary, events, context, write };
}
