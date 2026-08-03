#!/usr/bin/env node

import assert from "node:assert/strict";
import path from "node:path";

import { WorkGraphCompiler } from "../../scheduler/work-graph/index.mjs";
import { adaptGoInvocation } from "../../execution/runners/go.mjs";

const invocation = {
  rows: [{ row_id: "row-a", selectors: ["TestA"] }],
};
const passed = adaptGoInvocation(invocation, {
  status: 0,
  stdout: `${JSON.stringify({ Action: "pass", Test: "TestA", Elapsed: 0.01 })}\n`,
});
assert.deepEqual(passed.map((entry) => entry.terminal_state), ["passed"]);
const missing = adaptGoInvocation(invocation, { status: 1, stdout: "" });
assert.deepEqual(missing.map((entry) => entry.terminal_state), ["infrastructure_failed"]);
assert.deepEqual(missing.map((entry) => entry.exit_code), [3]);
const failed = adaptGoInvocation(invocation, {
  status: 1,
  stdout: `${JSON.stringify({ Action: "fail", Test: "TestA", Elapsed: 0.01 })}\n`,
});
assert.deepEqual(failed.map((entry) => entry.terminal_state), ["failed"]);
assert.deepEqual(failed.map((entry) => entry.exit_code), [10]);
const root = path.resolve(import.meta.dirname, "../../../..");
const graph = new WorkGraphCompiler(root).compile({ kind: "target", target: "backend-unit" });
const goUnits = graph.units.filter((unit) => unit.unit_id.startsWith("go:"));
assert.ok(goUnits.length > 0);
assert.ok(goUnits.every((unit) => unit.evidence_outputs.some((output) => output.startsWith("rows/"))));
assert.ok(
  goUnits.every((unit) =>
    unit.evidence_outputs.every(
      (output) => output.startsWith("rows/") || output.startsWith("unit-results/"),
    ),
  ),
);
