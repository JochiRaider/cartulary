// Private owner observations required by the harness contract fixtures.
export { createAcquisitionLaunch, recordAcquisitionProcess, readBrowserAcquisition, settleBrowserAcquisition } from "../browser/browser-acquisition.mjs";
export { validateFrontendAttachment } from "../browser/browser-session-evidence.mjs";
export { createRowSelectorResolver, resolveRowSelector } from "../test-catalog/selector-resolution.mjs";
export { loadCatalogMetadata } from "../test-catalog/test-catalog.mjs";
export { validatePostgresFixturePolicy } from "../test-catalog/postgres-fixture-policy.mjs";
