import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { TimelineFileLinkTransport } from "../features/evidence/timelineFileOperation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { buildAttachedEvidencePatchRequest } from "../timeline/adapters/timelineEvidenceRequestBuilders";
import { freezeWorkbookValue } from "../utils/freezeWorkbookValue";
import { sendWorkbookRecordMutation } from "./sendWorkbookRecordMutation";

export function createTimelineFileLinkTransport(
  apiBase: string | undefined,
  readScope?: WorkbookReadScopeSource,
): TimelineFileLinkTransport {
  return {
    capture(authority, source, evidenceRecordId, clientTxnId) {
      const request = buildAttachedEvidencePatchRequest(
        { rowVersion: source.row_version },
        evidenceRecordId,
        clientTxnId,
      );
      if (!request)
        throw new Error("The original Timeline record needs review.");
      return freezeWorkbookValue({
        authority: { ...authority },
        source: structuredClone(source),
        evidenceRecordId,
        clientTxnId,
        body: JSON.stringify(request),
        path: buildHTTPOperationPath("patchRecord", {
          record_id: source.record_id,
        }),
        apiBase,
      });
    },
    send(attempt, signal) {
      return sendWorkbookRecordMutation(
        {
          ...attempt,
          operationID: "patchRecord",
          viewSchemaId: timelineViewSchemaId,
          pathParameters: { record_id: attempt.source.record_id },
          recordId: attempt.source.record_id,
          baseRowVersion: attempt.source.row_version,
        },
        signal,
        readScope?.() ?? null,
      );
    },
  };
}
