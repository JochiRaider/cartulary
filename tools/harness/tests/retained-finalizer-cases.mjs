import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import os from "node:os";
import path from "node:path";
import { semanticJSONDigest } from "../contract/index.mjs";
import { createRetainedRunPreflight } from "../finalization/agent-finalize-retained-run.mjs";
import { writeCanonicalFixture } from "../observability/tests/retained-fixture.mjs";

const repository = path.resolve(import.meta.dirname, "../../..");
const hash = semanticJSONDigest;
const byteHash = (value) => `sha256:${createHash("sha256").update(value).digest("hex")}`;
const write = (file, value) => fs.writeFileSync(file, JSON.stringify(value));

function fixture(parent, name, second = 0) {
  const root = path.join(parent, name);
  const { manifest, summary, events } = writeCanonicalFixture(root);
  const row = { row_id: "harness.evidence_accounting.unit.finalizer_fixture", runner: "go" };
  const unit = { unit_id: "a", owner_id: "harness.evidence_accounting", kind: "runner",
    command: { executable: "true", args: [], environment: {} }, needs: [], resource_claims: { cpu: 1, process: 1 },
    shared_locks: [], exclusive_locks: [], fixture_lease: "none", service_dependencies: [], cache_policy: "none", timeout_ms: 1000,
    row_bindings: [{ row_id: row.row_id, runner: "go" }], current_run_evidence_outputs: [`rows/${row.row_id}.json`, "unit-results/a.json"], reusable_artifact_outputs: [],
    failure_policy: { block_descendants: true, continue_independent: true, aggregate_effect: "required" }, estimated_work_ms: 10 };
  unit.semantic_digest = hash(unit);
  const graph = { schema_id: "cartulary.harness_work_graph.v6", selector: "fixture", units: [unit], graph_digest: hash({ units: [unit] }) };
  const policyBytes = fs.readFileSync(path.join(repository, "tools/harness_instrumentation_policy.json"), "utf8");
  const inputs = [
    { path: "tools/harness_instrumentation_policy.json", kind: "file", mode: "0644", byte_digest: byteHash(policyBytes) },
    { path: "tools/toolchain_pins.json", kind: "file", mode: "0644", byte_digest: byteHash("toolchain fixture") },
  ];
  delete manifest.capability_snapshot.snapshot_digest;
  manifest.capability_snapshot.snapshot_digest = hash(manifest.capability_snapshot);
  Object.assign(manifest, { target: "check", command_id: "cartulary.harness.command.check.v2", source_state: "dirty", cache_mode: "normal",
    started_at: new Date(Date.UTC(2026, 9, 16, 0, 0, second)).toISOString(), source_digest: hash({ schema_id: "cartulary.source_snapshot_digest.v2", entries: inputs }),
    toolchain_digest: inputs[1].byte_digest, system_digest: manifest.capability_snapshot.snapshot_digest, graph_digest: graph.graph_digest,
    instrumentation: { mode: "off", policy_digest: hash(JSON.parse(policyBytes)) } });
  const context = { schema_id: "cartulary.harness_execution_context.v3", manifest_digest: hash(manifest), graph,
    projections: { check: { command_id: manifest.command_id, unit_ids: ["a"] } }, source_inputs: inputs,
    host_profile: { platform: "linux", architecture: "x64", kernel_release: "fixture", cpu_models: ["fixture CPU"] }, instrumentation_policy_bytes: policyBytes };
  const target = JSON.parse(fs.readFileSync(path.join(root, "target-summaries/test-slice.json"), "utf8"));
  Object.assign(target, { target: "check", command_id: manifest.command_id, evidence_refs: unit.current_run_evidence_outputs,
    workload_digest: hash({ target: "check", evidence_outputs: [`rows/${row.row_id}.json`] }) });
  Object.assign(summary, { target: "check", artifact_refs: ["run-manifest.json", "unit-events.ndjson", "target-summaries/check.json",
    "execution-context.json", "cleanup-results.json", ...unit.current_run_evidence_outputs] });
  events.find((event) => event.event === "queued").row_ids = [row.row_id];
  const save = (relative, value) => {
    fs.mkdirSync(path.dirname(path.join(root, relative)), { recursive: true });
    write(path.join(root, relative), value);
  };
  const alter = (relative, mutate) => {
    const value = JSON.parse(fs.readFileSync(path.join(root, relative), "utf8"));
    mutate(value); save(relative, value);
  };
  save("run-manifest.json", manifest); save("run-summary.json", summary); save("execution-context.json", context); save("target-summaries/check.json", target);
  fs.writeFileSync(path.join(root, "unit-events.ndjson"), events.map((event) => JSON.stringify(event)).join("\n") + "\n");
  save("unit-results/a.json", { schema_id: "cartulary.harness_unit_result.v1", unit_id: "a", semantic_digest: unit.semantic_digest,
    status: "passed", exit_code: 0, signal: null, failure_class: null, failure_reason: null, evidence_outputs: unit.current_run_evidence_outputs, missing_outputs: [] });
  save(`rows/${row.row_id}.json`, { schema_id: "cartulary.harness_row_result.v4", row_id: row.row_id, runner: "go", terminal_state: "passed",
    duration_ms: 1, exit_code: 0, failure_class: null, failure_reason: null, failure_diagnostic: null,
    started_at: "2026-10-16T00:00:00.000Z", finished_at: "2026-10-16T00:00:00.001Z", wall_duration_ms: 1 });
  save("cleanup-results.json", { schema_id: "cartulary.harness_cleanup_results.v1", run_id: name, results: [{ sequence: 1, operation: "host_release", unit_id: "a",
    fixture_lease_id: null, outcome: "completed", failure_class: null, failure_reason: null, failures: [], artifact_refs: [] }] });
  // Scan and instrumentation policies are independent identities in production.
  save("retained-secret-scan.json", { schema_id: "cartulary.harness_retained_secret_scan.v1", status: "pass", scanned_files: 10, policy_digest: byteHash("independent scanner fixture") });
  const plan = { graph: structuredClone(graph), projections: { check: ["a"] } };
  const ports = { sourceSnapshot: { digest: manifest.source_digest }, catalog: { rows: [row] }, surface: { targets: [{ name: "check", command_id: manifest.command_id }] }, capability: manifest.capability_snapshot, host: context.host_profile, compileRun: () => plan };
  const options = { allowOlderResultsDir: false, relToRepo: (value) => path.relative(parent, value), repoRoot: repository, resultsDirInput: root };
  return { root, row, plan, ports, options, save, alter, preflight: () => createRetainedRunPreflight(options, ports) };
}

