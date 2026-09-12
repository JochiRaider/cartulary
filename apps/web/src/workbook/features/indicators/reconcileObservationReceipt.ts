import { listWorkbookSurfaceContracts } from "@cartulary/view-contracts";
import type { WorkbookRecordHistoryOwner } from "../../history/WorkbookRecordHistoryOwner";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import { observationIndicatorView } from "./observationModel";
import {
  type ObservationAttempt,
  type ObservationReceipt,
  type ObservationScope,
  observationFailureIsAccessLoss,
  observationIntentSource,
} from "./observationOperation";
import type { WorkbookObservationOwner } from "./WorkbookObservationOwner";

/** Every returned first-class identity is refreshed, independently of visible filters. */
export async function reconcileObservationReceipt(
  owner: WorkbookObservationOwner,
  history: WorkbookRecordHistoryOwner,
  attempt: ObservationAttempt,
  receipt: ObservationReceipt,
  scope: ObservationScope,
) {
  const sourceId = observationIntentSource(attempt.intent);
  const sourceViews = listWorkbookSurfaceContracts().filter((surface) =>
    surface.contract.fields.some(
      (field) =>
        field.fieldKey === receipt.observation.source_field_key &&
        field.readKind === "text",
    ),
  );
  const sourceView =
    attempt.intent.action === "create"
      ? attempt.intent.source.viewSchemaId
      : sourceViews.length === 1
        ? sourceViews[0]?.viewSchemaId
        : null;
  if (!sourceView) throw new Error("Source view ownership is unavailable");
  const remaining = new Map(
    receipt.affected_records.map((record) => [
      record.record_id,
      record.row_version,
    ]),
  );
  for (const record of receipt.affected_records) {
    if (!scope.isCurrent())
      throw new Error("Observation reconciliation detached");
    const result = await history.load(record.record_id, scope.signal);
    if (
      scope.isCurrent() &&
      result.kind === "rejected" &&
      observationFailureIsAccessLoss(result.failure)
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
      throw new Error("Observation history presentation detached");
    owner.acceptVersion(
      record.record_id,
      history.latestVersion(record.record_id) ?? result.value.row_version,
    );
  }
  for (const view of new Set([sourceView, observationIndicatorView])) {
    const expected = new Set(
      [...remaining.keys()].filter(
        (id) =>
          (id === sourceId ? sourceView : observationIndicatorView) === view,
      ),
    );
    if (!expected.size) continue;
    const cursors = new Set<string>();
    let cursor: string | null = null;
    for (;;) {
      if (!scope.isCurrent())
        throw new Error("Observation reconciliation detached");
      const result = await owner.records(
        view,
        emptyWorkbookQueryState(),
        cursor,
        scope.signal,
      );
      if (!scope.isCurrent() || result.kind !== "accepted")
        throw new Error("Affected records could not be read");
      for (const row of result.value.items) {
        if (
          !expected.has(row.record_id) ||
          row.row_version <
            Math.max(
              remaining.get(row.record_id) ?? Infinity,
              owner.latestVersion(row.record_id) ?? 0,
            )
        )
          continue;
        owner.acceptRow(row);
        expected.delete(row.record_id);
        remaining.delete(row.record_id);
      }
      if (!expected.size) break;
      cursor = result.value.nextCursor;
      if (!result.value.hasMore || !cursor || cursors.has(cursor))
        throw new Error("Affected records still need refresh");
      cursors.add(cursor);
    }
  }
  if (
    !scope.isCurrent() ||
    remaining.size ||
    receipt.affected_records.some(
      (record) =>
        (owner.latestRow(record.record_id)?.row_version ?? 0) <
        (owner.latestVersion(record.record_id) ?? 0),
    )
  )
    throw new Error("Newer record evidence needs refresh");
}
