import type { ErrorEnvelope } from "@cartulary/protocol-ts/http";
import { publicErrorView } from "../../services/browserApi";
import { validatedPublicErrorReason } from "../../services/publicErrorIdentity";
import {
  type PublicErrorOperationFamily,
  resolvePublicErrorPresentation,
} from "../../shared/publicErrorPresentation";
import type {
  WorkbookOperationFailure,
  WorkbookOperationFieldFailure,
} from "../mutations/workbookOperationOutcome";
import { parseSameFieldConflict } from "../runtime/workbookConflictModel";
import type { WorkbookOperationID } from "./workbookOperationContract";
import { decodeWorkbookPublicError } from "./workbookPublicErrorDecoder";

type FailureKind = Exclude<
  WorkbookOperationFailure["kind"],
  "invalid_contract" | "same_field_conflict" | "validation"
>;

const operationFamilyByID = {
  appendIndicatorStateInterval: "field_mutation",
  applyWorkbookBulkMutation: "field_mutation",
  attachBlobToEvidenceRecord: "field_mutation",
  createIncidentSavedView: "field_mutation",
  createManualIndicatorObservation: "field_mutation",
  createObjectBlobSlot: "field_mutation",
  createRecordLinkedNote: "field_mutation",
  createViewRow: "field_mutation",
  deleteIncidentSavedView: "field_mutation",
  deleteRecord: "field_mutation",
  dismissIndicatorObservation: "field_mutation",
  getCurrentUserWorkbookPreferences: "field_mutation",
  getIncident: "field_mutation",
  getIncidentDefaultWorkbookPreferences: "field_mutation",
  getIncidentWorkbookStartup: "surface_load",
  getRecordHistory: "surface_load",
  getTimelineTimeConversionProfile: "field_mutation",
  issueEvidenceDownloadHandle: "evidence_download",
  issueEvidencePreviewHandle: "evidence_preview",
  listIncidentMemberships: "field_mutation",
  listIncidentSavedViews: "field_mutation",
  listIndicatorObservations: "field_mutation",
  listIndicatorStateIntervals: "field_mutation",
  listSourceRecordIndicatorObservations: "field_mutation",
  markTimelineRecordReviewed: "field_mutation",
  mergeEntityRecord: "field_mutation",
  pasteWorkbookClipboard: "field_mutation",
  patchIncidentSavedView: "field_mutation",
  patchRecord: "field_mutation",
  putCurrentUserWorkbookPreferences: "field_mutation",
  putIncidentDefaultWorkbookPreferences: "field_mutation",
  putTimelineTimeConversionProfile: "field_mutation",
  queryWorkbookView: "surface_load",
  resolveEntityMention: "field_mutation",
  resolveIndicatorObservation: "field_mutation",
  resolveRecordSameFieldConflict: "field_mutation",
  restoreIndicatorObservation: "field_mutation",
  restoreRecord: "field_mutation",
  rollbackRecord: "field_mutation",
  supersedeRecord: "field_mutation",
} as const satisfies Record<WorkbookOperationID, PublicErrorOperationFamily>;

const exactFailureKindByCode: Readonly<Record<string, FailureKind>> = {
  authorization_denied: "authorization_lost",
  client_txn_conflict: "client_txn_conflict",
  illegal_transition: "stale_target",
  indicator_not_found: "stale_target",
  indicator_observation_not_found: "stale_target",
  indicator_source_record_not_found: "stale_target",
  incident_not_found: "stale_target",
  record_already_deleted: "stale_target",
  record_deleted_use_restore: "stale_target",
  record_not_deleted: "stale_target",
  record_not_found: "stale_target",
  resolved_indicator_not_found: "stale_target",
  row_version_conflict: "stale_target",
};

function safeDetail(value: unknown): string | null {
  if (
    typeof value !== "string" &&
    typeof value !== "number" &&
    typeof value !== "boolean"
  ) {
    return null;
  }
  const text = String(value).trim();
  return text !== "" && text.length <= 120 ? text : null;
}

function validationFields(
  details: Readonly<Record<string, unknown>>,
): readonly WorkbookOperationFieldFailure[] | undefined {
  const field = safeDetail(details.field);
  if (field === null) return undefined;
  return [
    {
      field,
      message: safeDetail(details.reason_code) ?? "Invalid value.",
    },
  ];
}

