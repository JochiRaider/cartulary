#!/usr/bin/env node

import assert from "node:assert/strict";
import "./test-package-smoke-cleanup.mjs";
import path from "node:path";

import { WorkGraphCompiler } from "../../harness/scheduler/work-graph/index.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const compiler = new WorkGraphCompiler(root);
const plan = compiler.compileAggregatePlan("release-check");
assert.ok(plan.projections["release-check"].length > 0);
assert.ok(plan.projections["release-readiness-evidence"].length > 0);
assert.equal(plan.graph.units.filter((unit) => unit.unit_id === "target:release-readiness-evidence").length, 1);
assert.equal(plan.graph.units.some((unit) => unit.command.args.includes("release-check")), false, "release evidence must not nest the release aggregate");

const units = new Map(plan.graph.units.map((unit) => [unit.unit_id, unit]));
const readinessClosure = new Set();
function collectDependencies(id) {
  if (readinessClosure.has(id)) return;
  readinessClosure.add(id);
  for (const dependency of units.get(id).needs) collectDependencies(dependency);
}
collectDependencies("target:release-readiness-evidence");
for (const target of ["standup-package-smoke", "standup-operational-recovery-smoke", "standup-reference-pack-smoke"]) {
  const executions = plan.graph.units.filter((unit) => unit.command.environment.CARTULARY_TEST_TARGET === target);
  assert.equal(executions.length, 1, `${target} must execute exactly once in the release`);
  const unit = executions[0];
  assert.match(unit.unit_id, /^row:harness\.release\.behavior\./u, "package evidence needs a semantic owner");
  assert.ok(plan.projections[target].includes(unit.unit_id), "package target must retain its canonical projection");
  assert.ok(plan.projections["release-check"].includes(unit.unit_id), "package qualification must be release-required");
  assert.ok(readinessClosure.has(unit.unit_id), "readiness must wait for qualification and cleanup");
  assert.deepEqual(unit.service_dependencies, ["object_store", "postgres"]);
  assert.ok(unit.resource_claims.object_store >= 2 && unit.resource_claims.postgres >= 2, "admit both isolated deployments");
  assert.equal(unit.resource_claims.volume, 1, "serialize package-owned volume allocation");
  assert.equal(unit.cache_policy, "none", "package qualification needs fresh execution evidence");
  assert.ok(unit.needs.includes("target:deployable-shape"), "qualify the built deployable");
}
