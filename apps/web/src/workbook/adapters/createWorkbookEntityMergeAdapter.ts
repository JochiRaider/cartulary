import {
  buildHTTPOperationPath,
  httpOperationBindings,
  type MergeEntityRecordRequest,
  type MergeEntityRecordResponse,
} from "@cartulary/protocol-ts/http";
import { apiPath, fetchHTTPOperation } from "../../services/browserApi";
import type {
  EntityMergeAttempt,
  EntityMergeReceipt,
  WorkbookEntityMergePort,
} from "../features/entities/entityMergeOperation";
import { entityMergeIdentifierFields } from "../models/entityIdentifierClasses";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";

const operationID = "mergeEntityRecord";
function capturedPath(apiBase: string | undefined, survivorId: string): string {
  return apiPath(
    apiBase,
    buildHTTPOperationPath(operationID, { survivor_record_id: survivorId }),
  );
}
export function createWorkbookEntityMergeAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): WorkbookEntityMergePort {
  return {
    capture(review, id) {
      if (review.authority.incidentId !== options.incidentId)
        throw new Error("Merge incident changed");
      return {
        id,
        review: structuredClone(review),
        apiBase: options.apiBase,
        path: capturedPath(options.apiBase, review.survivor.recordId),
        body: JSON.stringify({
          loser_record_id: review.loser.recordId,
          survivor_base_row_version: review.survivor.baseRowVersion,
          loser_base_row_version: review.loser.baseRowVersion,
          client_txn_id: id,
          reason: review.reason,
        } satisfies MergeEntityRecordRequest),
      };
    },
    async send(attempt, signal) {
      if (
        attempt.review.authority.incidentId !== options.incidentId ||
        attempt.path !==
          capturedPath(attempt.apiBase, attempt.review.survivor.recordId)
      )
        return { kind: "uncertain" };
      let responseStatus: number | null = null;
      try {
        const result = await fetchHTTPOperation<MergeEntityRecordResponse>({
          apiBase: attempt.apiBase,
          operationID,
          pathParameters: {
            survivor_record_id: attempt.review.survivor.recordId,
          },
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
        const receipt = validatedEntityMergeReceipt(
          attempt,
          result.payload.data,
        );
        return receipt === null
          ? { kind: "uncertain" }
          : { kind: "acknowledged", receipt };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}

function validatedEntityMergeReceipt(
  attempt: EntityMergeAttempt,
  data: MergeEntityRecordResponse["data"],
): EntityMergeReceipt | null {
  const review = attempt.review;
  const summary = data.merge_summary;
  const classes = entityMergeIdentifierFields[review.entityType];
  const count = (value: unknown) =>
    typeof value === "number" && Number.isSafeInteger(value) && value >= 0;
  if (
    data.incident_id !== review.authority.incidentId ||
    data.record_type !== review.entityType ||
    data.survivor_record_id !== review.survivor.recordId ||
    data.loser_record_id !== review.loser.recordId ||
    data.merged_into_record_id !== review.survivor.recordId ||
    !count(data.survivor_row_version) ||
    data.survivor_row_version <= review.survivor.baseRowVersion ||
    !count(data.loser_row_version) ||
    data.loser_row_version <= review.loser.baseRowVersion ||
    !data.change_set_id ||
    summary.record_type !== review.entityType ||
    summary.exact_match_classes.length !== classes.length ||
    Object.entries(summary).some(
      ([key, value]) =>
        key !== "record_type" && key !== "exact_match_classes" && !count(value),
    ) ||
    summary.exact_match_classes.some(
      (entry, index) =>
        entry.identifier_class !== classes[index]?.identifierClass ||
        entry.promoted_count > 1 ||
        entry.blocked_conflict_count !== 0 ||
        Object.entries(entry).some(
          ([key, value]) => key !== "identifier_class" && !count(value),
        ),
    )
  )
    return null;
  return data;
}
