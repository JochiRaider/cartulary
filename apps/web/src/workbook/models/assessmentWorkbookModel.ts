import type { ViewContract } from "@cartulary/view-contracts";
import type {
  AssessmentConfidenceBand,
  AssessmentCreateDraft,
  TimelineApiRow,
} from "../timeline/models/workbookTimelineModel";
import { enumValuesFor } from "./genericWorkbookModel";

export function isAssessmentConfidenceBand(
  value: string,
): value is AssessmentConfidenceBand {
  return (
    value === "unset" ||
    value === "low" ||
    value === "medium" ||
    value === "high"
  );
}

export function assessmentColumnWidth(fieldKey: string): number {
  switch (fieldKey) {
    case "assessment.subject_ref":
      return 300;
    case "assessment.rationale":
      return 360;
    case "assessment.assessed_at":
      return 210;
    case "assessment.assessor":
      return 300;
    default:
      return 180;
  }
}

export function initialAssessmentDraft(
  assessmentsContract: ViewContract,
): AssessmentCreateDraft {
  const [assessmentState = "unknown"] = enumValuesFor(
    assessmentsContract,
    "assessment.assessment_state",
    ["unknown", "suspected", "confirmed", "disproven", "cleared"],
  );
  const confidenceBand = enumValuesFor(
    assessmentsContract,
    "assessment.confidence_band",
    ["unset", "low", "medium", "high"],
  ).find(isAssessmentConfidenceBand);
  return {
    assessedAt: "",
    assessmentState,
    confidenceBand: confidenceBand ?? "unset",
    rationale: "",
    subjectRecordId: "",
    subjectType: "host",
    supportRecordIds: [],
  };
}

export function supportRowLabel(row: TimelineApiRow): string {
  const summary = row.cells["timeline.summary"]?.value;
  return typeof summary === "string" && summary !== ""
    ? summary
    : row.record_id;
}
