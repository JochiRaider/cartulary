import type { HistoryReceipt } from "../history/workbookHistoryOperation";

export type WorkbookRecordHistoryOwnerEffects = {
  readonly refresh: () => Promise<void> | void;
  readonly deleteAccepted: (accepted: HistoryReceipt) => void;
  readonly restoreAccepted: (accepted: HistoryReceipt) => void;
  readonly rollbackAccepted: (accepted: HistoryReceipt) => void;
};
