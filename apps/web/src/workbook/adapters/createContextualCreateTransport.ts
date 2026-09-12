import {
  buildHTTPOperationPath,
  type CreateViewRowResponse,
} from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";
import {
  contextualCreateRequest,
  freezeContextualCreate,
} from "../features/coordination/contextualCreateModel";
import type { ContextualCreateTransport } from "../features/coordination/contextualCreateOperation";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";
import { acceptedRecordMutation } from "./workbookRecordPatchTransport";

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
      let status: number | null = null,
        requestId: string | null = null;
      try {
        const pathParameters = {
          incident_id: attempt.review.authority.incidentId,
          view_schema_id: attempt.review.draft.target.viewSchemaId,
        };
        if (
          attempt.path !==
            buildHTTPOperationPath("createViewRow", pathParameters) ||
          attempt.body !== JSON.stringify(attempt.request) ||
          attempt.request.client_txn_id !== attempt.clientTxnId
        )
          return { kind: "uncertain" };
        const result = await fetchHTTPOperation<CreateViewRowResponse>({
          apiBase: attempt.apiBase,
          operationID: "createViewRow",
          pathParameters,
          init: { method: "POST", body: attempt.body, signal },
          onResponse: (response) => {
            status = response.status;
            requestId = response.headers.get("X-Request-ID");
          },
        });
        if (!result.ok) {
          if (
            status === null ||
            status < 400 ||
            status >= 500 ||
            !result.payload.error?.code ||
            (requestId !== null &&
              requestId !== result.payload.error.request_id)
          )
            return { kind: "uncertain" };
          const failure = classifyWorkbookOperationFailure(
            result.status,
            result.payload,
            "createViewRow",
          );
          return failure.kind === "invalid_contract"
            ? { kind: "uncertain" }
            : { kind: "rejected", failure };
        }
        const receipt = result.payload;
        if (
          !receipt.meta.request_id ||
          (requestId !== null && requestId !== receipt.meta.request_id) ||
          !acceptedRecordMutation(receipt.data, pathParameters.view_schema_id)
        )
          return { kind: "uncertain" };
        return {
          kind: "accepted",
          receipt: freezeContextualCreate(structuredClone(receipt)),
        };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}
