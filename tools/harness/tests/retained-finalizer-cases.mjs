import assert from "node:assert/strict";
import fs from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import os from "node:os";
import path from "node:path";
import { semanticJSONDigest } from "../contract/index.mjs";
import { createRetainedRunPreflight } from "../finalization/agent-finalize-retained-run.mjs";
import { buildCompleteEvidence, buildEvidenceUnit, loadFixtureIdentity, persistEvidence, writeEvidenceFile } from "../observability/tests/retained-fixture.mjs";

const repository = path.resolve(import.meta.dirname, "../../..");
const hash = semanticJSONDigest;

function fixture(parent, name, second = 0) {
  const root = path.join(parent, name);
  const row = { row_id: "harness.evidence_accounting.unit.finalizer_fixture", runner: "go" };
  const evidence = buildCompleteEvidence({ runID: name, target: "check", identity: loadFixtureIdentity("check"),
    rows: [row], units: [buildEvidenceUnit("a", [row])], cacheMode: "normal",
    startedAt: new Date(Date.UTC(2026, 9, 16, 0, 0, second)).toISOString() });
  persistEvidence(root, evidence);
  const { manifest, context } = evidence;
  const graph = context.graph;
  const save = (relative, value) => writeEvidenceFile(root, relative, value);
  const alter = (relative, mutate) => {
    const value = JSON.parse(fs.readFileSync(path.join(root, relative), "utf8"));
    mutate(value); save(relative, value);
  };
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
      ["command", (f) => { f.ports.surface.targets[0].command_id = "cartulary.harness.command.foreign.v1"; }],
      ["unit-semantic", (f) => f.alter("unit-results/a.json", (x) => { x.semantic_digest = `sha256:${"b".repeat(64)}`; })],
      ["unit-exit", (f) => f.alter("unit-results/a.json", (x) => { x.exit_code = null; })],
      ["unit-evidence", (f) => f.alter("unit-results/a.json", (x) => { x.evidence_outputs = []; })],
      ["failed-unit", (f) => f.alter("unit-results/a.json", (x) => { x.status = "failed"; x.exit_code = 1; x.failure_class = "product"; x.failure_reason = "test_assertion_failure"; })],
      ["cancelled-unit", (f) => f.alter("unit-results/a.json", (x) => { x.status = "cancelled"; x.exit_code = 130; })],
      ["missing-unit", (f) => fs.unlinkSync(path.join(f.root, "unit-results/a.json"))],
      ["row-runner", (f) => f.alter(`rows/${f.row.row_id}.json`, (x) => { x.runner = "vitest"; })],
      ["row-exit", (f) => f.alter(`rows/${f.row.row_id}.json`, (x) => { x.exit_code = 1; })],
      ["failed-row", (f) => f.alter(`rows/${f.row.row_id}.json`, (x) => { x.terminal_state = "failed"; x.exit_code = 1; x.failure_class = "product"; x.failure_reason = "test_assertion_failure"; })],
      ["cancelled-row", (f) => f.alter(`rows/${f.row.row_id}.json`, (x) => { x.terminal_state = "cancelled"; x.exit_code = 130; })],
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
