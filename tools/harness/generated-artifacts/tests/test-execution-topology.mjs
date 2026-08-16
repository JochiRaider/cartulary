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
const quietGroups = browser.stages.flatMap((stage) => stage.groups)
  .filter((group) => group.resource_profile_id === "browser_measurement_quiet");
assert.ok(quietGroups.length > 0, "browser manifest must contain quiet measurement groups");
assert.ok(
  quietGroups.every((group) => group.selected_row_ids.length === 1),
  "every quiet measurement predicate must own one browser session",
);
for (const stage of browser.stages) {
  const stageQuietGroups = stage.groups.filter(
    (group) => group.resource_profile_id === "browser_measurement_quiet",
  );
  assert.equal(
    new Set(stageQuietGroups.map((group) => group.browser_session_group)).size,
    stageQuietGroups.length,
    `quiet measurement browser sessions must not be shared in ${stage.name}`,
  );
  assert.ok(
    stageQuietGroups
      .filter((group) => group.fixture_profile_id)
      .every((group) => group.reset_before === undefined),
    `immutable performance-fixture clones must not be reset after preparation in ${stage.name}`,
  );
}
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
