import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { createContractTestContext } from "../../tests/contract-test-context.mjs";
import { assertPublicTimezoneProjection } from "./test-json-shapes.mjs";
const root = path.resolve(import.meta.dirname, "../../../..");
function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}
export function registerBoundariesTests() {
  test("public_timezone_projection_boundary: packaged timezone projection excludes backend-only artifacts", async () => {
    assertPublicTimezoneProjection();
  });
  test("schema_attachment_uniqueness: active schema attachments are unique", async () => {
    {
      const attachments = readJSON("tools/harness_schema_attachments.json");
      const ids = attachments.attachments.map((entry) => entry.schema_id);
      assert.equal(new Set(ids).size, ids.length);
      return;
    }
  });
  test("current_schema_presence: current harness schemas are present", async () => {
    for (const schema of [
      "cartulary.harness_work_graph.v8.schema.json",
      "cartulary.harness_execution_context.v3.schema.json",
      "cartulary.harness_runtime_plan.v6.schema.json",
      "cartulary.harness_run_manifest.v3.schema.json",
      "cartulary.harness_unit_event.v4.schema.json",
      "cartulary.harness_run_summary.v1.schema.json",
    ])
      assert.ok(existsSync(path.join(root, "tools/schemas", schema)));
  });
  test("retired_schema_absence: retired harness schemas remain absent", async () => {
    for (const oldSchema of [
      "cartulary.harness_execution_context.v1.schema.json",
      "cartulary.harness_execution_context.v2.schema.json",
      "cartulary.harness_runtime_plan.v3.schema.json",
      "cartulary.harness_work_graph.v5.schema.json",
      "cartulary.harness_invocation_start.v1.schema.json",
      "cartulary.harness_observability_index.v1.schema.json",
      "cartulary.harness_sequence_event.v1.schema.json",
      "cartulary.task_surface_owner.v1.schema.json",
      "cartulary.test_family_manifest.v2.schema.json",
      "cartulary.test_family_manifest.v4.schema.json",
      "cartulary.scheduler_manifest.v2.schema.json",
    ])
      assert.equal(existsSync(path.join(root, "tools/schemas", oldSchema)), false);
  });
}
