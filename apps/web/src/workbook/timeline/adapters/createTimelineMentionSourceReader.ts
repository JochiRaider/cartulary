import type { QueryWorkbookViewResponse } from "@cartulary/protocol-ts/http";
import { requireViewContract } from "@cartulary/view-contracts";
import { boundedRead } from "../../../services/asyncObservation";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import { normalizeWorkbookViewRows } from "../../models/workbookContractRows";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import { acceptWorkbookRowObservation } from "../../query/acceptWorkbookRowObservation";
import type {
  WorkbookQueryRow,
  WorkbookReadScopeSource,
} from "../../query/WorkbookQueryRow";
import { sameWorkbookReadScope } from "../../query/workbookRowObservation";
export type TimelineMentionSourceReader = (
  recordId: string,
  signal: AbortSignal,
) => Promise<WorkbookQueryRow>;
/** Uses the existing paged Timeline view; no mention resource or read endpoint is invented. */
export function createTimelineMentionSourceReader(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly readScope?: WorkbookReadScopeSource;
}): TimelineMentionSourceReader {
  const operations = createWorkbookOperationExecutor(options),
    contract = requireViewContract(timelineViewSchemaId);
  return (recordId, signal) =>
    boundedRead(async (observedSignal) => {
      const scope = options.readScope?.() ?? null;
      let cursor: string | null = null;
      const cursors = new Set<string>();
      do {
        const result: WorkbookOperationOutcome<QueryWorkbookViewResponse> =
          await operations.execute({
            operationID: "queryWorkbookView",
            pathParameters: {
              incident_id: options.incidentId,
              view_schema_id: timelineViewSchemaId,
            },
            request: {
              limit: 100,
              ...(cursor ? { cursor_token: cursor } : {}),
            },
            signal: observedSignal,
          });
        if (
          observedSignal.aborted ||
          result.kind !== "accepted" ||
          (options.readScope &&
            !sameWorkbookReadScope(scope, options.readScope()))
        )
          throw new Error("Mention source refresh failed.");
        const data: QueryWorkbookViewResponse["data"] = result.value.data;
        const meta: QueryWorkbookViewResponse["meta"] = result.value.meta;
        if (
          data.incident_id !== options.incidentId ||
          data.view_schema_id !== timelineViewSchemaId ||
          !meta.paging ||
          data.rows.length > 100
        )
          throw new Error("Invalid mention source projection.");
        const row = normalizeWorkbookViewRows(
          contract,
          data.rows,
          "Mention source",
        ).find((row) => row.record_id === recordId);
        if (row) return acceptWorkbookRowObservation(row, scope);
        if (!meta.paging.has_more) break;
        cursor = meta.paging.next_cursor;
        if (!cursor || cursors.has(cursor))
          throw new Error("Invalid source paging.");
        cursors.add(cursor);
      } while (!observedSignal.aborted);
      throw new Error(
        "The mention source is no longer in the active Timeline. Open History for its changes.",
      );
    }, signal);
}
