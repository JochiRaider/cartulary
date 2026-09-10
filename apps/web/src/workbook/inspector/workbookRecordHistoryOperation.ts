import type { HistoryIntent } from "../history/workbookHistoryOperation";
import { historyOperationLabel } from "../history/workbookHistoryOperation";
import { workbookInspectorMessageFeedback } from "./workbookInspectorErrorModel";

export function workbookRecordHistoryCompletionFeedback(intent: HistoryIntent) {
  return workbookInspectorMessageFeedback(
    `${historyOperationLabel(intent)} completed.`,
    "polite",
  );
}
