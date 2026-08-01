import type {
  QueryWorkbookViewRequest,
  QueryWorkbookViewResponse,
} from "@cartulary/protocol-ts";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { buildQueryRequest } from "../models/workbookQuery";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../mutations/workbookOperationOutcome";
import type {
  WorkbookViewQueryPort,
  WorkbookViewQueryResult,
} from "../query/WorkbookViewQueryPort";
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
}): WorkbookViewQueryPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async query(input) {
      const viewSchemaId = input.contract.viewSchemaId;
      let outcome: WorkbookOperationOutcome<QueryWorkbookViewResponse>;
      try {
        outcome = await operations.execute({
          operationID: "queryWorkbookView",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: viewSchemaId,
          },
          request: buildQueryRequest(
            input.contract,
            input.queryState,
          ) as QueryWorkbookViewRequest,
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
        return {
          kind: "accepted",
          value: {
            incidentId: outcome.value.data.incident_id,
            rows: normalizeWorkbookViewRows(
              input.contract,
              outcome.value.data.rows,
              `${viewSchemaId} query response`,
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
