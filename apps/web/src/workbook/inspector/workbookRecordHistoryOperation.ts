import type {
  RecordLifecycleAccepted,
  RecordRouteCommandPort,
} from "../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import { workbookInspectorMessageFeedback } from "./workbookInspectorErrorModel";
import type { WorkbookRecordHistoryPendingAction } from "./workbookRecordHistoryModel";
import type { WorkbookRecordHistoryOwnerEffects } from "./workbookRecordHistoryOwnerEffects";

export function executeWorkbookRecordHistoryOperation(
  commands: RecordRouteCommandPort,
  pending: WorkbookRecordHistoryPendingAction,
): Promise<WorkbookOperationOutcome<RecordLifecycleAccepted>> {
  if (pending.kind === "rollback") {
    return commands.rollback({
      baseRowVersion: pending.rowVersion,
      reason: `Rollback ${pending.action} from the workbook inspector`,
      recordId: pending.recordId,
      target: pending.target,
    });
  }
  return commands.execute({
    action: pending.operation,
    baseRowVersion: pending.rowVersion,
    reason:
      pending.operation === "delete"
        ? "Deleted from the workbook inspector"
        : "Restored from the workbook inspector",
    recordId: pending.recordId,
  });
}

export async function applyWorkbookRecordHistoryOwnerEffect(
  ownerEffects: WorkbookRecordHistoryOwnerEffects,
  pending: WorkbookRecordHistoryPendingAction,
  accepted: RecordLifecycleAccepted,
): Promise<void> {
  if (pending.kind === "rollback") {
    await ownerEffects.rollbackAccepted(accepted);
  } else if (pending.operation === "delete") {
    await ownerEffects.deleteAccepted(accepted);
  } else {
    await ownerEffects.restoreAccepted(accepted);
  }
}

export function workbookRecordHistoryCompletionFeedback(
  pending: WorkbookRecordHistoryPendingAction,
  accepted: RecordLifecycleAccepted,
) {
  const message =
    pending.kind === "rollback"
      ? `Rolled back record ${accepted.recordId}.`
      : pending.operation === "delete"
        ? `Deleted record ${accepted.recordId}.`
        : `Restored record ${accepted.recordId}.`;
  return workbookInspectorMessageFeedback(message, "polite");
}
