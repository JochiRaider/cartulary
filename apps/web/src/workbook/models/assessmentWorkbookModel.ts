import {
  assessmentsViewSchemaId,
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import type { WorkbookProtocolCreateViewRowRequest } from "../adapters/workbookProtocolTypes";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import {
  buildGenericPatchChange,
  enumValuesFor,
  workbookCreateMinimumSatisfied,
  workbookCreationAvailable,
} from "./genericWorkbookModel";
import { decodeCreateViewRowRequest } from "./workbookRequestDecoders";

export type AssessmentSubjectType = "host" | "identity";
export type AssessmentConfidenceBand = "unset" | "low" | "medium" | "high";
export type AssessmentCreateRequest = Extract<
  WorkbookProtocolCreateViewRowRequest,
  { readonly "assessment.subject_ref": string }
>;
type AssessmentState = AssessmentCreateRequest["assessment.assessment_state"];

export type AssessmentCreateDraft = {
  assessedAt: string;
  assessmentState: string;
  confidenceBand: AssessmentConfidenceBand;
  rationale: string;
  subjectRecordId: string;
  subjectType: AssessmentSubjectType;
  supportRecordIds: string[];
  subjectDisplayText?: string;
  supportDisplayText?: Readonly<Record<string, string>>;
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

function isAssessmentSubjectType(
  value: string,
): value is AssessmentSubjectType {
  return value === "host" || value === "identity";
}

function isAssessmentState(value: string): value is AssessmentState {
  return (
    value === "unknown" ||
    value === "suspected" ||
    value === "confirmed" ||
    value === "disproven" ||
    value === "cleared"
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
    case "unset":
      return null;
  }
}

export function buildAssessmentCreatePayload(
  draft: AssessmentCreateDraft,
  clientTxnId: string,
  contract: ViewContract = requireViewContract(assessmentsViewSchemaId),
): AssessmentCreateRequest | null {
  if (Object.keys(assessmentCreateErrors(draft, contract)).length > 0)
    return null;
  const subjectRecordId = normalizedAssessmentValue(draft.subjectRecordId);
  const assessmentState = normalizedAssessmentValue(draft.assessmentState);
  const rationale = normalizedAssessmentValue(draft.rationale);
  if (
    subjectRecordId === "" ||
    !isAssessmentState(assessmentState) ||
    rationale === ""
  ) {
    return null;
  }

  const assessedAt = normalizedAssessmentValue(draft.assessedAt);
  const supportRecordIds = Array.from(
    new Set(
      draft.supportRecordIds
        .map((recordId) => normalizedAssessmentValue(recordId))
        .filter((recordId) => recordId !== ""),
    ),
  );
  const [firstSupportRecordId, ...remainingSupportRecordIds] = supportRecordIds;
  const request: AssessmentCreateRequest = {
    ...(assessedAt === "" ? {} : { "assessment.assessed_at": assessedAt }),
    "assessment.assessment_state": assessmentState,
    "assessment.confidence_score": confidenceScoreFromBand(
      draft.confidenceBand,
    ),
    "assessment.rationale": rationale,
    "assessment.subject_ref": subjectRecordId,
    "assessment.subject_type": draft.subjectType,
    ...(firstSupportRecordId === undefined
      ? {}
      : {
          "assessment.support_refs": {
            actions: [
              { linked_record_id: firstSupportRecordId, op: "add_record_ref" },
              ...remainingSupportRecordIds.map((recordId) => ({
                linked_record_id: recordId,
                op: "add_record_ref" as const,
              })),
            ],
            kind: "collection_actions_v1" as const,
          },
        }),
    client_txn_id: clientTxnId,
  };
  // Confidence null is an Assessment create value, independent of existing-row clearability.
  const { "assessment.confidence_score": score, ...withoutScore } = request;
  return decodeCreateViewRowRequest(
    contract,
    score === null ? withoutScore : request,
  ) === null
    ? null
    : request;
}

export function assessmentCreateErrors(
  draft: AssessmentCreateDraft,
  contract: ViewContract = requireViewContract(assessmentsViewSchemaId),
): Readonly<Record<string, string>> {
  const errors: Record<string, string> = {};
  if (!workbookCreationAvailable(contract))
    errors.form = "Assessment creation is unavailable.";
  if (!draft.subjectRecordId.trim()) errors.subject = "Select a subject.";
  if (!isAssessmentSubjectType(draft.subjectType))
    errors.subjectType = "Select a Host or Identity.";
  if (!isAssessmentState(draft.assessmentState.trim()))
    errors.assessmentState = "Select an assessment state.";
  if (!draft.rationale.trim()) errors.rationale = "Enter a rationale.";
  if (!isAssessmentConfidenceBand(draft.confidenceBand))
    errors.confidenceBand = "Select a confidence band.";
  const field = contract.fieldMap["assessment.assessed_at"];
  if (
    draft.assessedAt.trim() &&
    (!field || !buildGenericPatchChange(field, draft.assessedAt))
  )
    errors.assessedAt =
      "Enter an RFC3339 timestamp with a time zone, or leave it blank for the commit time.";
  if (new Set(draft.supportRecordIds).size > 64)
    errors.supportRecordIds = "Choose at most 64 supporting records.";
  if (draft.supportRecordIds.some((id) => !id.trim()))
    errors.supportRecordIds = "Remove the invalid support selection.";
  const values = {
    "assessment.subject_ref": draft.subjectRecordId,
    "assessment.subject_type": draft.subjectType,
    "assessment.assessment_state": draft.assessmentState,
    "assessment.rationale": draft.rationale,
  };
  if (
    !workbookCreateMinimumSatisfied(contract, values) &&
    !Object.keys(errors).length
  )
    errors.form = "Complete the required assessment fields.";
  for (const key of [
    ...Object.keys(values),
    "assessment.confidence_score",
    ...(draft.assessedAt.trim() ? ["assessment.assessed_at"] : []),
    ...(draft.supportRecordIds.length ? ["assessment.support_refs"] : []),
  ]) {
    if (!contract.fieldMap[key]?.createWritable)
      errors.form = "The current Assessment create contract is unavailable.";
  }
  return errors;
}

function normalizedAssessmentValue(value: unknown): string {
  return typeof value === "string" ? value.trim() : "";
}
