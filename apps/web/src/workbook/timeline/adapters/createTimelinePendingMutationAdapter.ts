import type {
  CreateViewRowRequest,
  CreateViewRowResponse,
  PatchRecordRequest,
} from "@cartulary/protocol-ts";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { createWorkbookOperationExecutor } from "../../mutations/workbookOperationExecutor";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../../mutations/workbookOperationOutcome";
import { normalizeTimelineFullRow } from "../models/workbookTimelineModel";
import type {
  TimelinePendingMutationOutcome,
  TimelinePendingMutationPort,
} from "../ports/TimelinePendingMutationPort";

const invalidMutationContract: WorkbookOperationFailure = {
  kind: "invalid_contract",
  message: "The Timeline mutation response was invalid.",
};

function invalidContract(): TimelinePendingMutationOutcome {
  return { kind: "rejected", failure: invalidMutationContract };
}

export function createTimelinePendingMutationAdapter(options: {
  readonly apiBase: string | undefined;
  readonly recordTiming?: (
    name: string,
    details?: Readonly<Record<string, unknown>>,
  ) => void;
}): TimelinePendingMutationPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    normalizeResolvedConflict(input) {
      if (input.viewSchemaId !== timelineViewSchemaId) {
        return invalidContract();
      }
      try {
        const row = normalizeTimelineFullRow(
          input.row,
          "conflict resolution response row",
        );
        return row.record_id === input.expectedRecordId
          ? {
              kind: "accepted",
              value: { row, viewSchemaId: input.viewSchemaId },
            }
          : invalidContract();
      } catch {
        return invalidContract();
      }
    },
    async execute({ committedRowVersion, unit }) {
      let outcome: WorkbookOperationOutcome<CreateViewRowResponse>;
      const timingDetails = {
        clientTxnId: unit.clientTxnId,
        kind: unit.kind,
        rowKey: unit.rowKey,
      };
      const observeTransport = {
        onJSONParsed: () => {
          options.recordTiming?.("pending_fetch_json_parsed", timingDetails);
        },
        onResponseStatus: (status: number) => {
          options.recordTiming?.("pending_fetch_response", {
            ...timingDetails,
            status,
          });
        },
      };
      try {
        outcome =
          unit.kind === "create"
            ? await operations.execute({
                operationID: "createViewRow",
                observeTransport,
                pathParameters: {
                  incident_id: unit.incidentId,
                  view_schema_id: unit.viewSchemaId,
                },
                request: {
                  ...unit.payloadIntent,
                  client_txn_id: unit.clientTxnId,
                } as CreateViewRowRequest,
              })
            : committedRowVersion === null || unit.recordId === null
              ? ({
                  kind: "rejected",
                  failure: {
                    kind: "stale_target",
                    message: "The Timeline row is no longer available.",
                  },
                } as const)
              : await operations.execute({
                  operationID: "patchRecord",
                  observeTransport,
                  pathParameters: { record_id: unit.recordId },
                  request: {
                    view_schema_id: unit.viewSchemaId,
                    base_row_version: committedRowVersion,
                    client_txn_id: unit.clientTxnId,
                    changes: Array.isArray(unit.payloadIntent.changes)
                      ? unit.payloadIntent.changes
                      : [],
                  } as PatchRecordRequest,
                });
      } catch {
        return {
          kind: "rejected",
          failure: {
            kind: "retryable",
            message: "The Timeline mutation could not be sent.",
          },
        };
      }
      if (outcome.kind === "rejected") return outcome;
      const response = outcome.value.data;
      if (
        response.view_schema_id !== timelineViewSchemaId ||
        response.view_schema_id !== unit.viewSchemaId ||
        (unit.kind === "patch" && response.row.record_id !== unit.recordId)
      ) {
        return invalidContract();
      }
      try {
        return {
          kind: "accepted",
          value: {
            changeSetId: response.change_set_id,
            row: normalizeTimelineFullRow(
              response.row,
              "pending mutation response row",
            ),
            viewSchemaId: response.view_schema_id,
          },
        };
      } catch {
        return invalidContract();
      }
    },
  };
}
