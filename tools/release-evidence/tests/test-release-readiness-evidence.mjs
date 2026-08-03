#!/usr/bin/env node

import assert from "node:assert/strict";
import path from "node:path";

import { WorkGraphCompiler } from "../../harness/scheduler/work-graph/index.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const compiler = new WorkGraphCompiler(root);
const plan = compiler.compileAggregatePlan("release-check");
assert.ok(plan.projections["release-check"].length > 0);
assert.ok(plan.projections["release-readiness-evidence"].length > 0);
assert.equal(plan.graph.units.filter((unit) => unit.unit_id === "target:release-readiness-evidence").length, 1);
assert.equal(plan.graph.units.some((unit) => unit.command.args.includes("release-check")), false, "release evidence must not nest the release aggregate");

