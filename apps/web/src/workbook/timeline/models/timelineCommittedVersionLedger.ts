import { timelineScalarBindings } from "./timelineFieldRegistry";
import type { WorkbookRow } from "./timelineRowModel";
import { decideWorkbookRecordFreshness } from "./workbookRecordFreshness";

type TimelineCommittedRowAcceptance = {
  readonly accepted: boolean;
  readonly row: WorkbookRow;
  readonly stale: boolean;
};

function committedTimelineProjection(row: WorkbookRow): WorkbookRow {
  const scalarValuesAreCommitted = timelineScalarBindings.every(
    (binding) => row.values[binding.key] === row.committedValues[binding.key],
  );
  if (
    row.pendingSignature === null &&
    scalarValuesAreCommitted &&
    row.collectionDrafts.hostRefs === "" &&
    row.collectionDrafts.identityRefs === "" &&
    row.collectionDrafts.tags === ""
  ) {
    return row;
  }
  return {
    ...row,
    collectionDrafts: { hostRefs: "", identityRefs: "", tags: "" },
    pendingSignature: null,
    values: { ...row.committedValues },
  };
}

export function createTimelineCommittedVersionLedger() {
  const rows = new Map<string, WorkbookRow>();
  const versions = new Map<string, number>();
  const retainedWork = new Set<string>();
  let epoch = 0;
  let inspectorRecord: string | null = null;

  const currentEpoch = () => epoch;
  const knownVersion = (recordId: string) => versions.get(recordId);
  const isStale = (recordId: string, rowVersion: number) =>
    decideWorkbookRecordFreshness(
      { recordId, rowVersion },
      knownVersion(recordId),
    ).stale;
  const current = (
    recordId: string,
    visibleRows: readonly WorkbookRow[],
  ): WorkbookRow | null => {
    const cached = rows.get(recordId);
    if (cached !== undefined) return cached;
    const visible = visibleRows.find(
      (candidate) =>
        candidate.recordId === recordId && candidate.rowVersion !== null,
    );
    return visible === undefined ? null : committedTimelineProjection(visible);
  };
  const accept = (
    row: WorkbookRow,
    visibleRows: readonly WorkbookRow[],
    retain = true,
  ): TimelineCommittedRowAcceptance => {
    if (row.recordId === null || row.rowVersion === null) {
      return { row, accepted: false, stale: false };
    }
    const recordId = row.recordId;
    if (retain) retainedWork.add(recordId);
    const rowVersion = row.rowVersion;
    const committed = committedTimelineProjection(row);
    const currentVersion = knownVersion(recordId);
    if (decideWorkbookRecordFreshness(committed, currentVersion).stale) {
      return {
        row: current(recordId, visibleRows) ?? committed,
        accepted: false,
        stale: true,
      };
    }
    if (currentVersion !== rowVersion) epoch += 1;
    versions.set(recordId, rowVersion);
    rows.set(recordId, committed);
    return { row: committed, accepted: true, stale: false };
  };
  // Version-only evidence fences later work; it does not establish saved fields
  // for that version. Only an accepted query or correlated row receipt does.
  const acceptVersion = (recordId: string, rowVersion: number) => {
    retainedWork.add(recordId);
    if (isStale(recordId, rowVersion)) return { accepted: false, stale: true };
    if (knownVersion(recordId) !== rowVersion) epoch += 1;
    versions.set(recordId, rowVersion);
    return { accepted: true, stale: false };
  };
  const latest = (
    recordId: string,
    visibleRows: readonly WorkbookRow[],
  ): WorkbookRow | null => {
    const visibleRow = visibleRows.find(
      (candidate) => candidate.recordId === recordId,
    );
    const currentVersion = knownVersion(recordId);
    if (
      visibleRow?.rowVersion !== null &&
      visibleRow?.rowVersion !== undefined &&
      (currentVersion === undefined || visibleRow.rowVersion >= currentVersion)
    ) {
      return accept(visibleRow, visibleRows).row;
    }
    const committedRow = rows.get(recordId);
    return committedRow?.rowVersion !== null &&
      committedRow?.rowVersion !== undefined &&
      (currentVersion === undefined ||
        committedRow.rowVersion >= currentVersion)
      ? committedRow
      : null;
  };

  return {
    clear: () => {
      rows.clear();
      versions.clear();
      retainedWork.clear();
      inspectorRecord = null;
      epoch += 1;
    },
    retainInspectorRecord: (recordId: string | null) => {
      inspectorRecord = recordId;
    },
    replaceQueryRows: (
      observed: readonly WorkbookRow[],
      visibleRows: readonly WorkbookRow[],
    ) => {
      const members = new Set(
        observed.flatMap((row) =>
          row.recordId === null ? [] : [row.recordId],
        ),
      );
      for (const row of visibleRows) {
        if (
          row.recordId &&
          (row.pendingSignature !== null ||
            row.collectionDrafts.hostRefs !== "" ||
            row.collectionDrafts.identityRefs !== "" ||
            row.collectionDrafts.tags !== "" ||
            timelineScalarBindings.some(
              (binding) =>
                row.values[binding.key] !== row.committedValues[binding.key],
            ))
        )
          retainedWork.add(row.recordId);
      }
      for (const id of rows.keys()) {
        if (
          !members.has(id) &&
          !retainedWork.has(id) &&
          inspectorRecord !== id
        ) {
          rows.delete(id);
          versions.delete(id);
        }
      }
      for (const row of observed) accept(row, visibleRows, false);
    },
    accept,
    acceptVersion,
    current,
    currentEpoch,
    isStale,
    knownVersion,
    latest,
  };
}
