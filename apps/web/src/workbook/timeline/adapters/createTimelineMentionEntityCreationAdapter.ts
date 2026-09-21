import {
  buildHTTPOperationPath,
  type CreateViewRowResponse,
} from "@cartulary/protocol-ts/http";
import { apiPath, fetchHTTPOperation } from "../../../services/browserApi";
import { classifyWorkbookOperationFailure } from "../../adapters/workbookOperationErrorPolicy";
import { normalizeWorkbookViewRows } from "../../models/workbookContractRows";
import { acceptWorkbookRowObservation } from "../../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../../query/WorkbookQueryRow";
import {
  mentionCreateRequest,
  mentionEntityContract,
} from "../actions/timelineMentionCreationModel";
import type { TimelineMentionEntityCreationPort } from "../ports/TimelineMentionPort";

/** Ordinary Host/Identity create/upsert, with author-reviewed fields and its own retained receipt. */
export function createTimelineMentionEntityCreationAdapter(options: {
  readonly apiBase: string | undefined;
  readonly readScope?: WorkbookReadScopeSource;
}): TimelineMentionEntityCreationPort {
  return {
    capture(review, id) {
      const request = mentionCreateRequest(review, id);
      if (!request)
        throw new Error("Review the entity fields before creating.");
      const viewSchemaId = mentionEntityContract(
        review.subject.entityType,
      ).viewSchemaId;
      return {
        id,
        review: structuredClone(review),
        operationID: "createViewRow",
        method: "POST",
        apiBase: options.apiBase,
        viewSchemaId,
        path: apiPath(
          options.apiBase,
          buildHTTPOperationPath("createViewRow", {
            incident_id: review.subject.incidentId,
            view_schema_id: viewSchemaId,
          }),
        ),
        body: JSON.stringify(request),
      };
    },
    async send(attempt, signal) {
      const scope = options.readScope?.() ?? null;
      if (
        attempt.path !==
          apiPath(
            attempt.apiBase,
            buildHTTPOperationPath("createViewRow", {
              incident_id: attempt.review.subject.incidentId,
              view_schema_id: attempt.viewSchemaId,
            }),
          ) ||
        attempt.method !== "POST"
      )
        return { kind: "uncertain" };
      let status: number | null = null;
      try {
        const result = await fetchHTTPOperation<CreateViewRowResponse>({
          apiBase: attempt.apiBase,
          operationID: "createViewRow",
          pathParameters: {
            incident_id: attempt.review.subject.incidentId,
            view_schema_id: attempt.viewSchemaId,
          },
          init: { method: "POST", body: attempt.body, signal },
          onResponse: (response) => {
            status = response.status;
          },
        });
        if (!result.ok) {
          if (
            status === null ||
            status < 400 ||
            status >= 500 ||
            !result.payload.error?.code
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
        const { data } = result.payload;
        if (data.view_schema_id !== attempt.viewSchemaId || !data.change_set_id)
          return { kind: "uncertain" };
        const contract = mentionEntityContract(
          attempt.review.subject.entityType,
        );
        normalizeWorkbookViewRows(
          contract,
          [data.row],
          "Created mention entity",
        );
        return {
          kind: "accepted",
          receipt: {
            ...result.payload,
            data: {
              ...data,
              row: acceptWorkbookRowObservation(data.row, scope),
            },
          },
        };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}
