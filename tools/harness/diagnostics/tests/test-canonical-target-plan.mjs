#!/usr/bin/env node

import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { spawnSync } from "node:child_process";
import path from "node:path";
import { mkdtempSync, writeFileSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";

const root = path.resolve(import.meta.dirname, "../../../..");
function invoke(...args) {
  const result = spawnSync(process.execPath, ["tools/harness/diagnostics/target-plan-cli.mjs", ...args], {
    cwd: root,
    encoding: "utf8",
    maxBuffer: 64 * 1024 * 1024,
  });
  assert.ifError(result.error);
  return result;
}
for (const target of ["test-fast", "check", "ci", "release-check"]) {
  const result = invoke("--json", "--target", target);
  assert.equal(result.status, 0, result.stderr);
  const repeated = invoke("--json", "--target", target);
  assert.equal(repeated.status, 0, repeated.stderr);
  // Compare every byte without dumping two complete graphs into failure evidence.
  const digest = (bytes) => createHash("sha256").update(bytes).digest("hex");
  assert.equal(digest(repeated.stdout), digest(result.stdout), `${target}: same selection has byte-identical plans`);
  const plan = JSON.parse(result.stdout);
  assert.equal(plan.schema_id, "cartulary.harness_target_plan.v5");
  assert.equal(plan.target, target);
  assert.ok(plan.units.length > 0);
  assert.ok(plan.projections[target].length > 0);
  assert.equal(new Set(plan.units.map((unit) => unit.unit_id)).size, plan.units.length);
  for (const unit of plan.units) {
    assert.equal(typeof unit.owner_id, "string");
    assert.ok(unit.owner_id.length > 0);
    assert.ok(unit.needs.every((id) => plan.units.some((candidate) => candidate.unit_id === id)));
    if (unit.service_dependencies.includes("postgres")) {
      // Command-owned integration checks acquire their own declared fixtures;
      // every consumer still reserves capacity in the enclosing graph.
      assert.ok(unit.resource_claims.postgres > 0, `${unit.unit_id}: Postgres work reserves capacity`);
    }
  }
  if (["check", "ci", "release-check"].includes(target)) {
    assert.ok(plan.units.some((unit) => unit.fixture_capability.startsWith("postgres_")),
      `${target}: service-backed gates retain declared PostgreSQL fixture work`);
  }
  if (["ci", "release-check"].includes(target)) {
    const gateID = "operation:harness.frontend_fallow_static";
    assert.ok(plan.projections["frontend-fallow-static"].includes(gateID));
    assert.ok(plan.projections[target].includes(gateID));
    assert.equal(plan.units.filter((unit) => unit.unit_id === gateID).length, 1,
      `${target} must include the bounded reachability gate exactly once`);
  }
}
const scratch = mkdtempSync(path.join(tmpdir(), "target-plan-inputs-"));
const file = path.join(scratch, "inputs.json");
writeFileSync(file, JSON.stringify({ SUITE: "backend-store" }));
const compact = invoke("--target", "test-slice", "--inputs", file);
const detailed = invoke("--view", "graph", "--target", "test-slice", "--inputs", file);
rmSync(scratch, { recursive: true });
assert.equal(compact.status, 0, compact.stderr);
assert.equal(detailed.status, 0, detailed.stderr);
assert.match(compact.stdout, /target=test-slice units=\d+ cached=\d+ fixtures=\d+ digest=sha256:/u);
assert.ok(compact.stdout.split("\n").length < detailed.stdout.split("\n").length);
assert.match(detailed.stdout, /fixture=postgres_/u);
assert.equal(invoke("--target", "unknown-target").status, 2);
assert.equal(invoke("--unknown-option").status, 2);
