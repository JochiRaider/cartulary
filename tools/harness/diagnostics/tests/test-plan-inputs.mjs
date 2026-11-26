import assert from "node:assert/strict";
import test from "node:test";
import fs, { writeFileSync, symlinkSync } from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import path from "node:path";
import { validateSchemaSync } from "../../contract/index.mjs";
import { planInvocation } from "../../runtime/harness-binary.mjs";
import { loadExecutionTopology } from "../../generated-artifacts/execution-topology.mjs";
import { loadTestCatalog } from "../../test-catalog/index.mjs";
import { buildTargetPlan, readPlanInputs, renderTargetPlan } from "../target-plan.mjs";
import { buildOwnerTaskGuide } from "../owner-diagnostics.mjs";
import { createScratch } from "../../test-support/scratch.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");

const unreachable = () => assert.fail("a rejected or explicit-free request was planned");

test("planning validates exact declared input files through the execution resolver", (t) => {
  const scratch = createScratch(t, "target-plan-inputs-");
  const file = path.join(scratch, "inputs.json");
  writeFileSync(file, '{"COMPONENT":"web"}');
  const inputs = readPlanInputs(root, file);
  const planned = [];
  const plan = buildTargetPlan(root, "build", inputs, { plan: (...args) => { planned.push(planInvocation(...args)); return planned.at(-1); } });
  validateSchemaSync(plan.schema_id, plan);
  assert.equal(planned.length, 1);
  assert.equal(plan.graph_digest, planned[0].graph.graph_digest);
  assert.deepEqual(plan.inputs, {
    CARTULARY_HARNESS_CACHE_MODE: "normal", HARNESS_DIAGNOSTICS: "off", HARNESS_ACCOUNTING: "off",
    COMPONENT: "web", SERVER_HARNESS_BIN: "$(CURDIR)/build/bin/server-harness", OPERATOR_BIN: "$(CURDIR)/build/bin/operator",
  });
  for (const detail of ["summary", "graph", "rows", "artifacts"]) assert.ok(renderTargetPlan(plan, detail));
  for (const invalid of ['{"COMPONENT":"web","COMPONENT":"server"}', '[]', 'null']) {
    writeFileSync(file, invalid);
    assert.throws(() => readPlanInputs(root, file));
  }
  symlinkSync(file, path.join(scratch, "linked.json"));
  assert.throws(() => readPlanInputs(root, path.join(scratch, "linked.json")));
  const catalog = loadTestCatalog(root);
  const ports = { plan: unreachable, surface: loadExecutionTopology({ root }).taskSurface, catalog };
  for (const invalid of [{ ARGS: "--unsafe" }, { COMPONENT: "unknown" }, { COMPONENT: 1 }]) {
    assert.throws(() => buildTargetPlan(root, "build", invalid, ports), /must (?:not )?be|not declared/u);
  }
  // Explicit-only rows that no explicit command binds are listed by no command.
  const guide = buildOwnerTaskGuide(root, "harness.command_surface", { plan: () => ({ graph: { units: [] } }) });
  validateSchemaSync(guide.schema_id, guide);
  assert.equal(guide.rows.length, guide.coverage.row_count);
  assert.ok(guide.rows.some((row) => row.row_id.endsWith("scheduler_smoke") && row.aggregate_placement === "fast"));
  assert.ok(guide.rows.some((row) => row.row_id.endsWith("native_runtime") && row.aggregate_placement === "full"));
  assert.deepEqual(guide.explicit_commands, []);
  assert.equal(Object.hasOwn(guide, "role"), false);
  // An owner without explicit-only rows never plans an explicit command.
  const explicitOwners = new Set(catalog.rows.filter((row) => row.aggregate_placement === "explicit_only").map((row) => row.owner_id));
  const explicitFree = catalog.rows.map((row) => row.owner_id).find((owner) => !explicitOwners.has(owner));
  assert.ok(explicitFree, "the catalog has an owner without explicit-only rows");
  assert.deepEqual(buildOwnerTaskGuide(root, explicitFree, { plan: unreachable }).explicit_commands, []);
});

