import type { WorkbookRecordHistoryOwner } from "../../history/WorkbookRecordHistoryOwner";
import { workbookOperationFailureIsAccessLoss } from "../../ports/WorkbookPortResult";
import type { TimelineCaptureReceipt } from "../adapters/timelineCaptureProtocol";
import type {
  TimelineReconciliationScope,
  WorkbookTimelineCaptureActionOwner,
} from "./WorkbookTimelineCaptureActionOwner";

export async function reconcileTimelineCaptureReceipt(
  owner: WorkbookTimelineCaptureActionOwner,
  history: WorkbookRecordHistoryOwner,
  receipt: TimelineCaptureReceipt,
  scope: TimelineReconciliationScope,
) {
  const target = receipt.data.record_id;
  const remaining = new Set([
    target,
    ...(receipt.data.replacement_record_id
      ? [receipt.data.replacement_record_id]
      : []),
  ]);
  const observed = new Map<string, number>();
  const cursors = new Set<string>();
  let cursor: string | null = null;
  do {
    if (!scope.isCurrent()) throw new Error("Timeline reconciliation detached");
    const page = await owner.page(cursor, scope.signal);
    if (!scope.isCurrent() || page.kind !== "accepted")
      throw new Error("Timeline result could not be read");
    for (const row of page.value.rows) {
      if (
        !remaining.has(row.recordId) ||
        row.rowVersion < (owner.latestVersion(row.recordId) ?? 0)
      )
        continue;
      if (
        row.recordId === target &&
        (row.rowVersion < receipt.data.row_version ||
          (row.rowVersion === receipt.data.row_version &&
            (row.captureState !== receipt.data.capture_state ||
              row.replacementRecordId !== receipt.data.replacement_record_id)))
      )
        continue;
      observed.set(row.recordId, row.rowVersion);
      remaining.delete(row.recordId);
    }
    if (!remaining.size) break;
    cursor = page.value.nextCursor;
    if (!page.value.hasMore || !cursor || cursors.has(cursor))
      throw new Error("Timeline result is missing or stale");
    cursors.add(cursor);
  } while (scope.isCurrent());
  if (remaining.size || !scope.isCurrent())
    throw new Error("Timeline reconciliation incomplete");
  const result = await history.load(target, scope.signal);
  if (
    scope.isCurrent() &&
    result.kind === "rejected" &&
    workbookOperationFailureIsAccessLoss(result.failure)
  )
    owner.loseAccess();
  if (
    !scope.isCurrent() ||
    result.kind !== "accepted" ||
    result.value.row_version <
      Math.max(receipt.data.row_version, owner.latestVersion(target) ?? 0)
  )
    throw new Error("Timeline history still needs refresh");
  owner.acceptVersion(target, result.value.row_version);
  await history.refreshRecordPresentation(target);
  owner.acceptVersion(
    target,
    history.latestVersion(target) ?? result.value.row_version,
  );
  if (
    !scope.isCurrent() ||
    [...observed].some(
      ([id, version]) => version < (owner.latestVersion(id) ?? 0),
    )
  )
    throw new Error("A newer Timeline version needs refresh");
}
