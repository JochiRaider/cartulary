import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";

export function createTimelineSocketTransactionAdapter(
  mutationRuntime: WorkbookMutationRuntime,
) {
  return {
    resolve: (clientTxnId: string | null | undefined) =>
      mutationRuntime.resolveSocketClientTxn(clientTxnId),
    track: (clientTxnId: string) =>
      mutationRuntime.rememberClientTransaction(clientTxnId),
  };
}
