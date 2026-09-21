import {
  buildHTTPOperationPath,
  httpOperationBindings,
  type MarkTimelineRecordReviewedRequest,
  type MarkTimelineRecordReviewedResponse,
  type SupersedeRecordRequest,
  type SupersedeRecordResponse,
} from "@cartulary/protocol-ts/http";
import { apiPath, fetchHTTPOperation } from "../../../services/browserApi";
import { classifyWorkbookOperationFailure } from "../../adapters/workbookOperationErrorPolicy";
import { acceptWorkbookRowObservation } from "../../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../../query/WorkbookQueryRow";
import { timelineCaptureReviewValid } from "../actions/timelineCaptureActionModel";
import type {
  TimelineCaptureAttempt,
  TimelineRecordActionPort,
} from "../ports/TimelineRecordActionPort";
import type { TimelineCaptureReceipt } from "./timelineCaptureProtocol";

export function createTimelineRecordActionAdapter(options: {
  readonly apiBase: string | undefined;
  readonly readScope?: WorkbookReadScopeSource;
}): TimelineRecordActionPort {
  return {
    capture(review, id) {
      if (!timelineCaptureReviewValid(review) || !id)
        throw new Error("Timeline review unavailable");
      const operationID =
        review.action === "mark-reviewed"
          ? "markTimelineRecordReviewed"
          : "supersedeRecord";
      const request =
        review.action === "mark-reviewed"
          ? ({
              base_row_version: review.target.rowVersion,
              client_txn_id: id,
            } satisfies MarkTimelineRecordReviewedRequest)
          : ({
              base_row_version: review.target.rowVersion,
              client_txn_id: id,
              reason: review.reason ?? "",
              ...(review.replacement === null
                ? {}
                : { replacement_record_id: review.replacement.recordId }),
            } satisfies SupersedeRecordRequest);
      return {
        id,
        review: structuredClone(review),
        operationID,
        method: "POST",
        apiBase: options.apiBase,
        path: apiPath(
          options.apiBase,
          buildHTTPOperationPath(operationID, {
            record_id: review.target.recordId,
          }),
        ),
        body: JSON.stringify(request),
      };
    },
    async send(attempt, signal) {
      const scope = options.readScope?.() ?? null;
      const operationID =
        attempt.review.action === "mark-reviewed"
          ? "markTimelineRecordReviewed"
          : "supersedeRecord";
      if (
        attempt.operationID !== operationID ||
        attempt.method !== httpOperationBindings[operationID].method ||
        attempt.path !==
          apiPath(
            attempt.apiBase,
            buildHTTPOperationPath(operationID, {
              record_id: attempt.review.target.recordId,
            }),
          )
      )
        return { kind: "uncertain" };
      let status: number | null = null;
      try {
        const result = await fetchHTTPOperation<
          MarkTimelineRecordReviewedResponse | SupersedeRecordResponse
        >({
          apiBase: attempt.apiBase,
          operationID,
          pathParameters: { record_id: attempt.review.target.recordId },
          init: { method: attempt.method, body: attempt.body, signal },
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
            operationID,
          );
          return failure.kind === "invalid_contract"
            ? { kind: "uncertain" }
            : { kind: "rejected", failure };
        }
        const receipt = validateTimelineCaptureReceipt(
          attempt,
          result.payload.data,
        );
        if (!receipt) return { kind: "uncertain" };
        const observation = acceptWorkbookRowObservation(
          {
            record_id: receipt.data.record_id,
            row_version: receipt.data.row_version,
            cells: {},
          },
          scope,
        ).observation;
        return {
          kind: "acknowledged",
          receipt: { ...receipt, ...(observation ? { observation } : {}) },
        };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}

function validateTimelineCaptureReceipt(
  attempt: TimelineCaptureAttempt,
  data: (MarkTimelineRecordReviewedResponse | SupersedeRecordResponse)["data"],
): TimelineCaptureReceipt | null {
  const { review } = attempt;
  if (
    !("record_id" in data) ||
    data.record_id !== review.target.recordId ||
    data.incident_id !== review.authority.incidentId ||
    !Number.isSafeInteger(data.row_version) ||
    data.row_version <= review.target.rowVersion ||
    !data.change_set_id ||
    data.capture_state !==
      (review.action === "mark-reviewed" ? "reviewed" : "superseded") ||
    data.reason !== review.reason ||
    data.replacement_record_id !== (review.replacement?.recordId ?? null)
  )
    return null;
  return { operation: review.action, data: { ...data } };
}