export async function assertRetainedFinalizerPreflight() {
  const parent = fs.mkdtempSync(path.join(os.tmpdir(), "cartulary-retained-finalizer-"));
  try {
    const good = fixture(parent, "valid");
    assert.equal((await good.preflight().validate(good.root, "scheduler_drift_validation")).ok, true);
    const internal = fixture(path.join(parent, "internal"), "run");
    const target = JSON.parse(fs.readFileSync(path.join(internal.root, "target-summaries/check.json"), "utf8"));
    target.target = "backend-integration-support";
    target.workload_digest = hash({ target: target.target, evidence_outputs: [`rows/${internal.row.row_id}.json`] });
    internal.save(`target-summaries/${target.target}.json`, target);
    internal.alter("execution-context.json", (x) => { x.projections[target.target] = { command_id: target.command_id, unit_ids: ["a"] }; });
    internal.alter("run-summary.json", (x) => { x.artifact_refs.push(`target-summaries/${target.target}.json`); });
    internal.plan.projections[target.target] = ["a"];
    internal.ports.surface.targets.push({ name: target.target, target_class: "check_internal" });
    assert.equal((await internal.preflight().validate(internal.root, "scheduler_drift_validation")).ok, true);
    const mutations = [
      ["cold", (f) => f.alter("run-manifest.json", (x) => { x.cache_mode = "cold"; }), "configuration_error"],
      ["off", (f) => f.alter("run-manifest.json", (x) => { x.cache_mode = "off"; }), "configuration_error"],
      ["slice", (f) => f.alter("run-manifest.json", (x) => { x.target = "test-slice"; }), "configuration_error"],
      ["source", (f) => { f.ports.sourceSnapshot = { digest: `sha256:${"b".repeat(64)}` }; }],
      ["host", (f) => { f.ports.host = { ...f.ports.host, kernel_release: "foreign" }; }],
      ["capacity", (f) => { f.ports.capability = { ...f.ports.capability, cpu_tokens: 1 }; }],
      ["partial-graph", (f) => { f.plan.graph = { ...f.plan.graph, graph_digest: `sha256:${"b".repeat(64)}` }; }],
      ["partial-target", (f) => { f.plan.projections.check = []; }],
      ["command", (f) => { f.ports.surface.targets[0].command_id = "cartulary.harness.command.check.v3"; }],
      ["unit-semantic", (f) => f.alter("unit-results/a.json", (x) => { x.semantic_digest = `sha256:${"b".repeat(64)}`; })],
      ["unit-exit", (f) => f.alter("unit-results/a.json", (x) => { x.exit_code = null; })],
      ["unit-evidence", (f) => f.alter("unit-results/a.json", (x) => { x.evidence_outputs = []; })],
      ["missing-unit", (f) => fs.unlinkSync(path.join(f.root, "unit-results/a.json"))],
      ["row-runner", (f) => f.alter(`rows/${f.row.row_id}.json`, (x) => { x.runner = "vitest"; })],
      ["row-exit", (f) => f.alter(`rows/${f.row.row_id}.json`, (x) => { x.exit_code = 1; })],
      ["missing-row", (f) => fs.unlinkSync(path.join(f.root, `rows/${f.row.row_id}.json`))],
      ["context", (f) => f.alter("execution-context.json", (x) => { x.manifest_digest = `sha256:${"b".repeat(64)}`; })],
      ["cleanup", (f) => f.alter("cleanup-results.json", (x) => { x.results = []; })],
      ["scan-policy", (f) => f.alter("retained-secret-scan.json", (x) => { x.policy_digest = "invalid"; })],
      ["failed-scan", (f) => f.alter("retained-secret-scan.json", (x) => { x.status = "fail"; })],
      ["row-symlink", (f) => { const file = path.join(f.root, `rows/${f.row.row_id}.json`); fs.unlinkSync(file); fs.symlinkSync("/dev/null", file); }],
    ];
    for (const [name, mutate, reason = "artifact_error"] of mutations) {
      const f = fixture(path.join(parent, name), "run"); mutate(f);
      const result = await f.preflight().validate(f.root, "scheduler_drift_validation");
      assert.equal(result.ok, false, name);
      assert.equal(result.failure.failure_reason, reason, `${name}: ${result.failure.headline}`);
    }
    const originalOpen = fs.openSync;
    try {
      fs.openSync = () => assert.fail("prohibited evidence must reject before filesystem traversal");
      syncBuiltinESMExports();
      for (const component of ["docs", "README", "human.Md"]) {
        const result = await good.preflight().validate(path.join(parent, component, "run"), "scheduler_drift_validation");
        assert.equal(result.ok, false);
        assert.match(result.failure.headline, /documentation/u);
      }
    } finally { fs.openSync = originalOpen; syncBuiltinESMExports(); }
    const newer = fixture(parent, "newer", 1);
    const old = await good.preflight().validate(good.root, "scheduler_drift_validation");
    assert.equal(old.selection.status, "older_rejected");
    good.options.allowOlderResultsDir = true;
    assert.equal((await good.preflight().validate(good.root, "scheduler_drift_validation")).selection.status, "older_with_override");
    const cold = fixture(parent, "latest-cold", 2);
    cold.alter("run-manifest.json", (x) => { x.cache_mode = "cold"; });
    assert.equal((await newer.preflight().validate(newer.root, "scheduler_drift_validation")).selection.status, "latest");
    const incomplete = fixture(parent, "latest-incomplete", 3);
    incomplete.alter("unit-results/a.json", (x) => { x.evidence_outputs = []; });
    assert.equal((await newer.preflight().validate(newer.root, "scheduler_drift_validation")).selection.status, "latest");
  } finally { fs.rmSync(parent, { recursive: true, force: true }); }
}
