import { WorkGraphCompiler } from "./compiler.mjs";
import { selectionFor, replayInputs } from "./preparation.mjs";
import { retainedGoWorkerBudget } from "./capability.mjs";

// Reconstruct only the execution contract. Historical validation must not
// acquire services, capture current capacity, inspect caches or start work.
export function compileRetainedRun(root, manifest) {
  const compiler = new WorkGraphCompiler(root);
  const recipe = compiler.taskSurface.make_recipes[manifest.target];
  const entry = compiler.taskSurface.targets.find((value) => value.name === manifest.target);
  if (!recipe || entry?.command_id !== manifest.command_id ||
      (recipe.type !== "work_graph" && recipe.graph_entry !== true)) {
    throw new Error("retained command is not a current graph entry");
  }
  compiler.configureCapacity(manifest.capability_snapshot, retainedGoWorkerBudget(manifest));
  const inputs = replayInputs(entry, manifest.declared_inputs);
  compiler.targetInputs[manifest.target] = inputs;
  return selectionFor({
    selection: recipe.selection ?? "target",
    target: manifest.target,
    owner: inputs.OWNER ?? "",
    rows: inputs.ROWS ? inputs.ROWS.split(",") : recipe.row_ids,
    serviceBackedOnly: recipe.service_backed_only === true,
  }, compiler);
}
