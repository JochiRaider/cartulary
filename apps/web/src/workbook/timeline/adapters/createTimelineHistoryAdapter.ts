import type { RollbackRecordRequest } from "@cartulary/protocol-ts";
import { createWorkbookOperationExecutor } from "../../adapters/workbookOperationExecutor";
import type {
  WorkbookOperationFailure,
  WorkbookOperationOutcome,
} from "../../mutations/workbookOperationOutcome";
import type {
  RecordHistoryData,
  RecordHistoryItem,
  RecordHistoryRollbackAction,
} from "../models/timelineHistoryModel";
import type {
  TimelineHistoryMutationAccepted,
  TimelineHistoryPort,
} from "../ports/TimelineHistoryPort";

const invalidHistoryContract: WorkbookOperationFailure = {
  kind: "invalid_contract",
  message: "Invalid row history response.",
};

function invalidContract<T>(): WorkbookOperationOutcome<T> {
  return { kind: "rejected", failure: invalidHistoryContract };
}

function retryable<T>(): WorkbookOperationOutcome<T> {
  return {
    kind: "rejected",
    failure: {
      kind: "retryable",
      message: "The row history operation could not be sent.",
    },
  };
}

const rollbackActionOrder = [
  "history_entry",
  "change_set",
  "row_restore",
] as const satisfies readonly RecordHistoryRollbackAction[];

function validItemSelector(
  item: RecordHistoryItem,
  action: RecordHistoryRollbackAction,
): boolean {
  if (action === "history_entry") {
    return (
      typeof item.history_entry_ref === "string" &&
      item.history_entry_ref !== ""
    );
  }
  if (action === "change_set") return item.change_set_id !== "";
  return Number.isInteger(item.revision_no) && (item.revision_no ?? 0) > 0;
}

function normalizeHistory(data: RecordHistoryData): RecordHistoryData | null {
  const seen = new Set<string>();
  for (const item of data.items) {
    if (
      item.history_item_ref.trim() === "" ||
      item.change_set_id.trim() === "" ||
      seen.has(item.history_item_ref)
    ) {
      return null;
    }
    seen.add(item.history_item_ref);
    let previous = -1;
    for (const action of item.available_rollback_actions) {
      const index = rollbackActionOrder.indexOf(action);
      if (index <= previous || !validItemSelector(item, action)) return null;
      previous = index;
    }
  }
  return data;
}

export function createTimelineHistoryAdapter(options: {
  readonly apiBase: string | undefined;
}): TimelineHistoryPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async load({ recordId }) {
      try {
        const outcome = await operations.execute({
          operationID: "getRecordHistory",
          pathParameters: { record_id: recordId },
        });
        if (outcome.kind === "rejected") {
          return outcome.failure.kind === "invalid_contract"
            ? invalidContract()
            : outcome;
        }
        const data = normalizeHistory(outcome.value.data);
        return data === null || data.record_id !== recordId
          ? invalidContract()
          : { kind: "accepted", value: data };
      } catch {
        return retryable();
      }
    },
    async deleteOrRestore(input) {
      try {
        const outcome = await operations.execute({
          operationID:
            input.operation === "delete" ? "deleteRecord" : "restoreRecord",
          pathParameters: { record_id: input.recordId },
          request: {
            base_row_version: input.baseRowVersion,
            client_txn_id: input.clientTxnId,
            reason:
              input.operation === "delete"
                ? "Deleted from workbook history"
                : "Restored from workbook history",
          },
        });
        return normalizeMutation(outcome, input.recordId);
      } catch {
        return retryable();
      }
    },
    async rollback(input) {
      try {
        const outcome = await operations.execute({
          operationID: "rollbackRecord",
          pathParameters: { record_id: input.recordId },
          request: {
            base_row_version: input.baseRowVersion,
            client_txn_id: input.clientTxnId,
            reason: "Rollback from workbook history",
            target: input.target as unknown as RollbackRecordRequest["target"],
          },
        });
        return normalizeMutation(outcome, input.recordId);
      } catch {
        return retryable();
      }
    },
  };
}

function normalizeMutation(
  outcome: WorkbookOperationOutcome<{
    readonly data: { readonly record_id: string; readonly row_version: number };
  }>,
  recordId: string,
): WorkbookOperationOutcome<TimelineHistoryMutationAccepted> {
  if (outcome.kind === "rejected") return outcome;
  return outcome.value.data.record_id === recordId
    ? {
        kind: "accepted",
        value: {
          recordId: outcome.value.data.record_id,
          rowVersion: outcome.value.data.row_version,
        },
      }
    : invalidContract();
}
