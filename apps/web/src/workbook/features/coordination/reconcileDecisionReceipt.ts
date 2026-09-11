import type { WorkbookRecordHistoryOwner } from "../../history/WorkbookRecordHistoryOwner";
import { workbookOperationFailureIsAccessLoss } from "../../ports/WorkbookPortResult";
import type { DecisionSupersessionReceipt } from "./decisionSupersessionOperation";
import type {
  DecisionReconciliationScope,
  WorkbookDecisionSupersessionOwner,
} from "./WorkbookDecisionSupersessionOwner";

export async function reconcileDecisionReceipt(
  owner: WorkbookDecisionSupersessionOwner,
  history: WorkbookRecordHistoryOwner,
  receipt: DecisionSupersessionReceipt,
  scope: DecisionReconciliationScope,
): Promise<void> {
  const remaining = new Map([
    [receipt.target_record_id, receipt.target_row_version],
    [receipt.superseding_record_id, receipt.superseding_row_version],
  ]);
  const cursors = new Set<string>();
  let cursor: string | null = null;
  do {
    if (!scope.isCurrent()) throw new Error("Decision reconciliation detached");
    const page = await owner.page(cursor, scope.signal);
    if (!scope.isCurrent() || page.kind !== "accepted")
      throw new Error("Current Decisions could not be read");
    for (const row of page.value.rows) {
      if (row.row_version < (owner.latestVersion(row.record_id) ?? 0)) continue;
      owner.acceptRow(row);
      if (row.row_version >= (remaining.get(row.record_id) ?? Infinity))
        remaining.delete(row.record_id);
    }
    if (!remaining.size) break;
    cursor = page.value.nextCursor;
    if (!page.value.hasMore || !cursor || cursors.has(cursor))
      throw new Error("Both affected Decisions could not be reconciled");
    cursors.add(cursor);
  } while (scope.isCurrent());
  if (remaining.size || !scope.isCurrent())
    throw new Error("Decision reconciliation incomplete");
  for (const [id, version] of [
    [receipt.target_record_id, receipt.target_row_version],
    [receipt.superseding_record_id, receipt.superseding_row_version],
  ] as const) {
    const result = await history.load(id, scope.signal);
    if (
      scope.isCurrent() &&
      result.kind === "rejected" &&
      workbookOperationFailureIsAccessLoss(result.failure)
    )
      owner.loseAccess();
    if (
      !scope.isCurrent() ||
      result.kind !== "accepted" ||
      result.value.row_version < Math.max(version, owner.latestVersion(id) ?? 0)
    )
      throw new Error("Affected Decision history still needs refresh");
    owner.acceptVersion(id, result.value.row_version);
    await history.refreshRecordPresentation(id);
    owner.acceptVersion(
      id,
      history.latestVersion(id) ?? result.value.row_version,
    );
  }
  for (const id of [receipt.target_record_id, receipt.superseding_record_id])
    if (
      (owner.latestRow(id)?.row_version ?? 0) < (owner.latestVersion(id) ?? 0)
    )
      throw new Error("A newer Decision version needs another refresh");
}
