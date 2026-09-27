import { entityCandidateDiscovery } from "@cartulary/protocol-ts/entities";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import type { TimelineMentionCandidatePort } from "../actions/TimelineMentionCandidatePort";

export const timelineMentionCandidatePolicy: TimelineMentionCandidatePort["policy"] =
  Object.freeze({
    pageSize: entityCandidateDiscovery.page_size,
    previousCursorLimit: entityCandidateDiscovery.previous_cursor_limit,
    settledInputMs: entityCandidateDiscovery.settled_input_ms,
  });

/** Authoritative entity discovery is independent of the displayed workbook page. */
export function createTimelineMentionCandidateReader(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): TimelineMentionCandidatePort {
  const operations = createWorkbookOperationExecutor(options);
  return {
    policy: timelineMentionCandidatePolicy,
    async page({ entityType, search, cursor }, signal) {
      if (signal.aborted) return { kind: "aborted" };
      let responseAccepted = false;
      try {
        const result = await operations.execute({
          operationID: "listEntityCandidates",
          pathParameters: { incident_id: options.incidentId },
          query: {
            entity_type: entityType,
            search,
            limit: entityCandidateDiscovery.page_size,
            ...(cursor === null ? {} : { cursor_token: cursor }),
          },
          signal,
        });
        if (signal.aborted) return { kind: "aborted" };
        if (result.kind === "rejected") return result;
        responseAccepted = true;
        const { data, meta } = result.value;
        if (
          data.incident_id !== options.incidentId ||
          data.entity_type !== entityType ||
          meta.paging.limit !== entityCandidateDiscovery.page_size ||
          data.candidates.length > entityCandidateDiscovery.page_size ||
          data.candidates.some(
            (candidate) => candidate.entity_type !== entityType,
          ) ||
          new Set(data.candidates.map((candidate) => candidate.record_id))
            .size !== data.candidates.length ||
          (meta.paging.has_more
            ? !meta.paging.next_cursor || meta.paging.next_cursor === cursor
            : meta.paging.next_cursor !== null)
        )
          throw new Error("Invalid target page");
        return {
          kind: "accepted",
          value: {
            candidates: data.candidates.map((candidate) => ({
              recordId: candidate.record_id,
              rowVersion: candidate.row_version,
              entityType: candidate.entity_type,
              displayText: candidate.display_name,
            })),
            nextCursor: meta.paging.next_cursor,
            hasMore: meta.paging.has_more,
          },
        };
      } catch {
        return signal.aborted
          ? { kind: "aborted" }
          : {
              kind: "rejected",
              failure: {
                kind: responseAccepted ? "invalid_contract" : "retryable",
                message: responseAccepted
                  ? "Target page was invalid. Restart target discovery."
                  : "Targets could not be loaded. Retry the read.",
              },
            };
      }
    },
  };
}
