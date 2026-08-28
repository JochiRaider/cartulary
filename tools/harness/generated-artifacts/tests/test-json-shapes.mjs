#!/usr/bin/env node

import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import path from "node:path";

import { validateSchemaSync } from "../../contract/index.mjs";

const root = path.resolve(import.meta.dirname, "../../../..");
const result = spawnSync(process.execPath, ["tools/harness/generated-artifacts/check-json-shapes.mjs"], {
  cwd: root,
  encoding: "utf8",
});
assert.equal(result.status, 0, result.stderr || result.stdout);
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
