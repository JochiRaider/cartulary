import type { HistoryOperation } from "./workbookHistoryOperation";

export function historyOperationStatus(entry: HistoryOperation): string {
  switch (entry.phase) {
    case "preparing":
      return "Checking the confirmed action.";
    case "submitting":
      return "Waiting for the action result.";
    case "uncertain":
      return "Outcome unknown. The action may have completed. Replay the exact action to recover its result.";
    case "rejected":
      return entry.failure?.kind === "client_txn_conflict"
        ? "A request identity conflict prevented this action. Review current history before confirming a replacement request."
        : "Action not completed. Review current history before trying again.";
    case "acknowledged":
      return entry.reconciliation === "required"
        ? "Action completed; refresh required."
        : entry.reconciliation === "complete"
          ? "Action completed."
          : "Action completed; refreshing the workbook.";
  }
}
