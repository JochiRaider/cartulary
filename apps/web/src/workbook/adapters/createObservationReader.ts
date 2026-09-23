import { requireViewContract } from "@cartulary/view-contracts";
import {
  observationOrder,
  observationTimeKey,
} from "../features/indicators/observationModel";
import type { ObservationReadPort } from "../features/indicators/observationOperation";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { buildQueryRequest } from "../models/workbookQuery";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { sameWorkbookReadScope } from "../query/workbookRowObservation";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

/** Explicit page size admitted by both existing collection and query contracts. */
const observationPageSize = 100;
const invalid = {
  kind: "rejected" as const,
  failure: {
    kind: "invalid_contract" as const,
    message:
      "The returned collection could not be verified. Restart this read.",
  },
};
const failed = {
  kind: "rejected" as const,
  failure: {
    kind: "retryable" as const,
    message: "The collection could not be loaded. Retry this read.",
  },
};
function validObservationPaging(
  paging:
    | { has_more: boolean; next_cursor: string | null; limit: number }
    | undefined,
  cursor: string | null,
) {
  return (
    !!paging &&
    paging.limit === observationPageSize &&
    (paging.has_more
      ? !!paging.next_cursor && paging.next_cursor !== cursor
      : paging.next_cursor === null)
  );
}
export function createObservationReader(options: {
  apiBase: string | undefined;
  incidentId: string;
  readScope?: WorkbookReadScopeSource;
}): ObservationReadPort {
  const operations = createWorkbookOperationExecutor(options);
  return {
    async observations(subject, cursor, signal) {
      try {
        const query = {
          limit: observationPageSize,
          ...(cursor === null ? {} : { cursor_token: cursor }),
        };
        const result =
          subject.kind === "source"
            ? await operations.execute({
                operationID: "listSourceRecordIndicatorObservations",
                pathParameters: { source_record_id: subject.recordId },
                query,
                signal,
              })
            : await operations.execute({
                operationID: "listIndicatorObservations",
                pathParameters: { indicator_id: subject.recordId },
                query,
                signal,
              });
        if (signal.aborted) return { kind: "aborted" };
        if (result.kind === "rejected") return result;
        const { data, meta } = result.value;
        if (!meta.paging || !validObservationPaging(meta.paging, cursor))
          return invalid;
        for (const [index, item] of data.observations.entries()) {
          if (
            item.incident_id !== options.incidentId ||
            (subject.kind === "source"
              ? item.source_record_id !== subject.recordId
              : item.resolution_status !== "resolved" ||
                item.resolved_indicator_record_id !== subject.recordId) ||
            !observationTimeKey(item.created_at) ||
            (item.resolved_at !== null && !observationTimeKey(item.resolved_at))
          )
            return invalid;
          const previous = data.observations[index - 1];
          if (previous && !observationOrder(previous, item)) return invalid;
        }
        return {
          kind: "accepted",
          value: {
            items: data.observations,
            hasMore: meta.paging.has_more,
            nextCursor: meta.paging.next_cursor,
          },
        };
      } catch {
        return signal.aborted ? { kind: "aborted" } : failed;
      }
    },
    async records(viewSchemaId, query, cursor, signal) {
      const scope = options.readScope?.() ?? null;
      try {
        const contract = requireViewContract(viewSchemaId);
        const result = await operations.execute({
          operationID: "queryWorkbookView",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: viewSchemaId,
          },
          request: {
            ...buildQueryRequest(contract, query),
            limit: observationPageSize,
            ...(cursor === null ? {} : { cursor_token: cursor }),
          },
          signal,
        });
        if (
          signal.aborted ||
          (options.readScope &&
            !sameWorkbookReadScope(scope, options.readScope()))
        )
          return { kind: "aborted" };
        if (result.kind === "rejected") return result;
        const { data, meta } = result.value;
        if (
          data.incident_id !== options.incidentId ||
          data.view_schema_id !== viewSchemaId ||
          !meta.paging ||
          !validObservationPaging(meta.paging, cursor)
        )
          return invalid;
        return {
          kind: "accepted",
          value: {
            items: normalizeWorkbookViewRows(
              contract,
              data.rows,
              "Observation records",
            ).map((row) => acceptWorkbookRowObservation(row, scope)),
            hasMore: meta.paging.has_more,
            nextCursor: meta.paging.next_cursor,
          },
        };
      } catch {
        return signal.aborted ? { kind: "aborted" } : failed;
      }
    },
  };
}
