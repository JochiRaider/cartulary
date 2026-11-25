import assert from "node:assert/strict";
import test from "node:test";
import childProcess, { spawnSync } from "node:child_process";
import path from "node:path";
import fs, { writeFileSync } from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import { createScratch } from "../../test-support/scratch.mjs";
import { redactionMarker } from "../../contract/index.mjs";
import { loadExecutionTopology } from "../../generated-artifacts/execution-topology.mjs";
import { loadTestCatalog } from "../../test-catalog/index.mjs";
import { buildTargetPlan, renderTargetPlan } from "../target-plan.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const digest = (character) => `sha256:${character.repeat(64)}`;

// A graph v9 unit with only the fields a test varies.
function stubUnit(unitID, fields = {}) {
  return {
    unit_id: unitID, owner_id: "harness.browser", kind: "operation",
    command: { executable: "node", args: [], environment: {} }, needs: [], resource_claims: { cpu: 1 },
    residents: [], shared_locks: [], exclusive_locks: [], fixture_lease: "none", service_dependencies: [],
    postgres_isolation: "none", cache_policy: "none", timeout_ms: 60000,
    current_run_evidence_outputs: [`unit-results/${unitID}.json`], row_bindings: [], reusable_artifact_outputs: [],
    runtime_binary_ids: [], failure_policy: { block_descendants: true }, estimated_work_ms: 10, semantic_digest: digest("a"),
    ...fields,
  };
}

// A planner port returning one service-backed runner unit, which binds the
// rows in reverse and twice, after an operation it needs.
function stubPlanner(rows) {
  const calls = [];
  const operation = stubUnit("operation:harness.test_service_images", { cache_policy: "content_addressed" });
  const runner = stubUnit("playwright:group", {
    kind: "runner", needs: [operation.unit_id], resource_claims: { cpu: 2, postgres: 1 }, fixture_lease: "browser_stack",
    fixture_profile_id: "browser_baseline_v1", snapshot_key: "b".repeat(64), service_dependencies: ["object_store", "postgres"],
    current_run_evidence_outputs: rows.map((row) => `rows/${row.row_id}.json`), estimated_work_ms: 2000, semantic_digest: digest("c"),
    row_bindings: [...rows, ...rows].reverse().map((row) => ({ row_id: row.row_id, runner: row.runner, ...(row.runner === "go" ? { selectors: ["TestFixture"] } : {}) })),
  });
  const plan = (args, options) => {
    calls.push({ args, options });
    return {
      schema_id: "cartulary.harness_runtime_plan.v10",
      graph: { schema_id: "cartulary.harness_work_graph.v10", graph_digest: digest("d"), units: [operation, runner] },
      manifest: { capability_snapshot: { cpu_tokens: 4, postgres_lanes: 2, snapshot_digest: digest("e") }, go_worker_budget: 3 },
      projections: {
        "test-slice": { command_id: "cartulary.harness.command.test_slice.v5", unit_ids: [operation.unit_id, runner.unit_id] },
        "backend-integration": { command_id: "cartulary.harness.command.test_slice.v5", unit_ids: [runner.unit_id] },
      },
    };
  };
  return { plan, calls, operation, runner };
}

const unreachable = () => assert.fail("target-plan planned a rejected request");

