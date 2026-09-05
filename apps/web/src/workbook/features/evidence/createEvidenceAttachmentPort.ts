import type { WorkbookOperationExecutor } from "../../adapters/workbookOperationContract";
import type {
  WorkbookProtocolAttachBlobRequest,
  WorkbookProtocolPatchRecordRequest,
} from "../../adapters/workbookProtocolTypes";
import { evidenceViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { SecureTransactionIdPort } from "../../mutations/secureTransactionId";
import type {
  EvidenceAttachOutcome,
  EvidenceCapabilityPort,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import { createUploadedEvidenceBlob } from "./createUploadedEvidenceBlob";

type AttachmentPort = Pick<EvidenceCapabilityPort, "attach">;

function operationIdentityFailure<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "terminal",
      message: "A secure transaction ID could not be created.",
    },
  };
}

function invalidOperationPayload<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: { kind: "validation", message: "invalid_mutation_payload" },
  };
}

function invalidOperationContract<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "The server returned an inconsistent Workbook operation result.",
    },
  };
}

function createId(
  transactionIds: SecureTransactionIdPort,
  prefix: string,
): string | null {
  try {
    return transactionIds.create(prefix);
  } catch {
    return null;
  }
}

function attachmentTransactionIDs(transactionIds: SecureTransactionIdPort) {
  const create = createId(transactionIds, "evidence-blob");
  const attach = createId(transactionIds, "evidence-attach");
  const available = createId(transactionIds, "evidence-available");
  return create === null || attach === null || available === null
    ? null
    : { attach, available, create };
}

async function attachBlob(
  operations: WorkbookOperationExecutor,
  input: Parameters<AttachmentPort["attach"]>[0],
  objectBlobId: string,
  clientTxnId: string,
): Promise<WorkbookOperationOutcome<{ readonly rowVersion: number }>> {
  const outcome = await operations.execute({
    operationID: "attachBlobToEvidenceRecord",
    pathParameters: { record_id: input.evidenceRecordId },
    request: {
      object_blob_id: objectBlobId,
      base_row_version: input.baseRowVersion,
      client_txn_id: clientTxnId,
    } satisfies WorkbookProtocolAttachBlobRequest,
  });
  if (outcome.kind === "rejected") return outcome;
  return outcome.value.data.row.record_id === input.evidenceRecordId &&
    outcome.value.data.object_blob_id === objectBlobId
    ? {
        kind: "accepted",
        value: { rowVersion: outcome.value.data.row.row_version },
      }
    : invalidOperationContract();
}

async function markEvidenceAvailable(
  operations: WorkbookOperationExecutor,
  evidenceRecordId: string,
  baseRowVersion: number,
  clientTxnId: string,
): Promise<EvidenceAttachOutcome> {
  const outcome = await operations.execute({
    operationID: "patchRecord",
    pathParameters: { record_id: evidenceRecordId },
    request: {
      view_schema_id: evidenceViewSchemaId,
      base_row_version: baseRowVersion,
      client_txn_id: clientTxnId,
      changes: [
        {
          field_key: "evidence.lifecycle_state",
          value: "available",
        },
      ],
    } satisfies WorkbookProtocolPatchRecordRequest,
  });
  if (outcome.kind === "rejected") return outcome;
  return outcome.value.data.view_schema_id === evidenceViewSchemaId &&
    outcome.value.data.row.record_id === evidenceRecordId &&
    outcome.value.data.row.cells["evidence.lifecycle_state"]?.value ===
      "available"
    ? { kind: "accepted", value: { evidenceRecordId } }
    : invalidOperationContract();
}

export function createEvidenceAttachmentPort(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly operations: WorkbookOperationExecutor;
  readonly transactionIds: SecureTransactionIdPort;
}): AttachmentPort {
  return {
    async attach(input) {
      const transactionIDs = attachmentTransactionIDs(options.transactionIds);
      if (transactionIDs === null) return operationIdentityFailure();
      if (input.file.size <= 0) return invalidOperationPayload();
      const blob = await createUploadedEvidenceBlob({
        ...options,
        file: input.file,
        clientTxnId: transactionIDs.create,
      });
      if (blob.kind === "rejected") return blob;
      const attachment = await attachBlob(
        options.operations,
        input,
        blob.value.objectBlobId,
        transactionIDs.attach,
      );
      if (attachment.kind === "rejected") return attachment;
      return markEvidenceAvailable(
        options.operations,
        input.evidenceRecordId,
        attachment.value.rowVersion,
        transactionIDs.available,
      );
    },
  };
}
