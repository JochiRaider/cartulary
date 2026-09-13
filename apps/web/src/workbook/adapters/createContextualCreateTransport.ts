import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";
import {
  contextualCreateRequest,
  freezeContextualCreate,
} from "../features/coordination/contextualCreateModel";
import type { ContextualCreateTransport } from "../features/coordination/contextualCreateOperation";
import { sendWorkbookRecordMutation } from "./sendWorkbookRecordMutation";

/** Ordinary create only: all relationship context is admitted atomically in this request. */
export function createContextualCreateTransport(
  apiBase: string | undefined,
): ContextualCreateTransport {
  return {
    capture(review, clientTxnId) {
      const request = contextualCreateRequest(review.draft, clientTxnId);
      if (!request)
        throw new Error("Complete the required fields before creating.");
      return freezeContextualCreate({
        clientTxnId,
        review: structuredClone(review),
        request,
        body: JSON.stringify(request),
        apiBase,
        path: buildHTTPOperationPath("createViewRow", {
          incident_id: review.authority.incidentId,
          view_schema_id: review.draft.target.viewSchemaId,
        }),
      });
    },
    async send(attempt, signal) {
      if (attempt.body !== JSON.stringify(attempt.request))
        return { kind: "uncertain" };
      return sendWorkbookRecordMutation(
        {
          ...attempt,
          operationID: "createViewRow",
          viewSchemaId: attempt.review.draft.target.viewSchemaId,
          pathParameters: {
            incident_id: attempt.review.authority.incidentId,
            view_schema_id: attempt.review.draft.target.viewSchemaId,
          },
        },
        signal,
      );
    },
  };
}
