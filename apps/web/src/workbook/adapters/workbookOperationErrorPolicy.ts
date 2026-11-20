import type { HTTPOperationErrorCode } from "@cartulary/protocol-ts/http";
import {
  type HTTPOperationError,
  type HTTPOperationPublicError,
  publicErrorView,
} from "../../services/browserApi";
import {
  type PublicErrorReasonCode,
  validatedPublicErrorReason,
} from "../../services/publicErrorIdentity";
import {
  isTemporaryUnavailability,
  isUnappliedRefusal,
  temporaryUnavailabilityText,
} from "../../shared/publicError";
import {
  type PublicErrorOperationFamily,
  resolvePublicErrorPresentation,
} from "../../shared/publicErrorPresentation";
import type {
  WorkbookOperationFailure,
  WorkbookOperationFieldFailure,
} from "../mutations/workbookOperationOutcome";
import { parseSameFieldConflictPayload } from "../runtime/workbookConflictModel";
import type { WorkbookOperationID } from "./workbookOperationContract";

/** A decoded public error of some Workbook operation (never the contract failure). */
type DecodedWorkbookError = HTTPOperationPublicError<WorkbookOperationID>;

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
  verifyObjectBlob: "field_mutation",
  getJob: "field_mutation",
  cancelJob: "field_mutation",
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
  getViewSchema: "surface_load",
  issueEvidenceDownloadHandle: "evidence_download",
  issueEvidencePreviewHandle: "evidence_preview",
  listEntityCandidates: "surface_load",
  listIncidentMemberships: "field_mutation",
  listIncidentSavedViews: "field_mutation",
  listIndicatorObservations: "field_mutation",
  listIndicatorStateIntervals: "field_mutation",
  listNoteAssociations: "surface_load",
  listSourceRecordIndicatorObservations: "field_mutation",
  listViewSchemas: "surface_load",
  markTimelineRecordReviewed: "field_mutation",
  mergeEntityRecord: "field_mutation",
  mutateNoteAssociations: "field_mutation",
  pasteWorkbookClipboard: "field_mutation",
  patchIncidentSavedView: "field_mutation",
  patchRecord: "field_mutation",
  putCurrentUserWorkbookPreferences: "field_mutation",
  putIncidentDefaultWorkbookPreferences: "field_mutation",
  putTimelineTimeConversionProfile: "field_mutation",
  queryEvidenceMetadata: "surface_load",
  queryWorkbookView: "surface_load",
  locateWorkbookViewRecord: "surface_load",
  resolveEntityMention: "field_mutation",
  resolveIndicatorObservation: "field_mutation",
  resolveRecordSameFieldConflict: "field_mutation",
  restoreIndicatorObservation: "field_mutation",
  restoreRecord: "field_mutation",
  rollbackRecord: "field_mutation",
  supersedeRecord: "field_mutation",
} as const satisfies Record<WorkbookOperationID, PublicErrorOperationFamily>;

/** Keys must be codes some Workbook operation declares. */
const exactFailureKindByCode: Readonly<
  Partial<Record<HTTPOperationErrorCode<WorkbookOperationID>, FailureKind>>
> = {
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
    // The closed merge error contract permits complete normalized collision
    // values. Preserve long values for review; React renders these as text.
    const value =
      field === "normalized_value" && typeof details[field] === "string"
        ? details[field]
        : safeDetail(details[field]);
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
  error: DecodedWorkbookError,
  message: string,
): WorkbookOperationFailure {
  const conflict = parseSameFieldConflictPayload(error.conflict);
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
  error: DecodedWorkbookError,
  message: string,
): WorkbookOperationFailure {
  const status = error.status;
  if (status === 401) return { kind: "authentication_required", message };
  const exactKind = exactFailureKindByCode[error.code];
  if (exactKind !== undefined) return { kind: exactKind, message };
  if (status === 403) return { kind: "authorization_lost", message };
  if (status === 400) {
    const fields = validationFields(error.details);
    return fields === undefined
      ? { kind: "validation", message }
      : { kind: "validation", message, fields };
  }
  return retryableFailure(error.retryable, status, message);
}

function publicMessage(
  error: DecodedWorkbookError,
  operationID: WorkbookOperationID,
): string {
  if (
    operationID === "issueEvidencePreviewHandle" ||
    operationID === "issueEvidenceDownloadHandle" ||
    operationID === "createObjectBlobSlot" ||
    operationID === "verifyObjectBlob" ||
    operationID === "attachBlobToEvidenceRecord"
  ) {
    return "Evidence request failed.";
  }
  return error.code === "invalid_mutation_payload"
    ? error.code
    : (publicErrorView(error)?.statusText ?? "Request failed.");
}

