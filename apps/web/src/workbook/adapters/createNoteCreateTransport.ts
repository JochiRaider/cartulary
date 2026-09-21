import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";
import { noteCreateView, prepareNote } from "../features/notes/noteCreateModel";
import type { NoteTransport } from "../features/notes/noteCreateOperation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { freezeWorkbookValue } from "../utils/freezeWorkbookValue";
import { sendWorkbookRecordMutation } from "./sendWorkbookRecordMutation";

/** Both Note operations are single atomic creates; the route owns the association. */
export function createNoteCreateTransport(
  apiBase: string | undefined,
  readScope?: WorkbookReadScopeSource,
): NoteTransport {
  return {
    capture(review, clientTxnId) {
      const request = prepareNote(review.draft, clientTxnId).request;
      if (!request)
        throw new Error("Complete the required Note fields before creating.");
      const operationID = review.draft.source
        ? "createRecordLinkedNote"
        : "createViewRow";
      const pathParameters: Record<string, string> = review.draft.source
        ? { record_id: review.draft.source.recordId }
        : {
            incident_id: review.authority.incidentId,
            view_schema_id: noteCreateView,
          };
      return freezeWorkbookValue({
        operationID,
        apiBase,
        pathParameters,
        path: buildHTTPOperationPath(operationID, pathParameters),
        request,
        body: JSON.stringify(request),
        clientTxnId,
        review: structuredClone(review),
      });
    },
    async send(attempt, signal) {
      if (attempt.body !== JSON.stringify(attempt.request))
        return { kind: "uncertain" };
      const result = await sendWorkbookRecordMutation(
        { ...attempt, viewSchemaId: noteCreateView },
        signal,
        readScope?.() ?? null,
      );
      if (
        result.kind === "accepted" &&
        attempt.operationID === "createRecordLinkedNote"
      ) {
        const data = result.receipt.data;
        if (
          !("source_record_id" in data) ||
          data.source_record_id !== attempt.review.draft.source?.recordId ||
          !("link_type" in data) ||
          data.link_type !== "references_artifact"
        )
          return { kind: "uncertain" };
      }
      return result;
    },
  };
}
