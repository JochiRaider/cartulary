import type {
  QueryWorkbookViewRequest,
  QueryWorkbookViewResponse,
} from "@cartulary/protocol-ts";
import type { ViewContract } from "@cartulary/view-contracts";
import { buildQueryRequest } from "../../models/workbookQuery";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { createWorkbookOperationExecutor } from "../../mutations/workbookOperationExecutor";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../../mutations/workbookOperationOutcome";
import {
  normalizeTimelineFullRow,
  rowFromApi,
  validateTimelineViewSchemaId,
} from "../models/workbookTimelineModel";
import type {
  TimelineViewQueryPort,
  TimelineViewQueryResult,
} from "../ports/TimelineViewQueryPort";

const invalidTimelineProjection: WorkbookOperationFailure = {
  kind: "invalid_contract",
  message: "Timeline projection load failed.",
};

function invalidProjection(): TimelineViewQueryResult {
  return { kind: "rejected", failure: invalidTimelineProjection };
}

export function createTimelineViewQueryAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly timelineContract: ViewContract;
}): TimelineViewQueryPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async query(input) {
      let outcome: WorkbookOperationOutcome<QueryWorkbookViewResponse>;
      try {
        outcome = await operations.execute({
          operationID: "queryWorkbookView",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: timelineViewSchemaId,
          },
          request: buildQueryRequest(
            options.timelineContract,
            input.queryState,
          ) as QueryWorkbookViewRequest,
          signal: input.signal,
        });
      } catch (error) {
        if (
          input.signal.aborted ||
          (error instanceof DOMException && error.name === "AbortError")
        ) {
          return { kind: "aborted" };
        }
        return {
          kind: "rejected",
          failure: {
            kind: "retryable",
            message: "Timeline projection load failed.",
          },
        };
      }
      if (outcome.kind === "rejected") {
        return outcome.failure.kind === "invalid_contract"
          ? invalidProjection()
          : outcome;
      }
      if (
        outcome.value.data.incident_id !== options.incidentId ||
        outcome.value.data.view_schema_id !== timelineViewSchemaId
      ) {
        return invalidProjection();
      }
      try {
        validateTimelineViewSchemaId(
          outcome.value.data.view_schema_id,
          "query response",
        );
        return {
          kind: "accepted",
          value: {
            incidentId: outcome.value.data.incident_id,
            rows: outcome.value.data.rows.map((row, index) =>
              rowFromApi(
                normalizeTimelineFullRow(row, `query response rows[${index}]`),
              ),
            ),
            viewSchemaId: outcome.value.data.view_schema_id,
          },
        };
      } catch {
        return invalidProjection();
      }
    },
  };
}
