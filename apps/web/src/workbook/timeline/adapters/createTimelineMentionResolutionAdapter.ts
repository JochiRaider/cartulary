import {
  buildHTTPOperationPath,
  type ResolveEntityMentionRequest,
  type ResolveEntityMentionResponse,
} from "@cartulary/protocol-ts/http";
import { apiPath, fetchHTTPOperation } from "../../../services/browserApi";
import { classifyWorkbookOperationFailure } from "../../adapters/workbookOperationErrorPolicy";
import { mentionReviewValid } from "../actions/timelineMentionOperationModel";
import type { TimelineMentionResolutionPort } from "../ports/TimelineMentionPort";
import { validateMentionReceipt } from "./timelineMentionProtocol";

export function createTimelineMentionResolutionAdapter(options: {
  readonly apiBase: string | undefined;
}): TimelineMentionResolutionPort {
  return {
    capture(review, id) {
      if (!mentionReviewValid(review) || !id)
        throw new Error("Mention review unavailable");
      const request = {
        action: review.intent.action,
        base_mention_row_version: review.subject.mentionRowVersion,
        client_txn_id: id,
        ...(review.intent.action === "resolve_item"
          ? { resolved_record_id: review.intent.resolvedRecordId }
          : {}),
      } satisfies ResolveEntityMentionRequest;
      return {
        id,
        review: structuredClone(review),
        operationID: "resolveEntityMention",
        method: "POST",
        apiBase: options.apiBase,
        path: apiPath(
          options.apiBase,
          buildHTTPOperationPath("resolveEntityMention", {
            entity_mention_id: review.subject.mentionId,
          }),
        ),
        body: JSON.stringify(request),
      };
    },
    async send(attempt, signal) {
      if (
        attempt.operationID !== "resolveEntityMention" ||
        attempt.method !== "POST" ||
        attempt.path !==
          apiPath(
            attempt.apiBase,
            buildHTTPOperationPath("resolveEntityMention", {
              entity_mention_id: attempt.review.subject.mentionId,
            }),
          )
      )
        return { kind: "uncertain" };
      let status: number | null = null;
      try {
        const result = await fetchHTTPOperation<ResolveEntityMentionResponse>({
          apiBase: attempt.apiBase,
          operationID: attempt.operationID,
          pathParameters: {
            entity_mention_id: attempt.review.subject.mentionId,
          },
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
            attempt.operationID,
          );
          return failure.kind === "invalid_contract"
            ? { kind: "uncertain" }
            : { kind: "rejected", failure };
        }
        const receipt = validateMentionReceipt(attempt, result.payload.data);
        return receipt ? { kind: "accepted", receipt } : { kind: "uncertain" };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}
