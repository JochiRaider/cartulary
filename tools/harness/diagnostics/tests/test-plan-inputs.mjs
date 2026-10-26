import assert from "node:assert/strict";
import test from "node:test";
import fs, { writeFileSync, symlinkSync } from "node:fs";
import { syncBuiltinESMExports } from "node:module";
import path from "node:path";
import { WorkGraphCompiler } from "../../scheduler/work-graph/index.mjs";
import { validateSchemaSync } from "../../contract/index.mjs";
import { buildTargetPlan, readPlanInputs, renderTargetPlan } from "../target-plan.mjs";
import { buildOwnerTaskGuide } from "../owner-diagnostics.mjs";
import { createScratch } from "../../test-support/scratch.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");

test("planning validates exact declared input files through the execution resolver", (t) => {
  const scratch = createScratch(t, "target-plan-inputs-");
  const file = path.join(scratch, "inputs.json");
  const compiler = new WorkGraphCompiler(root);
  writeFileSync(file, '{"COMPONENT":"web"}');
  const inputs = readPlanInputs(root, file);
  const plan = buildTargetPlan(root, "build", inputs, compiler);
  validateSchemaSync(plan.schema_id, plan);
  assert.equal(plan.graph_digest, compiler.compileCommand("build", inputs).graph.graph_digest);
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
  for (const invalid of [{ ARGS: "--unsafe" }, { COMPONENT: "unknown" }, { COMPONENT: 1 }]) {
    assert.throws(() => buildTargetPlan(root, "build", invalid, compiler));
  }
  assert.throws(() => buildTargetPlan(root, "bootstrap-tool-installations", {}, compiler), /unknown public target/u);
  assert.throws(() => buildTargetPlan(root, "absent", {}, compiler), /unknown public target/u);
  const guide = buildOwnerTaskGuide(root, "harness.command_surface");
  validateSchemaSync(guide.schema_id, guide);
  assert.equal(guide.rows.length, guide.coverage.row_count);
  assert.ok(guide.rows.some((row) => row.row_id.endsWith("scheduler_smoke") && row.aggregate_placement === "fast"));
  assert.ok(guide.rows.some((row) => row.row_id.endsWith("native_runtime") && row.aggregate_placement === "full"));
  assert.equal(Object.hasOwn(guide, "role"), false);
  for (const [owner, commands] of [
    ["module.networkflow", ["make browser-e2e SUITE=visual"]],
    ["harness.browser", ["make browser-e2e SUITE=visual", "make harness-contract SUITE=ui-review"]],
  ]) {
    const explicitGuide = buildOwnerTaskGuide(root, owner);
    validateSchemaSync(explicitGuide.schema_id, explicitGuide);
    assert.deepEqual(explicitGuide.explicit_commands, commands);
  }
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
  writeFileSync(file, JSON.stringify({ schema_id: "cartulary.harness_capacity_override.v1", cpu_tokens: 1, postgres_lanes: 1 }));
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
