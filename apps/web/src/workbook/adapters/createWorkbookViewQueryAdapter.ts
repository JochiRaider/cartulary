import type { QueryWorkbookViewResponse } from "@cartulary/protocol-ts/http";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { buildQueryRequest } from "../models/workbookQuery";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScope } from "../query/WorkbookQueryRow";
import type {
  WorkbookViewQueryPort,
  WorkbookViewQueryResult,
} from "../query/WorkbookViewQueryPort";
import { readWorkbookQueryMetadata } from "../query/workbookQueryMetadata";
import { sameWorkbookReadScope } from "../query/workbookRowObservation";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

const invalidProjectionFailure: WorkbookOperationFailure = {
  kind: "invalid_contract",
  message: "Workbook view load failed.",
};

function invalidProjection(): WorkbookViewQueryResult {
  return { kind: "rejected", failure: invalidProjectionFailure };
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException
    ? error.name === "AbortError"
    : error instanceof Error && error.name === "AbortError";
}

export function createWorkbookViewQueryAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly readScope?: () => WorkbookReadScope | null;
}): WorkbookViewQueryPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    readScope: options.readScope,
    async query(input) {
      const scope = options.readScope?.() ?? null;
      if (options.readScope && !scope) return { kind: "aborted" };
      const viewSchemaId = input.contract.viewSchemaId;
      const limit = input.limit ?? 100;
      if (!Number.isInteger(limit) || limit < 1 || limit > 500)
        return invalidProjection();
      const producingRequest = {
        queryState: structuredClone(input.queryState),
        limit,
        ...(input.cursorToken === undefined
          ? {}
          : { cursorToken: input.cursorToken }),
        ...(input.identity === undefined
          ? {}
          : { identity: { ...input.identity } }),
      };
      let outcome: WorkbookOperationOutcome<QueryWorkbookViewResponse>;
      try {
        outcome = await operations.execute({
          operationID: "queryWorkbookView",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: viewSchemaId,
          },
          request: {
            ...buildQueryRequest(input.contract, producingRequest.queryState),
            ...(input.limit === undefined ? {} : { limit }),
            ...(input.cursorToken === undefined
              ? {}
              : { cursor_token: input.cursorToken }),
          },
          signal: input.signal,
        });
      } catch (error) {
        if (input.signal.aborted || isAbortError(error)) {
          return { kind: "aborted" };
        }
        return {
          kind: "rejected",
          failure: {
            kind: "retryable",
            message: "Workbook view load failed.",
          },
        };
      }
      if (
        options.readScope &&
        !sameWorkbookReadScope(scope, options.readScope())
      )
        return { kind: "aborted" };
      if (outcome.kind === "rejected") {
        return outcome.failure.kind === "invalid_contract"
          ? invalidProjection()
          : outcome;
      }
      if (
        outcome.value.data.incident_id !== options.incidentId ||
        outcome.value.data.view_schema_id !== viewSchemaId
      ) {
        return invalidProjection();
      }
      try {
        if (outcome.value.data.rows.length > limit) return invalidProjection();
        const metadata = readWorkbookQueryMetadata(
          input.contract,
          outcome.value.meta,
          producingRequest.queryState,
          limit,
          input.cursorToken,
          input.expectedCanonicalQuery,
        );
        const rows = normalizeWorkbookViewRows(
          input.contract,
          outcome.value.data.rows,
          `${viewSchemaId} query response`,
        );
        if (new Set(rows.map((row) => row.record_id)).size !== rows.length)
          return invalidProjection();
        return {
          kind: "accepted",
          value: {
            incidentId: outcome.value.data.incident_id,
            rows: rows.map((row) => acceptWorkbookRowObservation(row, scope)),
            viewSchemaId: outcome.value.data.view_schema_id,
            ...metadata,
            producingRequest,
          },
        };
      } catch {
        return invalidProjection();
      }
    },
  };
}