const mergePreconditionDetailKeys = [
  "reason_code",
  "record_type",
  "identifier_class",
  "normalized_value",
  "blocking_record_id",
  "survivor_record_id",
  "loser_record_id",
  "survivor_base_row_version",
  "loser_base_row_version",
  "survivor_current_row_version",
  "loser_current_row_version",
] as const;

function mergePreconditionFailure(
  details: Readonly<Record<string, unknown>>,
  message: string,
): WorkbookOperationFailure {
  const fields = mergePreconditionDetailKeys.flatMap((field) => {
    const value = safeDetail(details[field]);
    return value === null ? [] : [{ field, message: value }];
  });
  const reason = safeDetail(details.reason_code);
  return {
    kind: "validation",
    message: reason === null ? message : `${message}: ${reason}`,
    ...(fields.length === 0 ? {} : { fields }),
  };
}

function conflictFailure(
  envelope: ErrorEnvelope,
  message: string,
): WorkbookOperationFailure {
  const conflict = parseSameFieldConflict(envelope);
  return conflict === null
    ? {
        kind: "invalid_contract",
        message: "The server returned an invalid conflict response.",
      }
    : { kind: "same_field_conflict", message, conflict };
}

function retryableFailure(
  declaredRetryable: boolean,
  status: number,
  message: string,
): WorkbookOperationFailure {
  return declaredRetryable || [429, 502, 503, 504].includes(status)
    ? { kind: "retryable", message }
    : { kind: "terminal", message };
}

function statusFailure(
  envelope: ErrorEnvelope,
  message: string,
  status: number,
): WorkbookOperationFailure {
  const error = envelope.error;
  if (status === 401) return { kind: "authentication_required", message };
  const exactKind = exactFailureKindByCode[error.code];
  if (exactKind !== undefined) return { kind: exactKind, message };
  if (status === 403) return { kind: "authorization_lost", message };
  if (status === 400 || status === 422) {
    const fields = validationFields(error.details);
    return fields === undefined
      ? { kind: "validation", message }
      : { kind: "validation", message, fields };
  }
  return retryableFailure(error.retryable, status, message);
}

function publicMessage(
  envelope: ErrorEnvelope,
  operationID: WorkbookOperationID,
  status: number,
): string {
  const error = envelope.error;
  if (
    operationID === "issueEvidencePreviewHandle" ||
    operationID === "issueEvidenceDownloadHandle" ||
    operationID === "createObjectBlobSlot" ||
    operationID === "attachBlobToEvidenceRecord"
  ) {
    return "Evidence request failed.";
  }
  return error.code === "invalid_mutation_payload"
    ? error.code
    : (publicErrorView(error, status)?.statusText ?? "Request failed.");
}

function classifyDecodedError(
  envelope: ErrorEnvelope,
  operationID: WorkbookOperationID,
  status: number,
): WorkbookOperationFailure {
  const error = envelope.error;
  const message = publicMessage(envelope, operationID, status);
  if (error.code === "same_field_conflict") {
    return conflictFailure(envelope, message);
  }
  if (
    operationID === "mergeEntityRecord" &&
    error.code === "merge_precondition_failed"
  ) {
    return mergePreconditionFailure(error.details, message);
  }
  return statusFailure(envelope, message, status);
}

export function classifyWorkbookOperationFailure(
  status: number,
  payload: unknown,
  operationID: WorkbookOperationID,
): WorkbookOperationFailure {
  const decoded = decodeWorkbookPublicError(payload);
  const code =
    decoded.kind === "decoded"
      ? decoded.envelope.error.code
      : "invalid_public_contract_response";
  const classified =
    decoded.kind === "decoded"
      ? classifyDecodedError(decoded.envelope, operationID, status)
      : { kind: decoded.kind, message: decoded.message };
  const publicReason =
    decoded.kind === "decoded"
      ? validatedPublicErrorReason(
          code,
          decoded.envelope.error.details.reason_code,
        )
      : undefined;
  return {
    ...classified,
    ...(publicReason === undefined ? {} : { publicReason }),
    ...(decoded.kind === "decoded" ? { publicCode: code } : {}),
    presentation: resolvePublicErrorPresentation({
      code,
      hasAuthorizedMaterialization: false,
      operationFamily: operationFamilyByID[operationID],
      reasonCode: publicReason,
      status,
    }),
  };
}
