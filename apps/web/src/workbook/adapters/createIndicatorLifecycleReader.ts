import { requireViewContract } from "@cartulary/view-contracts";
import {
  lifecycleReadTimestamp,
  lifecycleTimeKey,
} from "../features/indicators/indicatorLifecycleModel";
import type { IndicatorLifecycleReadPort } from "../features/indicators/indicatorLifecycleOperation";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { buildQueryRequest } from "../models/workbookQuery";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { sameWorkbookReadScope } from "../query/workbookRowObservation";
import { indicatorLifecycleConstraints } from "./indicatorLifecycleProtocol";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

export function validLifecyclePaging(
  paging:
    | { has_more: boolean; next_cursor: string | null; limit: number }
    | undefined,
  cursor: string | null,
): boolean {
  return (
    !!paging &&
    paging.limit === indicatorLifecycleConstraints.page.default &&
    (paging.has_more
      ? !!paging.next_cursor && paging.next_cursor !== cursor
      : paging.next_cursor === null)
  );
}
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
export function createIndicatorLifecycleReader(options: {
  apiBase: string | undefined;
  incidentId: string;
  readScope?: WorkbookReadScopeSource;
}): IndicatorLifecycleReadPort {
  const operations = createWorkbookOperationExecutor(options);
  return {
    async intervals(recordId, cursor, signal) {
      try {
        const result = await operations.execute({
          operationID: "listIndicatorStateIntervals",
          pathParameters: { indicator_id: recordId },
          query: {
            limit: indicatorLifecycleConstraints.page.default,
            ...(cursor === null ? {} : { cursor_token: cursor }),
          },
          signal,
        });
        if (signal.aborted) return { kind: "aborted" };
        if (result.kind === "rejected") return result;
        const { data, meta } = result.value;
        if (!meta.paging || !validLifecyclePaging(meta.paging, cursor))
          return invalid;
        let previous: (typeof data.intervals)[number] | undefined;
        const intervals: typeof data.intervals = [];
        for (const raw of data.intervals) {
          const from = lifecycleReadTimestamp(raw.valid_from),
            to =
              raw.valid_to === null
                ? null
                : lifecycleReadTimestamp(raw.valid_to),
            created = lifecycleReadTimestamp(raw.created_at),
            assessed = lifecycleReadTimestamp(raw.assessed_at);
          if (!from || (raw.valid_to !== null && !to) || !created || !assessed)
            return invalid;
          const item = {
            ...raw,
            valid_from: from,
            valid_to: to,
            created_at: created,
            assessed_at: assessed,
          };
          if (
            item.incident_id !== options.incidentId ||
            item.indicator_record_id !== recordId ||
            (item.valid_to !== null &&
              lifecycleTimeKey(item.valid_to) <
                lifecycleTimeKey(item.valid_from))
          )
            return invalid;
          if (
            previous &&
            (lifecycleTimeKey(item.valid_from) >
              lifecycleTimeKey(previous.valid_from) ||
              (item.valid_from === previous.valid_from &&
                item.interval_id > previous.interval_id))
          )
            return invalid;
          previous = item;
          intervals.push(item);
        }
        return {
          kind: "accepted",
          value: {
            items: intervals,
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
            limit: indicatorLifecycleConstraints.page.default,
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
          !validLifecyclePaging(meta.paging, cursor)
        )
          return invalid;
        return {
          kind: "accepted",
          value: {
            items: normalizeWorkbookViewRows(
              contract,
              data.rows,
              "Indicator supporting records",
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
