import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { timelineRelatedEvidenceRequest } from "../features/evidence/timelineRelatedEvidenceModel";
import type { RelatedEvidenceTransport } from "../features/evidence/timelineRelatedEvidenceOperation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { buildAttachedEvidencePatchRequest } from "../timeline/adapters/timelineEvidenceRequestBuilders";
import { freezeWorkbookValue } from "../utils/freezeWorkbookValue";
import { sendWorkbookRecordMutation } from "./sendWorkbookRecordMutation";

export function createTimelineRelatedEvidenceTransport(
  apiBase: string | undefined,
  readScope?: WorkbookReadScopeSource,
): RelatedEvidenceTransport {
  return {
    capture(stage, review, clientTxnId, evidenceRecordId) {
      const request =
        stage === "create"
          ? timelineRelatedEvidenceRequest(review.draft, clientTxnId)
          : evidenceRecordId
            ? buildAttachedEvidencePatchRequest(
                { rowVersion: review.source.row_version },
                evidenceRecordId,
                clientTxnId,
              )
            : null;
      if (!request || (stage === "link" && !evidenceRecordId))
        throw new Error(
          "Review the Evidence metadata and original Timeline record.",
        );
      return freezeWorkbookValue({
        stage,
        review: structuredClone(review),
        clientTxnId,
        evidenceRecordId,
        apiBase,
        body: JSON.stringify(request),
        path:
          stage === "create"
            ? buildHTTPOperationPath("createViewRow", {
                incident_id: review.authority.incidentId,
                view_schema_id: review.draft.target.viewSchemaId,
              })
            : buildHTTPOperationPath("patchRecord", {
                record_id: review.draft.source.recordId,
              }),
      });
    },
    send(attempt, signal) {
      return sendWorkbookRecordMutation(
        {
          ...attempt,
          operationID:
            attempt.stage === "create" ? "createViewRow" : "patchRecord",
          viewSchemaId:
            attempt.stage === "create"
              ? attempt.review.draft.target.viewSchemaId
              : timelineViewSchemaId,
          pathParameters:
            attempt.stage === "create"
              ? {
                  incident_id: attempt.review.authority.incidentId,
                  view_schema_id: attempt.review.draft.target.viewSchemaId,
                }
              : { record_id: attempt.review.draft.source.recordId },
          ...(attempt.stage === "link"
            ? {
                recordId: attempt.review.draft.source.recordId,
                baseRowVersion: attempt.review.source.row_version,
              }
            : {}),
        },
        signal,
        readScope?.() ?? null,
      );
    },
  };
}
