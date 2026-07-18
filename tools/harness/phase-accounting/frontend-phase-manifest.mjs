export {
  frontendPhaseNamespace,
  frontendPhaseRegistrySchemaID,
  frontendPhaseTestMapSchemaID,
} from "./frontend/constants.mjs";
export {
  frontendRegistryPath,
  frontendVisualFixtureRegistryPath,
  loadFrontendPhaseMap,
  loadFrontendPhaseRegistry,
} from "./frontend/registry-loader.mjs";
export { validateFrontendPhaseArtifacts } from "./frontend/phase-artifacts.mjs";
export { validateFrontendPhaseMap } from "./frontend/phase-map-validation.mjs";
export { validateFrontendVisualFixtureRegistry } from "./frontend/visual-fixture-registry.mjs";
export * from "./frontend/ledger.mjs";
export * from "./frontend/scenario-grep.mjs";
export * from "./frontend/phase-ids.mjs";

import { runFrontendPhaseManifestCLI } from "./frontend/cli.mjs";

if (import.meta.url === `file://${process.argv[1]}`) {
  runFrontendPhaseManifestCLI();
}
