import { uploadEvidenceObjectBlobTarget } from "../../../services/workbookEvidence";
import { resolvePublicErrorPresentation } from "../../../shared/publicErrorPresentation";
import type { WorkbookOperationExecutor } from "../../adapters/workbookOperationContract";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";

/** Shared slot/upload boundary; each caller retains its own finalization transaction. */
export async function createUploadedEvidenceBlob({
  apiBase,
  clientTxnId,
  file,
  incidentId,
  operations,
}: {
  readonly apiBase: string | undefined;
  readonly clientTxnId: string;
  readonly file: File;
  readonly incidentId: string;
  readonly operations: WorkbookOperationExecutor;
}): Promise<WorkbookOperationOutcome<{ readonly objectBlobId: string }>> {
  const request = {
    incident_id: incidentId,
    client_txn_id: clientTxnId,
    byte_size: file.size,
    filename_hint: file.name || null,
    content_type_hint: file.type || null,
  };
  const created = await operations.execute({
    operationID: "createObjectBlobSlot",
    request,
  });
  if (created.kind === "rejected") return created;
  const data = created.value.data;
  const accepted = data.accepted_contract;
  if (
    data.incident_id !== incidentId ||
    accepted.incident_id !== incidentId ||
    accepted.byte_size !== request.byte_size ||
    accepted.filename_hint !== request.filename_hint ||
    accepted.content_type_hint !== request.content_type_hint
  ) {
    return {
      kind: "rejected",
      failure: {
        kind: "invalid_contract",
        message: "The evidence upload response was invalid.",
      },
    };
  }
  const uploaded = await uploadEvidenceObjectBlobTarget(
    apiBase,
    data.upload_target,
    file,
  );
  if (uploaded.kind === "accepted")
    return { kind: "accepted", value: { objectBlobId: data.object_blob_id } };
  const status =
    uploaded.failure.cause === "http" ? uploaded.failure.status : 0;
  const presentation = resolvePublicErrorPresentation({
    code: "",
    status,
    operationFamily: "field_mutation",
    hasAuthorizedMaterialization: true,
  });
  return {
    kind: "rejected",
    failure: {
      kind:
        presentation.family === "authentication_required"
          ? "authentication_required"
          : "terminal",
      message: "The evidence file could not be uploaded.",
      presentation,
      uploadFailure: uploaded.failure,
    },
  };
}
