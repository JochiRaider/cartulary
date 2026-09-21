import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";
import { prepareCoordination } from "../features/coordination/coordinationCreateModel";
import type { CoordinationTransport } from "../features/coordination/coordinationCreateOperation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { freezeWorkbookValue } from "../utils/freezeWorkbookValue";
import { sendWorkbookRecordMutation } from "./sendWorkbookRecordMutation";

/** The ordinary route atomically creates target fields, references and source link. */
export function createCoordinationCreateTransport(
  apiBase: string | undefined,
  readScope?: WorkbookReadScopeSource,
): CoordinationTransport {
  return {
    capture(review, clientTxnId) {
      const request = prepareCoordination(review.draft, clientTxnId).request;
      if (!request)
        throw new Error(
          "Complete the required coordination fields before creating.",
        );
      const operationID = "createViewRow";
      const pathParameters = {
        incident_id: review.authority.incidentId,
        view_schema_id: review.draft.target.viewSchemaId,
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
        { ...attempt, viewSchemaId: attempt.review.draft.target.viewSchemaId },
        signal,
        readScope?.() ?? null,
      );
      if (result.kind === "accepted") {
        const data = result.receipt.data,
          source = attempt.review.draft.source;
        if (
          source
            ? !("source_record_id" in data) ||
              data.source_record_id !== source.recordId ||
              !("link_type" in data) ||
              data.link_type !== "references_artifact"
            : "source_record_id" in data || "link_type" in data
        )
          return { kind: "uncertain" };
      }
      return result;
    },
  };
}
