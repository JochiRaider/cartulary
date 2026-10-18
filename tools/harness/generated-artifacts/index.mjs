export {
  collectTaskSurfaceManifestErrors,
  defaultTaskSurfaceManifestPath,
  loadTaskSurfaceManifest,
  makeRecipeEntries,
  renderTaskSurfaceMake,
  repoRoot,
  targetEntryMap,
} from "./task-surface/index.mjs";
export {
  defaultExecutionTopologyManifestPath,
  loadExecutionTopology,
  renderBrowserBatchManifest,
  renderTaskSurfaceManifest,
  serviceRequirementForRuntimeProfile,
} from "./execution-topology.mjs";

export async function validateSourceShapes(context) {
  const validation = await import("./check-json-shapes.mjs");
  return validation.validateSourceShapes(context);
}
