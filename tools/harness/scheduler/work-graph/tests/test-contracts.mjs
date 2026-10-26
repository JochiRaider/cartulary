import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import { validateSchemaSync } from "../../../contract/index.mjs";
import { createContractTestContext } from "../../../tests/contract-test-context.mjs";
const root = path.resolve(import.meta.dirname, "../../../../..");
function readJSON(relative) {
  return JSON.parse(readFileSync(path.join(root, relative), "utf8"));
}
export function registerBoundariesTests() {
  test("work_graph_owner_validation: work graph owner validates against its current schema", async () => {
    validateSchemaSync("cartulary.harness_work_graph_owner.v8", readJSON("tools/harness_work_graph_owner.json"));
  });
}
export function registerCommandSurfaceTests() {
  test("graph_entrypoint_contract: graph entrypoints use one work-graph command model", async () => {
    const context = createContractTestContext(root);
    const graphTargets = ["test-slice", "browser-e2e", "test-fast", "check", "ci", "release-check", "lint"];
    assert.ok(graphTargets.every((target) => context.taskSurface.make_recipes[target]?.type === "work_graph"));
    assert.ok(context.taskSurface.observability_policy.required_targets.every((target) => {
      const recipe = context.taskSurface.make_recipes[target];
      return recipe?.type === "work_graph" || recipe?.graph_entry === true;
    }));
  });
  test("graph_input_contract: observable graph targets expose cache and capacity inputs", async () => {
    const context = createContractTestContext(root);
    for (const target of context.taskSurface.observability_policy.required_targets) {
      const inputNames = (context.taskSurface.targets.find((entry) => entry.name === target)?.input_contract?.inputs ?? []).map((input) => input.name);
      assert.ok(inputNames.includes("CARTULARY_HARNESS_CACHE_MODE"));
      assert.ok(inputNames.includes("CARTULARY_HARNESS_CAPACITY_OVERRIDE"));
    }
  });
}
