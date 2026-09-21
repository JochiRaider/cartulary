import {
  buildHTTPOperationPath,
  httpOperationBindings,
  type SupersedeRecordRequest,
  type SupersedeRecordResponse,
} from "@cartulary/protocol-ts/http";
import { apiPath, fetchHTTPOperation } from "../../services/browserApi";
import {
  decisionViewId,
  normalizeDecisionReason,
} from "../features/coordination/decisionSupersessionModel";
import type {
  DecisionSupersessionAttempt,
  DecisionSupersessionReceipt,
  DecisionSupersessionTransportPort,
} from "../features/coordination/decisionSupersessionOperation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { createDecisionCandidateReader } from "./createDecisionCandidateReader";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";

const operationID = "supersedeRecord";
const path = (apiBase: string | undefined, id: string) =>
  apiPath(apiBase, buildHTTPOperationPath(operationID, { record_id: id }));
export function createWorkbookDecisionSupersessionAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly readScope?: WorkbookReadScopeSource;
}): DecisionSupersessionTransportPort {
  return {
    ...createDecisionCandidateReader(options),
    capture(review, id) {
      if (review.authority.incidentId !== options.incidentId)
        throw new Error("Decision incident changed");
      return {
        id,
        review: structuredClone(review),
        apiBase: options.apiBase,
        path: path(options.apiBase, review.target.recordId),
        body: JSON.stringify({
          base_row_version: review.target.baseRowVersion,
          client_txn_id: id,
          replacement_record_id: review.replacement.recordId,
          reason: review.reason,
        } satisfies SupersedeRecordRequest),
      };
    },
    async send(attempt, signal) {
      if (
        attempt.review.authority.incidentId !== options.incidentId ||
        attempt.path !== path(attempt.apiBase, attempt.review.target.recordId)
      )
        return { kind: "uncertain" };
      let responseStatus: number | null = null;
      try {
        const result = await fetchHTTPOperation<SupersedeRecordResponse>({
          apiBase: attempt.apiBase,
          operationID,
          pathParameters: { record_id: attempt.review.target.recordId },
          init: {
            method: httpOperationBindings[operationID].method,
            body: attempt.body,
            signal,
          },
          onResponse: (response) => {
            responseStatus = response.status;
          },
        });
        if (!result.ok) {
          if (
            responseStatus === null ||
            responseStatus < 400 ||
            responseStatus >= 500 ||
            !result.payload.error?.code
          )
            return { kind: "uncertain" };
          const failure = classifyWorkbookOperationFailure(
            result.status,
            result.payload,
            operationID,
          );
          return failure.kind === "invalid_contract"
            ? { kind: "uncertain" }
            : { kind: "rejected", failure };
        }
        const receipt = validatedDecisionReceipt(attempt, result.payload.data);
        return receipt
          ? { kind: "acknowledged", receipt }
          : { kind: "uncertain" };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}

function validatedDecisionReceipt(
  attempt: DecisionSupersessionAttempt,
  data: SupersedeRecordResponse["data"],
): DecisionSupersessionReceipt | null {
  const { review } = attempt;
  if (
    !("target_record_id" in data) ||
    data.view_schema_id !== decisionViewId ||
    data.target_record_id !== review.target.recordId ||
    data.superseding_record_id !== review.replacement.recordId ||
    !Number.isSafeInteger(data.target_row_version) ||
    data.target_row_version <= review.target.baseRowVersion ||
    !Number.isSafeInteger(data.superseding_row_version) ||
    data.superseding_row_version < 1 ||
    !data.change_set_id ||
    data.target_status !==
      (review.target.status === "executed" ? "executed" : "superseded") ||
    data.reason !== review.reason ||
    normalizeDecisionReason(data.reason) !== data.reason
  )
    return null;
  // The replacement version is not an atomic public precondition.
  return {
    ...data,
    view_schema_id: decisionViewId,
    target_status: data.target_status,
  };
}
