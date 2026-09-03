import { timelineScalarBindings } from "./timelineFieldRegistry";
import type { WorkbookRow } from "./timelineRowModel";
import { decideWorkbookRecordFreshness } from "./workbookRecordFreshness";

export type TimelineCommittedRowAcceptance = {
  readonly accepted: boolean;
  readonly row: WorkbookRow;
  readonly stale: boolean;
};

export function committedTimelineProjection(row: WorkbookRow): WorkbookRow {
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

export class TimelineCommittedVersionLedger {
  readonly #rows = new Map<string, WorkbookRow>();
  readonly #versions = new Map<string, number>();
  #epoch = 0;

  currentEpoch() {
    return this.#epoch;
  }

  knownVersion(recordId: string) {
    return this.#versions.get(recordId);
  }

  isStale(recordId: string, rowVersion: number) {
    return decideWorkbookRecordFreshness(
      { recordId, rowVersion },
      this.knownVersion(recordId),
    ).stale;
  }

  current(
    recordId: string,
    visibleRows: readonly WorkbookRow[],
  ): WorkbookRow | null {
    const cached = this.#rows.get(recordId);
    if (cached !== undefined) return cached;
    const visible = visibleRows.find(
      (candidate) =>
        candidate.recordId === recordId && candidate.rowVersion !== null,
    );
    return visible === undefined ? null : committedTimelineProjection(visible);
  }

  accept(
    row: WorkbookRow,
    visibleRows: readonly WorkbookRow[],
  ): TimelineCommittedRowAcceptance {
    if (row.recordId === null || row.rowVersion === null) {
      return { row, accepted: false, stale: false };
    }
    const recordId = row.recordId;
    const rowVersion = row.rowVersion;
    const committed = committedTimelineProjection(row);
    const currentVersion = this.knownVersion(recordId);
    if (decideWorkbookRecordFreshness(committed, currentVersion).stale) {
      return {
        row: this.current(recordId, visibleRows) ?? committed,
        accepted: false,
        stale: true,
      };
    }
    if (currentVersion !== rowVersion) this.#epoch += 1;
    this.#versions.set(recordId, rowVersion);
    this.#rows.set(recordId, committed);
    return { row: committed, accepted: true, stale: false };
  }

  acceptVersion(
    recordId: string,
    rowVersion: number,
    visibleRows: readonly WorkbookRow[],
  ) {
    if (this.isStale(recordId, rowVersion)) {
      return { accepted: false, stale: true };
    }
    const existing = this.current(recordId, visibleRows);
    if (existing === null) {
      if (this.knownVersion(recordId) !== rowVersion) this.#epoch += 1;
      this.#versions.set(recordId, rowVersion);
      return { accepted: true, stale: false };
    }
    const accepted = this.accept(
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
  }

  latest(
    recordId: string,
    visibleRows: readonly WorkbookRow[],
  ): WorkbookRow | null {
    const visibleRow = visibleRows.find(
      (candidate) => candidate.recordId === recordId,
    );
    const knownVersion = this.knownVersion(recordId);
    if (
      visibleRow?.rowVersion !== null &&
      visibleRow?.rowVersion !== undefined &&
      (knownVersion === undefined || visibleRow.rowVersion >= knownVersion)
    ) {
      return this.accept(visibleRow, visibleRows).row;
    }
    const committedRow = this.#rows.get(recordId);
    return committedRow?.rowVersion !== null &&
      committedRow?.rowVersion !== undefined &&
      (knownVersion === undefined || committedRow.rowVersion >= knownVersion)
      ? committedRow
      : null;
  }
}
