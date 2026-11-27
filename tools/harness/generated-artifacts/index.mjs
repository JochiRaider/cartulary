export {
  canonicalInputPaths,
  collectTaskSurfaceManifestErrors,
  defaultTaskSurfaceManifestPath,
  entryRecipeFamily,
  graphForwardedInputNames,
  loadTaskSurfaceManifest,
  machineStateNames,
  publicMakeInputNames,
  requiresLiteralMakeTransport,
  stepForwardedInputNames,
  toolInputNames,
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
