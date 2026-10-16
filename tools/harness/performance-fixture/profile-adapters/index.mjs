import { validateEvidenceVerificationObservation } from "./evidence-verification.mjs";
import { validateEvidenceInspectorObservation } from "./evidence-inspector.mjs";
import { validateInteractiveMeasurementObservation } from "./interactive-large-grid.mjs";
import { validateTimelinePasteObservation } from "./timeline-paste.mjs";
import { validatePresenceObservation } from "./presence.mjs";
import { validateAnchorTraceObservation } from "./anchor-trace.mjs";
import { validateHistoryObservation } from "./history.mjs";

import { validateImportProgressObservation } from "./import-progress.mjs";

import { validateReportingProgressObservation } from "./reporting-progress.mjs";

import { validateReferencePackProgressObservation } from "./reference-pack-progress.mjs";

const observationValidators = new Map([
  ["timeline_large_grid_reporting_snapshot_v1", validateReportingProgressObservation],
  ["timeline_large_grid_import_snapshot_v1", validateImportProgressObservation],
  ["evidence_heavy_snapshot_v2", (root, profile, observation) => {
    if (observation.predicate_id === "perf.evidence_verification.visible_with_cancel.v1") validateEvidenceVerificationObservation(root, profile, observation);
    else validateEvidenceInspectorObservation(root, profile, observation);
  }],
  ["timeline_large_grid_snapshot_v2", (root, profile, observation) => {
    if (observation.criterion_id === "AC-033") {
      validateReferencePackProgressObservation(root, profile, observation);
    } else if (observation.criterion_id === "AC-047") {
      validateAnchorTraceObservation(root, profile, observation);
    } else if (observation.predicate_id === "perf.timeline_paste_20x5.v2") {
      validateTimelinePasteObservation(root, profile, observation);
    } else if (observation.criterion_id === "AC-008") {
      validatePresenceObservation(root, profile, observation);
    } else if (observation.criterion_id === "AC-011") {
      validateHistoryObservation(root, profile, observation);
    } else if (observation.criterion_id === "AC-044") {
      validateInteractiveMeasurementObservation(root, profile, observation, "contracts/performance/ac044.v1.json", "cartulary.performance_view_change_contract.v1");
    } else {
      validateInteractiveMeasurementObservation(root, profile, observation, "contracts/performance/ac043.v2.json", "cartulary.performance_interaction_contract.v2");
    }
  }],
]);

export function validateProfileMeasurementObservation(root, profile, observation) {
  const validate = observationValidators.get(profile.fixture_profile_id);
  if (validate === undefined) {
    throw new Error(
      `performance fixture profile lacks a measurement adapter: ${profile.fixture_profile_id}`,
    );
  }
  validate(root, profile, observation);
}
