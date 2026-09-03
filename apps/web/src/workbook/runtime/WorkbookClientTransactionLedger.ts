import type { WorkbookPendingQueueRuntime } from "./workbookPendingReplayRuntime";

const maximumRecentTransactionIds = 128;

export class WorkbookClientTransactionLedger {
  readonly #recent = new Set<string>();

  remember(clientTransactionId: string): void {
    this.#recent.add(clientTransactionId);
    if (this.#recent.size <= maximumRecentTransactionIds) return;
    const oldest = this.#recent.values().next().value;
    if (typeof oldest === "string") this.#recent.delete(oldest);
  }

  settle(
    clientTransactionId: string | null | undefined,
    pending: WorkbookPendingQueueRuntime,
  ): boolean {
    if (!clientTransactionId) return false;
    if (this.#recent.delete(clientTransactionId)) return true;
    return pending.model
      .snapshot()
      .units.some((unit) => unit.clientTxnId === clientTransactionId);
  }
}
