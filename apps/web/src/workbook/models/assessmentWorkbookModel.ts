import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { enumValuesFor } from "./genericWorkbookModel";

export type AssessmentSubjectType = "host" | "identity";
export type AssessmentConfidenceBand = "unset" | "low" | "medium" | "high";

export type AssessmentCreateDraft = {
  assessedAt: string;
  assessmentState: string;
  confidenceBand: AssessmentConfidenceBand;
  rationale: string;
  subjectRecordId: string;
  subjectType: AssessmentSubjectType;
  supportRecordIds: string[];
};

export type AssessmentSupportCandidate = {
  readonly displayText: string;
  readonly recordId: string;
};

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

export function isAssessmentSubjectType(
  value: string,
): value is AssessmentSubjectType {
  return value === "host" || value === "identity";
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
  seed?: {
    readonly subjectRecordId: string;
    readonly subjectType: AssessmentSubjectType;
  },
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
    subjectRecordId: seed?.subjectRecordId ?? "",
    subjectType: seed?.subjectType ?? "host",
    supportRecordIds: [],
  };
}

export function assessmentSupportCandidate(
  recordId: string,
  displayText: unknown,
): AssessmentSupportCandidate {
  return {
    recordId,
    displayText:
      typeof displayText === "string" && displayText !== ""
        ? displayText
        : recordId,
  };
}

export function followOnAssessmentDraft(
  assessmentsContract: ViewContract,
  selectedRow: WorkbookQueryRow,
): AssessmentCreateDraft | null {
  const subjectRecordId = normalizedAssessmentValue(
    selectedRow.cells["assessment.subject_ref"]?.value,
  );
  const subjectTypeValue = normalizedAssessmentValue(
    selectedRow.cells["assessment.subject_type"]?.value,
  );
  if (subjectRecordId === "" || !isAssessmentSubjectType(subjectTypeValue)) {
    return null;
  }
  return initialAssessmentDraft(assessmentsContract, {
    subjectRecordId,
    subjectType: subjectTypeValue,
  });
}

export function confidenceScoreFromBand(
  band: AssessmentConfidenceBand,
): number | null {
  switch (band) {
    case "low":
      return 25;
    case "medium":
      return 55;
    case "high":
      return 85;
    default:
      return null;
  }
}

export function buildAssessmentCreatePayload(
  draft: AssessmentCreateDraft,
  clientTxnId: string,
): (Record<string, unknown> & { readonly client_txn_id: string }) | null {
  const subjectRecordId = normalizedAssessmentValue(draft.subjectRecordId);
  const assessmentState = normalizedAssessmentValue(draft.assessmentState);
  const rationale = normalizedAssessmentValue(draft.rationale);
  if (subjectRecordId === "" || assessmentState === "" || rationale === "") {
    return null;
  }

  const payload: Record<string, unknown> & { readonly client_txn_id: string } =
    {
      client_txn_id: clientTxnId,
      "assessment.subject_ref": subjectRecordId,
      "assessment.subject_type": draft.subjectType,
      "assessment.assessment_state": assessmentState,
      "assessment.confidence_score": confidenceScoreFromBand(
        draft.confidenceBand,
      ),
      "assessment.rationale": rationale,
    };

  const assessedAt = normalizedAssessmentValue(draft.assessedAt);
  if (assessedAt !== "") {
    payload["assessment.assessed_at"] = assessedAt;
  }

  const supportRecordIds = Array.from(
    new Set(
      draft.supportRecordIds
        .map((recordId) => normalizedAssessmentValue(recordId))
        .filter((recordId) => recordId !== ""),
    ),
  );
  if (supportRecordIds.length > 0) {
    payload["assessment.support_refs"] = {
      kind: "collection_actions_v1",
      actions: supportRecordIds.map((recordId) => ({
        op: "add_record_ref",
        linked_record_id: recordId,
      })),
    };
  }

  return payload;
}

function normalizedAssessmentValue(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}
