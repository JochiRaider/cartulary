import type {
  HTTPOperationID,
  HTTPOperationRequest,
  HTTPOperationResponse,
  HTTPPathParameters,
  HTTPQueryValue,
} from "@cartulary/protocol-ts/http";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";

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
