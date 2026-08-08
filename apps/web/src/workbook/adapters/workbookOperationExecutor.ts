import {
  errorEnvelopeDecoder,
  type HTTPOperationID,
  type HTTPOperationRequest,
  type HTTPOperationResponse,
  type HTTPPathParameters,
  type HTTPQueryValue,
  httpOperationBindings,
} from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation, publicErrorView } from "../../services/browserApi";
import { resolvePublicErrorPresentation } from "../../shared/publicErrorPresentation";
import type {
  WorkbookOperationFailure,
  WorkbookOperationFieldFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";
import { parseSameFieldConflict } from "../runtime/workbookConflictModel";

const workbookOperationIDs = [
  "appendIndicatorStateInterval",
  "applyWorkbookBulkMutation",
  "attachBlobToEvidenceRecord",
  "createIncidentSavedView",
  "createManualIndicatorObservation",
  "createObjectBlobSlot",
  "createRecordLinkedNote",
  "createViewRow",
  "deleteIncidentSavedView",
  "deleteRecord",
  "dismissIndicatorObservation",
  "getCurrentUserWorkbookPreferences",
  "getIncident",
  "getIncidentDefaultWorkbookPreferences",
  "getIncidentWorkbookStartup",
  "getRecordHistory",
  "getTimelineTimeConversionProfile",
  "issueEvidenceDownloadHandle",
  "issueEvidencePreviewHandle",
  "listIncidentMemberships",
  "listIncidentSavedViews",
  "listIndicatorObservations",
  "listIndicatorStateIntervals",
  "listSourceRecordIndicatorObservations",
  "markTimelineRecordReviewed",
  "mergeEntityRecord",
  "pasteWorkbookClipboard",
  "patchIncidentSavedView",
  "patchRecord",
  "putCurrentUserWorkbookPreferences",
  "putIncidentDefaultWorkbookPreferences",
  "putTimelineTimeConversionProfile",
  "queryWorkbookView",
  "resolveEntityMention",
  "resolveIndicatorObservation",
  "resolveRecordSameFieldConflict",
  "restoreIndicatorObservation",
  "restoreRecord",
  "rollbackRecord",
  "supersedeRecord",
] as const satisfies readonly HTTPOperationID[];

export type WorkbookOperationID = (typeof workbookOperationIDs)[number];
export type WorkbookOperationResponse<OperationID extends WorkbookOperationID> =
  HTTPOperationResponse<OperationID>;

type WorkbookOperationRequestInput<OperationID extends WorkbookOperationID> =
  HTTPOperationRequest<OperationID> extends undefined
    ? { readonly request?: undefined }
    : { readonly request: HTTPOperationRequest<OperationID> };

export type WorkbookOperationExecution<
  OperationID extends WorkbookOperationID,
> = {
  readonly observeTransport?: {
    readonly onJSONParsed?: () => void;
    readonly onResponseStatus?: (status: number) => void;
  };
  readonly operationID: OperationID;
  readonly pathParameters?: HTTPPathParameters;
  readonly query?: Readonly<Record<string, HTTPQueryValue>>;
  readonly signal?: AbortSignal;
} & WorkbookOperationRequestInput<OperationID>;

export interface WorkbookOperationExecutor {
  execute<OperationID extends WorkbookOperationID>(
    input: WorkbookOperationExecution<OperationID>,
  ): Promise<WorkbookOperationOutcome<HTTPOperationResponse<OperationID>>>;
}

const staleTargetCodes = new Set([
  "illegal_transition",
  "indicator_not_found",
  "indicator_observation_not_found",
  "indicator_source_record_not_found",
  "incident_not_found",
  "record_already_deleted",
  "record_deleted_use_restore",
  "record_not_deleted",
  "record_not_found",
  "resolved_indicator_not_found",
  "row_version_conflict",
]);

const evidenceAccessReasonCodes = new Set([
  "blob_failed",
  "blob_missing",
  "blob_pending",
  "evidence_inconsistent",
  "evidence_quarantined",
  "no_visible_blob",
  "preview_payload_too_large",
  "unsupported_preview",
]);

function isInternalInvalidContractFailure(value: unknown): boolean {
  if (!value || typeof value !== "object" || !("error" in value)) {
    return false;
  }
  const error = value.error;
  if (!error || typeof error !== "object" || !("code" in error)) {
    return false;
  }
  return error.code === "invalid_public_contract_response";
}

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
  const reason = safeDetail(details.reason_code) ?? "Invalid value.";
  return [{ field, message: reason }];
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

function mergePreconditionFields(
  details: Readonly<Record<string, unknown>>,
): readonly WorkbookOperationFieldFailure[] {
  return mergePreconditionDetailKeys.flatMap((field) => {
    const value = safeDetail(details[field]);
    return value === null ? [] : [{ field, message: value }];
  });
}

function evidenceFailureMessage(
  operationID: WorkbookOperationID,
  code: string,
  details: Readonly<Record<string, unknown>>,
  fallback: string,
): string {
  if (
    code !== "evidence_access_unavailable" ||
    (operationID !== "issueEvidencePreviewHandle" &&
      operationID !== "issueEvidenceDownloadHandle")
  ) {
    return fallback;
  }
  const reason = safeDetail(details.reason_code);
  return reason !== null && evidenceAccessReasonCodes.has(reason)
    ? `${code}: ${reason}`
    : fallback;
}

function operationFailureWithoutPresentation(
  status: number,
  payload: unknown,
  operationID: WorkbookOperationID,
): WorkbookOperationFailure {
  if (isInternalInvalidContractFailure(payload)) {
    return {
      kind: "invalid_contract",
      message: "The server returned an invalid public contract response.",
    };
  }
  const decoded = errorEnvelopeDecoder.decode(payload);
  if (!decoded.ok) {
    return {
      kind: "invalid_contract",
      message: "The server returned an invalid public error response.",
    };
  }
  const error = decoded.value.error;
  const publicMessage =
    error.code === "invalid_mutation_payload"
      ? error.code
      : (publicErrorView(error, status)?.statusText ?? "Request failed.");
  const message = evidenceFailureMessage(
    operationID,
    error.code,
    error.details,
    publicMessage,
  );
  if (error.code === "same_field_conflict") {
    const conflict = parseSameFieldConflict(decoded.value);
    return conflict === null
      ? {
          kind: "invalid_contract",
          message: "The server returned an invalid conflict response.",
        }
      : { kind: "same_field_conflict", message, conflict };
  }
  if (error.code === "client_txn_conflict") {
    return { kind: "client_txn_conflict", message };
  }
  if (
    operationID === "mergeEntityRecord" &&
    error.code === "merge_precondition_failed"
  ) {
    const fields = mergePreconditionFields(error.details);
    const reason = safeDetail(error.details.reason_code);
    return {
      kind: "validation",
      message: reason === null ? message : `${message}: ${reason}`,
      ...(fields.length === 0 ? {} : { fields }),
    };
  }
  if (status === 401) {
    return { kind: "authentication_required", message };
  }
  if (status === 403 || error.code === "authorization_denied") {
    return { kind: "authorization_lost", message };
  }
  if (staleTargetCodes.has(error.code)) {
    return { kind: "stale_target", message };
  }
  if (status === 400 || status === 422) {
    const fields = validationFields(error.details);
    return fields === undefined
      ? { kind: "validation", message }
      : { kind: "validation", message, fields };
  }
  if (
    error.retryable ||
    status === 429 ||
    status === 502 ||
    status === 503 ||
    status === 504
  ) {
    return { kind: "retryable", message };
  }
  return { kind: "terminal", message };
}

function operationFailure(
  status: number,
  payload: unknown,
  operationID: WorkbookOperationID,
): WorkbookOperationFailure {
  const failure = operationFailureWithoutPresentation(
    status,
    payload,
    operationID,
  );
  const decoded = errorEnvelopeDecoder.decode(payload);
  const code = decoded.ok
    ? decoded.value.error.code
    : "invalid_public_contract_response";
  const operationFamily =
    operationID === "queryWorkbookView" ||
    operationID === "getIncidentWorkbookStartup" ||
    operationID === "getRecordHistory"
      ? "surface_load"
      : operationID === "issueEvidencePreviewHandle" ||
          operationID === "issueEvidenceDownloadHandle"
        ? "evidence_preview"
        : "field_mutation";
  return {
    ...failure,
    presentation: resolvePublicErrorPresentation({
      code,
      hasAuthorizedMaterialization: false,
      operationFamily,
      status,
    }),
  };
}

export function createWorkbookOperationExecutor(options: {
  readonly apiBase: string | undefined;
}): WorkbookOperationExecutor {
  return {
    async execute<OperationID extends WorkbookOperationID>(
      input: WorkbookOperationExecution<OperationID>,
    ): Promise<WorkbookOperationOutcome<HTTPOperationResponse<OperationID>>> {
      const request = input.request;
      const result = await fetchHTTPOperation<
        HTTPOperationResponse<OperationID>
      >({
        apiBase: options.apiBase,
        operationID: input.operationID,
        ...(input.observeTransport?.onJSONParsed === undefined
          ? {}
          : { onJSONParsed: input.observeTransport.onJSONParsed }),
        ...(input.observeTransport?.onResponseStatus === undefined
          ? {}
          : {
              onResponse: (response: Response) =>
                input.observeTransport?.onResponseStatus?.(response.status),
            }),
        pathParameters: input.pathParameters,
        query: input.query,
        init: {
          method: httpOperationBindings[input.operationID].method,
          ...(input.signal === undefined ? {} : { signal: input.signal }),
          ...(request === undefined ? {} : { body: JSON.stringify(request) }),
        },
      });
      return result.ok
        ? { kind: "accepted", value: result.payload }
        : {
            kind: "rejected",
            failure: operationFailure(
              result.status,
              result.payload,
              input.operationID,
            ),
          };
    },
  };
}
