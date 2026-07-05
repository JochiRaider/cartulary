export * from "./frontend/registry.mjs";
export * from "./frontend/validation.mjs";
export * from "./frontend/visual-fixtures.mjs";
export * from "./frontend/ledger.mjs";
export * from "./frontend/scenario-grep.mjs";
export * from "./frontend/phase-ids.mjs";

import { runFrontendPhaseManifestCLI } from "./frontend/cli.mjs";

if (import.meta.url === `file://${process.argv[1]}`) {
  runFrontendPhaseManifestCLI();
}
