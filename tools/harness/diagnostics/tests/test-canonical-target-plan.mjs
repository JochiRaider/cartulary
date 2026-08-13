#!/usr/bin/env node

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import path from "node:path";

const root = path.resolve(import.meta.dirname, "../../../..");
for (const target of ["test-fast", "check", "release-check"]) {
  const result = spawnSync(process.execPath, ["tools/harness/diagnostics/target-plan-cli.mjs", "--json", "--target", target], {
    cwd: root,
    encoding: "utf8",
  });
  assert.equal(result.status, 0, result.stderr);
  const plan = JSON.parse(result.stdout);
  assert.equal(plan.schema_id, "cartulary.harness_target_plan.v2");
  assert.equal(plan.target, target);
  assert.ok(plan.units.length > 0);
  assert.ok(plan.projections[target].length > 0);
}
