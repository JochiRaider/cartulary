import { runContractSuite } from "./contract-suite-support.mjs";
runContractSuite("graph");
import "./test-catalog-metadata-fixtures.mjs";

import assert from "node:assert/strict";
import test from "node:test";
import { WorkGraphCompiler } from "../scheduler/work-graph/index.mjs";

test("public assurance unions select each surviving assertion entrypoint once", () => {
  const compiler = new WorkGraphCompiler(new URL("../../../", import.meta.url).pathname);
  // Public assurance responsibilities are stable. Private entrypoint filenames
  // may move; duplicate registration remains forbidden.
  const common = ["harness-boundary-contract", "harness-command-surface-contract", "harness-evidence-contract", "harness-graph-contract", "harness-local-session-contract", "harness-scheduler-contract", "harness-ui-review-contract"];
  const ci = compiler.compileAggregatePlan("ci");
  for (const [target, additional] of [
    ["harness-contract", "harness-observation-contract"],
    ["harness-contract-tests", "harness-playwright-adapter-contract"],
  ]) {
    const leaves = compiler.owner.target_members[target];
    const files = leaves.flatMap((name) => compiler.taskSurface.make_recipes[name].command)
      .filter((arg) => arg.endsWith(".mjs")).map((arg) => arg.replace(/^\.\//u, ""));
    assert.deepEqual([...leaves].sort(), [...common, additional].sort(), target);
    assert.equal(new Set(files).size, files.length, "no entrypoint is registered twice within the assurance union");
    const direct = compiler.compileTarget(target);
    assert.ok(direct.units.every((unit) => unit.unit_id !== `target:${target}`), "unions do not launch a second scheduler");
    assert.deepEqual(ci.projections[target], direct.units.map((unit) => unit.unit_id).sort(), "CI retains the public union projection");
    for (const unit of direct.units) {
      assert.deepEqual(ci.graph.units.filter((entry) => entry.unit_id === unit.unit_id), [unit], "one canonical definition across direct and CI selection");
    }
  }
  const check = compiler.compileAggregatePlan("check");
  assert.equal(check.projections["harness-contract"], undefined, "partial selection cannot claim the full assurance union");
  const full = compiler.compileAggregatePlan("test");
  assert.equal(full.projections["harness-contract"], undefined, "observation-only policy is not added to the full test tier");
  assert.ok(full.projections["harness-contract-tests"], "full test retains its prior assertion union");
});
