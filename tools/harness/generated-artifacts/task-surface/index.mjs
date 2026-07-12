import { defaultTaskSurfaceManifestPath, readJSON, resolveRepoPath } from "./model.mjs";
import { validateTaskSurfaceManifest } from "./validation.mjs";

export * from "./model.mjs";
export * from "./validation.mjs";
export * from "./make-renderer.mjs";

export function loadTaskSurfaceManifest(file = defaultTaskSurfaceManifestPath) {
  const manifestPath = resolveRepoPath(file);
  const manifest = readJSON(manifestPath);
  validateTaskSurfaceManifest(manifest, manifestPath);
  return { manifest, manifestPath };
}
