import { requireViewContract } from "@cartulary/view-contracts";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import { normalizeWorkbookViewRows } from "../../models/workbookContractRows";
import {
  buildQueryRequest,
  emptyWorkbookQueryState,
} from "../../models/workbookQuery";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import type { TimelineMentionCandidatePort } from "../actions/TimelineMentionCandidatePort";

/** Eligible target discovery has no dependency on the currently displayed entity sheet. */
export function createTimelineMentionCandidateReader(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}): TimelineMentionCandidatePort {
  const operations = createWorkbookOperationExecutor(options);
  return {
    async page(entityType, cursor, signal) {
      if (signal.aborted) return { kind: "aborted" };
      const viewSchemaId =
        entityType === "host" ? hostsViewSchemaId : identitiesViewSchemaId;
      const contract = requireViewContract(viewSchemaId);
      const stateField = `${entityType}.${entityType}_state`;
      try {
        if (!contract.fieldMap[stateField]?.filterOps.includes("eq"))
          throw new Error("Eligible target filter unavailable");
        const result = await operations.execute({
          operationID: "queryWorkbookView",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: viewSchemaId,
          },
          request: {
            ...buildQueryRequest(contract, {
              ...emptyWorkbookQueryState(),
              filters: [
                {
                  fieldKey: stateField,
                  op: "eq",
                  arg: { values: ["stub", "canonical"] },
                },
              ],
            }),
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
          data.view_schema_id !== viewSchemaId ||
          !meta.paging ||
          data.rows.length > 100 ||
          (meta.paging.has_more
            ? !meta.paging.next_cursor || meta.paging.next_cursor === cursor
            : meta.paging.next_cursor !== null)
        )
          throw new Error("Invalid target page");
        const rows = normalizeWorkbookViewRows(
          contract,
          data.rows,
          "Mention targets",
        );
        if (
          rows.some(
            (row) =>
              !["stub", "canonical"].includes(
                String(row.cells[stateField]?.value),
              ),
          )
        )
          throw new Error("Ineligible target projection");
        return {
          kind: "accepted",
          value: {
            candidates: rows.map((row) => ({
              recordId: row.record_id,
              rowVersion: row.row_version,
              entityType,
              displayText: String(
                row.cells[`${entityType}.display_name`]?.value ?? row.record_id,
              ),
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
                kind: "retryable",
                message: "Targets could not be loaded. Retry the read.",
              },
            };
      }
    },
  };
}
