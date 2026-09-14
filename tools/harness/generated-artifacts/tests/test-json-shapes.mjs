#!/usr/bin/env node

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";

import { validateSchemaSync } from "../../contract/index.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const result = spawnSync(process.execPath, ["tools/harness/generated-artifacts/check-json-shapes.mjs"], {
  cwd: root,
  encoding: "utf8",
});
assert.equal(result.status, 0, result.stderr || result.stdout);
const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-projection-boundary-"));
try {
  const registry = JSON.parse(readFileSync(path.join(root, "contracts/index.json"), "utf8"));
  const timezoneProjection = registry.families.find((family) => family.family_id === "string-contracts").typescript_projections[0];
  assert.equal(timezoneProjection.artifact_path, "contracts/string-contracts/timezone_name_registry.v1.json");
  for (const [familyID, artifactPath, expected] of [
    ["string-contracts", "contracts/string-contracts/timezone_name_registry_provenance.v1.json", "permits only the public timezone-name registry"],
    ["string-contracts", "contracts/string-contracts/index.json", "permits only the public timezone-name registry"],
    ["parties", "contracts/parties/index.json", "must stay empty for protected backend-only inputs"],
  ]) {
    const candidate = structuredClone(registry);
    const family = candidate.families.find((entry) => entry.family_id === familyID);
    family.typescript_projections = [{ ...timezoneProjection, artifact_path: artifactPath }];
    const fixture = path.join(scratch, `${familyID}-${path.basename(artifactPath)}`);
    writeFileSync(fixture, JSON.stringify(candidate));
    const rejected = spawnSync(process.execPath, ["tools/harness/generated-artifacts/check-json-shapes.mjs", "--kind", "contract-family-registry", "--file", fixture], { cwd: root, encoding: "utf8" });
    assert.notEqual(rejected.status, 0);
    assert.ok(rejected.stderr.includes(expected), rejected.stderr || rejected.stdout);
  }
} finally {
  rmSync(scratch, { recursive: true, force: true });
}
for (const [file, schemaID] of [
  ["tools/execution_topology_manifest.json", "cartulary.execution_topology.v8"],
  ["tools/scheduler_manifest.json", "cartulary.scheduler_manifest.v3"],
  ["tools/browser_e2e_batch_manifest.json", "cartulary.browser_e2e_batch_manifest.v11"],
  ["tools/harness_work_graph_owner.json", "cartulary.harness_work_graph_owner.v2"],
]) {
  const value = JSON.parse(readFileSync(path.join(root, file), "utf8"));
  assert.equal(value.schema_id, schemaID);
  validateSchemaSync(schemaID, value);
}
const attachments = JSON.parse(readFileSync(path.join(root, "tools/harness_schema_attachments.json"), "utf8"));
for (const retired of [
  "cartulary.test_family_manifest.v2",
  "cartulary.execution_topology.v5",
  "cartulary.scheduler_manifest.v2",
]) {
  assert.equal(attachments.attachments.some((entry) => entry.schema_id === retired), false, `${retired} must be rejected as current input`);
}
