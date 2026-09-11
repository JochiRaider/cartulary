import type { IndicatorLifecycleReceipt } from "../../adapters/indicatorLifecycleProtocol";
import type { WorkbookRecordHistoryOwner } from "../../history/WorkbookRecordHistoryOwner";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import { workbookOperationFailureIsAccessLoss } from "../../ports/WorkbookPortResult";
import { indicatorLifecycleViewId } from "./indicatorLifecycleModel";
import type { LifecycleScope } from "./indicatorLifecycleOperation";
import type { WorkbookIndicatorLifecycleOwner } from "./WorkbookIndicatorLifecycleOwner";

/** Receipts are already accepted. Every failed read leaves reconciliation debt. */
export async function reconcileIndicatorLifecycleReceipt(
  owner: WorkbookIndicatorLifecycleOwner,
  history: WorkbookRecordHistoryOwner,
  receipt: IndicatorLifecycleReceipt,
  scope: LifecycleScope,
) {
  for (const record of receipt.affected_records) {
    if (!scope.isCurrent()) throw new Error("Interval reconciliation detached");
    const result = await history.load(record.record_id, scope.signal);
    if (
      scope.isCurrent() &&
      result.kind === "rejected" &&
      workbookOperationFailureIsAccessLoss(result.failure)
    )
      owner.loseAccess();
    if (
      !scope.isCurrent() ||
      result.kind !== "accepted" ||
      result.value.incident_id !== owner.incidentId ||
      result.value.record_id !== record.record_id ||
      result.value.row_version <
        Math.max(record.row_version, owner.latestVersion(record.record_id) ?? 0)
    )
      throw new Error("Affected record history needs refresh");
    owner.acceptVersion(record.record_id, result.value.row_version);
    await history.refreshRecordPresentation(record.record_id);
    if (!scope.isCurrent())
      throw new Error("Interval history presentation detached");
    owner.acceptVersion(
      record.record_id,
      history.latestVersion(record.record_id) ?? result.value.row_version,
    );
  }
  const remaining = new Map(
    receipt.affected_records.map((record) => [
      record.record_id,
      Math.max(record.row_version, owner.latestVersion(record.record_id) ?? 0),
    ]),
  );
  const cursors = new Set<string>();
  let cursor: string | null = null;
  do {
    if (!scope.isCurrent()) throw new Error("Interval reconciliation detached");
    const result = await owner.records(
      indicatorLifecycleViewId,
      emptyWorkbookQueryState(),
      cursor,
      scope.signal,
    );
    if (!scope.isCurrent() || result.kind !== "accepted")
      throw new Error("Current Indicators could not be read");
    for (const row of result.value.items) {
      if (
        !remaining.has(row.record_id) ||
        row.row_version < (owner.latestVersion(row.record_id) ?? 0)
      )
        continue;
      owner.acceptRow(row);
      if (row.row_version >= (remaining.get(row.record_id) ?? Infinity))
        remaining.delete(row.record_id);
    }
    if (!remaining.size) break;
    cursor = result.value.nextCursor;
    if (!result.value.hasMore || !cursor || cursors.has(cursor))
      throw new Error("Affected Indicators still need refresh");
    cursors.add(cursor);
  } while (scope.isCurrent());
  if (
    remaining.size ||
    !scope.isCurrent() ||
    receipt.affected_records.some(
      (record) =>
        (owner.latestRow(record.record_id)?.row_version ?? 0) <
        (owner.latestVersion(record.record_id) ?? 0),
    )
  )
    throw new Error("Newer Indicator evidence needs refresh");
}
