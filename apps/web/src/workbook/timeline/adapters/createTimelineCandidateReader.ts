import { requireViewContract } from "@cartulary/view-contracts";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import { normalizeWorkbookViewRows } from "../../models/workbookContractRows";
import {
  buildQueryRequest,
  emptyWorkbookQueryState,
} from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineCandidatePort } from "../actions/TimelineCandidatePort";
import { timelineCaptureSubject } from "../actions/timelineCaptureActionModel";

export function createTimelineCandidateReader(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): TimelineCandidatePort {
  const operations = createWorkbookOperationExecutor(options);
  const contract = requireViewContract(timelineViewSchemaId);
  return {
    async page(cursor, signal) {
      try {
        const result = await operations.execute({
          operationID: "queryWorkbookView",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: timelineViewSchemaId,
          },
          request: {
            ...buildQueryRequest(contract, emptyWorkbookQueryState()),
            limit: 100,
            ...(cursor === null ? {} : { cursor_token: cursor }),
          },
          signal,
        });
        if (signal.aborted) return { kind: "aborted" };
        if (result.kind === "rejected") return result;
        const { data, meta } = result.value;
        if (
          data.incident_id !== options.incidentId ||
          data.view_schema_id !== timelineViewSchemaId ||
          !meta.paging ||
          data.rows.length > 100 ||
          (meta.paging.has_more
            ? !meta.paging.next_cursor || meta.paging.next_cursor === cursor
            : meta.paging.next_cursor !== null)
        )
          throw new Error("Invalid Timeline candidate page");
        return {
          kind: "accepted",
          value: {
            rows: normalizeWorkbookViewRows(
              contract,
              data.rows,
              "Timeline candidates",
            ).map((row) => timelineCaptureSubject(row, options.incidentId)),
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
                kind: "retryable",
                message:
                  "Timeline replacements could not be loaded. Retry the read.",
              },
            };
      }
    },
  };
}
