// Public facade shared by selected-row and full frontend execution.
export {
  diagnosticFailure,
  failureDetailsSchemaID,
  indexVitestObservations,
  normalizeVitestObservations,
  readVitestJSON,
  reconcileVitestReport,
  selectedVitestObservations,
  vitestObservationKey,
  vitestFileKey,
  vitestReportProjection,
} from "./vitest-observations.mjs";
export { createVitestEvidence, readVitestEvidence, validateVitestRequest } from "./vitest-evidence.mjs";
