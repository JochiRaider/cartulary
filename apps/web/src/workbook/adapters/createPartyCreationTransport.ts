import {
  buildHTTPOperationPath,
  type CreateViewRowResponse,
} from "@cartulary/protocol-ts/http";
import { requireViewContract } from "@cartulary/view-contracts";
import { fetchHTTPOperation } from "../../services/browserApi";
import { buildGenericCreateRequest } from "../features/generic/genericCreateRequestBuilder";
import {
  type PartyReview,
  partyViewId,
} from "../features/parties/partyLinkModel";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";
import { acceptedRecordMutation } from "./workbookRecordPatchTransport";

export type PartyCreationReceipt = Readonly<CreateViewRowResponse>;
export type PartyCreationAttempt = Readonly<{
  id: string;
  review: PartyReview;
  draft: Readonly<Record<string, string>>;
  body: string;
  path: string;
  apiBase: string | undefined;
}>;
export type PartyCreationOutcome =
  | { kind: "accepted"; receipt: PartyCreationReceipt }
  | { kind: "uncertain" }
  | { kind: "rejected"; failure: WorkbookOperationFailure };
export interface PartyCreationTransport {
  capture(
    review: PartyReview,
    draft: Readonly<Record<string, string>>,
    id: string,
  ): PartyCreationAttempt;
  send(
    attempt: PartyCreationAttempt,
    signal: AbortSignal,
  ): Promise<PartyCreationOutcome>;
}
export function createPartyCreationTransport(
  apiBase: string | undefined,
  readScope?: WorkbookReadScopeSource,
): PartyCreationTransport {
  return {
    capture(review, draft, id) {
      const request = buildGenericCreateRequest(
        requireViewContract(partyViewId),
        draft,
        id,
      );
      if (!request) throw new Error("Review the required Party fields.");
      return Object.freeze({
        id,
        review: structuredClone(review),
        draft: structuredClone(draft),
        body: JSON.stringify(request),
        apiBase,
        path: buildHTTPOperationPath("createViewRow", {
          incident_id: review.authority.incidentId,
          view_schema_id: partyViewId,
        }),
      });
    },
    async send(attempt, signal) {
      const scope = readScope?.() ?? null;
      let status: number | null = null;
      try {
        const pathParameters = {
          incident_id: attempt.review.authority.incidentId,
          view_schema_id: partyViewId,
        };
        if (
          attempt.path !==
          buildHTTPOperationPath("createViewRow", pathParameters)
        )
          return { kind: "uncertain" };
        const result = await fetchHTTPOperation<CreateViewRowResponse>({
          apiBase: attempt.apiBase,
          operationID: "createViewRow",
          pathParameters,
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
        return acceptedRecordMutation(result.payload.data, partyViewId)
          ? {
              kind: "accepted",
              receipt: {
                ...result.payload,
                data: {
                  ...result.payload.data,
                  row: acceptWorkbookRowObservation(
                    result.payload.data.row,
                    scope,
                  ),
                },
              },
            }
          : { kind: "uncertain" };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}
