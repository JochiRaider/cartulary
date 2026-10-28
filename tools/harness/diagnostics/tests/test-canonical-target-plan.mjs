import assert from "node:assert/strict";
import test from "node:test";
import childProcess, { spawnSync } from "node:child_process";
import path from "node:path";
import fs, { writeFileSync } from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import { createScratch } from "../../test-support/scratch.mjs";
import { assertHarnessTierRouting } from "../../scheduler/work-graph/tests/tier-selection.mjs";
import { assertBlockingPolicy, assertSmokeContexts, assertAssuranceUnions, assertAggregateResources } from "../../scheduler/work-graph/tests/aggregate-assertions.mjs";
import { WorkGraphCompiler, compileRetainedRun, selectionFor } from "../../scheduler/work-graph/index.mjs";
import { buildTargetPlan, renderTargetPlan } from "../target-plan.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const rowIDs = (units) => [...new Set(units.flatMap((unit) => unit.evidence_outputs ?? unit.current_run_evidence_outputs)
  .filter((output) => output.startsWith("rows/")).map((output) => output.slice(5, -5)))].sort();

function assertComposition(target, supplied = {}) {
  const plan = buildTargetPlan(root, target, supplied);
  const compiler = new WorkGraphCompiler(root);
  compiler.configureCapacity(plan.capability_snapshot, plan.go_worker_budget);
  compiler.targetInputs[target] = supplied;
  const recipe = compiler.taskSurface.make_recipes[target];
  const execution = selectionFor({ selection: recipe.selection ?? "target", target,
    owner: supplied.OWNER, rows: supplied.ROWS?.split(",") ?? recipe.row_ids,
    serviceBackedOnly: recipe.service_backed_only === true }, compiler);
  const replay = compileRetainedRun(root, { target, command_id: plan.command_id, declared_inputs: plan.inputs,
    capability_snapshot: plan.capability_snapshot, go_worker_budget: plan.go_worker_budget });
  assert.equal(plan.schema_id, "cartulary.harness_target_plan.v7");
  for (const expected of [execution, replay]) {
    assert.equal(plan.graph_digest, expected.graph.graph_digest, `${target}: graph identity`);
    assert.deepEqual(plan.projections, expected.projections, `${target}: every projection and member`);
    assert.deepEqual(plan.units.map((unit) => [unit.unit_id, unit.semantic_digest]),
      expected.graph.units.map((unit) => [unit.unit_id, unit.semantic_digest]));
  }
  assert.deepEqual(plan, buildTargetPlan(root, target, supplied), `${target}: deterministic plan`);
  assert.ok(plan.go_worker_budget > 0 && plan.go_worker_budget <= plan.capability_snapshot.cpu_tokens);
  const byID = new Map(plan.units.map((unit) => [unit.unit_id, unit]));
  assert.equal(byID.size, plan.units.length);
  for (const [projection, ids] of Object.entries(plan.projections)) {
    assert.equal(new Set(ids).size, ids.length, projection);
    assert.ok(ids.every((id) => byID.has(id) && byID.get(id).needs.every((dependency) => ids.includes(dependency))), projection);
  }
  return { plan, compiler, execution };
}

test("aggregate planning, execution and replay preserve policy, contexts and resources", async (t) => {
  const plans = new Map();
  let matrixCompiler;
  for (const target of ["test-fast", "check", "ci", "release-check"]) {
    await t.test(`${target} planning has the complete execution and retained projection union`, () => {
      const { plan, compiler, execution } = assertComposition(target);
      plans.set(target, execution);
      matrixCompiler = compiler;
      assert.ok(plan.units.length > 0);
      assert.deepEqual(plan.projections[target], plan.units.map((unit) => unit.unit_id));
      for (const unit of plan.units) {
        assert.ok(unit.owner_id);
        if (unit.service_dependencies.includes("postgres")) assert.ok(unit.resource_claims.postgres > 0);
      }
      if (target !== "test-fast") assert.ok(plan.units.some((unit) => unit.fixture_capability.startsWith("postgres_")));
      if (["ci", "release-check"].includes(target)) {
        const gate = "operation:harness.frontend_fallow_static";
        assert.ok(plan.projections["frontend-fallow-static"].includes(gate));
        assert.ok(plan.projections[target].includes(gate));
        assert.equal(plan.units.filter((unit) => unit.unit_id === gate).length, 1);
      }
    });
  }
  assertHarnessTierRouting({ compiler: matrixCompiler }, plans);
  assertBlockingPolicy(matrixCompiler, plans);
  assertSmokeContexts(matrixCompiler, plans);
  assertAssuranceUnions(matrixCompiler, plans);
  assertAggregateResources(plans);
});

