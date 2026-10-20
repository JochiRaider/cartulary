import path from "node:path";

import { parseStrictJSON, semanticJSONDigest, validateSchemaSync } from "../../contract/index.mjs";
import { readLocalFile } from "../../runtime/secure-local-files.mjs";
import { validateWorkGraph } from "../../scheduler/work-graph/model.mjs";

// The already-published graph owns selection. Do not copy a cohort's potentially
// large row set into argv or an environment variable with a different bound.
export function readPublishedUnit(runRoot, runID, unitID) {
  const read = (name) => parseStrictJSON(readLocalFile(path.join(runRoot, name), { maximum: 64 * 1024 * 1024 }).toString("utf8"));
  const manifest = read("run-manifest.json"), context = read("execution-context.json");
  validateSchemaSync("cartulary.harness_run_manifest.v3", manifest);
  validateSchemaSync("cartulary.harness_execution_context.v3", context);
  if (manifest.run_id !== runID || context.manifest_digest !== semanticJSONDigest(manifest) ||
      context.graph.graph_digest !== manifest.graph_digest) throw new Error("row selection does not bind to its invocation");
  validateWorkGraph(context.graph);
  const unit = context.graph.units.find((candidate) => candidate.unit_id === unitID);
  if (!unit) throw new Error("selection requires its published work unit");
  return unit;
}

export function readUnitSelection(runRoot, runID, unitID) {
  const unit = readPublishedUnit(runRoot, runID, unitID);
  if (!unit || unit.command.args.length !== 2 || unit.command.args[0] !== "tools/harness/execution/runners/row-runner-cli.mjs" ||
      unit.command.args[1] !== "--unit-rows" || unit.row_bindings.length === 0 || unit.row_bindings.some((row) => row.runner !== "vitest")) {
    throw new Error("row selection requires its exact Vitest work unit");
  }
  return unit.row_bindings.map((row) => row.row_id);
}
