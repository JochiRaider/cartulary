import type {
  CreateViewRowRequest,
  CreateViewRowResponse,
  PatchRecordRequest,
} from "@cartulary/protocol-ts/http";
import {
  createUploadedEvidenceObjectBlob,
  evidenceAttachPublicErrorMessage,
} from "../../../services/workbookEvidence";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import {
  buildAttachedEvidenceCreatePayload,
  buildAttachedEvidencePatchPayload,
  normalizeTimelineFullRow,
} from "../models/workbookTimelineModel";
import type {
  TimelineEvidenceAttachmentAccepted,
  TimelineEvidenceAttachmentPort,
} from "../ports/TimelineEvidenceAttachmentPort";

function invalidContract(): WorkbookOperationOutcome<TimelineEvidenceAttachmentAccepted> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "The Timeline evidence response was invalid.",
    },
  };
}

function rejected(
  message: string,
): WorkbookOperationOutcome<TimelineEvidenceAttachmentAccepted> {
  return { kind: "rejected", failure: { kind: "terminal", message } };
}

function evidenceTitleFromFile(file: File): string {
  return file.name.trim() || "Workbook attachment";
}

export function createTimelineEvidenceAttachmentAdapter(options: {
  readonly apiBase: string | undefined;
  readonly createClientTxnId: () => string;
  readonly incidentId: string;
}): TimelineEvidenceAttachmentPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async attach({ file, onTimelineClientTxnId, target }) {
      let evidenceRecordId: string;
      try {
        const objectBlobId = await createUploadedEvidenceObjectBlob({
          apiBase: options.apiBase,
          createClientTxnId: options.createClientTxnId,
          file,
          incidentId: options.incidentId,
        });
        const rowClientTxnId = options.createClientTxnId();
        const request = {
          client_txn_id: rowClientTxnId,
          "evidence.collector_party_text": "Workbook upload",
          "evidence.initial_object_blob_id": objectBlobId,
          "evidence.lifecycle_state": "available",
          "evidence.title": evidenceTitleFromFile(file),
        } as CreateViewRowRequest;
        let createOutcome:
          | WorkbookOperationOutcome<CreateViewRowResponse>
          | undefined;
        for (let attempt = 0; attempt < 2; attempt += 1) {
          try {
            createOutcome = await operations.execute({
              operationID: "createViewRow",
              pathParameters: {
                incident_id: options.incidentId,
                view_schema_id: evidenceViewSchemaId,
              },
              request,
            });
            break;
          } catch {
            if (attempt === 1)
              throw new Error("evidence_row_transport_failure");
          }
        }
        if (createOutcome === undefined || createOutcome.kind === "rejected") {
          return {
            clientTxnId: null,
            outcome: createOutcome ?? rejected("Evidence row creation failed."),
          };
        }
        if (createOutcome.value.data.view_schema_id !== evidenceViewSchemaId) {
          return { clientTxnId: null, outcome: invalidContract() };
        }
        evidenceRecordId = createOutcome.value.data.row.record_id;
      } catch (error) {
        return {
          clientTxnId: null,
          outcome: rejected(evidenceAttachPublicErrorMessage(error)),
        };
      }

      let clientTxnId: string;
      try {
        clientTxnId = options.createClientTxnId();
      } catch {
        return {
          clientTxnId: null,
          outcome: rejected(
            "A secure request identifier could not be created.",
          ),
        };
      }
      const request =
        target.recordId === null
          ? buildAttachedEvidenceCreatePayload(evidenceRecordId, clientTxnId)
          : buildAttachedEvidencePatchPayload(
              target,
              evidenceRecordId,
              clientTxnId,
            );
      if (request === null) {
        return {
          clientTxnId,
          outcome: {
            kind: "rejected",
            failure: {
              kind: "stale_target",
              message: "The selected Timeline row is stale.",
            },
          },
        };
      }
      onTimelineClientTxnId(clientTxnId);
      try {
        const outcome =
          target.recordId === null
            ? await operations.execute({
                operationID: "createViewRow",
                pathParameters: {
                  incident_id: options.incidentId,
                  view_schema_id: timelineViewSchemaId,
                },
                request: request as CreateViewRowRequest,
              })
            : await operations.execute({
                operationID: "patchRecord",
                pathParameters: { record_id: target.recordId },
                request: request as PatchRecordRequest,
              });
        if (outcome.kind === "rejected") {
          return { clientTxnId, outcome };
        }
        const data = outcome.value.data;
        if (
          data.view_schema_id !== timelineViewSchemaId ||
          (target.recordId !== null && data.row.record_id !== target.recordId)
        ) {
          return { clientTxnId, outcome: invalidContract() };
        }
        try {
          return {
            clientTxnId,
            outcome: {
              kind: "accepted",
              value: {
                evidenceRecordId,
                row: normalizeTimelineFullRow(
                  data.row,
                  "evidence attachment response row",
                ),
                viewSchemaId: data.view_schema_id,
              },
            },
          };
        } catch {
          return { clientTxnId, outcome: invalidContract() };
        }
      } catch {
        return {
          clientTxnId,
          outcome: {
            kind: "rejected",
            failure: {
              kind: "retryable",
              message: "The Timeline evidence link could not be sent.",
            },
          },
        };
      }
    },
  };
}
