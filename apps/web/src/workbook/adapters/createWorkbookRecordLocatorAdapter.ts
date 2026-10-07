import type { HTTPOperationRequest } from "@cartulary/protocol-ts/http";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { buildQueryRequest } from "../models/workbookQuery";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import type { WorkbookRecordLocatorPort } from "../query/WorkbookRecordLocatorPort";
import { readWorkbookQueryMetadata } from "../query/workbookQueryMetadata";
import { sameWorkbookReadScope } from "../query/workbookRowObservation";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

export function createWorkbookRecordLocatorAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly readScope: WorkbookReadScopeSource;
}): WorkbookRecordLocatorPort {
  const operations = createWorkbookOperationExecutor(options);
  const invalid = () => ({
    kind: "rejected" as const,
    failure: {
      kind: "invalid_contract" as const,
      message: "Record navigation response could not be verified.",
    },
  });
  return {
    async locate(input) {
      const scope = options.readScope();
      if (!scope || input.signal.aborted) return { kind: "aborted" };
      if (
        input.queryState.sort.length > 8 ||
        input.queryState.filters.length > 16
      )
        return invalid();
      try {
        const outcome = await operations.execute({
          operationID: "locateWorkbookViewRecord",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: input.contract.viewSchemaId,
          },
          request: {
            ...buildQueryRequest(input.contract, input.queryState),
            record_id: input.recordId,
          } as HTTPOperationRequest<"locateWorkbookViewRecord">,
          signal: input.signal,
        });
        if (
          input.signal.aborted ||
          !sameWorkbookReadScope(scope, options.readScope())
        )
          return { kind: "aborted" };
        if (outcome.kind === "rejected") return outcome;
        const { data, meta } = outcome.value;
        if (data.outcome === "unavailable")
          return { kind: "accepted", value: { outcome: "unavailable" } };
        if (data.target_record_id !== input.recordId) return invalid();
        if (data.outcome === "outside_query")
          return { kind: "accepted", value: { outcome: "outside_query" } };
        if (
          data.incident_id !== options.incidentId ||
          data.view_schema_id !== input.contract.viewSchemaId ||
          data.rows.length < 1 ||
          data.rows.length > 100 ||
          data.rows[0]?.record_id !== input.recordId
        )
          return invalid();
        const producingRequest = {
          queryState: structuredClone(input.queryState),
          limit: 100,
          ...(data.window_start_cursor === null
            ? {}
            : { cursorToken: data.window_start_cursor }),
        };
        const metadata = readWorkbookQueryMetadata(
          input.contract,
          meta,
          input.queryState,
          100,
          producingRequest.cursorToken,
        );
        const rows = normalizeWorkbookViewRows(
          input.contract,
          data.rows,
          "record locator",
        );
        if (new Set(rows.map((row) => row.record_id)).size !== rows.length)
          return invalid();
        return {
          kind: "accepted",
          value: {
            outcome: "located",
            page: {
              incidentId: options.incidentId,
              viewSchemaId: input.contract.viewSchemaId,
              rows: rows.map((row) => acceptWorkbookRowObservation(row, scope)),
              ...metadata,
              producingRequest,
            },
          },
        };
      } catch {
        if (
          input.signal.aborted ||
          !sameWorkbookReadScope(scope, options.readScope())
        )
          return { kind: "aborted" };
        return invalid();
      }
    },
  };
}
