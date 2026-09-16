import { networkAnalysisSheetRef } from "../extensions/extensionWorkspaceIdentities";
import type { WorkbookRecoveryItem } from "../shared/workbookRecoveryNavigation";
import type { TableSnapshot } from "./NetworkFlowTableController";

export function networkFlowTableRecoveryItems(
  state: TableSnapshot,
): readonly WorkbookRecoveryItem[] {
  if (state.hidden) return [];
  const items = new Map<string, WorkbookRecoveryItem>();
  const draft = state.draft;
  if (draft)
    items.set(String(draft.id), {
      id: String(draft.id),
      label: draft.action === "rename" ? "Rename table" : "Delete table",
      summary:
        draft.reviewRequired || draft.error
          ? "Review retained table change"
          : "Table draft retained",
      origin: draft.target.display_name,
      sheetRef: networkAnalysisSheetRef(),
      attention: draft.reviewRequired || draft.error ? "attention" : "draft",
      order: draft.id,
    });
  const operation = state.operation;
  if (operation) {
    const completed =
      operation.status === "acknowledged" && state.loadState === "ready";
    items.set(String(operation.dialogId), {
      id: String(operation.dialogId),
      label:
        operation.attempt.action === "rename" ? "Rename table" : "Delete table",
      summary: completed
        ? "Table change saved"
        : operation.status === "acknowledged"
          ? "Saved; refresh required"
          : operation.status === "uncertain"
            ? "Outcome unconfirmed"
            : operation.status === "rejected"
              ? "Review rejected change"
              : "Saving table change",
      origin: operation.attempt.target.display_name,
      sheetRef: networkAnalysisSheetRef(),
      attention: completed
        ? "completed"
        : operation.status === "pending"
          ? "progress"
          : "attention",
      order: operation.dialogId,
    });
  }
  return [...items.values()];
}
