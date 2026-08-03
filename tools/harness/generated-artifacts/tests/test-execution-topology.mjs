#!/usr/bin/env node

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../../contract/index.mjs";
import { WorkGraphCompiler } from "../../scheduler/work-graph/index.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const topology = JSON.parse(readFileSync(path.join(root, "tools/execution_topology_manifest.json"), "utf8"));
const scheduler = JSON.parse(readFileSync(path.join(root, "tools/scheduler_manifest.json"), "utf8"));
const browser = JSON.parse(readFileSync(path.join(root, "tools/browser_e2e_batch_manifest.json"), "utf8"));
validateSchemaSync(topology.schema_id, topology);
validateSchemaSync(scheduler.schema_id, scheduler);
validateSchemaSync(browser.schema_id, browser);
for (const retired of [
  "fixture_profiles",
  "sequence_resource_profiles",
  "sequence_schedules",
  "check_schedules",
  "service_backed_schedules",
]) {
  assert.equal(Object.hasOwn(topology, retired), false, `${retired} must not survive the v6 cutover`);
}
assert.deepEqual(scheduler.schedules, []);
assert.equal(scheduler.generated.source_authoring.work_graph, "tools/harness_work_graph_owner.json");
const compiler = new WorkGraphCompiler(root);
for (const target of ["test-fast", "test", "check", "ci", "release-check"]) {
  const first = compiler.compileAggregatePlan(target);
  const second = compiler.compileAggregatePlan(target);
  assert.deepEqual(first, second, `${target} compilation must be deterministic`);
  assert.ok(first.graph.units.length > 0);
}
