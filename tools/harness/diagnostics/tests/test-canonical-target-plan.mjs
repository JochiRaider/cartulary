#!/usr/bin/env node

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import path from "node:path";

const root = path.resolve(import.meta.dirname, "../../../..");
for (const target of ["test-fast", "check", "ci", "release-check"]) {
  const result = spawnSync(process.execPath, ["tools/harness/diagnostics/target-plan-cli.mjs", "--json", "--target", target], {
    cwd: root,
    encoding: "utf8",
    maxBuffer: 64 * 1024 * 1024,
  });
  assert.ifError(result.error);
  assert.equal(result.status, 0, result.stderr);
  const plan = JSON.parse(result.stdout);
  assert.equal(plan.schema_id, "cartulary.harness_target_plan.v4");
  assert.equal(plan.target, target);
  assert.ok(plan.units.length > 0);
  assert.ok(plan.projections[target].length > 0);
  if (["check", "ci", "release-check"].includes(target)) {
    const gateID = "row:package.grid_adapter.boundary_support.manual_package_surface_reachability";
    assert.ok(plan.projections["frontend-fallow-static"].includes(gateID));
    assert.ok(plan.projections[target].includes(gateID));
    assert.equal(plan.units.filter((unit) => unit.unit_id === gateID).length, 1,
      `${target} must include the bounded reachability gate exactly once`);
  }
}
