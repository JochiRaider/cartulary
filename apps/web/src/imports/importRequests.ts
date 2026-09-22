import { clientTxnID } from "../services/browserApi";
import type {
  ApplyImportSessionRequest,
  CancelJobRequest,
  CreateImportUnitRegionRequest,
  ImportMappingRequest,
  ImportUploadMetadata,
  SelectImportUnitRequest,
} from "../services/importContractAdapter";

export type ImportScope = {
  readonly incidentId: string;
  readonly actorId: string;
  readonly lifetime: string;
};
export type ImportWriteAttempt =
  | {
      readonly kind: "upload";
      readonly scope: ImportScope;
      readonly metadata: ImportUploadMetadata;
      readonly file: Blob;
      readonly filename: string;
    }
  | {
      readonly kind: "mapping";
      readonly scope: ImportScope;
      readonly sessionId: string;
      readonly unitId: string;
      readonly body: ImportMappingRequest;
    }
  | {
      readonly kind: "select" | "skip";
      readonly scope: ImportScope;
      readonly sessionId: string;
      readonly unitId: string;
      readonly body: SelectImportUnitRequest;
    }
  | {
      readonly kind: "region";
      readonly scope: ImportScope;
      readonly sessionId: string;
      readonly unitId: string;
      readonly body: CreateImportUnitRegionRequest;
    }
  | {
      readonly kind: "apply";
      readonly scope: ImportScope;
      readonly sessionId: string;
      readonly body: ApplyImportSessionRequest;
    }
  | {
      readonly kind: "cancel";
      readonly scope: ImportScope;
      readonly jobId: string;
      readonly body: CancelJobRequest;
    };

export function immutableImportValue<T>(value: T): T {
  if (value !== null && typeof value === "object" && !(value instanceof Blob)) {
    for (const child of Object.values(value)) immutableImportValue(child);
    Object.freeze(value);
  }
  return value;
}

/** Captures bytes, not the mutable native input; filename is advisory only. */
export function captureWorkbookUpload(
  scope: ImportScope,
  file: File,
  transactionId = clientTxnID("workbook-import-upload"),
): Extract<ImportWriteAttempt, { kind: "upload" }> {
  const type = file.type.split(";")[0]?.trim().toLowerCase() ?? "";
  const mediaType = [
    "text/csv",
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    "application/octet-stream",
  ].includes(type)
    ? type
    : "application/octet-stream";
  return immutableImportValue({
    kind: "upload",
    scope: { ...scope },
    file: file.slice(0, file.size, mediaType),
    filename: file.name,
    metadata: {
      incident_id: scope.incidentId,
      assistant_profile: "workbook_import_v1",
      client_txn_id: transactionId,
    },
  });
}

export function captureImportWrite<
  T extends Exclude<ImportWriteAttempt, { kind: "upload" }>,
>(attempt: T): T {
  // JSON-shaped generated requests: detach all nested mappings/arrays from drafts.
  return immutableImportValue(JSON.parse(JSON.stringify(attempt)) as T);
}
