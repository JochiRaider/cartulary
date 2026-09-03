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
  let committedRawRow = row.rawRow;
  if (row.rawRow !== null) {
    let committedCells = row.rawRow.cells;
    for (const binding of timelineScalarBindings) {
      const cell = committedCells[binding.fieldKey];
      const committedValue = row.committedValues[binding.key];
      if (!Object.is(cell?.value, committedValue)) {
        if (committedCells === row.rawRow.cells) {
          committedCells = { ...row.rawRow.cells };
        }
        committedCells[binding.fieldKey] = {
          ...cell,
          value: committedValue,
        };
      }
    }
    if (committedCells !== row.rawRow.cells) {
      committedRawRow = { ...row.rawRow, cells: committedCells };
    }
  }
  if (
    row.pendingSignature === null &&
    scalarValuesAreCommitted &&
    committedRawRow === row.rawRow &&
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
    rawRow: committedRawRow,
    values: { ...row.committedValues },
  };
}

export function createTimelineCommittedVersionLedger() {
  const rows = new Map<string, WorkbookRow>();
  const versions = new Map<string, number>();
  let epoch = 0;

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
  ): TimelineCommittedRowAcceptance => {
    if (row.recordId === null || row.rowVersion === null) {
      return { row, accepted: false, stale: false };
    }
    const recordId = row.recordId;
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
  const acceptVersion = (
    recordId: string,
    rowVersion: number,
    visibleRows: readonly WorkbookRow[],
  ) => {
    if (isStale(recordId, rowVersion)) {
      return { accepted: false, stale: true };
    }
    const existing = current(recordId, visibleRows);
    if (existing === null) {
      if (knownVersion(recordId) !== rowVersion) epoch += 1;
      versions.set(recordId, rowVersion);
      return { accepted: true, stale: false };
    }
    const accepted = accept(
      {
        ...existing,
        rowVersion,
        rawRow:
          existing.rawRow === null
            ? null
            : { ...existing.rawRow, row_version: rowVersion },
      },
      visibleRows,
    );
    return { accepted: accepted.accepted, stale: accepted.stale };
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
    accept,
    acceptVersion,
    current,
    currentEpoch,
    isStale,
    knownVersion,
    latest,
  };
}
