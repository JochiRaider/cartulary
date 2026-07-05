export {
  assertAuthoritativeEvidenceNames,
  authoritativeEvidenceNameViolations,
  collectEntries,
  collectSupportGoEntries,
  entryClaimStatus,
  entryEvidenceNames,
  entryIsExecutable,
  goEntryScenarioSymbols,
  goEntrySymbols,
  playwrightEntryTitles,
  rowIDFragments,
  supportGoEntrySymbols,
  vitestEntryTitles,
} from "./phase-entry-evidence.mjs";
export {
  effectiveGoEntryPostgresFixtureBudget,
  effectiveGoEntryPostgresFixturePolicy,
  effectiveSupportGoEntryPostgresFixtureBudget,
  effectiveSupportGoEntryPostgresFixturePolicy,
  goEntryPostgresFixtureBudget,
  goEntryPostgresFixturePolicy,
  supportGoEntryPostgresFixtureBudget,
  supportGoEntryPostgresFixturePolicy,
} from "./phase-fixture-policy.mjs";
export { loadManifest, phaseManifestNames } from "./phase-manifest-loader.mjs";
export { validateManifest } from "./phase-manifest-validation.mjs";
export { loadPhasePolicyExceptions } from "./phase-policy-exceptions.mjs";
export {
  packageMatchesPattern,
  selectManifestEntries,
  selectPlaywrightEntries,
} from "./phase-selection.mjs";

if (import.meta.url === `file://${process.argv[1]}`) {
  try {
    const { runPhaseManifestCLI } = await import("./phase-manifest-cli.mjs");
    runPhaseManifestCLI(process.argv.slice(2));
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`${message}\n`);
    process.exit(1);
  }
}
