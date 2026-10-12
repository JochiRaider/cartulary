export { WorkGraphCompiler } from "./compiler.mjs";
export {
  browserStages,
  browserTargetStage,
  compileBrowserStageGraph,
} from "./browser.mjs";
export {
  buildWorkGraph,
  graphSemanticDigest,
  loadWorkGraphOwner,
  unitSemanticDigest,
  validateWorkGraph,
} from "./model.mjs";
export {
  assertFixtureServiceDependencies,
  assertServiceDependencies,
  requiredServicesForFixture,
  topologyResourceClaims,
} from "./resource-claims.mjs";
export {
  captureCapabilitySnapshot,
  cpuCapacityWithSafetyMargin,
  resourceCapacities,
} from "./capability.mjs";
export { loadCacheRegistry, CachePlan } from "./cache-plan.mjs";
export { resolveCacheDependencyClosure } from "./cache-dependencies.mjs";
export {
  assertScannerEvidenceParity,
  resolveVulnerabilityDatabaseRevision,
  scannerEvidenceFingerprint,
} from "./vulnerability.mjs";
