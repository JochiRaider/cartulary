import type { WorkbookRecordHistoryOwner } from "../../history/WorkbookRecordHistoryOwner";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import type { WorkbookCommittedRecordPort } from "../../query/WorkbookCommittedRecordPort";
import type { IndicatorCreateReceipt } from "./indicatorCreateOperation";
import { observationIndicatorView } from "./observationModel";
import type {
  ObservationReadPort,
  ObservationScope,
} from "./observationOperation";
import { observationFailureIsAccessLoss } from "./observationOperation";

/** Reconcile the canonical identity, independent of the visible sheet, filters and receipt age. */
export async function reconcileIndicatorCreateReceipt(
  records: WorkbookCommittedRecordPort,
  reader: Pick<ObservationReadPort, "records">,
  history: WorkbookRecordHistoryOwner,
  incidentId: string,
  receipt: IndicatorCreateReceipt,
  scope: ObservationScope,
  loseAccess: () => void,
) {
  const id = receipt.row.record_id;
  const current = () => scope.isCurrent() && !!records.getSnapshot().authority;
  const required = () =>
    Math.max(
      receipt.row.row_version,
      records.latestVersion(id) ?? 0,
      history.latestVersion(id) ?? 0,
    );
  if (!current()) throw new Error("Canonical refresh detached");
  const result = await history.load(id, scope.signal);
  if (
    current() &&
    result.kind === "rejected" &&
    observationFailureIsAccessLoss(result.failure)
  )
    loseAccess();
  if (
    !current() ||
    result.kind !== "accepted" ||
    result.value.incident_id !== incidentId ||
    result.value.record_id !== id ||
    result.value.row_version < required()
  )
    throw new Error("Canonical history needs refresh");
  history.acceptVersion(id, result.value.row_version);
  await history.refreshRecordPresentation(id);
  const cursors = new Set<string>();
  let cursor: string | null = null;
  for (;;) {
    if (!current()) throw new Error("Canonical refresh detached");
    const page = await reader.records(
      observationIndicatorView,
      emptyWorkbookQueryState(),
      cursor,
      scope.signal,
    );
    if (!current() || page.kind !== "accepted")
      throw new Error("Canonical record needs refresh");
    const row = page.value.items.find((row) => row.record_id === id);
    if (row && row.row_version >= required()) {
      records.acceptRow(row);
      if (current() && (records.latestRow(id)?.row_version ?? 0) >= required())
        return;
      throw new Error("Newer canonical evidence needs refresh");
    }
    cursor = page.value.nextCursor;
    if (!page.value.hasMore || !cursor || cursors.has(cursor))
      throw new Error("Canonical record is unavailable or needs refresh");
    cursors.add(cursor);
  }
}
