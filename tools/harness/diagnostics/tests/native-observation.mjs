import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import path from "node:path";
import { workspaceLayout } from "../../../workspace_layout.generated.mjs";
import { validateSchemaSync } from "../../contract/index.mjs";
const root = path.resolve(import.meta.dirname,"../../../..");
export const nativeBinary = path.join(root,workspaceLayout.toolbin,"cartulary-harness");
export async function observeRun(inputs) {
  const args = ["diagnostics", "--json"];
  for (const [key, flag] of [["resultsDir", "--results-dir"], ["runId", "--run-id"], ["target", "--target"], ["afterRevision", "--after-revision"], ["waitSeconds", "--wait-seconds"]]) if (inputs[key] !== undefined) args.push(flag, String(inputs[key]));
  const result = spawnSync(nativeBinary,args,{cwd:root,encoding:"utf8",timeout:35000});
  assert.ok(result.stdout,result.stderr);
  const observation=JSON.parse(result.stdout);
  validateSchemaSync("cartulary.harness_run_observation.v3",observation);
  assert.equal(result.status,observation.operation_exit_code);
  return observation;
}
export function renderObservation(observation) {
  const result=spawnSync(nativeBinary,["diagnostics","--results-dir",observation.run.result_root,...(observation.selected_target?["--target",observation.selected_target]:[])],{cwd:root,encoding:"utf8",timeout:10000});
  assert.equal(result.status,observation.operation_exit_code,result.stderr);
  return result.stdout;
}