test("task guides name the explicit commands whose plans bind the owner's explicit-only rows", () => {
  const visual = "harness.browser.boundary_support.accessibility_contrast";
  const review = "harness.browser.boundary_support.ui_review_artifacts";
  const bound = { "browser-e2e": [visual], "harness-contract": [review, "harness.browser.unit.fixture_broker_smoke"],
    "release-measure": ["module.reporting.measurement.render_progress"], "release-check": [] };
  const calls = [];
  const plan = (args, { environment }) => {
    const target = args[args.indexOf("--target") + 1];
    calls.push({ target, args, environment });
    return { graph: { units: [{ unit_id: `stub:${target}`, row_bindings: bound[target].map((id) => ({ row_id: id, runner: "playwright" })) }] } };
  };
  const guide = buildOwnerTaskGuide(root, "harness.browser", { plan });
  validateSchemaSync(guide.schema_id, guide);
  assert.deepEqual(guide.explicit_commands, ["make browser-e2e SUITE=visual", "make harness-contract SUITE=ui-review"]);
  assert.deepEqual(calls.map(({ target }) => target).sort(), Object.keys(bound).sort());
  const byTarget = Object.fromEntries(calls.map((call) => [call.target, call]));
  assert.deepEqual(byTarget["browser-e2e"].args.slice(-2), ["--target", "browser-e2e"]);
  assert.equal(byTarget["browser-e2e"].environment.SUITE, "visual");
  assert.equal(byTarget["browser-e2e"].environment.CARTULARY_MAKE_INPUT_SOURCES, "SUITE=cli");
  assert.equal(byTarget["harness-contract"].environment.SUITE, "ui-review");
  assert.equal(byTarget["release-measure"].environment.CARTULARY_MAKE_INPUT_SOURCES, "RELEASE_RESULTS_DIR=cli");
  assert.equal(byTarget["release-check"].environment.CARTULARY_MAKE_INPUT_SOURCES, "");
  assert.equal(Object.hasOwn(byTarget["release-check"].environment, "SUITE"), false);
});

test("planning excludes documentation before filesystem access", () => {
  const methods = new Map(["readFileSync", "openSync", "statSync", "lstatSync"].map((name) => [name, fs[name]]));
  for (const name of methods.keys()) fs[name] = () => { throw new Error("unexpected filesystem access"); };
  syncBuiltinESMExports();
  try {
    const directory = "docs";
    for (const file of [path.join("/unavailable", directory, "plan.json"), "/unavailable/README.json", "/unavailable/plan.md"]) {
      assert.throws(() => readPlanInputs(root, file), /documentation|JSON file/u);
    }
  } finally {
    for (const [name, method] of methods) fs[name] = method;
    syncBuiltinESMExports();
  }
});

test("planning capacity comes from declared target inputs, not another command's environment", (t) => {
  const scratch = createScratch(t, "target-plan-capacity-");
  const file = path.join(scratch, "capacity.json");
  writeFileSync(file, JSON.stringify({ schema_id: "cartulary.harness_capacity_override.v2", cpu_tokens: 1, postgres_lanes: 1 }));
  const original = process.env.CARTULARY_HARNESS_CAPACITY_OVERRIDE;
  process.env.CARTULARY_HARNESS_CAPACITY_OVERRIDE = path.join(scratch, "missing-ambient.json");
  try {
    const selection = { OWNER: "harness.command_surface", ROWS: "harness.command_surface.unit.scheduler_smoke" };
    const ordinary = buildTargetPlan(root, "test-slice", selection);
    assert.notEqual(ordinary.capability_snapshot.sources.cpu_tokens, "override");
    const plan = buildTargetPlan(root, "test-slice", { ...selection, CARTULARY_HARNESS_CAPACITY_OVERRIDE: file });
    validateSchemaSync(plan.schema_id, plan);
    assert.equal(plan.capability_snapshot.cpu_tokens, 1);
    assert.equal(plan.capability_snapshot.postgres_lanes, 1);
    assert.equal(plan.capability_snapshot.sources.cpu_tokens, "override");
    assert.equal(plan.go_worker_budget, 1);
    assert.throws(() => buildTargetPlan(root, "test-slice", { ...selection, CARTULARY_HARNESS_CAPACITY_OVERRIDE: process.env.CARTULARY_HARNESS_CAPACITY_OVERRIDE }), /capacity override/u);
  } finally {
    if (original === undefined) delete process.env.CARTULARY_HARNESS_CAPACITY_OVERRIDE;
    else process.env.CARTULARY_HARNESS_CAPACITY_OVERRIDE = original;
  }
});
