import type { ApplyWorkbookBulkMutationRequest } from "@cartulary/protocol-ts/http";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import { normalizeTimelineFullRow } from "../models/timelineRowModel";
import type {
  TimelineBulkTagAccepted,
  TimelineBulkTagCommandPort,
} from "../ports/TimelineBulkTagCommandPort";

export function createTimelineBulkTagCommandAdapter(options: {
  readonly apiBase: string | undefined;
  readonly createClientTxnId: () => string;
  readonly incidentId: string;
}): TimelineBulkTagCommandPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async assignTag(input) {
      let clientTxnId: string;
      try {
        clientTxnId = options.createClientTxnId();
      } catch {
        return identityFailure();
      }
      const request = bulkTagRequest(input, clientTxnId);
      if (request === null) return invalidPayload();
      try {
        const outcome = await operations.execute({
          operationID: "applyWorkbookBulkMutation",
          pathParameters: {
            incident_id: options.incidentId,
            view_schema_id: timelineViewSchemaId,
          },
          request,
        });
        if (outcome.kind === "rejected") return outcome;
        return normalizeBulkTagOutcome(outcome.value.data);
      } catch {
        return retryable();
      }
    },
  };
}

function bulkTagRequest(
  input: Parameters<TimelineBulkTagCommandPort["assignTag"]>[0],
  clientTxnId: string,
): ApplyWorkbookBulkMutationRequest | null {
  const [firstTarget, ...remainingTargets] = input.targets;
  if (input.tagName.trim() === "" || firstTarget === undefined) return null;
  return {
    client_txn_id: clientTxnId,
    kind: "multi_row_tag_assignment_v1",
    tag_name: input.tagName,
    targets: [
      {
        base_row_version: firstTarget.baseRowVersion,
        record_id: firstTarget.recordId,
      },
      ...remainingTargets.map((target) => ({
        base_row_version: target.baseRowVersion,
        record_id: target.recordId,
      })),
    ],
    view_schema_id: timelineViewSchemaId,
  };
}

function normalizeBulkTagOutcome(data: {
  readonly change_set_id?: string | null | undefined;
  readonly conflicts?: readonly unknown[] | undefined;
  readonly rows: readonly unknown[];
  readonly view_schema_id: string;
}): WorkbookOperationOutcome<TimelineBulkTagAccepted> {
  if (data.view_schema_id !== timelineViewSchemaId) return invalidContract();
  try {
    for (const row of data.rows) {
      normalizeTimelineFullRow(row, "bulk tag response row");
    }
  } catch {
    return invalidContract();
  }
  return {
    kind: "accepted",
    value: {
      affectedRowCount: data.rows.length,
      changeSetId: data.change_set_id ?? null,
      conflictCount: data.conflicts?.length ?? 0,
    },
  };
}

function identityFailure(): WorkbookOperationOutcome<TimelineBulkTagAccepted> {
  return {
    kind: "rejected",
    failure: {
      kind: "terminal",
      message: "A secure transaction ID could not be created.",
    },
  };
}

function invalidPayload(): WorkbookOperationOutcome<TimelineBulkTagAccepted> {
  return {
    kind: "rejected",
    failure: { kind: "validation", message: "invalid_mutation_payload" },
  };
}

function invalidContract(): WorkbookOperationOutcome<TimelineBulkTagAccepted> {
  return {
    kind: "rejected",
    failure: {
      kind: "invalid_contract",
      message: "The server returned an inconsistent Workbook operation result.",
    },
  };
}

function retryable(): WorkbookOperationOutcome<TimelineBulkTagAccepted> {
  return {
    kind: "rejected",
    failure: {
      kind: "retryable",
      message: "The Workbook operation could not be sent.",
    },
  };
}