function classifyDecodedError(
  error: DecodedWorkbookError,
  operationID: WorkbookOperationID,
): WorkbookOperationFailure {
  // Temporary unavailability precedes every operation-specific rule: nothing
  // was applied and the same request remains retryable (Core 03 REQ-03-086a).
  if (isTemporaryUnavailability(error))
    return { kind: "retryable", message: temporaryUnavailabilityText() };
  const status = error.status;
  const message = publicMessage(error, operationID);
  if (
    operationID === "issueEvidencePreviewHandle" ||
    operationID === "issueEvidenceDownloadHandle"
  ) {
    const reason = validatedPublicErrorReason(
      error.code,
      error.details.reason_code,
    );
    const invalid: WorkbookOperationFailure = {
      kind: "invalid_contract",
      message: "The server returned an invalid Evidence access response.",
    };
    // Status and code are already decoded against the operation contract;
    // retryability and the reason remain owner-projected identity checks.
    if (error.code === "object_store_unavailable")
      return error.retryable && reason
        ? { kind: "retryable", message }
        : invalid;
    if (error.code === "object_store_access_rejected")
      return !error.retryable && reason
        ? { kind: "terminal", message }
        : invalid;
    if (error.code === "evidence_access_unavailable")
      return !error.retryable &&
        reason &&
        (operationID === "issueEvidencePreviewHandle" ||
          (reason !== "unsupported_preview" &&
            reason !== "preview_payload_too_large"))
        ? { kind: "terminal", message }
        : invalid;
    if (status >= 500 && error.code !== "session_required")
      return { kind: "terminal", message };
  }
  if (error.code === "same_field_conflict") {
    return conflictFailure(error, message);
  }
  if (
    operationID === "mergeEntityRecord" &&
    error.code === "merge_precondition_failed"
  ) {
    return mergePreconditionFailure(error.details, message);
  }
  return statusFailure(error, message);
}

/**
 * Classifies one failure already decoded by `fetchHTTPOperation` against the
 * operation contract (Core 03 REQ-03-086a). The client contract failure never
 * selects a feature outcome and carries no public code.
 */
export function classifyWorkbookOperationFailure<
  OperationID extends WorkbookOperationID,
>(
  failure: HTTPOperationError<OperationID>,
  operationID: OperationID,
): WorkbookOperationFailure<OperationID> {
  const error = failure as HTTPOperationError<WorkbookOperationID>;
  const presentation = (publicReason?: PublicErrorReasonCode) =>
    resolvePublicErrorPresentation({
      code: error.code,
      hasAuthorizedMaterialization: false,
      operationFamily: operationFamilyByID[operationID],
      reasonCode: publicReason,
      status: error.status,
    });
  if (error.code === "invalid_public_contract_response")
    return {
      kind: "invalid_contract",
      message: "The server returned an invalid public contract response.",
      presentation: presentation(),
    };
  const classified = classifyDecodedError(error, operationID);
  const publicReason = validatedPublicErrorReason(
    error.code,
    error.details.reason_code,
  );
  const invalidEvidenceHandle =
    classified.kind === "invalid_contract" &&
    (operationID === "issueEvidencePreviewHandle" ||
      operationID === "issueEvidenceDownloadHandle");
  return {
    ...classified,
    ...(invalidEvidenceHandle || publicReason === undefined
      ? {}
      : { publicReason }),
    ...(invalidEvidenceHandle
      ? {}
      : {
          publicCode:
            error.code as WorkbookOperationFailure<OperationID>["publicCode"],
        }),
    presentation: presentation(publicReason),
  } as WorkbookOperationFailure<OperationID>;
}

/**
 * Whether a failed explicit Workbook write leaves its outcome unknown. A
 * response the client could not decode, an unexpected status or any server
 * fault stays uncertain and keeps exact replay. A decoded `503`
 * `service_unavailable` was not applied, so it is a definite rejection whose
 * retryable failure keeps the draft (Core 03 REQ-03-086a).
 */
export function explicitWorkbookWriteUncertain(
  status: number | null,
  failure: Pick<WorkbookOperationFailure, "kind" | "publicCode">,
): boolean {
  return (
    status === null ||
    status < 400 ||
    failure.kind === "invalid_contract" ||
    (status >= 500 &&
      !isUnappliedRefusal(status, { code: failure.publicCode, status }))
  );
}

/**
 * Settles one failed explicit Workbook write from its decoded error. A
 * response whose request identity differs from the observed one is
 * uncertain; otherwise `explicitWorkbookWriteUncertain` decides.
 */
export function explicitWorkbookFailureOutcome<
  OperationID extends WorkbookOperationID,
>(
  status: number | null,
  error: HTTPOperationError<OperationID> | null | undefined,
  operationID: OperationID,
  requestId: string | null = null,
):
  | { readonly kind: "uncertain" }
  | {
      readonly kind: "rejected";
      readonly failure: WorkbookOperationFailure<OperationID>;
    } {
  if (!error?.code || (requestId !== null && requestId !== error.request_id))
    return { kind: "uncertain" };
  const failure = classifyWorkbookOperationFailure(error, operationID);
  return explicitWorkbookWriteUncertain(status, failure)
    ? { kind: "uncertain" }
    : { kind: "rejected", failure };
}