test("target plans project the Go plan of exactly the supplied inputs and launch nothing themselves", () => {
  const catalog = loadTestCatalog(root);
  const surface = loadExecutionTopology({ root }).taskSurface;
  const rows = catalog.rows.filter((row) => row.owner_id === "harness.browser" && row.service_dependencies.length > 0).slice(0, 2);
  assert.equal(rows.length, 2);
  const { plan, calls, operation, runner } = stubPlanner(rows);
  const supplied = { OWNER: "harness.browser", SCOPE: "services", CARTULARY_TEST_SERVICES_MODE: "attach", CARTULARY_TEST_SERVICES_SESSION_FILE: "/private/session.json" };
  const ambient = { SUITE: process.env.SUITE, ROWS: process.env.ROWS };
  const saved = new Map();
  for (const [object, names] of [[childProcess, ["spawn", "spawnSync", "exec", "execSync", "execFile", "execFileSync"]],
    [fs, ["mkdirSync", "mkdtempSync", "writeFileSync", "appendFileSync", "renameSync"]]]) {
    for (const name of names) {
      saved.set([object, name], object[name]);
      object[name] = () => assert.fail(`target-plan performed ${name}`);
    }
  }
  syncBuiltinESMExports();
  let targetPlan;
  try {
    // Another command's inputs in the environment are not this command's.
    Object.assign(process.env, { SUITE: "backend-unit", ROWS: "harness.browser.unit.absent" });
    targetPlan = buildTargetPlan(root, "test-slice", supplied, { plan, surface, catalog });
  } finally {
    for (const [name, value] of Object.entries(ambient)) {
      if (value === undefined) delete process.env[name];
      else process.env[name] = value;
    }
    for (const [[object, name], original] of saved) object[name] = original;
    syncBuiltinESMExports();
  }
  assert.equal(calls.length, 1);
  assert.deepEqual(calls[0].args, ["--selection", surface.make_recipes["test-slice"].selection, "--target", "test-slice"]);
  assert.equal(calls[0].options.root, root);
  const environment = calls[0].options.environment;
  for (const [name, value] of Object.entries(supplied)) assert.equal(environment[name], value);
  assert.equal(environment.CARTULARY_MAKE_INPUT_SOURCES, Object.keys(supplied).map((name) => `${name}=cli`).join(" "));
  assert.equal(Object.hasOwn(environment, "SUITE"), false);
  assert.equal(Object.hasOwn(environment, "ROWS"), false);

  assert.equal(targetPlan.schema_id, "cartulary.harness_target_plan.v7");
  assert.equal(targetPlan.target, "test-slice");
  assert.equal(targetPlan.command_id, surface.targets.find((entry) => entry.name === "test-slice").command_id);
  assert.equal(targetPlan.inputs.CARTULARY_TEST_SERVICES_SESSION_FILE, redactionMarker());
  assert.equal(targetPlan.inputs.OWNER, "harness.browser");
  assert.equal(targetPlan.graph_digest, digest("d"));
  assert.deepEqual(targetPlan.capability_snapshot, { cpu_tokens: 4, postgres_lanes: 2, snapshot_digest: digest("e") });
  assert.equal(targetPlan.go_worker_budget, 3);
  assert.deepEqual(targetPlan.projections, { "test-slice": [operation.unit_id, runner.unit_id], "backend-integration": [runner.unit_id] });
  assert.deepEqual(targetPlan.rows, rows.map((row) => row.row_id).sort().map((id) => {
    const row = catalog.rowByID.get(id);
    return { row_id: id, owner_id: row.owner_id, runner: row.runner, evidence_class: row.evidence_class,
      aggregate_placement: row.aggregate_placement, fixture_capability: row.fixture_capability, service_dependencies: row.service_dependencies };
  }));
  assert.deepEqual(targetPlan.units, [
    { unit_id: operation.unit_id, owner_id: "harness.browser", kind: "operation", needs: [], resource_claims: { cpu: 1 },
      fixture_capability: "none", service_dependencies: [], cache_policy: "content_addressed", estimated_work_ms: 10,
      semantic_digest: digest("a"), evidence_outputs: operation.current_run_evidence_outputs },
    { unit_id: runner.unit_id, owner_id: "harness.browser", kind: "runner", needs: [operation.unit_id], resource_claims: { cpu: 2, postgres: 1 },
      fixture_capability: "browser_stack", fixture_profile_id: "browser_baseline_v1", snapshot_key: "b".repeat(64),
      service_dependencies: ["object_store", "postgres"], cache_policy: "none", estimated_work_ms: 2000,
      semantic_digest: digest("c"), evidence_outputs: runner.current_run_evidence_outputs },
  ]);

  assert.equal(renderTargetPlan(targetPlan), `target=test-slice units=2 cached=1 fixtures=1 digest=${digest("d")}`);
  const header = ["target test-slice", `graph ${digest("d")}`, `capacity ${digest("e")} cpu=4 go_workers=3 postgres=2`];
  assert.deepEqual(renderTargetPlan(targetPlan, "graph").split("\n"), [...header,
    `${operation.unit_id} kind=operation owner=harness.browser needs=- fixture=none services=- cache=content_addressed work_ms=10`,
    `${runner.unit_id} kind=runner owner=harness.browser needs=${operation.unit_id} fixture=browser_stack services=object_store,postgres cache=none work_ms=2000`]);
  assert.deepEqual(renderTargetPlan(targetPlan, "rows").split("\n").slice(3).map((line) => line.split(" ")[0]), targetPlan.rows.map((row) => row.row_id));
  assert.deepEqual(renderTargetPlan(targetPlan, "artifacts").split("\n").slice(3),
    [operation, runner].flatMap((unit) => unit.current_run_evidence_outputs.map((output) => `${unit.unit_id} ${output}`)));
  assert.throws(() => renderTargetPlan(targetPlan, "units"), /DETAIL must be/u);
});

test("target-plan rejects unknown targets, invalid inputs and commands without a work graph before planning", () => {
  const ports = { plan: unreachable, surface: loadExecutionTopology({ root }).taskSurface, catalog: loadTestCatalog(root) };
  for (const target of ["absent", "bootstrap-tool-installations"]) {
    assert.throws(() => buildTargetPlan(root, target, {}, ports), /unknown public target/u);
  }
  for (const target of ["lint-markdown", "help"]) {
    assert.throws(() => buildTargetPlan(root, target, {}, ports), new RegExp(`${target} runs no work graph`, "u"));
  }
  for (const invalid of [{}, { SUITE: "backend-unit", OWNER: "harness.browser" }, { ROWS: "harness.browser.unit.fixture_broker_smoke" }, { OWNER: "harness.browser", SCOPE: "everything" }]) {
    assert.throws(() => buildTargetPlan(root, "test-slice", invalid, ports), /exactly one of OWNER, SUITE|must be one of/u);
  }
});

test("target-plan CLI parses input files and reports rendering and usage failures", (t) => {
  const scratch = createScratch(t, "target-plan-cli-");
  const file = path.join(scratch, "inputs.json");
  writeFileSync(file, JSON.stringify({ CHECK: "schema" }));
  const invoke = (...args) => {
    const result = spawnSync(process.execPath, ["tools/harness/diagnostics/target-plan-cli.mjs", ...args], { cwd: root, encoding: "utf8", timeout: 30000 });
    assert.equal(result.error, undefined);
    assert.equal(result.signal, null);
    return result;
  };
  const compact = invoke("--target", "source-check", "--inputs", file);
  assert.equal(compact.status, 0, compact.stderr);
  assert.match(compact.stdout, /^target=source-check units=[1-9]\d* cached=\d+ fixtures=\d+ digest=sha256:[a-f0-9]{64}\n$/u);
  const detailed = invoke("--target", "source-check", "--inputs", file, "--view", "graph");
  assert.equal(detailed.status, 0, detailed.stderr);
  assert.ok(compact.stdout.split("\n").length < detailed.stdout.split("\n").length);
  assert.equal(invoke("--target", "unknown-target").status, 2);
  assert.equal(invoke("--unknown-option").status, 2);
});