test("direct planning preserves independently expected schema projections", () => {
  const { plan } = assertComposition("source-check", { CHECK: "schema" });
  const install = "operation:harness.frontend_install", drift = "operation:harness.toolchain_drift", schema = "operation:harness.json_shape_check";
  assert.deepEqual(plan.projections, {
    "source-check": [install, schema, drift],
    "harness.frontend_install": [install],
    "harness.json_shape_check": [install, schema, drift],
    "harness.toolchain_drift": [install, drift],
  });
  assert.deepEqual(plan.rows, []);
});

test("exact and service selections preserve their own rows and child projections", () => {
  for (const [owner, row, scope] of [
    ["harness.command_surface", "harness.command_surface.unit.scheduler_smoke", "all"],
    ["harness.browser", "harness.browser.integration.postgres_transaction_isolation", "services"],
  ]) {
    const { plan } = assertComposition("test-slice", { OWNER: owner, ROWS: row, SCOPE: scope });
    assert.deepEqual(rowIDs(plan.units), [row]);
    const child = scope === "all" ? "backend-unit" : "backend-integration";
    assert.deepEqual(rowIDs(plan.units.filter((unit) => plan.projections[child].includes(unit.unit_id))), [row]);
    assert.equal(plan.units.some((unit) => unit.unit_id === "operation:harness.test_service_images"), scope === "services");
  }
});

test("visual planning includes every screenshot row in its visual child projection", () => {
  const { plan, compiler } = assertComposition("browser-e2e", { SUITE: "visual" });
  const members = plan.projections["browser-e2e-visual"];
  assert.ok(members?.length, "the visual child must be published by planning");
  const expected = compiler.catalog.rows.filter((row) => row.runner === "playwright" && row.selector.stage === "visual")
    .map((row) => row.row_id);
  // The visual fixture registry is a required dependency of the screenshot units.
  expected.push("web.design.static.visual_fixture_registry");
  assert.deepEqual(rowIDs(plan.units.filter((unit) => members.includes(unit.unit_id))), expected.sort());
  assert.ok(plan.projections["browser-e2e"].includes("browser_target_summary:browser-e2e-visual"));
  assert.ok(members.includes("browser_target_summary:browser-e2e-visual"));
});

test("planning service-backed work does not acquire fixtures, write files or launch commands", () => {
  const compiler = new WorkGraphCompiler(root);
  const saved = new Map();
  for (const [object, names] of [[childProcess, ["spawn", "spawnSync", "exec", "execSync", "execFile", "execFileSync"]],
    [fs, ["mkdirSync", "mkdtempSync", "writeFileSync", "appendFileSync", "renameSync"]]]) {
    for (const name of names) {
      saved.set([object, name], object[name]);
      object[name] = () => assert.fail(`planning performed ${name}`);
    }
  }
  syncBuiltinESMExports();
  try {
    const plan = buildTargetPlan(root, "test-slice", { OWNER: "harness.browser", ROWS: "harness.browser.integration.postgres_transaction_isolation", SCOPE: "services" }, compiler);
    assert.deepEqual(plan.rows.map((row) => row.row_id), ["harness.browser.integration.postgres_transaction_isolation"]);
  } finally {
    for (const [[object, name], original] of saved) object[name] = original;
    syncBuiltinESMExports();
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
  assert.match(compact.stdout, /target=source-check units=3 cached=\d+ fixtures=\d+ digest=sha256:/u);
  const plan = buildTargetPlan(root, "source-check", { CHECK: "schema" });
  const detailed = renderTargetPlan(plan, "graph");
  assert.ok(compact.stdout.split("\n").length < detailed.split("\n").length);
  assert.equal(invoke("--target", "unknown-target").status, 2);
  assert.equal(invoke("--unknown-option").status, 2);
});
